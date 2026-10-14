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
	"github.com/snisid/pki-ht/internal/handler"
	"github.com/snisid/pki-ht/internal/kafka"
	"github.com/snisid/pki-ht/internal/repository"
	"github.com/snisid/pki-ht/internal/service"
)

func main() {
	dbHost := getEnv("PKI_DB_HOST", "localhost")
	dbPort := getEnv("PKI_DB_PORT", "26257")
	dbName := getEnv("PKI_DB_NAME", "snisid_pki")
	dbUser := getEnv("PKI_DB_USER", "root")
	kBrokers := getEnv("PKI_KAFKA_BROKERS", "localhost:9092")
	kTopic := getEnv("PKI_KAFKA_TOPIC", "snisid.pki.events")
	port := getEnv("PKI_SERVICE_PORT", "8086")

	dbURL := fmt.Sprintf("postgresql://%s@%s:%s/%s?sslmode=%s", dbUser, dbHost, dbPort, dbName, "disable")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	if err := runMigrations(db); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	producer := kafka.NewProducer([]string{kBrokers}, kTopic)
	defer producer.Close()

	repo := repository.NewPostgresRepo(db)
	svc := service.NewPKIService(repo, producer)
	svc.InitCA(context.Background())

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	h := handler.NewHandler(svc)
	api := r.Group("/api/v1/pki")
	h.RegisterRoutes(api)

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		log.Printf("pki-ht started on port %s", port)
		if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			log.Fatalf("pki-ht: %v", e)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func runMigrations(db *sql.DB) error {
	for _, m := range []string{
		`CREATE TYPE IF NOT EXISTS pki_ca_type AS ENUM ('ROOT','INTERMEDIATE_CITIZENS','INTERMEDIATE_SERVICES','INTERMEDIATE_AGENCIES')`,
		`CREATE TYPE IF NOT EXISTS pki_cert_status AS ENUM ('VALID','REVOKED','EXPIRED','SUSPENDED')`,
		`CREATE TABLE IF NOT EXISTS pki_certificate_authorities (ca_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), ca_type pki_ca_type NOT NULL, common_name VARCHAR(200) NOT NULL, serial_number VARCHAR(100) UNIQUE NOT NULL, public_key_pem TEXT NOT NULL, hsm_key_ref VARCHAR(200), parent_ca_id UUID REFERENCES pki_certificate_authorities(ca_id), valid_from TIMESTAMPTZ NOT NULL, valid_until TIMESTAMPTZ NOT NULL, is_active BOOLEAN DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS pki_issued_certificates (cert_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), serial_number VARCHAR(100) UNIQUE NOT NULL, issuing_ca_id UUID NOT NULL REFERENCES pki_certificate_authorities(ca_id), subject_type VARCHAR(30) NOT NULL, subject_ref UUID, common_name VARCHAR(200), status pki_cert_status NOT NULL DEFAULT 'VALID', valid_from TIMESTAMPTZ NOT NULL, valid_until TIMESTAMPTZ NOT NULL, revoked_at TIMESTAMPTZ, revocation_reason TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS pki_crl (crl_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), ca_id UUID NOT NULL REFERENCES pki_certificate_authorities(ca_id), crl_number BIGINT NOT NULL, revoked_serials TEXT[] DEFAULT '{}', published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), next_update TIMESTAMPTZ NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_pki_certs_subject ON pki_issued_certificates(subject_ref)`,
		`CREATE INDEX IF NOT EXISTS idx_pki_certs_status ON pki_issued_certificates(status)`,
	} {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration: %w", err)
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
