"""
SNISID Configuration Management
=================================
Centralized Pydantic-based configuration with Vault integration.
All secrets are fetched from HashiCorp Vault — never hardcoded.
"""
from __future__ import annotations

import os
from enum import Enum
from functools import lru_cache
from typing import Any

from pydantic import Field, field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict


class Environment(str, Enum):
    DEVELOPMENT = "development"
    STAGING = "staging"
    PRODUCTION = "production"


class DatabaseConfig(BaseSettings):
    """PostgreSQL database configuration."""

    model_config = SettingsConfigDict(env_prefix="DB_")

    host: str = "localhost"
    port: int = 5432
    name: str = "snisid"
    user: str = "snisid"
    password: str = Field(default="", description="Loaded from Vault in production")
    pool_size: int = 20
    max_overflow: int = 10
    pool_timeout: int = 30
    pool_recycle: int = 1800
    echo: bool = False
    ssl_mode: str = "prefer"

    @property
    def async_url(self) -> str:
        return (
            f"postgresql+asyncpg://{self.user}:{self.password}"
            f"@{self.host}:{self.port}/{self.name}"
        )

    @property
    def sync_url(self) -> str:
        return (
            f"postgresql+psycopg://{self.user}:{self.password}"
            f"@{self.host}:{self.port}/{self.name}"
        )


class RedisConfig(BaseSettings):
    """Redis configuration for cache, sessions, and rate limiting."""

    model_config = SettingsConfigDict(env_prefix="REDIS_")

    host: str = "localhost"
    port: int = 6379
    db: int = 0
    password: str = Field(default="", description="Loaded from Vault in production")
    ssl: bool = False
    max_connections: int = 50
    socket_timeout: int = 5
    socket_connect_timeout: int = 5
    retry_on_timeout: bool = True

    # Dedicated databases
    cache_db: int = 0
    session_db: int = 1
    rate_limit_db: int = 2
    celery_db: int = 3

    @property
    def url(self) -> str:
        scheme = "rediss" if self.ssl else "redis"
        auth = f":{self.password}@" if self.password else ""
        return f"{scheme}://{auth}{self.host}:{self.port}/{self.db}"

    def get_url(self, db: int | None = None) -> str:
        scheme = "rediss" if self.ssl else "redis"
        auth = f":{self.password}@" if self.password else ""
        return f"{scheme}://{auth}{self.host}:{self.port}/{db or self.db}"


class KafkaConfig(BaseSettings):
    """Apache Kafka configuration."""

    model_config = SettingsConfigDict(env_prefix="KAFKA_")

    bootstrap_servers: str = "localhost:9092"
    security_protocol: str = "PLAINTEXT"
    sasl_mechanism: str | None = None
    sasl_username: str | None = None
    sasl_password: str | None = None
    ssl_cafile: str | None = None
    ssl_certfile: str | None = None
    ssl_keyfile: str | None = None

    # Producer settings
    producer_acks: str = "all"
    producer_retries: int = 3
    producer_batch_size: int = 16384
    producer_linger_ms: int = 10

    # Consumer settings
    consumer_group_prefix: str = "snisid"
    consumer_auto_offset_reset: str = "earliest"
    consumer_max_poll_records: int = 500
    consumer_session_timeout_ms: int = 30000

    # Topic prefixes
    topic_prefix: str = "snisid"

    @property
    def bootstrap_list(self) -> list[str]:
        return [s.strip() for s in self.bootstrap_servers.split(",")]


class AuthConfig(BaseSettings):
    """Authentication and authorization configuration."""

    model_config = SettingsConfigDict(env_prefix="AUTH_")

    # JWT
    jwt_algorithm: str = "RS256"
    jwt_access_token_expire_minutes: int = 30
    jwt_refresh_token_expire_days: int = 7
    jwt_issuer: str = "snisid-auth"
    jwt_audience: str = "snisid-api"

    # RS256 Keys (loaded from Vault or file paths)
    jwt_private_key_path: str = ""
    jwt_public_key_path: str = ""
    jwt_private_key: str = Field(default="", description="PEM private key from Vault")
    jwt_public_key: str = Field(default="", description="PEM public key from Vault")

    # Account security
    max_failed_attempts: int = 5
    lockout_duration_minutes: int = 30
    password_min_length: int = 12

    # MFA
    mfa_issuer: str = "SNISID"
    mfa_required_roles: list[str] = Field(
        default_factory=lambda: ["SUPER_ADMIN", "ADMIN", "REGISTRAR"]
    )

    # OAuth2 / Keycloak
    keycloak_url: str = ""
    keycloak_realm: str = "snisid"
    keycloak_client_id: str = ""
    keycloak_client_secret: str = ""


