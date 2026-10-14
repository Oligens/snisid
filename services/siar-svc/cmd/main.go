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

	"github.com/snisid/platform/services/siar-svc/internal/handler"
	"github.com/snisid/platform/services/siar-svc/internal/kafka"
	"github.com/snisid/platform/services/siar-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/siar-svc/internal/service"
)

func main() {
	dbHost := getEnv("SIAR_DB_HOST", "localhost")
	dbPort := getEnv("SIAR_DB_PORT", "26257")
	dbName := getEnv("SIAR_DB_NAME", "snisid_siar")
	dbUser := getEnv("SIAR_DB_USER", "root")
	dbSSLMode := getEnv("SIAR_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("SIAR_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("SIAR_KAFKA_TOPIC", "snisid.siar.events")
	port := getEnv("SIAR_SERVICE_PORT", "8102")

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

	repo := postgres.NewFirearmRepo(pool)
	svc := service.NewSIARService(repo, logger)

	r := handler.SetupRouter(svc, logger)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("siar-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS siar_weapon_type AS ENUM ( 'HANDGUN','RIFLE','SHOTGUN','SUBMACHINE_GUN','ASSAULT_RIFLE', 'MACHINE_GUN','SNIPER','RPG','GRENADE','HOMEMADE','OTHER' );",
		"CREATE TYPE IF NOT EXISTS siar_status AS ENUM ( 'REGISTERED','REPORTED_STOLEN','SEIZED','DESTROYED', 'REPORTED_LOST','TRANSFERRED','DEACTIVATED' );",
		"CREATE TYPE IF NOT EXISTS siar_registration_type AS ENUM ( 'CIVILIAN','POLICE','MILITARY','SECURITY_COMPANY', 'EMBASSY','ILLEGAL_FOUND','HISTORICAL' );",
		"CREATE TABLE IF NOT EXISTS siar_firearms ( firearm_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(), national_siar_id    VARCHAR(25) UNIQUE NOT NULL,  -- SIAR-HT-NNNNNN serial_number       VARCHAR(100), make                VARCHAR(100) NOT NULL, model               VARCHAR(100) NOT NULL, caliber             VARCHAR(30) NOT NULL, weapon_type         siar_weapon_type NOT NULL, manufacture_year    SMALLINT, manufacture_country CHAR(3), status              siar_status NOT NULL DEFAULT 'REGISTERED', reg_type            siar_registration_type NOT NULL, owner_snisid_id     UUID, owner_entity_name   VARCHAR(200),     -- Si organisation license_number      VARCHAR(50), license_expiry      DATE, import_date         DATE, import_country      CHAR(3), import_permit_ref   VARCHAR(100), importer_name       VARCHAR(200), customs_entry_ref   VARCHAR(100), current_dept_code   CHAR(2), storage_location    TEXT, fir_record_id       UUID, gang_id             UUID, case_references     TEXT[] DEFAULT '{}', iarms_ref           VARCHAR(50), atf_etrace_ref      VARCHAR(50), notes               TEXT, created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS siar_licenses ( license_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(), license_number      VARCHAR(50) UNIQUE NOT NULL, holder_snisid_id    UUID NOT NULL, license_type        VARCHAR(50) NOT NULL,   -- CARRY, POSSESS, DEALER, COLLECTOR firearms_authorized INTEGER DEFAULT 1, issue_date          DATE NOT NULL, expiry_date         DATE NOT NULL, issuing_authority   VARCHAR(100) NOT NULL, is_active           BOOLEAN DEFAULT TRUE, revocation_reason   TEXT, revoked_at          TIMESTAMPTZ, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS siar_transfers ( transfer_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), firearm_id          UUID NOT NULL REFERENCES siar_firearms(firearm_id), from_owner_id       UUID, to_owner_id         UUID, transfer_type       VARCHAR(50),  -- SALE, GIFT, INHERITANCE, CONFISCATION transfer_date       DATE NOT NULL, permit_ref          VARCHAR(100), authorized_by       UUID, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS siar_seizures ( seizure_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(), firearm_id          UUID REFERENCES siar_firearms(firearm_id), seizure_date        TIMESTAMPTZ NOT NULL, seizing_unit        VARCHAR(50) NOT NULL, seizing_officer     UUID, location_desc       VARCHAR(300), dept_code           CHAR(2), context             TEXT,          -- Circonstances de saisie from_person_id      UUID,          -- Personne chez qui saisie gang_id             UUID, case_reference      VARCHAR(100), disposed_of         BOOLEAN DEFAULT FALSE, disposal_method     VARCHAR(50),   -- DESTROYED, KEPT_AS_EVIDENCE, RETURNED created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_siar_serial    ON siar_firearms(serial_number) WHERE serial_number IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_siar_status    ON siar_firearms(status);",
		"CREATE INDEX IF NOT EXISTS idx_siar_gang      ON siar_firearms(gang_id) WHERE gang_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_siar_owner     ON siar_firearms(owner_snisid_id) WHERE owner_snisid_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_siar_iarms     ON siar_firearms(iarms_ref) WHERE iarms_ref IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_siar_licenses  ON siar_licenses(holder_snisid_id, is_active);",
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
