package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/snisid/data-ht/internal/handler"
	"github.com/snisid/data-ht/internal/kafka"
	"github.com/snisid/data-ht/internal/repository"
	"github.com/snisid/data-ht/internal/service"
)

func main() {
	dbHost := getEnv("DATA_DB_HOST", "localhost")
	dbPort := getEnv("DATA_DB_PORT", "26257")
	dbName := getEnv("DATA_DB_NAME", "snisid_data")
	dbUser := getEnv("DATA_DB_USER", "root")
	dbSSLMode := getEnv("DATA_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("DATA_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("DATA_KAFKA_TOPIC", "snisid.data.events")
	port := getEnv("DATA_SERVICE_PORT", "8093")

	dbURL := fmt.Sprintf("postgresql://%s@%s:%s/%s?sslmode=%s", dbUser, dbHost, dbPort, dbName, dbSSLMode)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to CockroachDB: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping CockroachDB: %v", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	producer := kafka.NewProducer([]string{kafkaBrokers}, kafkaTopic)
	defer producer.Close()

	repo := repository.NewPostgresRepo(db)
	svc := service.NewDataService(repo, producer)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	h := handler.NewHandler(svc)
	api := r.Group("/api/v1/data")
	h.RegisterRoutes(api)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("data-ht service started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run data-ht: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down data-ht...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TYPE IF NOT EXISTS data_destination AS ENUM ('CLICKHOUSE', 'S3_PARQUET', 'FEATURE_STORE')`,
		`CREATE TABLE IF NOT EXISTS data_pipelines (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name         VARCHAR(100) NOT NULL,
			source_topics TEXT[],
			destination  data_destination NOT NULL,
			config       JSONB,
			is_active    BOOLEAN DEFAULT TRUE,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS data_ml_models (
			id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name             VARCHAR(100) NOT NULL,
			model_type       VARCHAR(50) NOT NULL,
			version          VARCHAR(20) NOT NULL,
			mlflow_run_id    VARCHAR(100) NOT NULL,
			bias_metric      VARCHAR(50),
			bias_score       DECIMAL(5,4),
			training_date    TIMESTAMPTZ,
			is_active        BOOLEAN DEFAULT TRUE,
			created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS data_governance_audit (
			id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			model_id      UUID NOT NULL REFERENCES data_ml_models(id),
			audit_type    VARCHAR(50) NOT NULL,
			findings      JSONB,
			conducted_by  UUID NOT NULL,
			conducted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_data_pipelines_dest ON data_pipelines(destination)`,
		`CREATE INDEX IF NOT EXISTS idx_data_models_type ON data_ml_models(model_type, is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_data_audit_model ON data_governance_audit(model_id, conducted_at DESC)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %s: %w", m[:60], err)
		}
	}
	return nil
}

func getEnv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
