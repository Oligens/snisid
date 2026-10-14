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

	"github.com/snisid/civil-ht/internal/handler"
	"github.com/snisid/civil-ht/internal/kafka"
	"github.com/snisid/civil-ht/internal/repository"
	"github.com/snisid/civil-ht/internal/service"
)

func main() {
	dbHost := getEnv("CIVIL_DB_HOST", "localhost")
	dbPort := getEnv("CIVIL_DB_PORT", "26257")
	dbName := getEnv("CIVIL_DB_NAME", "snisid_civil")
	dbUser := getEnv("CIVIL_DB_USER", "root")
	dbSSLMode := getEnv("CIVIL_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("CIVIL_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("CIVIL_KAFKA_TOPIC", "snisid.civil.events")
	port := getEnv("CIVIL_SERVICE_PORT", "8082")

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
	svc := service.NewCivilService(repo, producer)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	h := handler.NewHandler(svc)
	api := r.Group("/api/v1/civil")
	h.RegisterRoutes(api)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("civil-ht service started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run civil-ht: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down civil-ht...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TYPE IF NOT EXISTS civil_act_type AS ENUM (
			'BIRTH', 'DEATH', 'MARRIAGE', 'DIVORCE', 'ADOPTION', 'RECOGNITION_PATERNITY'
		)`,
		`CREATE TABLE IF NOT EXISTS civil_acts (
			act_id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			act_number           VARCHAR(30) UNIQUE NOT NULL,
			act_type             civil_act_type NOT NULL,
			citizen_id           UUID,
			registering_office   VARCHAR(150) NOT NULL,
			dept_code            CHAR(2) NOT NULL,
			commune              VARCHAR(100) NOT NULL,
			event_date           DATE NOT NULL,
			declared_date        DATE NOT NULL DEFAULT CURRENT_DATE,
			officer_name         VARCHAR(150),
			officer_id           UUID,
			is_late_declaration  BOOLEAN DEFAULT FALSE,
			is_reconstructed     BOOLEAN DEFAULT FALSE,
			created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS civil_birth_details (
			act_id                UUID PRIMARY KEY REFERENCES civil_acts(act_id),
			child_full_name       VARCHAR(200) NOT NULL,
			child_gender          VARCHAR(10),
			mother_citizen_id     UUID,
			father_citizen_id     UUID,
			birth_weight_g        INTEGER,
			birth_facility        VARCHAR(150),
			attending_professional VARCHAR(150)
		)`,
		`CREATE TABLE IF NOT EXISTS civil_marriage_details (
			act_id                UUID PRIMARY KEY REFERENCES civil_acts(act_id),
			spouse_a_citizen_id   UUID NOT NULL,
			spouse_b_citizen_id   UUID NOT NULL,
			marriage_regime       VARCHAR(50),
			prenuptial_agreement  BOOLEAN DEFAULT FALSE
		)`,
		`CREATE TABLE IF NOT EXISTS civil_death_details (
			act_id               UUID PRIMARY KEY REFERENCES civil_acts(act_id),
			deceased_citizen_id   UUID NOT NULL,
			cause_of_death        TEXT,
			death_location        VARCHAR(300),
			medical_certifier     VARCHAR(150),
			is_violent_death      BOOLEAN DEFAULT FALSE,
			fir_case_reference    VARCHAR(100)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_civil_acts_citizen ON civil_acts(citizen_id)`,
		`CREATE INDEX IF NOT EXISTS idx_civil_acts_type ON civil_acts(act_type, dept_code)`,
		`CREATE INDEX IF NOT EXISTS idx_civil_acts_date ON civil_acts(event_date)`,
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
