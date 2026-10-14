"""
SNISID Identity Service — CQRS Queries
========================================
Query handlers for identity read operations.
All queries hit the denormalized read model, not the event store.
"""
from __future__ import annotations

from datetime import datetime
from typing import Any, Sequence

from pydantic import BaseModel, Field
from sqlalchemy import func, or_, select
from sqlalchemy.ext.asyncio import AsyncSession

from shared.database.event_store import EventStore, StoredEvent
from shared.logging import get_logger
from services.identity.models import Citizen, CitizenReadModel, IdentityDocument, BiometricRecord

logger = get_logger(__name__)


# ── Query Definitions ─────────────────────────────────────────────────


class GetIdentityByIdQuery(BaseModel):
    """Query to get a single identity by ID."""
    identity_id: str


class GetIdentityByNationalIdQuery(BaseModel):
    """Query to get an identity by national ID."""
    national_id: str


class SearchIdentitiesQuery(BaseModel):
    """Query to search identities with filters and pagination."""
    search_term: str | None = None
    status: str | None = None
    agency_id: str | None = None
    nationality: str | None = None
    date_of_birth_from: str | None = None
    date_of_birth_to: str | None = None
    page: int = Field(default=1, ge=1)
    page_size: int = Field(default=20, ge=1, le=100)
    sort_by: str = "created_at"
    sort_order: str = "desc"


class GetIdentityHistoryQuery(BaseModel):
    """Query to get the event history of an identity."""
    identity_id: str
    after_version: int = 0
    limit: int = 100


class GetIdentityStatsQuery(BaseModel):
    """Query to get identity statistics for an agency."""
    agency_id: str | None = None


# ── Response Models ───────────────────────────────────────────────────


class IdentityResponse(BaseModel):
    """Full identity response with HATEOAS links."""
    id: str
    national_id: str
    nif: str | None = None
    first_name: str
    last_name: str
    full_name: str
    date_of_birth: str
    gender: str
    nationality: str
    status: str
    agency_id: str
    verified: bool
    photo_url: str | None = None
    document_count: int = 0
    has_biometrics: bool = False
    created_at: str
    updated_at: str
    links: dict[str, Any] = Field(default_factory=dict)


class IdentityDetailResponse(IdentityResponse):
    """Detailed identity response including documents and biometrics."""
    place_of_birth: str | None = None
    middle_name: str | None = None
    email: str | None = None
    phone: str | None = None
    marital_status: str | None = None
    address: dict[str, Any] | None = None
    documents: list[dict[str, Any]] = Field(default_factory=list)
    biometrics: list[dict[str, Any]] = Field(default_factory=list)
    verified_at: str | None = None
    verified_by: str | None = None
    version: int = 0


class PaginatedResponse(BaseModel):
    """Paginated list response."""
    items: list[IdentityResponse]
    total: int
    page: int
    page_size: int
    total_pages: int
    links: dict[str, Any] = Field(default_factory=dict)


class IdentityHistoryEntry(BaseModel):
    """Single event in the identity history."""
    event_id: str
    event_type: str
    version: int
    timestamp: str
    actor_id: str | None = None
    data: dict[str, Any]


class IdentityStatsResponse(BaseModel):
    """Identity statistics."""
    total: int = 0
    by_status: dict[str, int] = Field(default_factory=dict)
    by_nationality: dict[str, int] = Field(default_factory=dict)
    verified_count: int = 0
    biometric_enrolled_count: int = 0


# ── Query Handler ─────────────────────────────────────────────────────


