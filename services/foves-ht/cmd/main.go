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
	"github.com/snisid/foves-ht/internal/handler"
	"github.com/snisid/foves-ht/internal/kafka"
	"github.com/snisid/foves-ht/internal/repository"
	"github.com/snisid/foves-ht/internal/service"
)

func main() {
	dbHost := getEnv("FOVES_DB_HOST", "localhost")
	dbPort := getEnv("FOVES_DB_PORT", "26257")
	dbName := getEnv("FOVES_DB_NAME", "snisid_foves")
	dbUser := getEnv("FOVES_DB_USER", "root")
	dbSSLMode := getEnv("FOVES_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("FOVES_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("FOVES_KAFKA_TOPIC", "snisid.foves.events")
	port := getEnv("FOVES_SERVICE_PORT", "8095")

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
	svc := service.NewFovesService(repo, producer)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	h := handler.NewHandler(svc)
	api := r.Group("/api/v1/foves")
	h.RegisterRoutes(api)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("foves-ht service started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run foves-ht: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down foves-ht...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TYPE IF NOT EXISTS foves_vehicle_category AS ENUM (
			'PRIVATE_CAR', 'MOTORCYCLE', 'TAP_TAP', 'BUS', 'TRUCK',
			'COMMERCIAL', 'GOVERNMENT', 'DIPLOMATIC', 'AGRICULTURAL'
		)`,
		`CREATE TABLE IF NOT EXISTS foves_vehicles (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			plate_number      VARCHAR(20) UNIQUE NOT NULL,
			vin               VARCHAR(50) UNIQUE NOT NULL,
			make              VARCHAR(100) NOT NULL,
			model             VARCHAR(100) NOT NULL,
			year              INTEGER NOT NULL,
			color             VARCHAR(30),
			category          foves_vehicle_category NOT NULL,
			owner_citizen_id  UUID NOT NULL,
			is_stolen         BOOLEAN DEFAULT FALSE,
			is_active         BOOLEAN DEFAULT TRUE,
			registered_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS foves_ownership_transfers (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			vehicle_id        UUID NOT NULL REFERENCES foves_vehicles(id),
			from_citizen_id   UUID NOT NULL,
			to_citizen_id     UUID NOT NULL,
			transfer_date     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			contract_ref      VARCHAR(100),
			approved_by       VARCHAR(150),
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS foves_driver_licenses (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			citizen_id        UUID UNIQUE NOT NULL,
			license_number    VARCHAR(30) UNIQUE NOT NULL,
			category_a        BOOLEAN DEFAULT FALSE,
			category_b        BOOLEAN DEFAULT FALSE,
			category_c        BOOLEAN DEFAULT FALSE,
			category_d        BOOLEAN DEFAULT FALSE,
			category_e        BOOLEAN DEFAULT FALSE,
			category_f        BOOLEAN DEFAULT FALSE,
			issued_date       DATE NOT NULL DEFAULT CURRENT_DATE,
			expiry_date       DATE NOT NULL,
			points_balance    SMALLINT NOT NULL DEFAULT 12,
			is_suspended      BOOLEAN DEFAULT FALSE,
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_foves_vehicles_owner ON foves_vehicles(owner_citizen_id)`,
		`CREATE INDEX IF NOT EXISTS idx_foves_vehicles_plate ON foves_vehicles(plate_number)`,
		`CREATE INDEX IF NOT EXISTS idx_foves_vehicles_vin ON foves_vehicles(vin)`,
		`CREATE INDEX IF NOT EXISTS idx_foves_transfers_vehicle ON foves_ownership_transfers(vehicle_id)`,
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
