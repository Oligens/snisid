package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App        AppConfig
	HTTP       HTTPConfig
	DB         DBConfig
	Cockroach  CockroachConfig
	Redis      RedisConfig
	Neo4j      Neo4jConfig
	Kafka      KafkaConfig
	Milvus     MilvusConfig
	ClickHouse ClickHouseConfig
	JWT        JWTConfig
	HSM        HSMConfig
	SPIRE      SPIREConfig
	Vault      VaultConfig
	OTel       OTelConfig
	AI         AIConfig
	Features   FeatureConfig
}

type AppConfig struct {
	Env      string
	Name     string
	Version  string
	LogLevel string
}

type HTTPConfig struct {
	Host    string
	Port    int
	Timeout time.Duration
}

type DBConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
	PoolMin  int
	PoolMax  int
}

type CockroachConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type Neo4jConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

type KafkaConfig struct {
	Brokers       []string
	ClientID      string
	ConsumerGroup string
}

type MilvusConfig struct {
	Host string
	Port int
}

type ClickHouseConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

type JWTConfig struct {
	Secret     string
	TTL        time.Duration
	RefreshTTL time.Duration
}

type HSMConfig struct {
	PIN         string
	SlotID      int
	PKCS11Lib   string
	TrustDomain string
}

type SPIREConfig struct {
	AgentSocket string
	TrustDomain string
}

type VaultConfig struct {
	Addr   string
	Token  string
	KVPath string
}

type OTelConfig struct {
	ServiceName    string
	OTLPEndpoint   string
	PrometheusAddr string
}

type AIConfig struct {
	ModelEndpoint    string
	InferenceTimeout time.Duration
	FraudModelPath   string
}

type FeatureConfig struct {
	EnableDeepfakeDetection bool
	EnableFederatedSearch   bool
	EnableOfflineMode       bool
}

func Load() *Config {
	return &Config{
		App: AppConfig{
			Env:      getEnv("APP_ENV", "development"),
			Name:     getEnv("APP_NAME", "snisid-platform"),
			Version:  getEnv("APP_VERSION", "1.0.0"),
			LogLevel: getEnv("LOG_LEVEL", "debug"),
		},
		HTTP: HTTPConfig{
			Host:    getEnv("HTTP_HOST", "0.0.0.0"),
			Port:    getEnvInt("HTTP_PORT", 8080),
			Timeout: getEnvDuration("HTTP_TIMEOUT", 30*time.Second),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			Database: getEnv("DB_NAME", "snisid"),
			User:     getEnv("DB_USER", "snisid"),
			Password: getEnv("DB_PASSWORD", ""),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			PoolMin:  getEnvInt("DB_POOL_MIN", 2),
			PoolMax:  getEnvInt("DB_POOL_MAX", 10),
		},
		Cockroach: CockroachConfig{
			Host:     getEnv("COCKROACH_HOST", "localhost"),
			Port:     getEnvInt("COCKROACH_PORT", 26257),
			Database: getEnv("COCKROACH_DB", "snisid_ledger"),
			User:     getEnv("COCKROACH_USER", "snisid"),
			Password: getEnv("COCKROACH_PASSWORD", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Neo4j: Neo4jConfig{
			Host:     getEnv("NEO4J_HOST", "localhost"),
			Port:     getEnvInt("NEO4J_PORT", 7687),
			User:     getEnv("NEO4J_USER", "neo4j"),
			Password: getEnv("NEO4J_PASSWORD", ""),
		},
		Kafka: KafkaConfig{
			Brokers:       getEnvSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			ClientID:      getEnv("KAFKA_CLIENT_ID", "snisid-platform"),
			ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "snisid-core"),
		},
		Milvus: MilvusConfig{
			Host: getEnv("MILVUS_HOST", "localhost"),
			Port: getEnvInt("MILVUS_PORT", 19530),
		},
		ClickHouse: ClickHouseConfig{
			Host:     getEnv("CLICKHOUSE_HOST", "localhost"),
			Port:     getEnvInt("CLICKHOUSE_PORT", 9000),
			Database: getEnv("CLICKHOUSE_DB", "snisid"),
			User:     getEnv("CLICKHOUSE_USER", "default"),
			Password: getEnv("CLICKHOUSE_PASSWORD", ""),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "changer-moi"),
			TTL:        getEnvDuration("JWT_TTL", 15*time.Minute),
			RefreshTTL: getEnvDuration("JWT_REFRESH_TTL", 24*time.Hour),
		},
		HSM: HSMConfig{
			PIN:         getEnv("HSM_PIN", ""),
			SlotID:      getEnvInt("HSM_SLOT_ID", 0),
			PKCS11Lib:   getEnv("HSM_PKCS11_LIB", "/usr/lib/softhsm/libsofthsm2.so"),
			TrustDomain: getEnv("TRUST_DOMAIN", "snisid.gouv.ht"),
		},
		SPIRE: SPIREConfig{
			AgentSocket: getEnv("SPIRE_AGENT_SOCKET", "/tmp/spire-agent/public/api.sock"),
			TrustDomain: getEnv("SPIRE_TRUST_DOMAIN", "snisid.gouv.ht"),
		},
		Vault: VaultConfig{
			Addr:   getEnv("VAULT_ADDR", "http://localhost:8200"),
			Token:  getEnv("VAULT_TOKEN", ""),
			KVPath: getEnv("VAULT_KV_PATH", "secret/snisid"),
		},
		OTel: OTelConfig{
			ServiceName:    getEnv("OTEL_SERVICE_NAME", "snisid-platform"),
			OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			PrometheusAddr: getEnv("PROMETHEUS_ADDR", "0.0.0.0:2112"),
		},
		AI: AIConfig{
			ModelEndpoint:    getEnv("AI_MODEL_ENDPOINT", "http://localhost:8501"),
			InferenceTimeout: getEnvDuration("AI_INFERENCE_TIMEOUT", 5*time.Second),
			FraudModelPath:   getEnv("FRAUD_MODEL_PATH", "/models/fraud/v1"),
		},
		Features: FeatureConfig{
			EnableDeepfakeDetection: getEnvBool("FF_ENABLE_DEEPFAKE_DETECTION", false),
			EnableFederatedSearch:   getEnvBool("FF_ENABLE_FEDERATED_SEARCH", false),
			EnableOfflineMode:       getEnvBool("FF_ENABLE_OFFLINE_MODE", true),
		},
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_min_conns=%d pool_max_conns=%d",
		c.DB.Host, c.DB.Port, c.DB.User, c.DB.Password, c.DB.Database, c.DB.SSLMode, c.DB.PoolMin, c.DB.PoolMax)
}

func (c *Config) CockroachDSN() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		c.Cockroach.User, c.Cockroach.Password, c.Cockroach.Host, c.Cockroach.Port, c.Cockroach.Database)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

func (c *Config) ClickHouseAddr() string {
	return fmt.Sprintf("%s:%d", c.ClickHouse.Host, c.ClickHouse.Port)
}

func (c *Config) Neo4jAddr() string {
	return fmt.Sprintf("bolt://%s:%d", c.Neo4j.Host, c.Neo4j.Port)
}

func (c *Config) MilvusAddr() string {
	return fmt.Sprintf("%s:%d", c.Milvus.Host, c.Milvus.Port)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		return strings.Split(v, ",")
	}
	return fallback
}
