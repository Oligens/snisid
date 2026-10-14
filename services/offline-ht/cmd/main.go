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
	"github.com/snisid/offline-ht/internal/handler"
	"github.com/snisid/offline-ht/internal/kafka"
	"github.com/snisid/offline-ht/internal/repository"
	"github.com/snisid/offline-ht/internal/service"
)

func main() {
	dbHost := getEnv("OFFLINE_DB_HOST", "localhost")
	dbPort := getEnv("OFFLINE_DB_PORT", "26257")
	dbName := getEnv("OFFLINE_DB_NAME", "snisid_offline")
	dbUser := getEnv("OFFLINE_DB_USER", "root")
	dbSSLMode := getEnv("OFFLINE_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("OFFLINE_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("OFFLINE_KAFKA_TOPIC", "snisid.offline.events")
	port := getEnv("OFFLINE_SERVICE_PORT", "8091")

	dbURL := fmt.Sprintf("postgresql://%s@%s:%s/%s?sslmode=%s", dbUser, dbHost, dbPort, dbName, dbSSLMode)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to CockroachDB: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping CockroachDB: %v", err)
	}
	db.SetMaxOpenConns(25)

	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	producer := kafka.NewProducer([]string{kafkaBrokers}, kafkaTopic)
	defer producer.Close()

	repo := repository.NewPostgresRepo(db)
	svc := service.NewOfflineService(repo, producer)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	h := handler.NewHandler(svc)
	api := r.Group("/api/v1/offline")
	h.RegisterRoutes(api)

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		log.Printf("offline-ht service started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run offline-ht: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down offline-ht...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TYPE IF NOT EXISTS offline_sync_status AS ENUM ('PENDING', 'SYNCING', 'SYNCED', 'CONFLICT', 'FAILED')`,
		`CREATE TABLE IF NOT EXISTS offline_sync_queue (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			terminal_id  UUID NOT NULL,
			entity_type  VARCHAR(100) NOT NULL,
			entity_id    VARCHAR(255) NOT NULL,
			action       VARCHAR(50) NOT NULL,
			payload      TEXT NOT NULL,
			status       offline_sync_status NOT NULL DEFAULT 'PENDING',
			retry_count  INT NOT NULL DEFAULT 0,
			error_msg    TEXT,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			synced_at    TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_offline_queue_terminal ON offline_sync_queue(terminal_id)`,
		`CREATE INDEX IF NOT EXISTS idx_offline_queue_status ON offline_sync_queue(status)`,
		`CREATE TABLE IF NOT EXISTS offline_terminals (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name         VARCHAR(255) NOT NULL,
			location     VARCHAR(255) NOT NULL,
			last_sync_at TIMESTAMPTZ,
			firmware_ver VARCHAR(50) NOT NULL DEFAULT '1.0.0',
			is_online    BOOLEAN NOT NULL DEFAULT FALSE,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_offline_terminal_online ON offline_terminals(is_online)`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration: %s: %w", m[:60], err)
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
