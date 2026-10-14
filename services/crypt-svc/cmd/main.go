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

	"github.com/snisid/platform/services/crypt-svc/internal/handler"
	"github.com/snisid/platform/services/crypt-svc/internal/kafka"
	"github.com/snisid/platform/services/crypt-svc/internal/repository/postgres"
	"github.com/snisid/platform/services/crypt-svc/internal/service"
)

func main() {
	dbHost := getEnv("CRYPT_DB_HOST", "localhost")
	dbPort := getEnv("CRYPT_DB_PORT", "26257")
	dbName := getEnv("CRYPT_DB_NAME", "snisid_crypt")
	dbUser := getEnv("CRYPT_DB_USER", "root")
	dbSSLMode := getEnv("CRYPT_DB_SSLMODE", "disable")
	kafkaBrokers := getEnv("CRYPT_KAFKA_BROKERS", "localhost:9092")

	port := getEnv("CRYPT_SERVICE_PORT", "8117")

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

	repo := postgres.NewWalletRepo(pool)
	svc := service.NewCryptService(repo, logger)
	_ = svc

	h := handler.NewHealthHandler(pool, logger)
	r := gin.Default()
	r.GET("/healthz", h.Healthz)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		logger.Info("crypt-svc started", zap.String("port", port))
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
		"CREATE TYPE IF NOT EXISTS crypt_asset_type AS ENUM ( 'BITCOIN', 'ETHEREUM', 'USDT', 'USDC', 'MONERO', 'ZCASH', 'LITECOIN', 'OTHER_ERC20', 'UNKNOWN' );",
		"CREATE TYPE IF NOT EXISTS crypt_suspicion_type AS ENUM ( 'RANSOM_RECEIPT', 'SANCTIONS_EVASION', 'DARKWEB_PAYMENT', 'MIXER_SERVICE', 'PEER_TO_PEER_UNREGULATED', 'EXCHANGE_HIGH_RISK', 'GANG_PAYMENT', 'UNKNOWN' );",
		"CREATE TABLE IF NOT EXISTS crypt_flagged_wallets ( wallet_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(), national_crypt_id   VARCHAR(25) UNIQUE NOT NULL,  -- CRYPT-HT-NNNNNN wallet_address      VARCHAR(200) NOT NULL, asset_type          crypt_asset_type NOT NULL, blockchain_network  VARCHAR(50),                  -- Bitcoin, Ethereum, Tron, etc. suspicion_type      crypt_suspicion_type NOT NULL, snisid_person_id    UUID, gang_id             UUID, estimated_balance_usd DECIMAL(18,2), total_received_usd  DECIMAL(18,2), total_sent_usd      DECIMAL(18,2), first_tx_date       TIMESTAMPTZ, last_tx_date        TIMESTAMPTZ, is_sanctioned       BOOLEAN DEFAULT FALSE, ofac_sdn_ref        VARCHAR(50), chainalysis_ref     VARCHAR(100),                 -- Ref rapport Chainalysis elliptic_ref        VARCHAR(100),                 -- Ref rapport Elliptic source_intel        TEXT, linked_cases        UUID[] DEFAULT '{}', is_frozen           BOOLEAN DEFAULT FALSE, freeze_jurisdiction VARCHAR(50), created_by          UUID NOT NULL, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS crypt_transactions ( tx_id               UUID PRIMARY KEY DEFAULT gen_random_uuid(), wallet_id           UUID REFERENCES crypt_flagged_wallets(wallet_id), tx_hash             VARCHAR(100) NOT NULL, asset_type          crypt_asset_type NOT NULL, direction           VARCHAR(10) NOT NULL,          -- INCOMING, OUTGOING from_address        VARCHAR(200), to_address          VARCHAR(200), amount_crypto       DECIMAL(30,18), amount_usd_at_tx    DECIMAL(18,2), tx_timestamp        TIMESTAMPTZ NOT NULL, block_number        BIGINT, is_mixer_involved   BOOLEAN DEFAULT FALSE, mixer_service       VARCHAR(100), risk_score          SMALLINT, suspicion_flags     TEXT[] DEFAULT '{}', extors_case_id      UUID,                          -- Lien ranÃ§on si applicable ucref_str_id        UUID, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE TABLE IF NOT EXISTS crypt_exchange_accounts ( exchange_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), snisid_person_id    UUID, exchange_name       VARCHAR(100) NOT NULL,         -- Binance, Coinbase, LocalBitcoins exchange_country    CHAR(3), account_ref         VARCHAR(200),                  -- Partiellement masque kyc_level           VARCHAR(20),                   -- NONE, BASIC, FULL total_volume_usd    DECIMAL(18,2), is_flagged          BOOLEAN DEFAULT FALSE, flagging_reason     TEXT, legal_hold_request  BOOLEAN DEFAULT FALSE, created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW() );",
		"CREATE INDEX IF NOT EXISTS idx_crypt_wallets_address ON crypt_flagged_wallets(wallet_address);",
		"CREATE INDEX IF NOT EXISTS idx_crypt_wallets_gang    ON crypt_flagged_wallets(gang_id) WHERE gang_id IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_crypt_wallets_sanctioned ON crypt_flagged_wallets(is_sanctioned) WHERE is_sanctioned = TRUE;",
		"CREATE INDEX IF NOT EXISTS idx_crypt_tx_wallet       ON crypt_transactions(wallet_id, tx_timestamp DESC);",
		"CREATE INDEX IF NOT EXISTS idx_crypt_tx_hash         ON crypt_transactions(tx_hash);",
		"CREATE INDEX IF NOT EXISTS idx_crypt_tx_mixer        ON crypt_transactions(is_mixer_involved) WHERE is_mixer_involved = TRUE;",
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
