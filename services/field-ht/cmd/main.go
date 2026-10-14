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
	"github.com/snisid/field-ht/internal/handler"
	"github.com/snisid/field-ht/internal/kafka"
	"github.com/snisid/field-ht/internal/repository"
	"github.com/snisid/field-ht/internal/service"
)

func main() {
	dbHost := getEnv("FIELD_DB_HOST", "localhost")
	dbPort := getEnv("FIELD_DB_PORT", "26257")
	dbName := getEnv("FIELD_DB_NAME", "snisid_field")
	dbUser := getEnv("FIELD_DB_USER", "root")
	dbSSLMode := getEnv("FIELD_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("FIELD_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("FIELD_KAFKA_TOPIC", "snisid.field.events")
	port := getEnv("FIELD_SERVICE_PORT", "8092")

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
	svc := service.NewFieldService(repo, producer)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	h := handler.NewHandler(svc)
	api := r.Group("/api/v1/field")
	h.RegisterRoutes(api)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("field-ht service started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run field-ht: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down field-ht...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TYPE IF NOT EXISTS field_mission_status AS ENUM ('PLANNED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')`,
		`CREATE TABLE IF NOT EXISTS field_mobile_units (
			id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			unit_code     VARCHAR(20) NOT NULL UNIQUE,
			team_members  TEXT[],
			equipment     JSONB,
			location      VARCHAR(150),
			is_active     BOOLEAN DEFAULT TRUE,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS field_missions (
			id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title            VARCHAR(200) NOT NULL,
			description      TEXT,
			status           field_mission_status NOT NULL DEFAULT 'PLANNED',
			assigned_unit_id UUID REFERENCES field_mobile_units(id),
			dept_code        CHAR(2) NOT NULL,
			started_at       TIMESTAMPTZ,
			completed_at     TIMESTAMPTZ,
			created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS field_mission_logs (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			mission_id UUID NOT NULL REFERENCES field_missions(id),
			logged_by  UUID NOT NULL,
			action     VARCHAR(100) NOT NULL,
			latitude   DECIMAL(10,7),
			longitude  DECIMAL(10,7),
			notes      TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_field_missions_status ON field_missions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_field_missions_dept ON field_missions(dept_code)`,
		`CREATE INDEX IF NOT EXISTS idx_field_mission_logs_mission ON field_mission_logs(mission_id, created_at DESC)`,
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
