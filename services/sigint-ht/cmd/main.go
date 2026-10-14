package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/snisid/sigint-ht/internal/handler"
	sigintKafka "github.com/snisid/sigint-ht/internal/kafka"
	"github.com/snisid/sigint-ht/internal/repository"
	"github.com/snisid/sigint-ht/internal/service"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TYPE IF NOT EXISTS sigint_target_type AS ENUM ('PHONE_NUMBER', 'EMAIL', 'SOCIAL_MEDIA', 'RADIO_FREQUENCY', 'IP_ADDRESS')`,
		`CREATE TYPE IF NOT EXISTS sigint_target_status AS ENUM ('ACTIVE', 'SUSPENDED', 'EXPIRED', 'REVOKED')`,
		`CREATE TYPE IF NOT EXISTS sigint_comm_type AS ENUM ('CALL', 'SMS', 'EMAIL', 'RADIO', 'SOCIAL_MEDIA')`,
		`CREATE TABLE IF NOT EXISTS sigint_targets (
			id UUID PRIMARY KEY,
			target_type sigint_target_type NOT NULL,
			status sigint_target_status NOT NULL DEFAULT 'ACTIVE',
			authorization_ref UUID NOT NULL,
			judge_name VARCHAR(255) NOT NULL,
			issuing_court VARCHAR(255) NOT NULL,
			start_date TIMESTAMPTZ NOT NULL,
			end_date TIMESTAMPTZ NOT NULL,
			target_identifier TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS sigint_intercepted_comms (
			id UUID PRIMARY KEY,
			source_target_id UUID NOT NULL REFERENCES sigint_targets(id),
			comm_type sigint_comm_type NOT NULL,
			metadata JSONB DEFAULT '{}',
			content_ref TEXT NOT NULL,
			intercepted_at TIMESTAMPTZ NOT NULL,
			collector_node VARCHAR(255) NOT NULL,
			case_number VARCHAR(255),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS sigint_cdr_analysis (
			id UUID PRIMARY KEY,
			caller VARCHAR(255) NOT NULL,
			callee VARCHAR(255) NOT NULL,
			duration INT NOT NULL,
			tower_location VARCHAR(255),
			imsi VARCHAR(64),
			imei VARCHAR(64),
			timestamp TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sigint_targets_status ON sigint_targets(status)`,
		`CREATE INDEX IF NOT EXISTS idx_sigint_targets_identifier ON sigint_targets(target_identifier)`,
		`CREATE INDEX IF NOT EXISTS idx_sigint_comms_target ON sigint_intercepted_comms(source_target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sigint_cdr_caller ON sigint_cdr_analysis(caller)`,
		`CREATE INDEX IF NOT EXISTS idx_sigint_cdr_callee ON sigint_cdr_analysis(callee)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %s: %w", m[:60], err)
		}
	}
	return nil
}

func main() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("SIGINT_DB_HOST", "localhost"),
		getEnv("SIGINT_DB_PORT", "26257"),
		getEnv("SIGINT_DB_USER", "root"),
		getEnv("SIGINT_DB_PASSWORD", ""),
		getEnv("SIGINT_DB_NAME", "snisid_sigint"),
		getEnv("SIGINT_DB_SSLMODE", "disable"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("migrations completed successfully")

	kafkaBrokers := getEnv("SIGINT_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("SIGINT_KAFKA_TOPIC", "sigint-events")

	producer := sigintKafka.NewProducer([]string{kafkaBrokers}, kafkaTopic)
	defer producer.Close()

	repo := repository.NewPostgresRepo(db)
	svc := service.NewSigintService(repo, producer)
	h := handler.NewSigintHandler(svc)

	ginMode := getEnv("SIGINT_GIN_MODE", "release")
	gin.SetMode(ginMode)
	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		if err := repo.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	h.RegisterRoutes(router)

	port := getEnv("SIGINT_PORT", "8301")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("sigint-ht server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down sigint-ht server...")
}
