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
	"github.com/snisid/bio-ht/internal/handler"
	"github.com/snisid/bio-ht/internal/kafka"
	"github.com/snisid/bio-ht/internal/milvus"
	"github.com/snisid/bio-ht/internal/repository"
	"github.com/snisid/bio-ht/internal/service"
)

func main() {
	dbHost := getEnv("BIO_DB_HOST", "localhost")
	dbPort := getEnv("BIO_DB_PORT", "26257")
	dbName := getEnv("BIO_DB_NAME", "snisid_bio")
	dbUser := getEnv("BIO_DB_USER", "root")
	dbSSLMode := getEnv("BIO_DB_SSLMODE", "disable")
	milvusAddr := getEnv("BIO_MILVUS_ADDR", "localhost:19530")
	kafkaBrokers := getEnv("BIO_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("BIO_KAFKA_TOPIC", "snisid.bio.events")
	port := getEnv("BIO_SERVICE_PORT", "8083")

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

	milvusClient, err := milvus.NewClient(milvusAddr)
	if err != nil {
		log.Fatalf("failed to connect to Milvus: %v", err)
	}
	defer milvusClient.Close()

	producer := kafka.NewProducer([]string{kafkaBrokers}, kafkaTopic)
	defer producer.Close()

	repo := repository.NewPostgresRepo(db)
	svc := service.NewBioService(repo, milvusClient, producer)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	h := handler.NewHandler(svc)
	api := r.Group("/api/v1/bio")
	h.RegisterRoutes(api)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("bio-ht service started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run bio-ht: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down bio-ht...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TYPE IF NOT EXISTS bio_modality AS ENUM ('FINGERPRINT', 'FACE', 'IRIS', 'VOICE')`,
		`CREATE TABLE IF NOT EXISTS bio_templates (
			template_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			citizen_id             UUID NOT NULL,
			modality               bio_modality NOT NULL,
			milvus_vector_id       VARCHAR(100) NOT NULL,
			quality_score          DECIMAL(5,2),
			capture_device         VARCHAR(100),
			capture_location       VARCHAR(150),
			captured_by            UUID NOT NULL,
			is_active              BOOLEAN DEFAULT TRUE,
			superseded_by_template_id UUID,
			created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS bio_verification_log (
			verification_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			citizen_id        UUID,
			modality          bio_modality NOT NULL,
			requesting_module VARCHAR(50) NOT NULL,
			match_score       DECIMAL(5,4),
			is_match          BOOLEAN,
			verified_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_bio_templates_citizen ON bio_templates(citizen_id) WHERE is_active = TRUE`,
		`CREATE INDEX IF NOT EXISTS idx_bio_verif_citizen ON bio_verification_log(citizen_id, verified_at DESC)`,
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
