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

	"github.com/snisid/platform/services/trait-svc/internal/handler"
	"github.com/snisid/platform/services/trait-svc/internal/kafka"
	"github.com/snisid/platform/services/trait-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/trait-svc/internal/service"
)

func main() {
	dbHost := getEnv("TRAIT_DB_HOST", "localhost")
	dbPort := getEnv("TRAIT_DB_PORT", "26257")
	dbName := getEnv("TRAIT_DB_NAME", "snisid_trait")
	dbUser := getEnv("TRAIT_DB_USER", "root")
	dbSSLMode := getEnv("TRAIT_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("TRAIT_KAFKA_BROKERS", "localhost:9092")

	port := getEnv("TRAIT_SERVICE_PORT", "8122")

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

	repo := postgres.NewTraffickingRepo(pool)
	svc := service.NewTraiffickingService(repo, logger)
	_ = svc

	h := handler.NewHealthHandler(pool, logger)
	r := gin.Default()
	r.GET("/healthz", h.Healthz)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("trait-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS trait_type AS ENUM ( 'LABOR_EXPLOITATION', 'SEXUAL_EXPLOITATION', 'FORCED_MARRIAGE', 'CHILD_DOMESTIC_SERVITUDE', 'GANG_RECRUITMENT_FORCED', 'IRREGULAR_MIGRATION_FACILITATION', 'ORGAN_TRAFFICKING', 'OTHER' );",
		"CREATE TYPE IF NOT EXISTS trait_victim_status AS ENUM ( 'IDENTIFIED_VICTIM', 'POTENTIAL_VICTIM', 'WITNESS', 'RESCUED', 'REPATRIATED', 'DECEASED', 'MISSING' );",
		"CREATE TABLE IF NOT EXISTS trait_cases ( case_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(), national_trait_id   VARCHAR(25) UNIQUE NOT NULL,  -- TRAIT-HT-AAAA-NNNNNN trait_type          trait_type NOT NULL, status              VARCHAR(20) DEFAULT 'OPEN', victim_count        SMALLINT DEFAULT 1, minor_count         SMALLINT DEFAULT 0, origin_country      CHAR(3) DEFAULT 'HTI', transit_countries   CHAR(3)[] DEFAULT '{}', destination_country CHAR(3), route_description   TEXT, transport_mode      TEXT[] DEFAULT '{}',     -- BOAT, BUS, FOOT, AIR mar_incident_id     UUID,                    -- Lien MAR-HT si maritime sifr_crossing_ids   UUID[] DEFAULT '{}',     -- Postes frontiere impliques gang_id             UUID,                    -- Si gang facilite recruiter_ids       UUID[] DEFAULT '{}',     -- Passeurs / recruteurs SNISID total_amount_paid   DECIMAL(12,2), amount_per_person   DECIMAL(10,2), currency            CHAR(3) DEFAULT 'USD', investigating_unit  VARCHAR(50), case_reference      VARCHAR(100), iom_case_ref        VARCHAR(50), created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS trait_victims ( victim_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(), case_id             UUID NOT NULL REFERENCES trait_cases(case_id), snisid_person_id    UUID, victim_status       trait_victim_status NOT NULL, full_name           VARCHAR(200), nationality         CHAR(3) DEFAULT 'HTI', dob                 DATE, gender              VARCHAR(10), is_minor            BOOLEAN DEFAULT FALSE, exploitation_type   TEXT, rescue_date         TIMESTAMPTZ, rescue_location     VARCHAR(300), current_location    TEXT, assistance_provided TEXT[] DEFAULT '{}',  -- SHELTER, LEGAL, MEDICAL, REPATRIATION dipe_case_id        UUID,                 -- Lien DIPE si disparu initialement afis_subject_id     UUID, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS trait_networks ( network_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(), network_name        VARCHAR(150), primary_route       TEXT, origin_dept         CHAR(2), known_members       UUID[] DEFAULT '{}',   -- SNISID person IDs gang_affiliations   UUID[] DEFAULT '{}', monthly_volume_est  INTEGER, fee_per_person_usd  DECIMAL(10,2), is_active           BOOLEAN DEFAULT TRUE, intel_confidence    SMALLINT, linked_cases        UUID[] DEFAULT '{}', created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_trait_cases_type    ON trait_cases(trait_type, status);",
		"CREATE INDEX IF NOT EXISTS idx_trait_cases_gang    ON trait_cases(gang_id) WHERE gang_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_trait_victims_case  ON trait_victims(case_id);",
		"CREATE INDEX IF NOT EXISTS idx_trait_victims_minor ON trait_victims(is_minor) WHERE is_minor = TRUE;",
		"CREATE INDEX IF NOT EXISTS idx_trait_victims_snisid ON trait_victims(snisid_person_id) WHERE snisid_person_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_trait_networks_active ON trait_networks(is_active) WHERE is_active = TRUE;",
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
