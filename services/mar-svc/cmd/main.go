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

	"github.com/snisid/platform/services/mar-svc/internal/handler"
	"github.com/snisid/platform/services/mar-svc/internal/kafka"
	"github.com/snisid/platform/services/mar-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/mar-svc/internal/service"
)

func main() {
	dbHost := getEnv("MAR_DB_HOST", "localhost")
	dbPort := getEnv("MAR_DB_PORT", "26257")
	dbName := getEnv("MAR_DB_NAME", "snisid_mar")
	dbUser := getEnv("MAR_DB_USER", "root")
	dbSSLMode := getEnv("MAR_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("MAR_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("MAR_KAFKA_TOPIC", "snisid.mar.events")
	port := getEnv("MAR_SERVICE_PORT", "8107")

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

	repo := postgres.NewMaritimeRepo(pool, logger)
	svc := service.NewMaritimeService(repo, logger)

	r := handler.SetupRouter(svc, logger)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("mar-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS mar_vessel_type AS ENUM ( 'CARGO_SHIP','TANKER','FISHING_BOAT','GO_FAST', 'SAILBOAT','YACHT','FERRY','PATROL_BOAT', 'WOODEN_BOAT','CANOE','UNKNOWN' );",
		"CREATE TYPE IF NOT EXISTS mar_vessel_status AS ENUM ( 'REGISTERED','STOLEN','SUSPECTED','DETAINED', 'SUNK','DESTROYED','MISSING','INTERPOL_ALERT' );",
		"CREATE TYPE IF NOT EXISTS mar_incident_type AS ENUM ( 'DRUG_SEIZURE','ARMS_SEIZURE','MIGRANT_INTERDICTION', 'SMUGGLING','SUSPICIOUS_ACTIVITY','DISTRESS', 'PIRACY','ILLEGAL_FISHING','HUMAN_TRAFFICKING' );",
		"CREATE TABLE IF NOT EXISTS mar_vessels ( vessel_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(), national_mar_id     VARCHAR(25) UNIQUE NOT NULL,   -- MAR-HT-NNNNNN vessel_name         VARCHAR(150), imo_number          VARCHAR(20),                   -- IMO unique ID mmsi                VARCHAR(15),                   -- AIS transponder ID call_sign           VARCHAR(15), vessel_type         mar_vessel_type NOT NULL, flag_country        CHAR(3), hull_color          VARCHAR(50), length_m            DECIMAL(8,2), tonnage_gt          INTEGER, engine_count        SMALLINT, horsepower          INTEGER,                       -- Critique pour go-fasts owner_name          VARCHAR(200), owner_snisid_id     UUID, registration_number VARCHAR(50), registration_port   VARCHAR(100), status              mar_vessel_status NOT NULL DEFAULT 'REGISTERED', gang_id             UUID, interpol_svd_ref    VARCHAR(50), notes               TEXT, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS mar_ais_sightings ( sighting_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), vessel_id           UUID REFERENCES mar_vessels(vessel_id), mmsi                VARCHAR(15), vessel_name         VARCHAR(150), sighting_timestamp  TIMESTAMPTZ NOT NULL, lat                 DECIMAL(10,7) NOT NULL, lng                 DECIMAL(10,7) NOT NULL, speed_knots         DECIMAL(5,2), heading_degrees     SMALLINT, destination         VARCHAR(100), source_type         VARCHAR(30),     -- AIS_TERRESTRIAL, AIS_SATELLITE, RADAR, VISUAL zone_code           VARCHAR(20),     -- WINDWARD_PASS, TORTUE, GONAVE, etc. alert_triggered     BOOLEAN DEFAULT FALSE, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS mar_incidents ( incident_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), vessel_id           UUID REFERENCES mar_vessels(vessel_id), incident_type       mar_incident_type NOT NULL, incident_date       TIMESTAMPTZ NOT NULL, lat                 DECIMAL(10,7), lng                 DECIMAL(10,7), zone_desc           VARCHAR(100), responding_unit     VARCHAR(50),     -- GCH, USCG, JIATF-South outcome             TEXT, persons_involved    INTEGER DEFAULT 0, snisid_person_ids   UUID[] DEFAULT '{}', drug_types          TEXT[] DEFAULT '{}', drug_weight_kg      DECIMAL(12,3), weapons_found       BOOLEAN DEFAULT FALSE, weapons_count       INTEGER DEFAULT 0, migrants_count      INTEGER DEFAULT 0, biar_refs           UUID[] DEFAULT '{}', case_reference      VARCHAR(100), photo_refs          TEXT[] DEFAULT '{}', created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS mar_watch_vessels ( watch_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(), vessel_id           UUID REFERENCES mar_vessels(vessel_id), mmsi                VARCHAR(15), vessel_name         VARCHAR(150), watch_reason        TEXT NOT NULL, alert_level         VARCHAR(20) DEFAULT 'CAUTION', requesting_unit     VARCHAR(50), is_active           BOOLEAN DEFAULT TRUE, expiry_date         TIMESTAMPTZ, created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_mar_ais_timestamp   ON mar_ais_sightings(sighting_timestamp DESC);",
		"CREATE INDEX IF NOT EXISTS idx_mar_ais_mmsi        ON mar_ais_sightings(mmsi);",
		"CREATE INDEX IF NOT EXISTS idx_mar_ais_coords      ON mar_ais_sightings(lat, lng);",
		"CREATE INDEX IF NOT EXISTS idx_mar_incidents_date  ON mar_incidents(incident_date DESC);",
		"CREATE INDEX IF NOT EXISTS idx_mar_incidents_type  ON mar_incidents(incident_type);",
		"CREATE INDEX IF NOT EXISTS idx_mar_watch_active    ON mar_watch_vessels(is_active) WHERE is_active = TRUE;",
		"CREATE INDEX IF NOT EXISTS idx_mar_vessels_status  ON mar_vessels(status);",
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
