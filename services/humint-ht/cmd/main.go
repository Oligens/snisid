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

	"github.com/snisid/humint-ht/internal/handler"
	humintKafka "github.com/snisid/humint-ht/internal/kafka"
	"github.com/snisid/humint-ht/internal/repository"
	"github.com/snisid/humint-ht/internal/service"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TYPE IF NOT EXISTS humint_source_status AS ENUM ('ACTIVE', 'COMPROMISED', 'TERMINATED', 'DEAD')`,
		`CREATE TYPE IF NOT EXISTS humint_payment_freq AS ENUM ('ONE_TIME', 'MONTHLY', 'PER_REPORT')`,
		`CREATE TYPE IF NOT EXISTS humint_risk_level AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')`,
		`CREATE TYPE IF NOT EXISTS humint_report_class AS ENUM ('UNCLASSIFIED', 'CONFIDENTIAL', 'SECRET', 'TOP_SECRET')`,
		`CREATE TYPE IF NOT EXISTS humint_debrief_method AS ENUM ('IN_PERSON', 'PHONE', 'ENCRYPTED_APP', 'DEAD_DROP')`,
		`CREATE TABLE IF NOT EXISTS humint_sources (
			code_name VARCHAR(255) PRIMARY KEY,
			credibility_rating INT NOT NULL CHECK (credibility_rating BETWEEN 1 AND 6),
			reliability_rating CHAR(1) NOT NULL CHECK (reliability_rating IN ('A','B','C','D','E','F')),
			status humint_source_status NOT NULL DEFAULT 'ACTIVE',
			handling_officer_id UUID NOT NULL,
			payment_amount DECIMAL(12,2) DEFAULT 0,
			payment_frequency humint_payment_freq DEFAULT 'ONE_TIME',
			risk_level humint_risk_level NOT NULL DEFAULT 'MEDIUM',
			compartment VARCHAR(50),
			reports_count INT DEFAULT 0,
			first_recruited_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_contact_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS humint_reports (
			id UUID PRIMARY KEY,
			source_code VARCHAR(255) NOT NULL REFERENCES humint_sources(code_name),
			classification humint_report_class NOT NULL,
			content_hash TEXT NOT NULL,
			threat_actors TEXT[] DEFAULT '{}',
			sectors_targeted TEXT[] DEFAULT '{}',
			veracity_score DECIMAL(3,2) DEFAULT 0,
			verified_by TEXT[] DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS humint_debriefings (
			id UUID PRIMARY KEY,
			source_code VARCHAR(255) NOT NULL REFERENCES humint_sources(code_name),
			officer_id UUID NOT NULL,
			session_date TIMESTAMPTZ NOT NULL,
			location_method humint_debrief_method NOT NULL,
			topics_covered TEXT[] DEFAULT '{}',
			next_meeting_planned_at TIMESTAMPTZ,
			risk_assessment TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_humint_sources_status ON humint_sources(status)`,
		`CREATE INDEX IF NOT EXISTS idx_humint_sources_risk ON humint_sources(risk_level)`,
		`CREATE INDEX IF NOT EXISTS idx_humint_reports_source ON humint_reports(source_code)`,
		`CREATE INDEX IF NOT EXISTS idx_humint_debrief_source ON humint_debriefings(source_code)`,
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
		getEnv("HUMINT_DB_HOST", "localhost"),
		getEnv("HUMINT_DB_PORT", "26257"),
		getEnv("HUMINT_DB_USER", "root"),
		getEnv("HUMINT_DB_PASSWORD", ""),
		getEnv("HUMINT_DB_NAME", "snisid_humint"),
		getEnv("HUMINT_DB_SSLMODE", "disable"),
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

	kafkaBrokers := getEnv("HUMINT_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("HUMINT_KAFKA_TOPIC", "humint-events")

	producer := humintKafka.NewProducer([]string{kafkaBrokers}, kafkaTopic)
	defer producer.Close()

	repo := repository.NewPostgresRepo(db)
	svc := service.NewHumintService(repo, producer)
	h := handler.NewHumintHandler(svc)

	ginMode := getEnv("HUMINT_GIN_MODE", "release")
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

	port := getEnv("HUMINT_PORT", "8302")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("humint-ht server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down humint-ht server...")
}
