"""mTLS-protected, resilient outbound connectors for government institutions."""

from __future__ import annotations

import asyncio
import os
import secrets
import ssl
import time
from abc import ABC, abstractmethod
from collections import deque
from dataclasses import dataclass
from pathlib import Path
from typing import Any
from urllib.parse import quote, urlsplit

import httpx
from jose import jwt

from shared.config import get_settings
from shared.resilience import CircuitBreaker, RetryPolicy


class ConnectorConfigurationError(ValueError):
    """Raised when an institution connector is missing secure configuration."""


class InstitutionConnectorError(RuntimeError):
    """Base exception for institution connector failures."""


class RetryableInstitutionError(InstitutionConnectorError):
    """Transient remote response eligible for bounded retry."""


class InstitutionResponseError(InstitutionConnectorError):
    """Non-transient error returned by an institution."""

    def __init__(self, institution: str, status_code: int) -> None:
        self.institution = institution
        self.status_code = status_code
        super().__init__(f"{institution} returned HTTP {status_code}")


class InstitutionPayloadError(InstitutionConnectorError):
    """Institution returned a response that does not match the JSON contract."""


@dataclass(frozen=True, slots=True)
class InstitutionConnectorConfig:
    institution: str
    base_url: str
    audience: str
    citizen_record_path: str
    ca_bundle_path: str
    client_certificate_path: str
    client_private_key_path: str
    requests_per_second: int = 5
    max_retries: int = 2
    connect_timeout_seconds: float = 3.0
    read_timeout_seconds: float = 8.0
    write_timeout_seconds: float = 3.0
    pool_timeout_seconds: float = 2.0
    retry_base_delay_seconds: float = 0.25
    retry_max_delay_seconds: float = 2.0
    circuit_failure_threshold: int = 5
    circuit_recovery_timeout_seconds: float = 30.0

    def __post_init__(self) -> None:
        institution = self.institution.upper()
        object.__setattr__(self, "institution", institution)
        parsed_url = urlsplit(self.base_url)
        if parsed_url.scheme != "https" or not parsed_url.hostname:
            raise ConnectorConfigurationError(
                f"{institution} base URL must be an absolute HTTPS URL"
            )
        if parsed_url.username or parsed_url.password or parsed_url.query or parsed_url.fragment:
            raise ConnectorConfigurationError(
                f"{institution} base URL cannot contain credentials, query, or fragment"
            )
        if not self.audience.strip():
            raise ConnectorConfigurationError(f"{institution} OAuth audience is required")
        if (
            not self.citizen_record_path.startswith("/")
            or "{identity_id}" not in self.citizen_record_path
            or ".." in self.citizen_record_path.split("/")
            or "?" in self.citizen_record_path
            or "#" in self.citizen_record_path
        ):
            raise ConnectorConfigurationError(
                f"{institution} citizen record path must be an absolute path template "
                "containing {identity_id}, without query or traversal segments"
            )
        if self.requests_per_second < 1:
            raise ConnectorConfigurationError("requests_per_second must be positive")
        if self.max_retries < 0:
            raise ConnectorConfigurationError("max_retries cannot be negative")

    @classmethod
    def from_env(
        cls,
        institution: str,
        environ: dict[str, str] | None = None,
    ) -> InstitutionConnectorConfig:
        """Load non-secret endpoint and X.509 paths from institution-scoped env vars."""
        source = os.environ if environ is None else environ
        prefix = f"SNISID_INTEROP_{institution.upper()}_"
        required = (
            "BASE_URL",
            "AUDIENCE",
            "CITIZEN_RECORD_PATH",
            "CA_BUNDLE_PATH",
            "CLIENT_CERTIFICATE_PATH",
            "CLIENT_PRIVATE_KEY_PATH",
        )
        missing = [name for name in required if not source.get(prefix + name)]
        if missing:
            names = ", ".join(prefix + name for name in missing)
            raise ConnectorConfigurationError(f"Missing connector configuration: {names}")

        return cls(
            institution=institution,
            base_url=source[prefix + "BASE_URL"],
            audience=source[prefix + "AUDIENCE"],
            citizen_record_path=source[prefix + "CITIZEN_RECORD_PATH"],
            ca_bundle_path=source[prefix + "CA_BUNDLE_PATH"],
            client_certificate_path=source[prefix + "CLIENT_CERTIFICATE_PATH"],
            client_private_key_path=source[prefix + "CLIENT_PRIVATE_KEY_PATH"],
            requests_per_second=int(source.get(prefix + "REQUESTS_PER_SECOND", "5")),
            max_retries=int(source.get(prefix + "MAX_RETRIES", "2")),
            retry_base_delay_seconds=float(source.get(prefix + "RETRY_BASE_DELAY_SECONDS", "0.25")),
            retry_max_delay_seconds=float(source.get(prefix + "RETRY_MAX_DELAY_SECONDS", "2.0")),
            circuit_failure_threshold=int(source.get(prefix + "CIRCUIT_FAILURE_THRESHOLD", "5")),
            circuit_recovery_timeout_seconds=float(
                source.get(prefix + "CIRCUIT_RECOVERY_TIMEOUT_SECONDS", "30")
            ),
        )


