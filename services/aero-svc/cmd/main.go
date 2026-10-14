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

	"github.com/snisid/platform/services/aero-svc/internal/handler"
	"github.com/snisid/platform/services/aero-svc/internal/kafka"
	"github.com/snisid/platform/services/aero-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/aero-svc/internal/service"
)

func main() {
	dbHost := getEnv("AERO_DB_HOST", "localhost")
	dbPort := getEnv("AERO_DB_PORT", "26257")
	dbName := getEnv("AERO_DB_NAME", "snisid_aero")
	dbUser := getEnv("AERO_DB_USER", "root")
	dbSSLMode := getEnv("AERO_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("AERO_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("AERO_KAFKA_TOPIC", "snisid.aero.events")
	port := getEnv("AERO_SERVICE_PORT", "8109")

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

	repo := postgres.NewAircraftRepo(pool)
	svc := service.NewAeroService(repo, logger)

	r := handler.SetupRouter(svc, logger)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("aero-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS aero_aircraft_type AS ENUM ( 'COMMERCIAL_JET','TURBOPROP','PISTON_SINGLE','PISTON_TWIN', 'HELICOPTER','ULTRALIGHT','DRONE_LARGE','UNKNOWN' );",
		"CREATE TYPE IF NOT EXISTS aero_strip_status AS ENUM ( 'ACTIVE','INACTIVE','DESTROYED','LEGALIZED','UNDER_SURVEILLANCE' );",
		"CREATE TABLE IF NOT EXISTS aero_aircraft_registry ( aircraft_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), registration_mark   VARCHAR(20), icao_hex_code       VARCHAR(10), aircraft_type       aero_aircraft_type NOT NULL, make                VARCHAR(100), model               VARCHAR(100), manufacture_year    SMALLINT, flag_country        CHAR(3), owner_name          VARCHAR(200), owner_snisid_id     UUID, operator_name       VARCHAR(200), is_registered       BOOLEAN DEFAULT FALSE, is_suspected        BOOLEAN DEFAULT FALSE, is_stolen           BOOLEAN DEFAULT FALSE, gang_id             UUID, drug_trafficking    BOOLEAN DEFAULT FALSE, interpol_ref        VARCHAR(50), faa_registry_ref    VARCHAR(50), notes               TEXT, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS aero_clandestine_strips ( strip_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(), strip_name          VARCHAR(150), dept_code           CHAR(2) NOT NULL, commune             VARCHAR(100), lat                 DECIMAL(10,7) NOT NULL, lng                 DECIMAL(10,7) NOT NULL, length_m            INTEGER, surface_type        VARCHAR(30),   -- GRASS, DIRT, ASPHALT, GRAVEL status              aero_strip_status NOT NULL DEFAULT 'ACTIVE', capable_aircraft    TEXT[] DEFAULT '{}', gang_id             UUID, first_detected      DATE, last_activity_date  DATE, source_intel        TEXT, satellite_image_ref VARCHAR(500), created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS aero_suspicious_flights ( flight_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(), aircraft_id         UUID REFERENCES aero_aircraft_registry(aircraft_id), registration_mark   VARCHAR(20), flight_date         TIMESTAMPTZ NOT NULL, origin_airport      VARCHAR(10), destination_airport VARCHAR(10), origin_country      CHAR(3), destination_country CHAR(3) DEFAULT 'HTI', landing_strip_id    UUID REFERENCES aero_clandestine_strips(strip_id), landing_location    VARCHAR(300), flight_type         VARCHAR(30),   -- DRUG_RUN, ARMS_DELIVERY, UNKNOWN cargo_suspected     TEXT, source_radar        VARCHAR(50), source_informant    BOOLEAN DEFAULT FALSE, case_reference      VARCHAR(100), created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_aero_registry_mark    ON aero_aircraft_registry(registration_mark);",
		"CREATE INDEX IF NOT EXISTS idx_aero_registry_gang    ON aero_aircraft_registry(gang_id) WHERE gang_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_aero_strips_dept      ON aero_clandestine_strips(dept_code) WHERE status = 'ACTIVE';",
		"CREATE INDEX IF NOT EXISTS idx_aero_strips_coords    ON aero_clandestine_strips(lat, lng);",
		"CREATE INDEX IF NOT EXISTS idx_aero_flights_date     ON aero_suspicious_flights(flight_date DESC);",
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
