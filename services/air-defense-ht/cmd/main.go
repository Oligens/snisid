package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/snisid/air-defense-ht/internal/handler"
	"github.com/snisid/air-defense-ht/internal/kafka"
	"github.com/snisid/air-defense-ht/internal/repository"
	"github.com/snisid/air-defense-ht/internal/service"
)

func main() {
	dbHost := getEnv("AIRDEF_DB_HOST", "localhost")
	dbPort := getEnv("AIRDEF_DB_PORT", "5432")
	dbUser := getEnv("AIRDEF_DB_USER", "postgres")
	dbPass := getEnv("AIRDEF_DB_PASSWORD", "postgres")
	dbName := getEnv("AIRDEF_DB_NAME", "snisid_airdefense")
	kafkaBrokers := []string{getEnv("AIRDEF_KAFKA_BROKERS", "localhost:9092")}
	httpPort := getEnv("AIRDEF_HTTP_PORT", "8303")

	dsn := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser + " password=" + dbPass + " dbname=" + dbName + " sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Printf("warning: database not reachable: %v", err)
	}

	initTables(db)

	producer := kafka.NewProducer(kafkaBrokers, "airdefense-events")
	defer producer.Close()

	repo := repository.NewPostgresRepo(db)
	svc := service.NewAirDefenseService(repo, producer)
	h := handler.NewAirDefenseHandler(svc)

	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api/v1/airdef")
	{
		v1.POST("/tracks", h.IngestTrack)
		v1.GET("/tracks/active", h.GetActiveTracks)
		v1.GET("/tracks/:track_id", h.GetTrackByID)
		v1.POST("/incidents", h.OpenIncident)
		v1.PATCH("/incidents/:id/resolve", h.ResolveIncident)
		v1.POST("/no-fly", h.AddNoFly)
		v1.GET("/no-fly/check", h.CheckNoFly)
	}

	srv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
	}

	go func() {
		log.Printf("air-defense-ht listening on :%s", httpPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
}

func initTables(db *sql.DB) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS airdef_radar_contacts (
			contact_id UUID PRIMARY KEY,
			track_number VARCHAR(20) NOT NULL,
			contact_type VARCHAR(20) NOT NULL DEFAULT 'UNKNOWN',
			latitude DECIMAL(10,7) NOT NULL,
			longitude DECIMAL(10,7) NOT NULL,
			altitude_m INT NOT NULL DEFAULT 0,
			speed_kmh DECIMAL(6,2) NOT NULL DEFAULT 0,
			heading_deg INT NOT NULL DEFAULT 0,
			source_radar VARCHAR(100) NOT NULL,
			identified BOOLEAN NOT NULL DEFAULT FALSE,
			squawk_code VARCHAR(4) DEFAULT '',
			flight_plan_ref VARCHAR(50) DEFAULT '',
			threat_assessment VARCHAR(20) NOT NULL DEFAULT 'UNKNOWN',
			operator_notes TEXT DEFAULT '',
			first_detected_at TIMESTAMPTZ NOT NULL,
			last_updated_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS airdef_incidents (
			incident_id UUID PRIMARY KEY,
			severity VARCHAR(10) NOT NULL DEFAULT 'INFO',
			status VARCHAR(20) NOT NULL DEFAULT 'DETECTED',
			aircraft_id UUID NOT NULL,
			interception_asset VARCHAR(100) DEFAULT '',
			pilot_response VARCHAR(20) NOT NULL DEFAULT 'COMPLIANT',
			engagement_rules_applied BOOLEAN NOT NULL DEFAULT FALSE,
			duration_minutes INT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS airdef_no_fly_list (
			entry_id UUID PRIMARY KEY,
			identity_ref VARCHAR(255) NOT NULL,
			full_name VARCHAR(255) NOT NULL,
			document_number VARCHAR(100) DEFAULT '',
			reason TEXT NOT NULL,
			added_by VARCHAR(255) NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			interpol_notice_ref VARCHAR(100) DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL
		)`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Printf("table init error: %v", err)
		}
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