class CeleryConfig(BaseSettings):
    """Celery task queue configuration."""

    model_config = SettingsConfigDict(env_prefix="CELERY_")

    broker_url: str = "redis://localhost:6379/3"
    result_backend: str = "redis://localhost:6379/4"
    task_serializer: str = "json"
    result_serializer: str = "json"
    accept_content: list[str] = Field(default_factory=lambda: ["json"])
    timezone: str = "UTC"
    task_track_started: bool = True
    task_time_limit: int = 300
    task_soft_time_limit: int = 240
    worker_prefetch_multiplier: int = 1
    worker_max_tasks_per_child: int = 1000
    task_acks_late: bool = True
    task_reject_on_worker_lost: bool = True


class VaultConfig(BaseSettings):
    """HashiCorp Vault configuration."""

    model_config = SettingsConfigDict(env_prefix="VAULT_")

    enabled: bool = False
    url: str = "http://localhost:8200"
    token: str = ""
    mount_point: str = "secret"
    path_prefix: str = "snisid"
    namespace: str | None = None
    tls_verify: bool = True


class ObservabilityConfig(BaseSettings):
    """Observability and metrics configuration."""

    model_config = SettingsConfigDict(env_prefix="OTEL_")

    enabled: bool = True
    service_name: str = "snisid"
    exporter_endpoint: str = "http://localhost:4317"
    log_level: str = "INFO"
    metrics_port: int = 9090


class Settings(BaseSettings):
    """
    Master configuration aggregating all sub-configs.
    Secrets are resolved from Vault when VAULT_ENABLED=true.
    """

    model_config = SettingsConfigDict(
        env_prefix="SNISID_",
        env_nested_delimiter="__",
        case_sensitive=False,
    )

    # Global
    environment: Environment = Environment.DEVELOPMENT
    service_name: str = "snisid"
    service_version: str = "1.0.0"
    debug: bool = False
    host: str = Field(default_factory=lambda: os.getenv("SNISID_HOST", "127.0.0.1"))
    port: int = 8000
    workers: int = 4
    cors_origins: list[str] = Field(default_factory=lambda: ["*"])

    # Sub-configs
    database: DatabaseConfig = Field(default_factory=DatabaseConfig)
    redis: RedisConfig = Field(default_factory=RedisConfig)
    kafka: KafkaConfig = Field(default_factory=KafkaConfig)
    auth: AuthConfig = Field(default_factory=AuthConfig)
    celery: CeleryConfig = Field(default_factory=CeleryConfig)
    vault: VaultConfig = Field(default_factory=VaultConfig)
    observability: ObservabilityConfig = Field(default_factory=ObservabilityConfig)

    @field_validator("environment", mode="before")
    @classmethod
    def parse_environment(cls, v: Any) -> Environment:
        if isinstance(v, str):
            return Environment(v.lower())
        return v

    def load_secrets_from_vault(self) -> None:
        """Load secrets from HashiCorp Vault if enabled."""
        if not self.vault.enabled:
            return

        try:
            import hvac

            client = hvac.Client(
                url=self.vault.url,
                token=self.vault.token,
                namespace=self.vault.namespace,
                verify=self.vault.tls_verify,
            )

            if not client.is_authenticated():
                raise RuntimeError("Vault authentication failed")

            prefix = self.vault.path_prefix

            # Load database secrets
            db_secrets = client.secrets.kv.v2.read_secret_version(
                path=f"{prefix}/database",
                mount_point=self.vault.mount_point,
            )
            if db_secrets and "data" in db_secrets.get("data", {}):
                data = db_secrets["data"]["data"]
                self.database.password = data.get("password", self.database.password)
                self.database.user = data.get("username", self.database.user)

            # Load Redis secrets
            redis_secrets = client.secrets.kv.v2.read_secret_version(
                path=f"{prefix}/redis",
                mount_point=self.vault.mount_point,
            )
            if redis_secrets and "data" in redis_secrets.get("data", {}):
                data = redis_secrets["data"]["data"]
                self.redis.password = data.get("password", self.redis.password)

            # Load JWT keys
            jwt_secrets = client.secrets.kv.v2.read_secret_version(
                path=f"{prefix}/jwt",
                mount_point=self.vault.mount_point,
            )
            if jwt_secrets and "data" in jwt_secrets.get("data", {}):
                data = jwt_secrets["data"]["data"]
                self.auth.jwt_private_key = data.get("private_key", "")
                self.auth.jwt_public_key = data.get("public_key", "")

            # Load Kafka secrets
            kafka_secrets = client.secrets.kv.v2.read_secret_version(
                path=f"{prefix}/kafka",
                mount_point=self.vault.mount_point,
            )
            if kafka_secrets and "data" in kafka_secrets.get("data", {}):
                data = kafka_secrets["data"]["data"]
                self.kafka.sasl_username = data.get("username")
                self.kafka.sasl_password = data.get("password")

        except ImportError:
            raise RuntimeError("hvac package required for Vault integration")
        except Exception as e:
            raise RuntimeError(f"Failed to load secrets from Vault: {e}")


@lru_cache(maxsize=1)
def get_settings() -> Settings:
    """Get cached application settings singleton."""
    settings = Settings()
    settings.load_secrets_from_vault()
    return settings
