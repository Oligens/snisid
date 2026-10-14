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

	"github.com/snisid/platform/services/vict-svc/internal/handler"
	"github.com/snisid/platform/services/vict-svc/internal/kafka"
	"github.com/snisid/platform/services/vict-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/vict-svc/internal/service"
)

func main() {
	dbHost := getEnv("VICT_DB_HOST", "localhost")
	dbPort := getEnv("VICT_DB_PORT", "26257")
	dbName := getEnv("VICT_DB_NAME", "snisid_vict")
	dbUser := getEnv("VICT_DB_USER", "root")
	dbSSLMode := getEnv("VICT_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("VICT_KAFKA_BROKERS", "localhost:9092")

	port := getEnv("VICT_SERVICE_PORT", "8123")

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

	repo := postgres.NewVictimRepo(pool)
	svc := service.NewVictimService(repo, logger)
	_ = svc

	h := handler.NewHealthHandler(pool, logger)
	r := gin.Default()
	r.GET("/healthz", h.Healthz)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("vict-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS vict_crime_type AS ENUM ( 'HOMICIDE', 'MASS_KILLING', 'RAPE', 'GANG_RAPE', 'TORTURE', 'FORCED_DISAPPEARANCE', 'EXTRAJUDICIAL_KILLING', 'KIDNAPPING_VICTIM', 'MUTILATION', 'OTHER_GRAVE' );",
		"CREATE TYPE IF NOT EXISTS vict_victim_status AS ENUM ( 'ALIVE_SURVIVOR', 'DECEASED_IDENTIFIED', 'DECEASED_UNIDENTIFIED', 'MISSING_PRESUMED_DEAD' );",
		"CREATE TABLE IF NOT EXISTS vict_victims ( victim_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(), national_vict_id    VARCHAR(25) UNIQUE NOT NULL,  -- VICT-HT-AAAA-NNNNNN snisid_person_id    UUID, crime_type          vict_crime_type NOT NULL, victim_status       vict_victim_status NOT NULL, full_name           VARCHAR(200), dob                 DATE, gender              VARCHAR(10), nationality         CHAR(3) DEFAULT 'HTI', occupation          VARCHAR(100), incident_date       TIMESTAMPTZ NOT NULL, incident_location   VARCHAR(300), dept_code           CHAR(2), commune             VARCHAR(100), lat                 DECIMAL(10,7), lng                 DECIMAL(10,7), perpetrator_ids     UUID[] DEFAULT '{}', gang_id             UUID, case_reference      VARCHAR(100), parquet_ref         VARCHAR(100), medical_report_ref  VARCHAR(200), autopsy_ref         VARCHAR(200), dna_sample_ref      VARCHAR(100), afis_subject_id     UUID, rvin_case_id        UUID, iachr_ref           VARCHAR(50),       -- CIDH/IACHR reference un_special_rap_ref  VARCHAR(50), needs_reparation    BOOLEAN DEFAULT FALSE, created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS vict_mass_incidents ( mass_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(), incident_name       VARCHAR(200) NOT NULL,   -- Ex: Massacre de CitÃ© Soleil 2022 crime_type          vict_crime_type NOT NULL, incident_date       TIMESTAMPTZ NOT NULL, dept_code           CHAR(2), commune             VARCHAR(100), lat                 DECIMAL(10,7), lng                 DECIMAL(10,7), victim_count        INTEGER NOT NULL, survivor_count      INTEGER DEFAULT 0, perpetrator_gang_id UUID, description         TEXT, documented_by       TEXT[] DEFAULT '{}',    -- RNDDH, HRW, MSF, ONU, etc. iachr_case_ref      VARCHAR(50), linked_victim_ids   UUID[] DEFAULT '{}', created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_vict_crime_type ON vict_victims(crime_type, victim_status);",
		"CREATE INDEX IF NOT EXISTS idx_vict_dept       ON vict_victims(dept_code, incident_date DESC);",
		"CREATE INDEX IF NOT EXISTS idx_vict_gang       ON vict_victims(gang_id) WHERE gang_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_vict_snisid     ON vict_victims(snisid_person_id) WHERE snisid_person_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_vict_mass       ON vict_mass_incidents(dept_code, incident_date DESC);",
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
