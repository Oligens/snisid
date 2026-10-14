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
	"github.com/snisid/interop-ht/internal/handler"
	"github.com/snisid/interop-ht/internal/kafka"
	"github.com/snisid/interop-ht/internal/repository"
	"github.com/snisid/interop-ht/internal/service"
)

func main() {
	dbURL := fmt.Sprintf("postgresql://root@%s:26257/%s?sslmode=disable", getEnv("INTEROP_DB_HOST", "localhost"), getEnv("INTEROP_DB_NAME", "snisid_interop"))
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	for _, m := range []string{
		`CREATE TABLE IF NOT EXISTS interop_agencies (agency_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), agency_code VARCHAR(20) UNIQUE NOT NULL, agency_name VARCHAR(150) NOT NULL, security_server_url VARCHAR(300) NOT NULL, public_key_cert_ref VARCHAR(200), is_active BOOLEAN DEFAULT TRUE, onboarded_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS interop_data_exchange_agreements (agreement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), provider_agency_id UUID NOT NULL REFERENCES interop_agencies(agency_id), consumer_agency_id UUID NOT NULL REFERENCES interop_agencies(agency_id), service_name VARCHAR(150) NOT NULL, allowed_fields TEXT[] DEFAULT '{}', legal_basis TEXT, rate_limit_per_min INTEGER DEFAULT 1000, is_active BOOLEAN DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS interop_exchange_log (log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), agreement_id UUID NOT NULL REFERENCES interop_data_exchange_agreements(agreement_id), request_hash VARCHAR(64) NOT NULL, response_size_bytes INTEGER, status_code SMALLINT, duration_ms INTEGER, exchanged_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE INDEX IF NOT EXISTS idx_interop_log_agreement ON interop_exchange_log(agreement_id, exchanged_at DESC)`,
	} {
		if _, err := db.Exec(m); err != nil {
			log.Fatalf("migration: %v", err)
		}
	}

	p := kafka.NewProducer([]string{getEnv("INTEROP_KAFKA_BROKERS", "localhost:9092")}, "snisid.interop.events")
	defer p.Close()

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	h := handler.NewHandler(service.NewInteropService(repository.NewPostgresRepo(db), p))
	h.RegisterRoutes(r.Group("/api/v1/interop"))

	srv := &http.Server{Addr: ":" + getEnv("INTEROP_SERVICE_PORT", "8088"), Handler: r}
	go func() { log.Printf("interop-ht on :%s", getEnv("INTEROP_SERVICE_PORT", "8088")); srv.ListenAndServe() }()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
