# Institutional interoperability connectors

This package is an internal outbound adapter layer. It deliberately does not
publish a citizen-record proxy route: callers must apply the existing
`shared.auth` authorization policy before invoking `InteroperabilityGateway`.
Institutional URLs and record paths are deployment-owned contracts and are not
hard-coded here.

## Required deployment configuration

The central service signing key must be supplied through the existing
`SNISID_AUTH__JWT_PRIVATE_KEY` setting or the `SNISID_AUTH__JWT_PRIVATE_KEY_PATH`
file-path setting (or populated by the existing Vault integration). It must be
an RS256 private key held by the deployment secret manager; never place its
value in source control or an image. The matching public key must be registered
with the institution for the configured audience.

Configure each adapter independently with these variables (`DCPJ` or `ONI`):

```text
SNISID_INTEROP_DCPJ_BASE_URL=https://<institution-host>
SNISID_INTEROP_DCPJ_AUDIENCE=<audience-assigned-by-DCPJ>
SNISID_INTEROP_DCPJ_CITIZEN_RECORD_PATH=<institution-contract-path>/{identity_id}
SNISID_INTEROP_DCPJ_CA_BUNDLE_PATH=/run/secrets/state-institution-ca.pem
SNISID_INTEROP_DCPJ_CLIENT_CERTIFICATE_PATH=/run/secrets/snisid-dcpj-client.pem
SNISID_INTEROP_DCPJ_CLIENT_PRIVATE_KEY_PATH=/run/secrets/snisid-dcpj-client-key.pem
SNISID_INTEROP_DCPJ_REQUESTS_PER_SECOND=5
SNISID_INTEROP_DCPJ_MAX_RETRIES=2
SNISID_INTEROP_DCPJ_RETRY_BASE_DELAY_SECONDS=0.25
SNISID_INTEROP_DCPJ_RETRY_MAX_DELAY_SECONDS=2
SNISID_INTEROP_DCPJ_CIRCUIT_FAILURE_THRESHOLD=5
SNISID_INTEROP_DCPJ_CIRCUIT_RECOVERY_TIMEOUT_SECONDS=30
```

Repeat with `ONI` and values provided by that institution. The CA bundle must
trust the institution's server certificate; the client certificate and key
must be issued for SNISID by the agreed government PKI. TLS certificate and
hostname verification cannot be disabled. No fallback to plain HTTP exists.

## Internal use

```python
from services.interoperability import create_example_gateway

gateway = create_example_gateway()
try:
    record = await gateway.fetch_citizen_record("DCPJ", identity_id)
finally:
    await gateway.aclose()
```

Create the gateway once in the application lifespan and close it at shutdown
for connection pooling. Do not create it per request. The connector applies
per-institution request throttling, bounded exponential retries for transport,
429, and 5xx failures, and a circuit breaker. 4xx responses are not retried.
Response bodies and tokens are intentionally excluded from logs.

The institution must separately define the `citizen_record_path`, OAuth
audience, accepted JWT claims, and response schema before enabling a production
connector. The returned JSON is passed through unchanged; callers must validate
it against an institution-specific schema before using it.