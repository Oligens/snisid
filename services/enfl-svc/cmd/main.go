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

	"github.com/snisid/platform/services/enfl-svc/internal/handler"
	"github.com/snisid/platform/services/enfl-svc/internal/kafka"
	"github.com/snisid/platform/services/enfl-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/enfl-svc/internal/service"
)

func main() {
	dbHost := getEnv("ENFL_DB_HOST", "localhost")
	dbPort := getEnv("ENFL_DB_PORT", "26257")
	dbName := getEnv("ENFL_DB_NAME", "snisid_enfl")
	dbUser := getEnv("ENFL_DB_USER", "root")
	dbSSLMode := getEnv("ENFL_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("ENFL_KAFKA_BROKERS", "localhost:9092")

	port := getEnv("ENFL_SERVICE_PORT", "8119")

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

	repo := postgres.NewChildRepo(pool)
	svc := service.NewChildService(repo, logger)
	_ = svc

	h := handler.NewHealthHandler(pool, logger)
	r := gin.Default()
	r.GET("/healthz", h.Healthz)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("enfl-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS enfl_risk_category AS ENUM ( 'MISSING_ABDUCTION', 'GANG_RECRUITMENT', 'DOMESTIC_SERVITUDE_RESTAVEK', 'SEXUAL_EXPLOITATION', 'TRAFFICKING', 'UNACCOMPANIED_MIGRANT', 'SEPARATED_DISASTER', 'STREET_CHILD', 'OTHER' );",
		"CREATE TYPE IF NOT EXISTS enfl_status AS ENUM ( 'AT_RISK', 'MISSING', 'LOCATED_SAFE', 'LOCATED_AT_RISK', 'IN_CARE', 'REPATRIATED', 'DECEASED' );",
		"CREATE TABLE IF NOT EXISTS enfl_children ( child_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(), national_enfl_id    VARCHAR(25) UNIQUE NOT NULL,  -- ENFL-HT-AAAA-NNNNNN snisid_person_id    UUID, dipe_case_id        UUID,           -- Lien DIPE-HT si disparu trait_case_id       UUID,           -- Lien TRAIT-HT si traite risk_category       enfl_risk_category NOT NULL, status              enfl_status NOT NULL DEFAULT 'MISSING', full_name           VARCHAR(200) NOT NULL, dob                 DATE NOT NULL, age_at_registration SMALLINT, gender              VARCHAR(10), nationality         CHAR(3) DEFAULT 'HTI', photo_refs          TEXT[] DEFAULT '{}', distinguishing_marks TEXT, height_cm           SMALLINT, skin_tone           VARCHAR(30), guardian_name       VARCHAR(200), guardian_phone      VARCHAR(30), guardian_snisid_id  UUID, last_known_location VARCHAR(300), dept_code           CHAR(2), commune             VARCHAR(100), disappearance_date  TIMESTAMPTZ, gang_id             UUID,           -- Si gang implique recruiter_snisid_id UUID,           -- Si recruteur identifie afis_subject_id     UUID, dna_profile_id      UUID, interpol_icse_ref   VARCHAR(50),    -- INTERPOL Crimes Against Children ncmec_ref           VARCHAR(50), ibesr_ref           VARCHAR(50), assistance_type     TEXT[] DEFAULT '{}', current_shelter     VARCHAR(200), assigned_caseworker UUID, created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS enfl_restaveks ( restavek_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), child_id            UUID NOT NULL REFERENCES enfl_children(child_id), employing_household VARCHAR(300), household_dept      CHAR(2), household_commune   VARCHAR(100), employing_person_id UUID, reported_conditions TEXT, school_attendance   BOOLEAN DEFAULT FALSE, ibesr_inspection    BOOLEAN DEFAULT FALSE, last_inspection_date DATE, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_enfl_status      ON enfl_children(status, risk_category);",
		"CREATE INDEX IF NOT EXISTS idx_enfl_dept        ON enfl_children(dept_code) WHERE status IN ('MISSING','AT_RISK');",
		"CREATE INDEX IF NOT EXISTS idx_enfl_gang        ON enfl_children(gang_id) WHERE gang_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_enfl_name_fts    ON enfl_children USING gin(to_tsvector('simple', full_name));",
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
