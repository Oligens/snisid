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
	"github.com/snisid/iam-ht/internal/handler"
	"github.com/snisid/iam-ht/internal/kafka"
	"github.com/snisid/iam-ht/internal/repository"
	"github.com/snisid/iam-ht/internal/service"
)

func main() {
	dbHost := getEnv("IAM_DB_HOST", "localhost")
	dbPort := getEnv("IAM_DB_PORT", "26257")
	dbName := getEnv("IAM_DB_NAME", "snisid_iam")
	port := getEnv("IAM_SERVICE_PORT", "8087")
	kBrokers := getEnv("IAM_KAFKA_BROKERS", "localhost:9092")
	kTopic := getEnv("IAM_KAFKA_TOPIC", "snisid.iam.events")

	dbURL := fmt.Sprintf("postgresql://%s@%s:%s/%s?sslmode=%s", "root", dbHost, dbPort, dbName, "disable")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	for _, m := range []string{
		`CREATE TYPE IF NOT EXISTS iam_assurance_level AS ENUM ('IAL1_SELF_ASSERTED','IAL2_BIOMETRIC_VERIFIED','IAL3_IN_PERSON')`,
		`CREATE TABLE IF NOT EXISTS iam_identity_assurance (assurance_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), citizen_id UUID NOT NULL, keycloak_user_id VARCHAR(100) UNIQUE NOT NULL, assurance_level iam_assurance_level NOT NULL DEFAULT 'IAL1_SELF_ASSERTED', biometric_verified_at TIMESTAMPTZ, mfa_enrolled BOOLEAN DEFAULT FALSE, last_login_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS iam_agency_clients (client_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), agency_name VARCHAR(150) NOT NULL, oauth_client_id VARCHAR(100) UNIQUE NOT NULL, allowed_scopes TEXT[] DEFAULT '{}', redirect_uris TEXT[] DEFAULT '{}', required_assurance_level iam_assurance_level DEFAULT 'IAL1_SELF_ASSERTED', is_active BOOLEAN DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS iam_access_log (log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), citizen_id UUID, client_id UUID REFERENCES iam_agency_clients(client_id), action VARCHAR(50) NOT NULL, ip_hash VARCHAR(64), accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
	} {
		if _, err := db.Exec(m); err != nil {
			log.Fatalf("migration: %v", err)
		}
	}

	producer := kafka.NewProducer([]string{kBrokers}, kTopic)
	defer producer.Close()

	repo := repository.NewPostgresRepo(db)
	svc := service.NewIAMService(repo, producer)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	h := handler.NewHandler(svc)
	api := r.Group("/api/v1/iam")
	h.RegisterRoutes(api)

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		log.Printf("iam-ht started on port %s", port)
		if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			log.Fatalf("iam-ht: %v", e)
		}
	}()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