class IdentityQueryHandler:
    """
    Handles all identity read queries against the denormalized read model.
    Falls back to the write model when needed.
    """

    def __init__(self, session: AsyncSession) -> None:
        self._session = session
        self._event_store = EventStore(session)

    async def get_by_id(self, query: GetIdentityByIdQuery) -> IdentityDetailResponse:
        """Get a full identity by its ID."""
        result = await self._session.execute(
            select(Citizen)
            .where(Citizen.id == query.identity_id, Citizen.is_deleted == False)
        )
        citizen = result.scalar_one_or_none()
        if not citizen:
            raise IdentityQueryNotFoundError(f"Identity {query.identity_id} not found")

        return self._to_detail_response(citizen)

    async def get_by_national_id(
        self, query: GetIdentityByNationalIdQuery
    ) -> IdentityDetailResponse:
        """Get an identity by national ID."""
        result = await self._session.execute(
            select(Citizen)
            .where(Citizen.national_id == query.national_id, Citizen.is_deleted == False)
        )
        citizen = result.scalar_one_or_none()
        if not citizen:
            raise IdentityQueryNotFoundError(
                f"Identity with national_id {query.national_id} not found"
            )

        return self._to_detail_response(citizen)

    async def search(self, query: SearchIdentitiesQuery) -> PaginatedResponse:
        """Search identities with filters and pagination."""
        stmt = select(CitizenReadModel).where(True == True)  # noqa: E712

        # Apply filters
        if query.search_term:
            search = f"%{query.search_term}%"
            stmt = stmt.where(
                or_(
                    CitizenReadModel.full_name.ilike(search),
                    CitizenReadModel.national_id.ilike(search),
                )
            )

        if query.status:
            stmt = stmt.where(CitizenReadModel.status == query.status)

        if query.agency_id:
            stmt = stmt.where(CitizenReadModel.agency_id == query.agency_id)

        if query.nationality:
            stmt = stmt.where(CitizenReadModel.nationality == query.nationality)

        # Count total
        count_stmt = select(func.count()).select_from(stmt.subquery())
        total_result = await self._session.execute(count_stmt)
        total = total_result.scalar_one()

        # Apply sorting
        sort_column = getattr(CitizenReadModel, query.sort_by, CitizenReadModel.created_at)
        if query.sort_order == "desc":
            stmt = stmt.order_by(sort_column.desc())
        else:
            stmt = stmt.order_by(sort_column.asc())

        # Apply pagination
        offset = (query.page - 1) * query.page_size
        stmt = stmt.offset(offset).limit(query.page_size)

        result = await self._session.execute(stmt)
        items = result.scalars().all()

        total_pages = (total + query.page_size - 1) // query.page_size

        return PaginatedResponse(
            items=[self._to_detail_response_from_readmodel(item) for item in items],
            total=total,
            page=query.page,
            page_size=query.page_size,
            total_pages=total_pages,
            links=self._build_pagination_links(query, total_pages),
        )

    async def get_history(
        self, query: GetIdentityHistoryQuery
    ) -> list[IdentityHistoryEntry]:
        """Get the event history of an identity."""
        events = await self._event_store.get_events(
            aggregate_id=query.identity_id,
            after_version=query.after_version,
            limit=query.limit,
        )

        return [
            IdentityHistoryEntry(
                event_id=e.event_id,
                event_type=e.event_type,
                version=e.version,
                timestamp=e.timestamp.isoformat(),
                actor_id=e.actor_id,
                data=e.event_data,
            )
            for e in events
        ]

    async def get_stats(self, query: GetIdentityStatsQuery) -> IdentityStatsResponse:
        """Get aggregate statistics."""
        base_stmt = select(CitizenReadModel).where(CitizenReadModel.is_deleted == False)  # noqa: E712
        if query.agency_id:
            base_stmt = base_stmt.where(CitizenReadModel.agency_id == query.agency_id)

        # Total count
        total_result = await self._session.execute(
            select(func.count()).select_from(base_stmt.subquery())
        )
        total = total_result.scalar_one()

        # By status
        status_result = await self._session.execute(
            select(CitizenReadModel.status, func.count())
            .where(CitizenReadModel.is_deleted == False)  # noqa: E712
            .group_by(CitizenReadModel.status)
        )
        by_status = {row[0]: row[1] for row in status_result.all()}

        # By nationality
        nationality_result = await self._session.execute(
            select(CitizenReadModel.nationality, func.count())
            .where(CitizenReadModel.is_deleted == False)  # noqa: E712
            .group_by(CitizenReadModel.nationality)
        )
        by_nationality = {row[0]: row[1] for row in nationality_result.all()}

        # Verified count
        verified_result = await self._session.execute(
            select(func.count())
            .select_from(CitizenReadModel)
            .where(CitizenReadModel.is_deleted == False, CitizenReadModel.verified == True)  # noqa: E712
        )
        verified_count = verified_result.scalar_one()

        return IdentityStatsResponse(
            total=total,
            by_status=by_status,
            by_nationality=by_nationality,
            verified_count=verified_count,
        )

    # ── Private Helpers ───────────────────────────────────────────────

    def _to_detail_response(self, citizen: Citizen) -> IdentityDetailResponse:
        """Convert a Citizen model to a detailed response."""
        documents = [
            {
                "id": doc.id,
                "document_type": doc.document_type.value if doc.document_type else None,
                "document_number": doc.document_number,
                "issue_date": doc.issue_date.isoformat() if doc.issue_date else None,
                "expiry_date": doc.expiry_date.isoformat() if doc.expiry_date else None,
                "status": doc.status.value if doc.status else None,
                "issuing_agency": doc.issuing_agency,
            }
            for doc in (citizen.documents or [])
        ]

        biometrics = [
            {
                "id": bio.id,
                "biometric_type": bio.biometric_type.value if bio.biometric_type else None,
                "quality_score": bio.quality_score,
                "captured_at": bio.captured_at.isoformat() if bio.captured_at else None,
                "is_primary": bio.is_primary,
            }
            for bio in (citizen.biometrics or [])
        ]

        return IdentityDetailResponse(
            id=citizen.id,
            national_id=citizen.national_id,
            nif=citizen.nif,
            first_name=citizen.first_name,
            last_name=citizen.last_name,
            full_name=f"{citizen.first_name} {citizen.last_name}",
            middle_name=citizen.middle_name,
            date_of_birth=citizen.date_of_birth.isoformat(),
            place_of_birth=citizen.place_of_birth,
            gender=citizen.gender.value if citizen.gender else "",
            nationality=citizen.nationality,
            status=citizen.status.value if citizen.status else "",
            agency_id=citizen.agency_id,
            verified=citizen.verified_at is not None,
            verified_at=citizen.verified_at.isoformat() if citizen.verified_at else None,
            verified_by=citizen.verified_by,
            photo_url=citizen.photo_url,
            email=citizen.email,
            phone=citizen.phone,
            marital_status=citizen.marital_status,
            address=citizen.address,
            document_count=len(documents),
            has_biometrics=len(biometrics) > 0,
            documents=documents,
            biometrics=biometrics,
            version=citizen.version,
            created_at=citizen.created_at.isoformat(),
            updated_at=citizen.updated_at.isoformat(),
            links=self._build_identity_links(citizen.id),
        )

    @staticmethod
    def _build_identity_links(identity_id: str) -> dict[str, Any]:
        """Build HATEOAS links for an identity."""
        base = f"/v1/identities/{identity_id}"
        return {
            "self": {"href": base, "method": "GET"},
            "update": {"href": base, "method": "PUT"},
            "verify": {"href": f"{base}/verify", "method": "POST"},
            "suspend": {"href": f"{base}/suspend", "method": "POST"},
            "revoke": {"href": base, "method": "DELETE"},
            "history": {"href": f"{base}/history", "method": "GET"},
            "documents": {"href": f"{base}/documents", "method": "GET"},
            "biometrics": {"href": f"{base}/biometrics", "method": "GET"},
        }

    @staticmethod
    def _build_pagination_links(
        query: SearchIdentitiesQuery, total_pages: int
    ) -> dict[str, Any]:
        """Build HATEOAS pagination links preserving filter parameters."""
        def _href(page: int) -> str:
            params = [f"page={page}", f"page_size={query.page_size}"]
            if query.search_term:
                params.append(f"search_term={query.search_term}")
            if query.status:
                params.append(f"status={query.status}")
            if query.agency_id:
                params.append(f"agency_id={query.agency_id}")
            if query.nationality:
                params.append(f"nationality={query.nationality}")
            return "/v1/identities?" + "&".join(params)

        links: dict[str, Any] = {
            "self": {"href": _href(query.page)},
        }
        if query.page > 1:
            links["prev"] = {"href": _href(query.page - 1)}
        if query.page < total_pages:
            links["next"] = {"href": _href(query.page + 1)}
        links["first"] = {"href": _href(1)}
        links["last"] = {"href": _href(total_pages)}
        return links

    @staticmethod
    def _to_detail_response_from_readmodel(record: CitizenReadModel) -> IdentityDetailResponse:
        """Convert a CitizenReadModel to a detail response using available fields."""
        return IdentityDetailResponse(
            id=record.id,
            national_id=record.national_id,
            nif=record.nif,
            first_name=record.first_name,
            last_name=record.last_name,
            full_name=record.full_name,
            date_of_birth=record.date_of_birth.isoformat(),
            gender=record.gender,
            nationality=record.nationality,
            status=record.status,
            agency_id=record.agency_id,
            verified=record.verified,
            photo_url=record.photo_url,
            document_count=record.document_count,
            has_biometrics=record.has_biometrics,
            created_at=record.created_at.isoformat(),
            updated_at=record.updated_at.isoformat(),
            links=IdentityQueryHandler._build_identity_links(record.id),
        )


class IdentityQueryNotFoundError(Exception):
    """Raised when an identity is not found in the read model."""
    pass
