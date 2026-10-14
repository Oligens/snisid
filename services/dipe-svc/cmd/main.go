package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/dipe-svc/internal/handler"
	"github.com/snisid/platform/services/dipe-svc/internal/kafka"
	"github.com/snisid/platform/services/dipe-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/dipe-svc/internal/service"
)

func main() {
	dbHost := getEnv("DIPE_DB_HOST", "localhost")
	dbPort := getEnv("DIPE_DB_PORT", "26257")
	dbName := getEnv("DIPE_DB_NAME", "snisid_dipe")
	dbUser := getEnv("DIPE_DB_USER", "root")
	dbSSLMode := getEnv("DIPE_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("DIPE_KAFKA_BROKERS", "localhost:9092")

	port := getEnv("DIPE_SERVICE_PORT", "8118")

	dbURL := fmt.Sprintf("postgresql://%s@%s:%s/%s?sslmode=%s", dbUser, dbHost, dbPort, dbName, dbSSLMode)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping: %v", err)
	}
	pool.Config().MaxConns = 25

	if err := runMigrations(ctx, pool); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	producer, err := kafka.NewProducer([]string{kafkaBrokers}, logger)
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
	}
	defer producer.Close()

	repo := postgres.NewMissingRepo(pool)
	svc := service.NewMissingPersonService(repo, logger)
	_ = svc

	h := handler.NewHealthHandler(pool, logger)
	r := gin.Default()
	r.GET("/healthz", h.Healthz)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("dipe-svc started", zap.String("port", port))
		if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			logger.Fatal("error", zap.Error(e))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	logger.Info("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		"CREATE TYPE IF NOT EXISTS dipe_case_type AS ENUM ( 'KIDNAPPING_SUSPECTED', 'VOLUNTARY_DISAPPEARANCE', 'DISASTER_RELATED', 'GANG_VIOLENCE', 'MIGRATION_RELATED', 'CHILD_ABDUCTION', 'TRAFFICKING_SUSPECTED', 'UNKNOWN' );",
		"CREATE TYPE IF NOT EXISTS dipe_case_status AS ENUM ( 'OPEN', 'LOCATED_ALIVE', 'BODY_IDENTIFIED', 'BODY_UNIDENTIFIED', 'CANCELLED', 'COLD_CASE' );",
		"CREATE TABLE IF NOT EXISTS dipe_missing_persons ( case_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(), national_dipe_id    VARCHAR(25) UNIQUE NOT NULL,  -- DIPE-HT-AAAA-NNNNNN case_type           dipe_case_type NOT NULL, status              dipe_case_status NOT NULL DEFAULT 'OPEN', snisid_person_id    UUID, full_name           VARCHAR(200) NOT NULL, aliases             TEXT[] DEFAULT '{}', dob                 DATE, gender              VARCHAR(10), nationality         CHAR(3) DEFAULT 'HTI', occupation          VARCHAR(100), photo_refs          TEXT[] DEFAULT '{}', height_cm           SMALLINT, weight_kg           SMALLINT, skin_tone           VARCHAR(30), eye_color           VARCHAR(30), hair_color          VARCHAR(30), distinguishing_marks TEXT, clothing_last_seen  TEXT, last_seen_date      TIMESTAMPTZ NOT NULL, last_seen_location  VARCHAR(300), last_seen_dept_code CHAR(2), last_seen_commune   VARCHAR(100), last_seen_lat       DECIMAL(10,7), last_seen_lng       DECIMAL(10,7), circumstances       TEXT, sivc_alert_id       UUID,            -- Vehicule kidnapping SIVC-HT gang_id             UUID, extors_case_id      UUID,            -- Lien ranÃ§on EXTORS-HT reported_by_name    VARCHAR(200), reported_by_phone   VARCHAR(30), reported_by_snisid  UUID, report_date         TIMESTAMPTZ NOT NULL, reporting_unit      VARCHAR(50), afis_subject_id     UUID, dna_sample_ref      VARCHAR(100), dna_profile_id      UUID, interpol_notice_ref VARCHAR(50), ncmec_ref           VARCHAR(50),     -- Pour enfants resolution_date     TIMESTAMPTZ, resolution_notes    TEXT, rvin_case_id        UUID,            -- Lien RVIN si corps non identifie created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS dipe_sightings ( sighting_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), case_id             UUID NOT NULL REFERENCES dipe_missing_persons(case_id), sighting_date       TIMESTAMPTZ NOT NULL, location_desc       VARCHAR(300), dept_code           CHAR(2), lat                 DECIMAL(10,7), lng                 DECIMAL(10,7), reported_by         UUID, report_method       VARCHAR(30),    -- TIP_LINE, LAPI, FIELD_OFFICER, PUBLIC confidence          SMALLINT, photo_ref           VARCHAR(500), verified            BOOLEAN DEFAULT FALSE, verified_by         UUID, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS dipe_disaster_missing ( disaster_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), case_id             UUID NOT NULL REFERENCES dipe_missing_persons(case_id), disaster_type       VARCHAR(30) NOT NULL,  -- EARTHQUAKE, HURRICANE, FLOOD disaster_name       VARCHAR(100), disaster_date       DATE NOT NULL, last_known_address  TEXT, shelter_checked     TEXT[] DEFAULT '{}', hospital_checked    TEXT[] DEFAULT '{}', morgue_checked      TEXT[] DEFAULT '{}', rc_haiti_ref        VARCHAR(50),           -- Reference Croix-Rouge HaÃ¯tienne created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_dipe_status      ON dipe_missing_persons(status, last_seen_date DESC);",
		"CREATE INDEX IF NOT EXISTS idx_dipe_type        ON dipe_missing_persons(case_type) WHERE status = 'OPEN';",
		"CREATE INDEX IF NOT EXISTS idx_dipe_dept        ON dipe_missing_persons(last_seen_dept_code) WHERE status = 'OPEN';",
		"CREATE INDEX IF NOT EXISTS idx_dipe_person      ON dipe_missing_persons(snisid_person_id) WHERE snisid_person_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_dipe_sightings   ON dipe_sightings(case_id, sighting_date DESC);",
		"CREATE INDEX IF NOT EXISTS idx_dipe_name_fts    ON dipe_missing_persons USING gin(to_tsvector('simple', full_name));",
	}
	for _, m := range migrations {
		if _, err := pool.Exec(ctx, m); err != nil {
			return fmt.Errorf("migration: %s: %w", m[:60], err)
		}
	}
	return nil
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
