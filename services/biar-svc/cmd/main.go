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

	"github.com/snisid/platform/services/biar-svc/internal/handler"
	"github.com/snisid/platform/services/biar-svc/internal/kafka"
	"github.com/snisid/platform/services/biar-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/biar-svc/internal/service"
)

func main() {
	dbHost := getEnv("BIAR_DB_HOST", "localhost")
	dbPort := getEnv("BIAR_DB_PORT", "26257")
	dbName := getEnv("BIAR_DB_NAME", "snisid_biar")
	dbUser := getEnv("BIAR_DB_USER", "root")
	dbSSLMode := getEnv("BIAR_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("BIAR_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("BIAR_KAFKA_TOPIC", "snisid.biar.events")
	port := getEnv("BIAR_SERVICE_PORT", "8103")

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

	repo := postgres.NewWeaponRepo(pool)
	svc := service.NewBIARService(repo, logger)

	r := handler.SetupRouter(svc, logger)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("biar-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS biar_recovery_context AS ENUM ( 'POLICE_OPERATION','CHECKPOINT','PORT_SEIZURE','AIRPORT_SEIZURE', 'COMMUNITY_SURRENDER','CRIME_SCENE','RAID','BORDER_SEIZURE','OTHER' );",
		"CREATE TYPE IF NOT EXISTS biar_weapon_disposition AS ENUM ( 'HELD_AS_EVIDENCE','DESTROYED','RETURNED_TO_OWNER', 'TRANSFERRED_TO_POLICE','SENT_TO_INTERPOL','PENDING' );",
		"CREATE TABLE IF NOT EXISTS biar_illicit_weapons ( weapon_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(), national_biar_id    VARCHAR(25) UNIQUE NOT NULL,     -- BIAR-HT-NNNNNN serial_number       VARCHAR(100), serial_obliterated  BOOLEAN DEFAULT FALSE,           -- Numero efface (crime) make                VARCHAR(100), model               VARCHAR(100), caliber             VARCHAR(30), weapon_type         VARCHAR(50) NOT NULL, manufacture_country CHAR(3), estimated_manufacture_year SMALLINT, recovery_date       TIMESTAMPTZ NOT NULL, recovery_context    biar_recovery_context NOT NULL, recovery_location   VARCHAR(300), recovery_dept_code  CHAR(2), recovery_commune    VARCHAR(100), recovery_lat        DECIMAL(10,7), recovery_lng        DECIMAL(10,7), seizing_unit        VARCHAR(50) NOT NULL, seizing_officer     UUID, case_reference      VARCHAR(100), from_person_id      UUID,                            -- Personne chez qui saisie gang_id             UUID, crime_category      VARCHAR(50), associated_cases    TEXT[] DEFAULT '{}', origin_country      CHAR(3), transit_countries   CHAR(3)[] DEFAULT '{}', trafficking_route   TEXT, import_method       TEXT,                            -- Conteneur, bagages, go-fast... iarms_ref           VARCHAR(50), atf_etrace_ref      VARCHAR(50), reported_to_interpol BOOLEAN DEFAULT FALSE, interpol_reported_at TIMESTAMPTZ, disposition         biar_weapon_disposition DEFAULT 'HELD_AS_EVIDENCE', disposal_date       TIMESTAMPTZ, disposal_auth       UUID, quantity_ammunition INTEGER DEFAULT 0, ammunition_type     VARCHAR(50), photos_refs         TEXT[] DEFAULT '{}', notes               TEXT, created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS biar_batch_seizures ( batch_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(), batch_reference     VARCHAR(50) UNIQUE NOT NULL, operation_name      TEXT, seizure_date        TIMESTAMPTZ NOT NULL, location_desc       VARCHAR(300), dept_code           CHAR(2), total_weapons       INTEGER NOT NULL, weapon_ids          UUID[] DEFAULT '{}', seizing_unit        VARCHAR(50) NOT NULL, lead_officer        UUID, partnering_agencies TEXT[] DEFAULT '{}', notes               TEXT, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS biar_iarms_sync_log ( sync_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(), weapon_id           UUID REFERENCES biar_illicit_weapons(weapon_id), direction           VARCHAR(10) NOT NULL,     -- OUTBOUND / INBOUND iarms_ref           VARCHAR(50), sync_status         VARCHAR(20) DEFAULT 'PENDING', synced_at           TIMESTAMPTZ, error_message       TEXT, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_biar_serial    ON biar_illicit_weapons(serial_number) WHERE serial_number IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_biar_gang      ON biar_illicit_weapons(gang_id) WHERE gang_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_biar_dept      ON biar_illicit_weapons(recovery_dept_code);",
		"CREATE INDEX IF NOT EXISTS idx_biar_date      ON biar_illicit_weapons(recovery_date DESC);",
		"CREATE INDEX IF NOT EXISTS idx_biar_iarms     ON biar_illicit_weapons(iarms_ref) WHERE iarms_ref IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_biar_origin    ON biar_illicit_weapons(origin_country);",
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
