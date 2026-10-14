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

	"github.com/snisid/platform/services/extors-svc/internal/handler"
	"github.com/snisid/platform/services/extors-svc/internal/kafka"
	"github.com/snisid/platform/services/extors-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/extors-svc/internal/service"
)

func main() {
	dbHost := getEnv("EXTORS_DB_HOST", "localhost")
	dbPort := getEnv("EXTORS_DB_PORT", "26257")
	dbName := getEnv("EXTORS_DB_NAME", "snisid_extors")
	dbUser := getEnv("EXTORS_DB_USER", "root")
	dbSSLMode := getEnv("EXTORS_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("EXTORS_KAFKA_BROKERS", "localhost:9092")

	port := getEnv("EXTORS_SERVICE_PORT", "8116")

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

	repo := postgres.NewExtortionRepo(pool)
	svc := service.NewExtorsService(repo, logger)
	_ = svc

	h := handler.NewHealthHandler(pool, logger)
	r := gin.Default()
	r.GET("/healthz", h.Healthz)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("extors-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS extors_type AS ENUM ( 'KIDNAPPING_RANSOM', 'ROAD_TOLL_ILLEGAL', 'BUSINESS_PROTECTION_RACKET', 'REAL_ESTATE_EXTORTION', 'PUBLIC_SERVANT_EXTORTION', 'NGO_EXTORTION', 'FUEL_TRUCK_HIJACK', 'OTHER' );",
		"CREATE TYPE IF NOT EXISTS extors_payment_channel AS ENUM ( 'MONCASH', 'NATCASH', 'DIGICEL_MONEY', 'WIRE_TRANSFER', 'CASH_DROP', 'CRYPTOCURRENCY', 'INTERMEDIARY', 'UNKNOWN' );",
		"CREATE TYPE IF NOT EXISTS extors_status AS ENUM ( 'ACTIVE','PAID','REFUSED','NEGOTIATING', 'LAW_ENFORCEMENT_INVOLVED','RESOLVED','VICTIM_HARMED' );",
		"CREATE TABLE IF NOT EXISTS extors_cases ( case_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(), national_extors_id  VARCHAR(25) UNIQUE NOT NULL,  -- EXTORS-HT-AAAA-NNNNNN extors_type         extors_type NOT NULL, status              extors_status NOT NULL DEFAULT 'ACTIVE', gang_id             UUID, gang_name           VARCHAR(150), perpetrator_ids     UUID[] DEFAULT '{}',       -- CHEF-HT member IDs chef_member_ids     UUID[] DEFAULT '{}', victim_count        SMALLINT DEFAULT 1, victim_snisid_ids   UUID[] DEFAULT '{}', victim_types        TEXT[] DEFAULT '{}',       -- INDIVIDUAL, BUSINESS, NGO, GOVERNMENT victim_nationality  CHAR(3)[] DEFAULT '{}', is_foreigner_victim BOOLEAN DEFAULT FALSE, incident_location   VARCHAR(300), dept_code           CHAR(2), commune             VARCHAR(100), lat                 DECIMAL(10,7), lng                 DECIMAL(10,7), route_number        VARCHAR(10),               -- RN1, RN2 etc pour les peages demanded_amount     DECIMAL(15,2), demanded_currency   CHAR(3) DEFAULT 'USD', paid_amount         DECIMAL(15,2), paid_currency       CHAR(3), payment_channel     extors_payment_channel, payment_ref         VARCHAR(200),              -- Numero de transaction MonCash etc. payment_date        TIMESTAMPTZ, first_contact_date  TIMESTAMPTZ NOT NULL, resolution_date     TIMESTAMPTZ, case_reference      VARCHAR(100), investigating_unit  VARCHAR(50), ucref_str_id        UUID,                      -- Lien STR si ranÃ§on tracÃ©e blan_case_id        UUID,                      -- Lien dossier blanchiment notes               TEXT, created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS extors_road_toll_points ( toll_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(), gang_id             UUID NOT NULL, location_desc       VARCHAR(300) NOT NULL, route_number        VARCHAR(10), dept_code           CHAR(2) NOT NULL, commune             VARCHAR(100), lat                 DECIMAL(10,7), lng                 DECIMAL(10,7), daily_revenue_usd   DECIMAL(10,2), vehicle_types_taxed TEXT[] DEFAULT '{}', toll_rates          JSONB,                     -- {moto: 50, voiture: 200, camion: 500} active_since        DATE, is_active           BOOLEAN DEFAULT TRUE, source_intel        TEXT, last_confirmed_at   TIMESTAMPTZ, created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS extors_negotiations ( neg_id              UUID PRIMARY KEY DEFAULT gen_random_uuid(), case_id             UUID NOT NULL REFERENCES extors_cases(case_id), negotiation_date    TIMESTAMPTZ NOT NULL, contact_method      VARCHAR(50),               -- PHONE, INTERMEDIARY, DROP_NOTE contact_number      VARCHAR(30), demand_updated      DECIMAL(15,2), demand_currency     CHAR(3), position_update     TEXT, recorded_by         UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_extors_cases_gang    ON extors_cases(gang_id) WHERE gang_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_extors_cases_type    ON extors_cases(extors_type, status);",
		"CREATE INDEX IF NOT EXISTS idx_extors_cases_dept    ON extors_cases(dept_code, first_contact_date DESC);",
		"CREATE INDEX IF NOT EXISTS idx_extors_cases_channel ON extors_cases(payment_channel) WHERE paid_amount IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_extors_tolls_route   ON extors_road_toll_points(route_number) WHERE is_active = TRUE;",
		"CREATE INDEX IF NOT EXISTS idx_extors_tolls_dept    ON extors_road_toll_points(dept_code) WHERE is_active = TRUE;",
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
