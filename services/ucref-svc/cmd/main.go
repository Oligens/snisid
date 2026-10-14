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

	"github.com/snisid/platform/services/ucref-svc/internal/handler"
	"github.com/snisid/platform/services/ucref-svc/internal/kafka"
	"github.com/snisid/platform/services/ucref-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/ucref-svc/internal/service"
)

func main() {
	dbHost := getEnv("UCREF_DB_HOST", "localhost")
	dbPort := getEnv("UCREF_DB_PORT", "26257")
	dbName := getEnv("UCREF_DB_NAME", "snisid_ucref")
	dbUser := getEnv("UCREF_DB_USER", "root")
	dbSSLMode := getEnv("UCREF_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("UCREF_KAFKA_BROKERS", "localhost:9092")

	port := getEnv("UCREF_SERVICE_PORT", "8112")

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

	producer, err := kafka.NewProducer([]string{kafkaBrokers}, logger)
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
	}
	defer producer.Close()

	repo := postgres.NewSTRRepository(pool)
	svc := service.NewUCREFService(repo, logger)
	_ = svc

	h := handler.NewHealthHandler(pool, logger)
	r := gin.Default()
	r.GET("/healthz", h.Healthz)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("ucref-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS ucref_str_status AS ENUM ( 'RECEIVED','UNDER_ANALYSIS','DISSEMINATED','ARCHIVED','NO_ACTION' );",
		"CREATE TYPE IF NOT EXISTS ucref_report_type AS ENUM ( 'STR',          -- Suspicious Transaction Report 'CTR',          -- Cash Transaction Report (> HTG 500,000) 'INTERNATIONAL_WIRE', 'REAL_ESTATE', 'MONCASH_PATTERN', 'CRYPTO_PATTERN' );",
		"CREATE TABLE IF NOT EXISTS ucref_str_reports ( str_id              UUID PRIMARY KEY DEFAULT gen_random_uuid(), national_str_id     VARCHAR(25) UNIQUE NOT NULL,   -- STR-HT-AAAA-NNNNNN report_type         ucref_report_type NOT NULL, status              ucref_str_status NOT NULL DEFAULT 'RECEIVED', reporting_institution VARCHAR(200) NOT NULL, institution_type    VARCHAR(30),    -- BANK, MSB, MONCASH, INSURANCE, CASINO report_date         TIMESTAMPTZ NOT NULL, transaction_date    TIMESTAMPTZ, transaction_amount  DECIMAL(18,2), transaction_currency CHAR(3) DEFAULT 'HTG', transaction_amount_usd DECIMAL(18,2), subject_snisid_ids  UUID[] DEFAULT '{}', subject_names       TEXT[] DEFAULT '{}', subject_accounts    TEXT[] DEFAULT '{}', suspicious_activity TEXT NOT NULL, ml_typology         VARCHAR(100),   -- Smurfing, Trade-Based ML, Real Estate, etc. predicate_crime     VARCHAR(100),   -- Crime sous-jacent suspecte gang_id             UUID,           -- Si lien gang identifie fpr_person_ids      UUID[] DEFAULT '{}', sanc_match_ids      UUID[] DEFAULT '{}', analyst_id          UUID, analysis_notes      TEXT, disseminated_to     TEXT[] DEFAULT '{}',   -- DCPJ, MJSP, PARQUET, etc. disseminated_at     TIMESTAMPTZ, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS ucref_financial_profiles ( profile_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(), snisid_person_id    UUID NOT NULL UNIQUE, total_str_count     INTEGER DEFAULT 0, total_ctr_count     INTEGER DEFAULT 0, estimated_illegal_assets_usd DECIMAL(18,2), known_accounts      JSONB,          -- [{institution, account_type, country}] known_properties    JSONB,          -- [{address, value, acquisition_date}] known_businesses    TEXT[] DEFAULT '{}', ml_risk_score       SMALLINT CHECK (ml_risk_score BETWEEN 0 AND 100), is_pep              BOOLEAN DEFAULT FALSE,   -- Politically Exposed Person last_updated        TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS ucref_moncash_patterns ( pattern_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(), str_id              UUID REFERENCES ucref_str_reports(str_id), phone_number        VARCHAR(20) NOT NULL,    -- Numero MonCash snisid_person_id    UUID, pattern_type        VARCHAR(50),    -- STRUCTURING, RAPID_TRANSFERS, RANSOM_RECEIPT transaction_count   INTEGER, total_amount_htg    DECIMAL(18,2), period_start        TIMESTAMPTZ, period_end          TIMESTAMPTZ, notes               TEXT, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_ucref_str_status  ON ucref_str_reports(status, report_date DESC);",
		"CREATE INDEX IF NOT EXISTS idx_ucref_str_gang    ON ucref_str_reports(gang_id) WHERE gang_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_ucref_str_subjects ON ucref_str_reports USING gin(subject_snisid_ids);",
		"CREATE INDEX IF NOT EXISTS idx_ucref_profiles    ON ucref_financial_profiles(snisid_person_id);",
		"CREATE INDEX IF NOT EXISTS idx_ucref_moncash     ON ucref_moncash_patterns(phone_number);",
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
