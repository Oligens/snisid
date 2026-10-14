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

	"github.com/snisid/platform/services/dpide-svc/internal/handler"
	"github.com/snisid/platform/services/dpide-svc/internal/kafka"
	"github.com/snisid/platform/services/dpide-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/dpide-svc/internal/service"
)

func main() {
	dbHost := getEnv("DPIDE_DB_HOST", "localhost")
	dbPort := getEnv("DPIDE_DB_PORT", "26257")
	dbName := getEnv("DPIDE_DB_NAME", "snisid_dpide")
	dbUser := getEnv("DPIDE_DB_USER", "root")
	dbSSLMode := getEnv("DPIDE_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("DPIDE_KAFKA_BROKERS", "localhost:9092")

	port := getEnv("DPIDE_SERVICE_PORT", "8121")

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

	repo := postgres.NewIDPRepo(pool)
	svc := service.NewIDPService(repo, logger)
	_ = svc

	h := handler.NewHealthHandler(pool, logger)
	r := gin.Default()
	r.GET("/healthz", h.Healthz)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("dpide-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS dpide_displacement_cause AS ENUM ( 'GANG_VIOLENCE', 'EARTHQUAKE', 'HURRICANE', 'FLOOD', 'FIRE', 'POLITICAL_VIOLENCE', 'OTHER' );",
		"CREATE TYPE IF NOT EXISTS dpide_idp_status AS ENUM ( 'DISPLACED', 'IN_CAMP', 'WITH_HOST_FAMILY', 'RELOCATED', 'RETURNED_HOME', 'EMIGRATED', 'DECEASED' );",
		"CREATE TABLE IF NOT EXISTS dpide_idps ( idp_id              UUID PRIMARY KEY DEFAULT gen_random_uuid(), national_dpide_id   VARCHAR(25) UNIQUE NOT NULL,  -- DPIDE-HT-AAAA-NNNNNN snisid_person_id    UUID, full_name           VARCHAR(200) NOT NULL, dob                 DATE, gender              VARCHAR(10), household_size      SMALLINT DEFAULT 1, minors_count        SMALLINT DEFAULT 0, displacement_cause  dpide_displacement_cause NOT NULL, displacement_date   TIMESTAMPTZ NOT NULL, origin_address      TEXT, origin_dept_code    CHAR(2) NOT NULL, origin_commune      VARCHAR(100), status              dpide_idp_status NOT NULL DEFAULT 'DISPLACED', current_location    TEXT, current_dept_code   CHAR(2), current_commune     VARCHAR(100), current_lat         DECIMAL(10,7), current_lng         DECIMAL(10,7), camp_id             UUID, shelter_type        VARCHAR(50),   -- CAMP, HOST_FAMILY, RENTED, SPONTANEOUS has_nfi             BOOLEAN DEFAULT FALSE,      -- Non-Food Items receives_food_aid   BOOLEAN DEFAULT FALSE, has_latrines        BOOLEAN DEFAULT FALSE, has_water_access    BOOLEAN DEFAULT FALSE, medical_needs       TEXT[] DEFAULT '{}', iom_dtm_ref         VARCHAR(50), ocha_ref            VARCHAR(50), created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS dpide_camps ( camp_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(), camp_name           VARCHAR(150) NOT NULL, dept_code           CHAR(2) NOT NULL, commune             VARCHAR(100), lat                 DECIMAL(10,7), lng                 DECIMAL(10,7), displacement_cause  dpide_displacement_cause, managing_org        VARCHAR(150), capacity            INTEGER, current_population  INTEGER DEFAULT 0, is_active           BOOLEAN DEFAULT TRUE, has_medical_post    BOOLEAN DEFAULT FALSE, has_school          BOOLEAN DEFAULT FALSE, water_source        TEXT, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_dpide_status      ON dpide_idps(status, displacement_cause);",
		"CREATE INDEX IF NOT EXISTS idx_dpide_dept        ON dpide_idps(current_dept_code) WHERE status IN ('DISPLACED','IN_CAMP');",
		"CREATE INDEX IF NOT EXISTS idx_dpide_cause       ON dpide_idps(displacement_cause);",
		"CREATE INDEX IF NOT EXISTS idx_dpide_camp        ON dpide_idps(camp_id) WHERE camp_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_dpide_snisid      ON dpide_idps(snisid_person_id) WHERE snisid_person_id IS NOT NULL;",
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
