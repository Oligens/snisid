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

	"github.com/snisid/platform/services/expl-svc/internal/handler"
	"github.com/snisid/platform/services/expl-svc/internal/kafka"
	"github.com/snisid/platform/services/expl-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/expl-svc/internal/service"
)

func main() {
	dbHost := getEnv("EXPL_DB_HOST", "localhost")
	dbPort := getEnv("EXPL_DB_PORT", "26257")
	dbName := getEnv("EXPL_DB_NAME", "snisid_expl")
	dbUser := getEnv("EXPL_DB_USER", "root")
	dbSSLMode := getEnv("EXPL_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("EXPL_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("EXPL_KAFKA_TOPIC", "snisid.expl.events")
	port := getEnv("EXPL_SERVICE_PORT", "8104")

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

	repo := postgres.NewIncidentRepository(pool, logger)
	svc := service.NewExplService(repo, logger)

	r := handler.SetupRouter(svc, logger)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("expl-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS expl_type AS ENUM ( 'IED','GRENADE','RPG','MORTAR','LANDMINE','DYNAMITE', 'BLASTING_CAP','AMMUNITION_BULK','MILITARY_ORDNANCE','UNKNOWN' );",
		"CREATE TYPE IF NOT EXISTS expl_status AS ENUM ( 'RECOVERED','DESTROYED','DETONATED','STORED_EVIDENCE','TRANSFERRED' );",
		"CREATE TABLE IF NOT EXISTS expl_incidents ( incident_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), national_expl_id    VARCHAR(25) UNIQUE NOT NULL, incident_type       VARCHAR(30) NOT NULL,     -- FIND, DETONATION, SEIZURE, SURRENDER explosive_type      expl_type NOT NULL, status              expl_status NOT NULL DEFAULT 'RECOVERED', quantity            INTEGER DEFAULT 1, weight_kg           DECIMAL(10,3), manufacturer        VARCHAR(100), lot_number          VARCHAR(50), manufacture_country CHAR(3), estimated_date      DATE, incident_date       TIMESTAMPTZ NOT NULL, location_desc       VARCHAR(300), dept_code           CHAR(2), commune             VARCHAR(100), lat                 DECIMAL(10,7), lng                 DECIMAL(10,7), responding_unit     VARCHAR(50), eod_officer         UUID, casualties          SMALLINT DEFAULT 0, gang_id             UUID, from_person_id      UUID, case_reference      VARCHAR(100), dna_sample_taken    BOOLEAN DEFAULT FALSE, bio_sample_ref      VARCHAR(100), photo_refs          TEXT[] DEFAULT '{}', interpol_exploint_ref VARCHAR(50), notes               TEXT, created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS expl_legal_stocks ( stock_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(), holder_entity       VARCHAR(200) NOT NULL, holder_type         VARCHAR(30),     -- MINING, CONSTRUCTION, MILITARY, POLICE explosive_type      expl_type NOT NULL, quantity_kg         DECIMAL(12,3) NOT NULL, storage_location    TEXT NOT NULL, dept_code           CHAR(2), license_ref         VARCHAR(50), last_audit_date     DATE, next_audit_date     DATE, is_secured          BOOLEAN DEFAULT TRUE, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_expl_type  ON expl_incidents(explosive_type, incident_date DESC);",
		"CREATE INDEX IF NOT EXISTS idx_expl_dept  ON expl_incidents(dept_code);",
		"CREATE INDEX IF NOT EXISTS idx_expl_gang  ON expl_incidents(gang_id) WHERE gang_id IS NOT NULL;",
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
