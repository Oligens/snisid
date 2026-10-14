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

	"github.com/snisid/platform/services/trafar-svc/internal/handler"
	"github.com/snisid/platform/services/trafar-svc/internal/kafka"
	"github.com/snisid/platform/services/trafar-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/trafar-svc/internal/service"
)

func main() {
	dbHost := getEnv("TRAFAR_DB_HOST", "localhost")
	dbPort := getEnv("TRAFAR_DB_PORT", "26257")
	dbName := getEnv("TRAFAR_DB_NAME", "snisid_trafar")
	dbUser := getEnv("TRAFAR_DB_USER", "root")
	dbSSLMode := getEnv("TRAFAR_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("TRAFAR_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("TRAFAR_KAFKA_TOPIC", "snisid.trafar.events")
	port := getEnv("TRAFAR_SERVICE_PORT", "8105")

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

	repo := postgres.NewRouteRepository(pool)
	svc := service.NewTrafarService(repo, logger)

	r := handler.SetupRouter(svc, logger)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("trafar-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS trafar_route_type AS ENUM ( 'MARITIME_DIRECT','MARITIME_VIA_BAHAMAS', 'AIR_CARGO','AIR_PASSENGER', 'LAND_BORDER_DOM','LAND_BORDER_OTHER', 'POSTAL','MIXED' );",
		"CREATE TYPE IF NOT EXISTS trafar_method AS ENUM ( 'STRAW_PURCHASE','STOLEN_DIVERTED','CORRUPT_OFFICIAL', 'FALSE_END_USER','DARK_WEB','DIPLOMATIC_POUCH', 'CONCEALED_CARGO','DRUGS_FOR_GUNS_SWAP','UNKNOWN' );",
		"CREATE TABLE IF NOT EXISTS trafar_routes ( route_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(), route_name          VARCHAR(150) NOT NULL, route_type          trafar_route_type NOT NULL, trafficking_method  trafar_method NOT NULL, origin_country      CHAR(3) NOT NULL, origin_city         VARCHAR(100), transit_points      JSONB,       -- [{country, city, transport_mode}] entry_point_haiti   VARCHAR(100), entry_dept_code     CHAR(2), associated_gang_ids UUID[] DEFAULT '{}', known_suppliers     TEXT[] DEFAULT '{}', activity_level      VARCHAR(20) DEFAULT 'ACTIVE', estimated_volume_monthly INTEGER, weapon_types        TEXT[] DEFAULT '{}', intel_confidence    SMALLINT CHECK (intel_confidence BETWEEN 1 AND 10), first_detected      DATE, last_confirmed      DATE, linked_case_refs    TEXT[] DEFAULT '{}', biar_weapon_ids     UUID[] DEFAULT '{}', atf_case_refs       TEXT[] DEFAULT '{}', unodc_ref           VARCHAR(50), analyst_notes       TEXT, created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS trafar_shipments ( shipment_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), route_id            UUID REFERENCES trafar_routes(route_id), shipment_date       TIMESTAMPTZ NOT NULL, intercepted         BOOLEAN DEFAULT FALSE, interception_date   TIMESTAMPTZ, interception_location VARCHAR(300), interception_unit   VARCHAR(50), weapons_count       INTEGER, weapons_types       TEXT[] DEFAULT '{}', estimated_value_usd DECIMAL(12,2), linked_persons      UUID[] DEFAULT '{}', port_ht_ref         UUID,       -- Lien PORT-HT si interception portuaire mar_ht_ref          UUID,       -- Lien MAR-HT si interception maritime case_reference      VARCHAR(100), notes               TEXT, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS trafar_suppliers ( supplier_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), supplier_name       VARCHAR(200), supplier_type       VARCHAR(50),   -- DEALER, CARTEL, CORRUPT_OFFICIAL, UNKNOWN country             CHAR(3) NOT NULL, city                VARCHAR(100), snisid_person_id    UUID,          -- Si identifie dans SNISID linked_routes       UUID[] DEFAULT '{}', atf_subject_ref     VARCHAR(50), interpol_notice_ref VARCHAR(50), is_active           BOOLEAN DEFAULT TRUE, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_trafar_routes_type     ON trafar_routes(route_type, activity_level);",
		"CREATE INDEX IF NOT EXISTS idx_trafar_routes_origin   ON trafar_routes(origin_country);",
		"CREATE INDEX IF NOT EXISTS idx_trafar_routes_entry    ON trafar_routes(entry_dept_code);",
		"CREATE INDEX IF NOT EXISTS idx_trafar_shipments_route ON trafar_shipments(route_id);",
		"CREATE INDEX IF NOT EXISTS idx_trafar_shipments_date  ON trafar_shipments(shipment_date DESC);",
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
