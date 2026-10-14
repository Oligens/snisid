"""
SNISID Backend — FastAPI Application Entry Point
=================================================
Main FastAPI application with CQRS/Event-Sourcing identity management.
"""
from __future__ import annotations

import os
import uuid
from contextlib import asynccontextmanager
from typing import Any, AsyncGenerator

from fastapi import FastAPI, Depends, Query, status
from pydantic import BaseModel, Field
from sqlalchemy.ext.asyncio import AsyncSession

from shared.cache import close_redis_client, get_redis_client
from shared.config import get_settings
from shared.database import init_database, close_database, get_db_session
from shared.health import HealthCheck, create_health_router, check_database, check_kafka, check_redis
from shared.logging import configure_logging, get_logger
from shared.middleware import setup_middleware
from shared.telemetry import close_telemetry, init_telemetry

from services.agency.commands import (
    AgencyCommandHandler,
    CreateAgencyCommand,
    UpdateAgencyCommand,
    DeactivateAgencyCommand,
)
from services.bio_adn.kafka import init_bio_adn_kafka, shutdown_bio_adn_kafka
from services.agency.queries import (
    AgencyQueryHandler,
    GetAgencyByIdQuery,
    ListAgenciesQuery,
)
from services.did.api import router as did_router
from services.didcomm.api import router as didcomm_router
from services.credential_flow.api import router as credential_flow_router
from services.siopv2.api import router as siopv2_router
from services.wallet.api import router as wallet_router
from services.chapi.api import router as chapi_router
from services.credential_manifest.api import router as credential_manifest_router
from services.revocation.api import router as revocation_router
from services.didcomm_mediator.api import router as didcomm_mediator_router
from services.pex.api import router as pex_router
from services.bio_adn import bio_adn_router, init_bio_adn_kafka
from services.identity.commands import (
    IdentityCommandHandler,
    CreateIdentityCommand,
    UpdateIdentityCommand,
    VerifyIdentityCommand,
    SuspendIdentityCommand,
    RevokeIdentityCommand,
    EnrollBiometricCommand,
    IssueDocumentCommand,
)
from services.identity.queries import (
    IdentityQueryHandler,
    GetIdentityByIdQuery,
    GetIdentityByNationalIdQuery,
    SearchIdentitiesQuery,
    GetIdentityHistoryQuery,
    GetIdentityStatsQuery,
)
from services.identity.search import router as identity_search_router
from services.pki.api import create_pki_router
from services.sd_jwt.api import router as sd_jwt_router
from services.status_list.api import router as status_list_router
from services.vc.api import create_vc_router
from services.vp.api import router as vp_router

logger = get_logger(__name__)
settings = get_settings()


@asynccontextmanager
async def lifespan(app: FastAPI):
    configure_logging(service_name=settings.service_name)
    init_telemetry()
    await init_database()
    await get_redis_client()
    await init_bio_adn_kafka()
    logger.info("snisid_backend_started", version=settings.service_version)
    yield
    await shutdown_bio_adn_kafka()
    await close_database()
    await close_redis_client()
    close_telemetry()
    logger.info("snisid_backend_stopped")


app = FastAPI(
    title="SNISID Identity API",
    version=settings.service_version,
    description="National Identity Management — CQRS/Event-Sourcing Backend",
    lifespan=lifespan,
)

setup_middleware(app)

# ── Global Exception Handlers ──────────────────────────────────────────

from fastapi import HTTPException as FastAPIHTTPException
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse

@app.exception_handler(RequestValidationError)
async def validation_exception_handler(_request, exc: RequestValidationError):
    return JSONResponse(
        status_code=422,
        content={
            "detail": "Request validation failed",
            "errors": exc.errors(),
        },
    )

@app.exception_handler(FastAPIHTTPException)
async def http_exception_handler(_request, exc: FastAPIHTTPException):
    return JSONResponse(
        status_code=exc.status_code,
        content={"detail": exc.detail},
    )

# ── Optional API Key Authentication ────────────────────────────────────

if settings.environment.value != "development":
    from fastapi import Request
    from fastapi.responses import JSONResponse

    _API_KEY = os.environ.get("API_KEY")
    if not _API_KEY:
        raise RuntimeError("API_KEY environment variable is required in non-development mode")

    @app.middleware("http")
    async def api_key_middleware(request: Request, call_next):
        if request.url.path in ("/health", "/ready", "/metrics", "/docs", "/openapi.json", "/redoc"):
            return await call_next(request)
        api_key = request.headers.get("X-API-Key")
        if not api_key or api_key != _API_KEY:
            return JSONResponse(status_code=401, content={"detail": "Invalid or missing API key"})
        return await call_next(request)

# ── Health Probes (exempt from API key auth) ──────────────────────────

health = HealthCheck()
health.register("database", check_database)
health.register("redis", check_redis)
health.register("kafka", check_kafka)
app.include_router(create_health_router(health))

# ── SSI Routers ────────────────────────────────────────────────────────

