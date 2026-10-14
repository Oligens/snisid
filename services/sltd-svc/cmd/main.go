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

	"github.com/snisid/platform/services/sltd-svc/internal/handler"
	"github.com/snisid/platform/services/sltd-svc/internal/kafka"
	"github.com/snisid/platform/services/sltd-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/sltd-svc/internal/service"
)

func main() {
	dbHost := getEnv("SLTD_DB_HOST", "localhost")
	dbPort := getEnv("SLTD_DB_PORT", "26257")
	dbName := getEnv("SLTD_DB_NAME", "snisid_sltd")
	dbUser := getEnv("SLTD_DB_USER", "root")
	dbSSLMode := getEnv("SLTD_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("SLTD_KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("SLTD_KAFKA_TOPIC", "snisid.sltd.events")
	port := getEnv("SLTD_SERVICE_PORT", "8108")

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

	repo := postgres.NewDocumentRepo(pool)
	svc := service.NewSLTDService(repo, logger)

	r := handler.SetupRouter(svc, logger)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("sltd-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS sltd_doc_type AS ENUM ( 'PASSPORT','NATIONAL_ID','TRAVEL_DOCUMENT', 'VISA','RESIDENCE_PERMIT','REFUGEE_DOCUMENT','LAISSEZ_PASSER' );",
		"CREATE TYPE IF NOT EXISTS sltd_doc_status AS ENUM ( 'LOST','STOLEN','REVOKED','EXPIRED','FOUND','RECOVERED','CANCELLED' );",
		"CREATE TABLE IF NOT EXISTS sltd_documents ( doc_id              UUID PRIMARY KEY DEFAULT gen_random_uuid(), national_sltd_id    VARCHAR(25) UNIQUE NOT NULL,  -- SLTD-HT-NNNNNN doc_type            sltd_doc_type NOT NULL, document_number     VARCHAR(100) NOT NULL, issuing_country     CHAR(3) NOT NULL DEFAULT 'HTI', holder_name         VARCHAR(200), holder_snisid_id    UUID, holder_dob          DATE, holder_nationality  CHAR(3) DEFAULT 'HTI', issue_date          DATE, expiry_date         DATE, status              sltd_doc_status NOT NULL, reported_date       TIMESTAMPTZ NOT NULL DEFAULT NOW(), reported_by         UUID NOT NULL, reporting_dept_code CHAR(2), theft_context       TEXT, found_date          TIMESTAMPTZ, found_location      VARCHAR(300), interpol_sltd_ref   VARCHAR(50), reported_to_interpol BOOLEAN DEFAULT FALSE, interpol_reported_at TIMESTAMPTZ, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS sltd_check_log ( check_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(), document_number     VARCHAR(100) NOT NULL, doc_type            sltd_doc_type, checked_by          UUID NOT NULL, check_location      VARCHAR(100), post_id             UUID, result              VARCHAR(20) NOT NULL,    -- CLEAR, LOST, STOLEN, REVOKED source              VARCHAR(20) NOT NULL,    -- LOCAL, INTERPOL_SLTD, BOTH sltd_doc_id         UUID REFERENCES sltd_documents(doc_id), checked_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_sltd_doc_number ON sltd_documents(document_number, issuing_country) WHERE status IN ('LOST','STOLEN','REVOKED');",
		"CREATE INDEX IF NOT EXISTS idx_sltd_holder    ON sltd_documents(holder_snisid_id) WHERE holder_snisid_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_sltd_status    ON sltd_documents(status);",
		"CREATE INDEX IF NOT EXISTS idx_sltd_check_log ON sltd_check_log(document_number, checked_at DESC);",
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