class ServiceTokenSigner:
    """Issues short-lived RS256 service JWTs for one configured audience."""

    def __init__(
        self,
        *,
        issuer: str,
        subject: str,
        private_key_pem: str | None = None,
        private_key_path: str | None = None,
        lifetime_seconds: int = 60,
    ) -> None:
        if bool(private_key_pem) == bool(private_key_path):
            raise ConnectorConfigurationError(
                "Provide exactly one central JWT private key source"
            )
        if not issuer.strip() or not subject.strip():
            raise ConnectorConfigurationError("JWT issuer and subject are required")
        if not 1 <= lifetime_seconds <= 300:
            raise ConnectorConfigurationError("Service JWT lifetime must be between 1 and 300 seconds")

        try:
            self._private_key = (
                private_key_pem
                if private_key_pem is not None
                else Path(private_key_path or "").read_text(encoding="utf-8")
            )
        except OSError as exc:
            raise ConnectorConfigurationError("Unable to read central JWT signing key") from exc
        self._issuer = issuer
        self._subject = subject
        self._lifetime_seconds = lifetime_seconds

    @classmethod
    def from_settings(cls) -> ServiceTokenSigner:
        auth = get_settings().auth
        return cls(
            issuer=auth.jwt_issuer,
            subject=os.getenv("SNISID_INTEROP_SERVICE_SUBJECT", "snisid-interoperability"),
            private_key_pem=auth.jwt_private_key or None,
            private_key_path=auth.jwt_private_key_path or None,
        )

    def issue(self, audience: str) -> str:
        now = int(time.time())
        claims = {
            "iss": self._issuer,
            "sub": self._subject,
            "aud": audience,
            "iat": now,
            "nbf": now,
            "exp": now + self._lifetime_seconds,
            "jti": secrets.token_urlsafe(18),
        }
        return jwt.encode(claims, self._private_key, algorithm="RS256")


class AsyncRateLimiter:
    """A sliding-window limiter that caps requests per institution and per process."""

    def __init__(self, requests_per_second: int) -> None:
        if requests_per_second < 1:
            raise ValueError("requests_per_second must be positive")
        self._limit = requests_per_second
        self._window_seconds = 1.0
        self._timestamps: deque[float] = deque()
        self._lock = asyncio.Lock()

    async def acquire(self) -> None:
        loop = asyncio.get_running_loop()
        while True:
            async with self._lock:
                now = loop.time()
                while self._timestamps and now - self._timestamps[0] >= self._window_seconds:
                    self._timestamps.popleft()
                if len(self._timestamps) < self._limit:
                    self._timestamps.append(now)
                    return
                delay = self._window_seconds - (now - self._timestamps[0])
            await asyncio.sleep(delay)


def build_mtls_context(config: InstitutionConnectorConfig) -> ssl.SSLContext:
    """Build strict TLS with the configured State CA bundle and client X.509 identity."""
    paths = (
        Path(config.ca_bundle_path),
        Path(config.client_certificate_path),
        Path(config.client_private_key_path),
    )
    if any(not path.is_file() for path in paths):
        raise ConnectorConfigurationError(
            f"{config.institution} CA bundle, client certificate, and private key must exist"
        )

    try:
        context = ssl.create_default_context(ssl.Purpose.SERVER_AUTH, cafile=str(paths[0]))
        context.minimum_version = ssl.TLSVersion.TLSv1_2
        context.verify_mode = ssl.CERT_REQUIRED
        context.check_hostname = True
        context.load_cert_chain(certfile=str(paths[1]), keyfile=str(paths[2]))
        return context
    except (OSError, ssl.SSLError) as exc:
        raise ConnectorConfigurationError(
            f"Invalid X.509 mTLS material for {config.institution}"
        ) from exc


