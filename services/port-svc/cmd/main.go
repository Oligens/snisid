package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/port-svc/internal/handler"
	"github.com/snisid/platform/services/port-svc/internal/kafka"
	"github.com/snisid/platform/services/port-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/port-svc/internal/service"
)

func main() {
	dbHost := getEnv("PORT_DB_HOST", "localhost")
	dbPort := getEnv("PORT_DB_PORT", "26257")
	dbName := getEnv("PORT_DB_NAME", "snisid_port")
	dbUser := getEnv("PORT_DB_USER", "root")
	dbSSLMode := getEnv("PORT_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("PORT_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("PORT_KAFKA_TOPIC", "snisid.port.events")
	port := getEnv("PORT_SERVICE_PORT", "8111")

	dbURL := fmt.Sprintf("postgresql://%s@%s:%s/%s?sslmode=%s", dbUser, dbHost, dbPort, dbName, dbSSLMode)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping: %v", err)
	}
	pool.Config().MaxConns = 25

	if err := runMigrations(ctx, pool); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	producer := kafka.NewProducer([]string{kafkaBrokers}, kafkaTopic, logger)
	defer producer.Close()

	repo := postgres.NewContainerRepo(pool)
	svc := service.NewPortService(repo, logger)

	r := handler.SetupRouter(svc, logger)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("port-svc started", zap.String("port", port))
		if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			logger.Fatal("error", zap.Error(e))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	logger.Info("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		"CREATE TYPE IF NOT EXISTS port_risk_level AS ENUM ('LOW','MEDIUM','HIGH','CRITICAL');",
		"CREATE TYPE IF NOT EXISTS port_container_status AS ENUM ( 'PENDING_INSPECTION','CLEARED','HELD_FOR_INSPECTION', 'SEIZED','RELEASED_AFTER_INSPECTION' );",
		"CREATE TABLE IF NOT EXISTS port_vessels_arrivals ( arrival_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(), port_code           VARCHAR(10) NOT NULL,  -- PAP, CAP, GON, CAY vessel_imo          VARCHAR(20), vessel_name         VARCHAR(150) NOT NULL, flag_country        CHAR(3), shipping_company    VARCHAR(200), arrival_date        TIMESTAMPTZ NOT NULL, origin_port         VARCHAR(100), origin_country      CHAR(3), container_count     INTEGER DEFAULT 0, manifest_ref        VARCHAR(100), mar_vessel_id       UUID,              -- Lien MAR-HT si vessel suspecte risk_score          SMALLINT DEFAULT 0, risk_level          port_risk_level DEFAULT 'LOW', cbp_targeting_ref   VARCHAR(50),       -- US Customs Pre-Targeting created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS port_containers ( container_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(), arrival_id          UUID NOT NULL REFERENCES port_vessels_arrivals(arrival_id), container_number    VARCHAR(20) NOT NULL, container_type      VARCHAR(10),       -- 20GP, 40HC, REEFER, etc. declared_content    TEXT NOT NULL, declared_weight_kg  DECIMAL(12,3), declared_value_usd  DECIMAL(15,2), shipper_name        VARCHAR(200), shipper_country     CHAR(3), consignee_name      VARCHAR(200), consignee_snisid_id UUID, status              port_container_status NOT NULL DEFAULT 'PENDING_INSPECTION', risk_score          SMALLINT DEFAULT 0, risk_level          port_risk_level DEFAULT 'LOW', risk_flags          TEXT[] DEFAULT '{}', selected_for_scan   BOOLEAN DEFAULT FALSE, scan_date           TIMESTAMPTZ, scan_result         TEXT, seized              BOOLEAN DEFAULT FALSE, seizure_description TEXT, case_reference      VARCHAR(100), cbp_targeting_match BOOLEAN DEFAULT FALSE, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS port_risk_factors ( factor_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(), container_id        UUID NOT NULL REFERENCES port_containers(container_id), factor_type         VARCHAR(50) NOT NULL, description         TEXT NOT NULL, weight_score        SMALLINT NOT NULL,  -- Points ajoutes au risk_score source              VARCHAR(50),        -- BLKL, BLAN, GANG, CBP, INTEL created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"BEGIN IF p_shipper_country IN ('COL','MEX','VEN','ECU') THEN score := score + 30; END IF;",
		"IF p_consignee_id IS NOT NULL AND EXISTS ( SELECT 1 FROM blkl_blacklist WHERE snisid_person_id = p_consignee_id AND is_active = TRUE ) THEN score := score + 50; END IF;",
		"IF lower(p_declared_content) ~ 'general cargo|mixed goods|used items' THEN score := score + 15; END IF;",
		"RETURN LEAST(score, 100);",
		"END;",
		"$$ LANGUAGE plpgsql;",
		"CREATE INDEX IF NOT EXISTS idx_port_containers_risk   ON port_containers(risk_level, status);",
		"CREATE INDEX IF NOT EXISTS idx_port_containers_arrival ON port_containers(arrival_id);",
		"CREATE INDEX IF NOT EXISTS idx_port_arrivals_date     ON port_vessels_arrivals(arrival_date DESC);",
		"CREATE INDEX IF NOT EXISTS idx_port_arrivals_port     ON port_vessels_arrivals(port_code, arrival_date DESC);",
	}
	for _, m := range migrations {
		if _, err := pool.Exec(ctx, m); err != nil {
			return fmt.Errorf("migration: %s: %w", m[:60], err)
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
