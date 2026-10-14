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
	"github.com/snisid/infra-ht/internal/handler"
	"github.com/snisid/infra-ht/internal/kafka"
	"github.com/snisid/infra-ht/internal/repository"
	"github.com/snisid/infra-ht/internal/service"
)

func main() {
	dbURL := fmt.Sprintf("postgresql://root@%s:26257/%s?sslmode=disable", getEnv("INFRA_DB_HOST", "localhost"), getEnv("INFRA_DB_NAME", "snisid_infra"))
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	for _, m := range []string{
		`CREATE TABLE IF NOT EXISTS infra_datacenters (dc_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), dc_name VARCHAR(100) NOT NULL, dc_role VARCHAR(30) NOT NULL, dept_code CHAR(2) NOT NULL, tier_rating VARCHAR(10), power_capacity_kw DECIMAL(10,2), has_generator_backup BOOLEAN DEFAULT TRUE, has_redundant_internet BOOLEAN DEFAULT TRUE, rack_count INTEGER, is_active BOOLEAN DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS infra_k8s_clusters (cluster_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), dc_id UUID NOT NULL REFERENCES infra_datacenters(dc_id), cluster_name VARCHAR(100) NOT NULL, distro VARCHAR(30) DEFAULT 'RKE2', node_count INTEGER, kubernetes_version VARCHAR(20), is_production BOOLEAN DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS infra_dr_drills (drill_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), drill_date TIMESTAMPTZ NOT NULL, scenario TEXT NOT NULL, rto_target_minutes INTEGER DEFAULT 15, rto_actual_minutes INTEGER, rpo_target_minutes INTEGER DEFAULT 1, rpo_actual_minutes INTEGER, success BOOLEAN, notes TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
	} {
		if _, err := db.Exec(m); err != nil {
			log.Fatalf("migration: %v", err)
		}
	}

	p := kafka.NewProducer([]string{getEnv("INFRA_KAFKA_BROKERS", "localhost:9092")}, "snisid.infra.events")
	defer p.Close()

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	h := handler.NewHandler(service.NewInfraService(repository.NewPostgresRepo(db), p))
	h.RegisterRoutes(r.Group("/api/v1/infra"))

	srv := &http.Server{Addr: ":" + getEnv("INFRA_SERVICE_PORT", "8089"), Handler: r}
	go srv.ListenAndServe()
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
