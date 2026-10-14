from __future__ import annotations

from datetime import UTC, datetime, timedelta

import httpx
import pytest
from cryptography import x509
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.x509.oid import ExtendedKeyUsageOID, NameOID
from jose import jwt

from services.interoperability.connectors import (
    ConnectorConfigurationError,
    DcpjConnector,
    InstitutionConnectorConfig,
    InstitutionResponseError,
    OniConnector,
    RetryableInstitutionError,
    ServiceTokenSigner,
)
from shared.resilience import CircuitOpenError


@pytest.fixture
def security_material(tmp_path):
    ca_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
    ca_name = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, "SNISID test CA")])
    now = datetime.now(UTC)
    ca_certificate = (
        x509.CertificateBuilder()
        .subject_name(ca_name)
        .issuer_name(ca_name)
        .public_key(ca_key.public_key())
        .serial_number(x509.random_serial_number())
        .not_valid_before(now - timedelta(minutes=1))
        .not_valid_after(now + timedelta(days=1))
        .add_extension(x509.BasicConstraints(ca=True, path_length=0), critical=True)
        .sign(ca_key, hashes.SHA256())
    )

    client_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
    client_name = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, "SNISID test client")])
    client_certificate = (
        x509.CertificateBuilder()
        .subject_name(client_name)
        .issuer_name(ca_name)
        .public_key(client_key.public_key())
        .serial_number(x509.random_serial_number())
        .not_valid_before(now - timedelta(minutes=1))
        .not_valid_after(now + timedelta(days=1))
        .add_extension(
            x509.ExtendedKeyUsage([ExtendedKeyUsageOID.CLIENT_AUTH]),
            critical=True,
        )
        .sign(ca_key, hashes.SHA256())
    )

    ca_path = tmp_path / "state-ca.pem"
    certificate_path = tmp_path / "snisid-client.pem"
    key_path = tmp_path / "snisid-client-key.pem"
    ca_path.write_bytes(ca_certificate.public_bytes(serialization.Encoding.PEM))
    certificate_path.write_bytes(client_certificate.public_bytes(serialization.Encoding.PEM))
    key_path.write_bytes(
        client_key.private_bytes(
            serialization.Encoding.PEM,
            serialization.PrivateFormat.PKCS8,
            serialization.NoEncryption(),
        )
    )
    signer = ServiceTokenSigner(
        issuer="snisid-test",
        subject="snisid-interoperability-test",
        private_key_pem=ca_key.private_bytes(
            serialization.Encoding.PEM,
            serialization.PrivateFormat.PKCS8,
            serialization.NoEncryption(),
        ).decode(),
    )
    return {
        "ca": ca_path,
        "certificate": certificate_path,
        "key": key_path,
        "signer": signer,
        "public_key": ca_key.public_key().public_bytes(
            serialization.Encoding.PEM,
            serialization.PublicFormat.SubjectPublicKeyInfo,
        ),
    }


def make_config(material, institution: str, **overrides) -> InstitutionConnectorConfig:
    values = {
        "institution": institution,
        "base_url": "https://institution.example.invalid",
        "audience": f"snisid-{institution.lower()}-api",
        "citizen_record_path": "/configured/records/{identity_id}",
        "ca_bundle_path": str(material["ca"]),
        "client_certificate_path": str(material["certificate"]),
        "client_private_key_path": str(material["key"]),
        "requests_per_second": 100,
        "max_retries": 1,
        "retry_base_delay_seconds": 0,
        "retry_max_delay_seconds": 0,
    }
    values.update(overrides)
    return InstitutionConnectorConfig(**values)


def test_config_rejects_plain_http_and_unconfigured_path(security_material):
    with pytest.raises(ConnectorConfigurationError, match="HTTPS"):
        make_config(security_material, "DCPJ", base_url="http://dcpj.example.invalid")

    with pytest.raises(ConnectorConfigurationError, match="path template"):
        make_config(security_material, "DCPJ", citizen_record_path="/records")


def test_environment_configuration_requires_state_mtls_material():
    with pytest.raises(ConnectorConfigurationError, match="Missing connector configuration"):
        InstitutionConnectorConfig.from_env("DCPJ", environ={})


async def test_dcpj_fetch_uses_signed_audience_token_and_encoded_identity(security_material):
    observed: list[httpx.Request] = []

    def handler(request: httpx.Request) -> httpx.Response:
        observed.append(request)
        return httpx.Response(200, json={"institution": "DCPJ", "record_id": "record-1"})

    connector = DcpjConnector(
        security_material["signer"],
        config=make_config(security_material, "DCPJ"),
        transport=httpx.MockTransport(handler),
    )
    try:
        result = await connector.fetch_citizen_record("NIF/2025 01")
    finally:
        await connector.aclose()

    assert result == {"institution": "DCPJ", "record_id": "record-1"}
    request = observed[0]
    assert request.method == "GET"
    assert request.url.raw_path.endswith(b"NIF%2F2025%2001")
    token = request.headers["Authorization"].removeprefix("Bearer ")
    claims = jwt.decode(
        token,
        security_material["public_key"],
        algorithms=["RS256"],
        audience="snisid-dcpj-api",
        issuer="snisid-test",
    )
    assert claims["sub"] == "snisid-interoperability-test"
    assert claims["exp"] - claims["iat"] == 60


async def test_oni_retries_transient_5xx_and_returns_record(security_material):
    attempts = 0

    def handler(_request: httpx.Request) -> httpx.Response:
        nonlocal attempts
        attempts += 1
        if attempts == 1:
            return httpx.Response(503)
        return httpx.Response(200, json={"institution": "ONI", "record_id": "record-2"})

    connector = OniConnector(
        security_material["signer"],
        config=make_config(security_material, "ONI"),
        transport=httpx.MockTransport(handler),
    )
    try:
        result = await connector.fetch_citizen_record("identity-2")
    finally:
        await connector.aclose()

    assert attempts == 2
    assert result["record_id"] == "record-2"


async def test_dcpj_does_not_retry_permanent_4xx(security_material):
    attempts = 0

    def handler(_request: httpx.Request) -> httpx.Response:
        nonlocal attempts
        attempts += 1
        return httpx.Response(403)

    connector = DcpjConnector(
        security_material["signer"],
        config=make_config(security_material, "DCPJ"),
        transport=httpx.MockTransport(handler),
    )
    try:
        with pytest.raises(InstitutionResponseError) as error:
            await connector.fetch_citizen_record("identity-3")
    finally:
        await connector.aclose()

    assert error.value.status_code == 403
    assert attempts == 1
    assert connector._circuit_breaker.failure_count == 0


async def test_oni_circuit_breaker_opens_after_transient_failures(security_material):
    def handler(_request: httpx.Request) -> httpx.Response:
        return httpx.Response(503)

    connector = OniConnector(
        security_material["signer"],
        config=make_config(
            security_material,
            "ONI",
            max_retries=0,
            circuit_failure_threshold=1,
        ),
        transport=httpx.MockTransport(handler),
    )
    try:
        with pytest.raises(RetryableInstitutionError, match="ONI temporarily returned HTTP 503"):
            await connector.fetch_citizen_record("identity-4")
        with pytest.raises(CircuitOpenError):
            await connector.fetch_citizen_record("identity-4")
    finally:
        await connector.aclose()


def test_connector_refuses_missing_mtls_files(security_material, tmp_path):
    config = make_config(
        security_material,
        "ONI",
        client_private_key_path=str(tmp_path / "missing-key.pem"),
    )
    with pytest.raises(ConnectorConfigurationError, match="must exist"):
        OniConnector(security_material["signer"], config=config)