app.include_router(create_vc_router())
app.include_router(create_pki_router())
app.include_router(sd_jwt_router)
app.include_router(did_router)
app.include_router(vp_router)
app.include_router(status_list_router)
app.include_router(didcomm_router)
app.include_router(credential_flow_router)
app.include_router(siopv2_router)
app.include_router(wallet_router)
app.include_router(chapi_router)
app.include_router(credential_manifest_router)
app.include_router(revocation_router)
app.include_router(didcomm_mediator_router)
app.include_router(pex_router)
app.include_router(bio_adn_router)
app.include_router(identity_search_router)


# ── Request/Response Schemas ───────────────────────────────────────────

class CreateIdentityRequest(BaseModel):
    national_id: str = Field(..., min_length=5, max_length=20)
    nif: str | None = Field(default=None, min_length=3, max_length=32)
    first_name: str = Field(..., min_length=1, max_length=100)
    last_name: str = Field(..., min_length=1, max_length=100)
    date_of_birth: str = Field(..., pattern=r"^\d{4}-\d{2}-\d{2}$")
    place_of_birth: str = Field(..., min_length=1, max_length=200)
    gender: str = Field(..., pattern=r"^(male|female|other)$")
    nationality: str = Field(..., min_length=3, max_length=3)
    agency_id: str
    actor_id: str = "system"
    correlation_id: str | None = None

class UpdateIdentityRequest(BaseModel):
    identity_id: str
    changes: dict[str, Any]
    actor_id: str = "system"
    correlation_id: str | None = None

class VerifyIdentityRequest(BaseModel):
    identity_id: str
    verification_method: str = "biometric"
    verifier_id: str
    actor_id: str = "system"

class SuspendIdentityRequest(BaseModel):
    identity_id: str
    reason: str = Field(..., min_length=10)
    actor_id: str = "system"

class RevokeIdentityRequest(BaseModel):
    identity_id: str
    reason: str = Field(..., min_length=10)
    actor_id: str = "system"

class EnrollBiometricRequest(BaseModel):
    identity_id: str
    biometric_type: str = Field(..., pattern=r"^(fingerprint|iris|face|voice)$")
    template_hash: str = Field(..., min_length=32)
    quality_score: float = Field(..., ge=0.0, le=1.0)
    actor_id: str = "system"

class IssueDocumentRequest(BaseModel):
    identity_id: str
    document_type: str
    document_number: str = Field(..., min_length=1)
    issue_date: str = Field(..., pattern=r"^\d{4}-\d{2}-\d{2}$")
    expiry_date: str = Field(..., pattern=r"^\d{4}-\d{2}-\d{2}$")
    issuing_agency: str
    actor_id: str = "system"


def get_command_handler(db: AsyncSession = Depends(get_db_session)) -> IdentityCommandHandler:
    return IdentityCommandHandler(db)

def get_query_handler(db: AsyncSession = Depends(get_db_session)) -> IdentityQueryHandler:
    return IdentityQueryHandler(db)


# ── Identity Commands (Write Side - CQRS) ─────────────────────────────

@app.post("/v1/identities", status_code=status.HTTP_201_CREATED)
async def create_identity(
    req: CreateIdentityRequest,
    handler: IdentityCommandHandler = Depends(get_command_handler),
):
    cmd = CreateIdentityCommand(
        national_id=req.national_id,
        nif=req.nif,
        first_name=req.first_name,
        last_name=req.last_name,
        date_of_birth=req.date_of_birth,
        place_of_birth=req.place_of_birth,
        gender=req.gender,
        nationality=req.nationality,
        agency_id=req.agency_id,
        actor_id=req.actor_id,
        correlation_id=req.correlation_id or str(uuid.uuid4()),
    )
    return await handler.handle_create(cmd)


@app.put("/v1/identities/{identity_id}")
async def update_identity(
    identity_id: str,
    changes: dict[str, Any],
    actor_id: str = "system",
    handler: IdentityCommandHandler = Depends(get_command_handler),
):
    cmd = UpdateIdentityCommand(
        identity_id=identity_id,
        changes=changes,
        actor_id=actor_id,
    )
    return await handler.handle_update(cmd)


@app.post("/v1/identities/{identity_id}/verify")
async def verify_identity(
    identity_id: str,
    verification_method: str = "biometric",
    verifier_id: str = "system",
    handler: IdentityCommandHandler = Depends(get_command_handler),
):
    cmd = VerifyIdentityCommand(
        identity_id=identity_id,
        verification_method=verification_method,
        verifier_id=verifier_id,
        actor_id=verifier_id,
    )
    return await handler.handle_verify(cmd)


@app.post("/v1/identities/{identity_id}/suspend")
async def suspend_identity(
    identity_id: str,
    reason: str = "Administrative suspension",
    handler: IdentityCommandHandler = Depends(get_command_handler),
):
    cmd = SuspendIdentityCommand(
        identity_id=identity_id,
        reason=reason,
        actor_id="system",
    )
    return await handler.handle_suspend(cmd)


@app.post("/v1/identities/{identity_id}/revoke")
async def revoke_identity(
    identity_id: str,
    reason: str = "Permanent revocation",
    handler: IdentityCommandHandler = Depends(get_command_handler),
):
    cmd = RevokeIdentityCommand(
        identity_id=identity_id,
        reason=reason,
        actor_id="system",
    )
    return await handler.handle_revoke(cmd)