class InstitutionConnector(ABC):
    """Common async connector contract for an institutional citizen-record API."""

    def __init__(
        self,
        config: InstitutionConnectorConfig,
        token_signer: ServiceTokenSigner,
        *,
        transport: httpx.AsyncBaseTransport | None = None,
    ) -> None:
        self.config = config
        self._token_signer = token_signer
        self._rate_limiter = AsyncRateLimiter(config.requests_per_second)
        self._retry_policy = RetryPolicy(
            max_retries=config.max_retries,
            base_delay=config.retry_base_delay_seconds,
            max_delay=config.retry_max_delay_seconds,
            exponential_base=2.0,
            jitter=True,
            retryable_exceptions=(httpx.TransportError, RetryableInstitutionError),
        )
        self._circuit_breaker = CircuitBreaker(
            name=f"institution:{config.institution.lower()}",
            failure_threshold=config.circuit_failure_threshold,
            recovery_timeout=config.circuit_recovery_timeout_seconds,
            half_open_max_calls=1,
        )
        timeout = httpx.Timeout(
            connect=config.connect_timeout_seconds,
            read=config.read_timeout_seconds,
            write=config.write_timeout_seconds,
            pool=config.pool_timeout_seconds,
        )
        self._client = httpx.AsyncClient(
            verify=build_mtls_context(config),
            timeout=timeout,
            limits=httpx.Limits(max_connections=10, max_keepalive_connections=5),
            follow_redirects=False,
            transport=transport,
            headers={"Accept": "application/json"},
        )

    @abstractmethod
    async def fetch_citizen_record(self, identity_id: str) -> dict[str, Any]:
        """Fetch a citizen record from this institution using the configured identity."""

    async def _fetch_record(self, identity_id: str) -> dict[str, Any]:
        if not identity_id or not identity_id.strip():
            raise ValueError("identity_id is required")
        encoded_id = quote(identity_id, safe="")
        path = self.config.citizen_record_path.replace("{identity_id}", encoded_id)
        url = f"{self.config.base_url.rstrip('/')}/{path.lstrip('/')}"

        async def request_once() -> httpx.Response:
            await self._rate_limiter.acquire()
            token = self._token_signer.issue(self.config.audience)
            response = await self._client.get(
                url,
                headers={"Authorization": f"Bearer {token}"},
            )
            if response.status_code == 429 or response.status_code >= 500:
                raise RetryableInstitutionError(
                    f"{self.config.institution} temporarily returned HTTP {response.status_code}"
                )
            return response

        response = await self._circuit_breaker.call(
            self._retry_policy.execute,
            request_once,
        )
        if not response.is_success:
            raise InstitutionResponseError(self.config.institution, response.status_code)
        try:
            payload = response.json()
        except ValueError as exc:
            raise InstitutionPayloadError(
                f"{self.config.institution} returned invalid JSON"
            ) from exc
        if not isinstance(payload, dict):
            raise InstitutionPayloadError(
                f"{self.config.institution} returned a non-object citizen record"
            )
        return payload

    async def aclose(self) -> None:
        await self._client.aclose()


class DcpjConnector(InstitutionConnector):
    """DCPJ adapter; its URL, audience, certificate chain, and path are deployment config."""

    def __init__(
        self,
        token_signer: ServiceTokenSigner,
        *,
        config: InstitutionConnectorConfig | None = None,
        environ: dict[str, str] | None = None,
        transport: httpx.AsyncBaseTransport | None = None,
    ) -> None:
        resolved = config or InstitutionConnectorConfig.from_env("DCPJ", environ)
        if resolved.institution != "DCPJ":
            raise ConnectorConfigurationError("DcpjConnector requires DCPJ configuration")
        super().__init__(resolved, token_signer, transport=transport)

    async def fetch_citizen_record(self, identity_id: str) -> dict[str, Any]:
        return await self._fetch_record(identity_id)


class OniConnector(InstitutionConnector):
    """ONI adapter; its URL, audience, certificate chain, and path are deployment config."""

    def __init__(
        self,
        token_signer: ServiceTokenSigner,
        *,
        config: InstitutionConnectorConfig | None = None,
        environ: dict[str, str] | None = None,
        transport: httpx.AsyncBaseTransport | None = None,
    ) -> None:
        resolved = config or InstitutionConnectorConfig.from_env("ONI", environ)
        if resolved.institution != "ONI":
            raise ConnectorConfigurationError("OniConnector requires ONI configuration")
        super().__init__(resolved, token_signer, transport=transport)

    async def fetch_citizen_record(self, identity_id: str) -> dict[str, Any]:
        return await self._fetch_record(identity_id)


class InteroperabilityGateway:
    """Internal registry for invoking only explicitly configured institution adapters."""

    def __init__(self, connectors: list[InstitutionConnector]) -> None:
        self._connectors = {connector.config.institution: connector for connector in connectors}

    async def fetch_citizen_record(self, institution: str, identity_id: str) -> dict[str, Any]:
        connector = self._connectors.get(institution.upper())
        if connector is None:
            raise ConnectorConfigurationError(
                f"No connector is configured for institution {institution!r}"
            )
        return await connector.fetch_citizen_record(identity_id)

    async def aclose(self) -> None:
        await asyncio.gather(*(connector.aclose() for connector in self._connectors.values()))


def create_example_gateway(
    *,
    environ: dict[str, str] | None = None,
    token_signer: ServiceTokenSigner | None = None,
) -> InteroperabilityGateway:
    """Create DCPJ and ONI connectors from environment and central RS256 settings."""
    signer = token_signer or ServiceTokenSigner.from_settings()
    return InteroperabilityGateway(
        [
            DcpjConnector(signer, environ=environ),
            OniConnector(signer, environ=environ),
        ]
    )
