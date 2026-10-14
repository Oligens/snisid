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
	"github.com/snisid/api-ht/internal/handler"
	"github.com/snisid/api-ht/internal/kafka"
	"github.com/snisid/api-ht/internal/repository"
	"github.com/snisid/api-ht/internal/service"
)

func main() {
	dbHost := getEnv("API_DB_HOST", "localhost")
	dbPort := getEnv("API_DB_PORT", "26257")
	dbName := getEnv("API_DB_NAME", "snisid_api")
	dbUser := getEnv("API_DB_USER", "root")
	dbSSLMode := getEnv("API_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("API_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("API_KAFKA_TOPIC", "snisid.api.events")
	port := getEnv("API_SERVICE_PORT", "8094")

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
	svc := service.NewAPIService(repo, producer)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	h := handler.NewHandler(svc)
	api := r.Group("/api/v1/devportal")
	h.RegisterRoutes(api)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("api-ht service started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run api-ht: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down api-ht...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TYPE IF NOT EXISTS api_sensitivity AS ENUM (
			'PUBLIC', 'RESTRICTED', 'CONFIDENTIAL', 'SECRET'
		)`,
		`CREATE TABLE IF NOT EXISTS api_catalog (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			path              VARCHAR(255) NOT NULL,
			method            VARCHAR(10) NOT NULL,
			description       TEXT,
			sensitivity       api_sensitivity NOT NULL DEFAULT 'PUBLIC',
			module_source     VARCHAR(100),
			base_path         VARCHAR(100) NOT NULL,
			is_active         BOOLEAN DEFAULT TRUE,
			version           VARCHAR(20) NOT NULL DEFAULT 'v1',
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS api_developer_accounts (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email             VARCHAR(255) UNIQUE NOT NULL,
			org_name          VARCHAR(200),
			contact_name      VARCHAR(200) NOT NULL,
			contact_phone     VARCHAR(30),
			is_approved       BOOLEAN DEFAULT FALSE,
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			account_id        UUID NOT NULL REFERENCES api_developer_accounts(id),
			key_value         VARCHAR(100) UNIQUE NOT NULL,
			description       TEXT,
			is_active         BOOLEAN DEFAULT TRUE,
			expires_at        TIMESTAMPTZ,
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			revoked_at        TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS api_usage_log (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			key_id            UUID NOT NULL REFERENCES api_keys(id),
			endpoint          VARCHAR(255) NOT NULL,
			method            VARCHAR(10) NOT NULL,
			status            INTEGER NOT NULL,
			latency_ms        INTEGER NOT NULL DEFAULT 0,
			ip_address        VARCHAR(45),
			user_agent        TEXT,
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_account ON api_keys(account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_value ON api_keys(key_value)`,
		`CREATE INDEX IF NOT EXISTS idx_api_usage_key ON api_usage_log(key_id)`,
		`CREATE INDEX IF NOT EXISTS idx_api_usage_created ON api_usage_log(created_at)`,
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
