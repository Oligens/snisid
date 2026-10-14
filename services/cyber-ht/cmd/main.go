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
	"github.com/snisid/cyber-ht/internal/handler"
	"github.com/snisid/cyber-ht/internal/kafka"
	"github.com/snisid/cyber-ht/internal/repository"
	"github.com/snisid/cyber-ht/internal/service"
)

func main() {
	dbHost := getEnv("CYBER_DB_HOST", "localhost")
	dbPort := getEnv("CYBER_DB_PORT", "26257")
	dbName := getEnv("CYBER_DB_NAME", "snisid_cyber")
	dbUser := getEnv("CYBER_DB_USER", "root")
	dbSSLMode := getEnv("CYBER_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("CYBER_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("CYBER_KAFKA_TOPIC", "snisid.cyber.events")
	port := getEnv("CYBER_SERVICE_PORT", "8090")

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
	svc := service.NewCyberService(repo, producer)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	h := handler.NewHandler(svc)
	api := r.Group("/api/v1/cyber")
	h.RegisterRoutes(api)

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		log.Printf("cyber-ht service started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run cyber-ht: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down cyber-ht...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TYPE IF NOT EXISTS cyber_severity AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')`,
		`CREATE TYPE IF NOT EXISTS cyber_incident_status AS ENUM ('DETECTED', 'TRIAGING', 'CONTAINED', 'ERADICATED', 'RECOVERED', 'CLOSED')`,
		`CREATE TABLE IF NOT EXISTS cyber_incidents (
			id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title           VARCHAR(255) NOT NULL,
			description     TEXT,
			severity        cyber_severity NOT NULL,
			status          cyber_incident_status NOT NULL DEFAULT 'DETECTED',
			source_ip       VARCHAR(45),
			target_asset    VARCHAR(255),
			detected_by     VARCHAR(255) NOT NULL,
			assigned_to     VARCHAR(255),
			created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			closed_at       TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cyber_incidents_status ON cyber_incidents(status)`,
		`CREATE INDEX IF NOT EXISTS idx_cyber_incidents_severity ON cyber_incidents(severity)`,
		`CREATE TABLE IF NOT EXISTS cyber_zero_trust_policies (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name        VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,
			policy_type VARCHAR(100) NOT NULL,
			rules       TEXT[] NOT NULL,
			enabled     BOOLEAN NOT NULL DEFAULT TRUE,
			created_by  VARCHAR(255) NOT NULL,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cyber_zt_policies_type ON cyber_zero_trust_policies(policy_type)`,
		`CREATE TABLE IF NOT EXISTS cyber_threat_indicators (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			indicator    VARCHAR(500) UNIQUE NOT NULL,
			type         VARCHAR(100) NOT NULL,
			threat_level VARCHAR(50) NOT NULL,
			source       VARCHAR(255) NOT NULL,
			description  TEXT,
			tags         TEXT[] DEFAULT '{}',
			expires_at   TIMESTAMPTZ,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cyber_threat_indicator ON cyber_threat_indicators(indicator)`,
		`CREATE INDEX IF NOT EXISTS idx_cyber_threat_expires ON cyber_threat_indicators(expires_at)`,
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
