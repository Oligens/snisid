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

	"github.com/snisid/mil-c2-ht/internal/handler"
	"github.com/snisid/mil-c2-ht/internal/kafka"
	"github.com/snisid/mil-c2-ht/internal/repository"
	"github.com/snisid/mil-c2-ht/internal/service"
)

func main() {
	dbHost := getEnv("MILC2_DB_HOST", "localhost")
	dbPort := getEnv("MILC2_DB_PORT", "5432")
	dbUser := getEnv("MILC2_DB_USER", "postgres")
	dbPass := getEnv("MILC2_DB_PASSWORD", "postgres")
	dbName := getEnv("MILC2_DB_NAME", "snisid_milc2")
	kafkaBrokers := []string{getEnv("MILC2_KAFKA_BROKERS", "localhost:9092")}
	httpPort := getEnv("MILC2_HTTP_PORT", "8304")

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

	producer := kafka.NewProducer(kafkaBrokers, "milc2-events")
	defer producer.Close()

	repo := repository.NewPostgresRepo(db)
	svc := service.NewMilC2Service(repo, producer)
	h := handler.NewMilC2Handler(svc)

	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api/v1/milc2")
	{
		v1.POST("/units", h.CreateUnit)
		v1.GET("/units/deployed", h.GetDeployedUnits)
		v1.POST("/operations", h.CreateOperation)
		v1.GET("/operations/active", h.GetActiveOperations)
		v1.POST("/operations/:id/reports", h.SubmitReport)
		v1.GET("/operations/:id/timeline", h.GetOperationTimeline)
		v1.GET("/common-operating-picture", h.GetCommonOperatingPicture)
	}

	srv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
	}

	go func() {
		log.Printf("mil-c2-ht listening on :%s", httpPort)
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
		`CREATE TABLE IF NOT EXISTS milc2_units (
			unit_id UUID PRIMARY KEY,
			unit_name VARCHAR(255) NOT NULL,
			branch VARCHAR(20) NOT NULL,
			parent_unit_id UUID,
			commander_name VARCHAR(255) DEFAULT '',
			personnel_count INT NOT NULL DEFAULT 0,
			location_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
			location_lng DECIMAL(10,7) NOT NULL DEFAULT 0,
			operational_status VARCHAR(20) NOT NULL DEFAULT 'STANDBY',
			equipment_summary TEXT DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS milc2_operations (
			operation_id UUID PRIMARY KEY,
			operation_name VARCHAR(255) NOT NULL,
			operation_type VARCHAR(20) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'PLANNING',
			commander_id UUID NOT NULL,
			start_date TIMESTAMPTZ NOT NULL,
			expected_end_date TIMESTAMPTZ,
			operational_area TEXT DEFAULT '',
			rules_of_engagement TEXT DEFAULT '',
			mission_objective TEXT DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS milc2_tactical_reports (
			report_id UUID PRIMARY KEY,
			operation_id UUID NOT NULL,
			reporting_unit_id UUID NOT NULL,
			report_type VARCHAR(10) NOT NULL,
			position_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
			position_lng DECIMAL(10,7) NOT NULL DEFAULT 0,
			enemy_activity TEXT DEFAULT '',
			civilian_interactions TEXT DEFAULT '',
			casualties INT NOT NULL DEFAULT 0,
			detainees INT NOT NULL DEFAULT 0,
			equipment_status TEXT DEFAULT '',
			submitted_at TIMESTAMPTZ NOT NULL
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