@app.post("/v1/identities/{identity_id}/biometrics")
async def enroll_biometric(
    identity_id: str,
    req: EnrollBiometricRequest,
    handler: IdentityCommandHandler = Depends(get_command_handler),
):
    cmd = EnrollBiometricCommand(
        identity_id=identity_id,
        biometric_type=req.biometric_type,
        template_hash=req.template_hash,
        quality_score=req.quality_score,
        actor_id=req.actor_id,
    )
    return await handler.handle_enroll_biometric(cmd)


@app.post("/v1/identities/{identity_id}/documents")
async def issue_document(
    identity_id: str,
    req: IssueDocumentRequest,
    handler: IdentityCommandHandler = Depends(get_command_handler),
):
    cmd = IssueDocumentCommand(
        identity_id=identity_id,
        document_type=req.document_type,
        document_number=req.document_number,
        issue_date=req.issue_date,
        expiry_date=req.expiry_date,
        issuing_agency=req.issuing_agency,
        actor_id=req.actor_id,
    )
    return await handler.handle_issue_document(cmd)


# ── Identity Queries (Read Side - CQRS) ──────────────────────────────
# IMPORTANT: static routes MUST be registered before parameterised
# routes to avoid path-parameter capture (e.g. "stats" → identity_id).

@app.get("/v1/identities/stats")
async def get_identity_stats(
    agency_id: str | None = Query(None),
    handler: IdentityQueryHandler = Depends(get_query_handler),
):
    query = GetIdentityStatsQuery(agency_id=agency_id)
    return await handler.get_stats(query)


@app.get("/v1/identities/by-national-id/{national_id}")
async def get_identity_by_national_id(
    national_id: str,
    handler: IdentityQueryHandler = Depends(get_query_handler),
):
    query = GetIdentityByNationalIdQuery(national_id=national_id)
    return await handler.get_by_national_id(query)


@app.get("/v1/identities/{identity_id}/history")
async def get_identity_history(
    identity_id: str,
    after_version: int = Query(0),
    handler: IdentityQueryHandler = Depends(get_query_handler),
):
    query = GetIdentityHistoryQuery(identity_id=identity_id, after_version=after_version)
    return await handler.get_history(query)


@app.get("/v1/identities/{identity_id}")
async def get_identity(
    identity_id: str,
    handler: IdentityQueryHandler = Depends(get_query_handler),
):
    query = GetIdentityByIdQuery(identity_id=identity_id)
    return await handler.get_by_id(query)


@app.get("/v1/identities")
async def search_identities(
    search_term: str | None = Query(None),
    status: str | None = Query(None),
    agency_id: str | None = Query(None),
    nationality: str | None = Query(None),
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
    handler: IdentityQueryHandler = Depends(get_query_handler),
):
    query = SearchIdentitiesQuery(
        search_term=search_term,
        status=status,
        agency_id=agency_id,
        nationality=nationality,
        page=page,
        page_size=page_size,
    )
    return await handler.search(query)


# ── Agency Commands ────────────────────────────────────────────────────

def get_agency_command_handler(db: AsyncSession = Depends(get_db_session)) -> AgencyCommandHandler:
    return AgencyCommandHandler(db)


def get_agency_query_handler(db: AsyncSession = Depends(get_db_session)) -> AgencyQueryHandler:
    return AgencyQueryHandler(db)


@app.post("/v1/agencies", status_code=status.HTTP_201_CREATED)
async def create_agency(
    req: CreateAgencyCommand,
    handler: AgencyCommandHandler = Depends(get_agency_command_handler),
):
    return await handler.handle_create(req)


@app.put("/v1/agencies/{agency_id}")
async def update_agency(
    agency_id: str,
    changes: dict[str, Any],
    handler: AgencyCommandHandler = Depends(get_agency_command_handler),
):
    cmd = UpdateAgencyCommand(agency_id=agency_id, changes=changes)
    return await handler.handle_update(cmd)


@app.post("/v1/agencies/{agency_id}/deactivate")
async def deactivate_agency(
    agency_id: str,
    reason: str = "Administrative",
    handler: AgencyCommandHandler = Depends(get_agency_command_handler),
):
    cmd = DeactivateAgencyCommand(agency_id=agency_id, reason=reason)
    return await handler.handle_deactivate(cmd)


@app.get("/v1/agencies/{agency_id}")
async def get_agency(
    agency_id: str,
    handler: AgencyQueryHandler = Depends(get_agency_query_handler),
):
    query = GetAgencyByIdQuery(agency_id=agency_id)
    return await handler.get_by_id(query)


@app.get("/v1/agencies")
async def list_agencies(
    status: str | None = Query(None),
    agency_type: str | None = Query(None),
    department: str | None = Query(None),
    search_term: str | None = Query(None),
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
    handler: AgencyQueryHandler = Depends(get_agency_query_handler),
):
    query = ListAgenciesQuery(
        status=status, agency_type=agency_type,
        department=department, search_term=search_term,
        page=page, page_size=page_size,
    )
    return await handler.list_all(query)
