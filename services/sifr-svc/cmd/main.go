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

	"github.com/snisid/platform/services/sifr-svc/internal/handler"
	"github.com/snisid/platform/services/sifr-svc/internal/kafka"
	"github.com/snisid/platform/services/sifr-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/sifr-svc/internal/service"
)

func main() {
	dbHost := getEnv("SIFR_DB_HOST", "localhost")
	dbPort := getEnv("SIFR_DB_PORT", "26257")
	dbName := getEnv("SIFR_DB_NAME", "snisid_sifr")
	dbUser := getEnv("SIFR_DB_USER", "root")
	dbSSLMode := getEnv("SIFR_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("SIFR_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("SIFR_KAFKA_TOPIC", "snisid.sifr.events")
	port := getEnv("SIFR_SERVICE_PORT", "8106")

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

	producer := kafka.NewProducer([]string{kafkaBrokers}, kafkaTopic, logger)
	defer producer.Close()

	repo := postgres.NewBorderRepo(pool)
	svc := service.NewBorderService(repo, logger)

	r := handler.SetupRouter(svc, logger)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("sifr-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS sifr_crossing_direction AS ENUM ('ENTRY', 'EXIT');",
		"CREATE TYPE IF NOT EXISTS sifr_doc_type AS ENUM ( 'PASSPORT', 'NATIONAL_ID', 'LAISSEZ_PASSER', 'BIRTH_CERTIFICATE', 'TRAVEL_DOCUMENT', 'NONE' );",
		"CREATE TYPE IF NOT EXISTS sifr_alert_type AS ENUM ( 'WANTED_PERSON', 'STOLEN_DOCUMENT', 'BLACKLIST', 'ACTIVE_WARRANT', 'SANCTIONS', 'CUSTOMS_ALERT' );",
		"CREATE TABLE IF NOT EXISTS sifr_border_posts ( post_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(), post_code           VARCHAR(10) UNIQUE NOT NULL,  -- MLP, OUN, BLD, AAP, etc. name                VARCHAR(150) NOT NULL, dept_code           CHAR(2) NOT NULL, border_country      CHAR(3) NOT NULL DEFAULT 'DOM', post_lat            DECIMAL(10,7), post_lng            DECIMAL(10,7), is_official         BOOLEAN DEFAULT TRUE, is_active           BOOLEAN DEFAULT TRUE, lanes_count         SMALLINT DEFAULT 2, has_biometric_scanner BOOLEAN DEFAULT FALSE, has_vehicle_scanner BOOLEAN DEFAULT FALSE, operating_hours     VARCHAR(50), commanding_officer  UUID, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS sifr_crossings ( crossing_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), post_id             UUID NOT NULL REFERENCES sifr_border_posts(post_id), direction           sifr_crossing_direction NOT NULL, crossing_datetime   TIMESTAMPTZ NOT NULL, snisid_person_id    UUID, document_type       sifr_doc_type NOT NULL DEFAULT 'PASSPORT', document_number     VARCHAR(100), document_country    CHAR(3), document_expiry     DATE, traveler_name       VARCHAR(200) NOT NULL, traveler_dob        DATE, traveler_nationality CHAR(3), vehicle_plate       VARCHAR(20), lane_number         SMALLINT, processing_officer  UUID NOT NULL, alert_triggered     BOOLEAN DEFAULT FALSE, alert_type          sifr_alert_type, alert_action_taken  TEXT, processing_time_sec INTEGER, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS sifr_alerts_log ( alert_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(), crossing_id         UUID REFERENCES sifr_crossings(crossing_id), post_id             UUID NOT NULL, alert_type          sifr_alert_type NOT NULL, snisid_person_id    UUID, document_number     VARCHAR(100), vehicle_plate       VARCHAR(20), alert_source        VARCHAR(50),     -- FPR, BLKL, SLTD, OPR, SANC source_record_id    UUID, notified_units      TEXT[] DEFAULT '{}', action_taken        TEXT, resolved            BOOLEAN DEFAULT FALSE, resolved_by         UUID, resolved_at         TIMESTAMPTZ, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS sifr_clandestine_crossings ( report_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(), location_desc       VARCHAR(300), dept_code           CHAR(2), lat                 DECIMAL(10,7), lng                 DECIMAL(10,7), reported_date       TIMESTAMPTZ NOT NULL, crossing_type       VARCHAR(50),     -- FOOT, VEHICLE, BOAT estimated_persons   INTEGER, gang_related        BOOLEAN DEFAULT FALSE, gang_id             UUID, trafficking_type    TEXT, reported_by         UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_sifr_crossings_datetime ON sifr_crossings(crossing_datetime DESC);",
		"CREATE INDEX IF NOT EXISTS idx_sifr_crossings_person   ON sifr_crossings(snisid_person_id) WHERE snisid_person_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_sifr_crossings_doc      ON sifr_crossings(document_number);",
		"CREATE INDEX IF NOT EXISTS idx_sifr_crossings_alert    ON sifr_crossings(alert_triggered) WHERE alert_triggered = TRUE;",
		"CREATE INDEX IF NOT EXISTS idx_sifr_crossings_post     ON sifr_crossings(post_id, crossing_datetime DESC);",
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
