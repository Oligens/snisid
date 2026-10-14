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
	"github.com/snisid/card-ht/internal/handler"
	"github.com/snisid/card-ht/internal/kafka"
	"github.com/snisid/card-ht/internal/repository"
	"github.com/snisid/card-ht/internal/service"
)

func main() {
	dbHost := getEnv("CARD_DB_HOST", "localhost")
	dbPort := getEnv("CARD_DB_PORT", "26257")
	dbName := getEnv("CARD_DB_NAME", "snisid_card")
	dbUser := getEnv("CARD_DB_USER", "root")
	dbSSLMode := getEnv("CARD_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("CARD_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("CARD_KAFKA_TOPIC", "snisid.card.events")
	port := getEnv("CARD_SERVICE_PORT", "8084")

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
	svc := service.NewCardService(repo, producer)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	h := handler.NewHandler(svc)
	api := r.Group("/api/v1/card")
	h.RegisterRoutes(api)

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		log.Printf("card-ht service started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run card-ht: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down card-ht...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TYPE IF NOT EXISTS card_doc_type AS ENUM ('NATIONAL_ID', 'PASSPORT', 'RESIDENCE_PERMIT', 'REFUGEE_DOC')`,
		`CREATE TYPE IF NOT EXISTS card_status AS ENUM ('ISSUED', 'ACTIVE', 'EXPIRED', 'REVOKED', 'LOST', 'STOLEN', 'RENEWED')`,
		`CREATE TABLE IF NOT EXISTS card_documents (
			document_id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			document_number          VARCHAR(20) UNIQUE NOT NULL,
			doc_type                 card_doc_type NOT NULL,
			citizen_id               UUID NOT NULL,
			status                   card_status NOT NULL DEFAULT 'ISSUED',
			chip_serial              VARCHAR(50) UNIQUE,
			mrz_line1                VARCHAR(44),
			mrz_line2                VARCHAR(44),
			public_key_cert_ref      VARCHAR(200),
			issue_date               DATE NOT NULL,
			expiry_date              DATE NOT NULL,
			issuing_office           VARCHAR(150),
			personalization_facility VARCHAR(150) DEFAULT 'Imprimerie Nationale PAP',
			photo_ref                VARCHAR(500),
			signature_ref            VARCHAR(500),
			sltd_reported            BOOLEAN DEFAULT FALSE,
			created_by               UUID NOT NULL,
			created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_card_citizen ON card_documents(citizen_id)`,
		`CREATE INDEX IF NOT EXISTS idx_card_status ON card_documents(status)`,
		`CREATE INDEX IF NOT EXISTS idx_card_chip ON card_documents(chip_serial)`,
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
