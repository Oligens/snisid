"""Permission-scoped, indexed identity search and optional source aggregation."""
from __future__ import annotations

import asyncio
from datetime import date
from typing import Annotated, Any, Protocol

from fastapi import APIRouter, Depends, HTTPException, Request, status
from pydantic import BaseModel, Field, field_validator, model_validator
from sqlalchemy import and_, case, false, func, literal, or_, select, text
from sqlalchemy.ext.asyncio import AsyncSession

from services.identity.models import CitizenReadModel
from shared.auth.dependencies import CurrentUser, require_permission
from shared.database import get_db_session
from shared.logging import get_logger

logger = get_logger(__name__)
router = APIRouter(prefix="/api/v1/search", tags=["identity-search"])

TRIGRAM_THRESHOLD = 0.3
MAX_PAGE_SIZE = 50
INSTITUTION_PERMISSIONS = {
    "OAVCT": "institution:oavct:identity:read",
    "MCI": "institution:mci:identity:read",
    "DCPJ": "institution:dcpj:identity:read",
}


class IdentitySearchFilters(BaseModel):
    """Optional identity criteria; identifiers are exact, names are fuzzy."""

    nom: str | None = Field(default=None, min_length=2, max_length=100)
    prenom: str | None = Field(default=None, min_length=2, max_length=100)
    nif: str | None = Field(default=None, min_length=3, max_length=32)
    nin: str | None = Field(default=None, min_length=5, max_length=20)
    date_naissance: date | None = None
    page: int = Field(default=1, ge=1)
    page_size: int = Field(default=20, ge=1, le=MAX_PAGE_SIZE)

    @field_validator("nom", "prenom", mode="before")
    @classmethod
    def normalize_name(cls, value: str | None) -> str | None:
        return " ".join(value.split()) if value else None

    @field_validator("nif", "nin", mode="before")
    @classmethod
    def normalize_identifier(cls, value: str | None) -> str | None:
        return value.strip().upper() if value else None

    @model_validator(mode="after")
    def require_search_criterion(self) -> IdentitySearchFilters:
        if not any((self.nom, self.prenom, self.nif, self.nin)):
            raise ValueError("Provide at least one of nom, prenom, nif, or nin")
        return self


SearchFiltersDependency = Annotated[IdentitySearchFilters, Depends()]
SearchSessionDependency = Annotated[AsyncSession, Depends(get_db_session)]
IdentitySearchUserDependency = Annotated[
    CurrentUser,
    Depends(require_permission("identity:search")),
]


class IdentitySearchHit(BaseModel):
    identity_id: str
    nom: str
    prenom: str
    date_naissance: date
    statut: str
    score: float
    correspondances: list[str]
    nif: str | None = None
    nin: str | None = None


class InstitutionalSourceResult(BaseModel):
    source: str
    status: str
    data: dict[str, Any] | None = None


class IdentitySearchResponse(BaseModel):
    items: list[IdentitySearchHit]
    total: int
    page: int
    page_size: int
    total_pages: int
    institutional_data: list[InstitutionalSourceResult]


class InstitutionRecordProvider(Protocol):
    """Adapter contract installed by the application for an approved source."""

    async def fetch_citizen_record(self, identity_id: str) -> dict[str, Any]: ...


def _name_search_expressions(filters: IdentitySearchFilters):
    """Build indexable pg_trgm predicates and ranking expressions."""
    first_name = func.lower(CitizenReadModel.first_name)
    last_name = func.lower(CitizenReadModel.last_name)
    full_name = func.lower(CitizenReadModel.full_name)
    first_query = filters.prenom.lower() if filters.prenom else None
    last_query = filters.nom.lower() if filters.nom else None

    predicates = []
    similarities = []
    if first_query and last_query:
        predicates.extend([
            (first_name.op("%")(first_query)) & (last_name.op("%")(last_query)),
            (first_name.op("%")(last_query)) & (last_name.op("%")(first_query)),
            full_name.op("%")(f"{first_query} {last_query}"),
            full_name.op("%")(f"{last_query} {first_query}"),
        ])
        similarities.extend([
            (func.similarity(first_name, first_query) + func.similarity(last_name, last_query)) / 2,
            (func.similarity(first_name, last_query) + func.similarity(last_name, first_query)) / 2,
            func.similarity(full_name, f"{first_query} {last_query}"),
            func.similarity(full_name, f"{last_query} {first_query}"),
        ])
    else:
        query = first_query or last_query
        if query:
            predicates.extend([
                first_name.op("%")(query),
                last_name.op("%")(query),
                full_name.op("%")(query),
            ])
            similarities.extend([
                func.similarity(first_name, query),
                func.similarity(last_name, query),
                func.similarity(full_name, query),
            ])

    if not predicates:
        return false(), literal(0.0)
    return or_(*predicates), func.greatest(*similarities)


def build_identity_search_statement(filters: IdentitySearchFilters, user: CurrentUser):
    """Build a PostgreSQL query whose trigram predicates can use the GIN indexes."""
    exact_nif = CitizenReadModel.nif == filters.nif if filters.nif else false()
    exact_nin = CitizenReadModel.national_id == filters.nin if filters.nin else false()
    provided_identifiers = [
        expression for expression, value in ((exact_nif, filters.nif), (exact_nin, filters.nin))
        if value is not None
    ]
    exact_identifier = (
        provided_identifiers[0]
        if len(provided_identifiers) == 1
        else and_(*provided_identifiers) if provided_identifiers
        else false()
    )
    fuzzy_name, name_similarity = _name_search_expressions(filters)
    candidate_match = exact_identifier if filters.nif or filters.nin else fuzzy_name

    rank = case(
        (exact_identifier, literal(1.0)),
        (fuzzy_name, literal(0.5) + literal(0.4) * name_similarity),
        else_=literal(0.0),
    )
    statement = select(
        CitizenReadModel.id.label("identity_id"),
        CitizenReadModel.first_name.label("prenom"),
        CitizenReadModel.last_name.label("nom"),
        CitizenReadModel.date_of_birth.label("date_naissance"),
        CitizenReadModel.status.label("statut"),
        CitizenReadModel.nif.label("nif"),
        CitizenReadModel.national_id.label("nin"),
        exact_nif.label("nif_exact"),
        exact_nin.label("nin_exact"),
        fuzzy_name.label("name_fuzzy"),
        rank.label("score"),
    ).where(
        CitizenReadModel.is_deleted.is_(False),
        candidate_match,
    )

    if filters.date_naissance:
        statement = statement.where(
            CitizenReadModel.date_of_birth == filters.date_naissance
        )
    if not user.has_role("SUPER_ADMIN"):
        statement = statement.where(CitizenReadModel.agency_id == user.agency_id)

    return statement


async def _aggregate_institutional_records(
    request: Request,
    user: CurrentUser,
    identity_id: str,
) -> list[InstitutionalSourceResult]:
    providers: dict[str, InstitutionRecordProvider] = getattr(
        request.app.state, "identity_search_providers", {}
    )
    results: list[InstitutionalSourceResult] = []

    async def fetch(source: str, provider: InstitutionRecordProvider) -> InstitutionalSourceResult:
        try:
            data = await asyncio.wait_for(
                provider.fetch_citizen_record(identity_id),
                timeout=3.0,
            )
            return InstitutionalSourceResult(source=source, status="AVAILABLE", data=data)
        except Exception:
            logger.warning(
                "identity_search_institution_unavailable",
                source=source,
                identity_id=identity_id,
            )
            return InstitutionalSourceResult(source=source, status="UNAVAILABLE")

    tasks = []
    for source, permission in INSTITUTION_PERMISSIONS.items():
        if not user.has_permission(permission):
            results.append(InstitutionalSourceResult(source=source, status="FORBIDDEN"))
            continue
        provider = providers.get(source)
        if provider is None:
            results.append(InstitutionalSourceResult(source=source, status="NOT_CONFIGURED"))
            continue
        tasks.append(fetch(source, provider))

    if tasks:
        results.extend(await asyncio.gather(*tasks))
    return results


@router.get("/identity", response_model=IdentitySearchResponse)
async def search_identity(
    request: Request,
    filters: SearchFiltersDependency,
    session: SearchSessionDependency,
    user: IdentitySearchUserDependency,
) -> IdentitySearchResponse:
    """Search the authorized identity register by exact identifiers or fuzzy name."""
    if filters.nif and not user.has_permission("identity:search:nif"):
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="NIF search is not permitted")
    if filters.nin and not user.has_permission("identity:search:nin"):
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="NIN search is not permitted")
    if not user.has_role("SUPER_ADMIN") and not user.agency_id:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="User has no agency scope")

    if session.get_bind().dialect.name == "postgresql":
        await session.execute(text("SET LOCAL pg_trgm.similarity_threshold = 0.3"))

    statement = build_identity_search_statement(filters, user)
    count_statement = select(func.count()).select_from(statement.order_by(None).subquery())
    total = (await session.execute(count_statement)).scalar_one()
    rows = (
        await session.execute(
            statement.order_by(text("score DESC"), CitizenReadModel.last_name, CitizenReadModel.first_name)
            .offset((filters.page - 1) * filters.page_size)
            .limit(filters.page_size)
        )
    ).mappings().all()

    can_read_identifiers = user.has_permission("identity:read:identifiers")
    items = [
        IdentitySearchHit(
            identity_id=row["identity_id"],
            nom=row["nom"],
            prenom=row["prenom"],
            date_naissance=row["date_naissance"],
            statut=row["statut"],
            score=round(float(row["score"]), 4),
            correspondances=[
                name
                for name, matched in (
                    ("NIF exact", row["nif_exact"]),
                    ("NIN exact", row["nin_exact"]),
                    ("Nom/prénom flous", row["name_fuzzy"]),
                    ("Date de naissance exacte", filters.date_naissance is not None),
                )
                if matched
            ],
            nif=row["nif"] if can_read_identifiers else None,
            nin=row["nin"] if can_read_identifiers else None,
        )
        for row in rows
    ]

    exact_identifier_hits = [
        row for row in rows if row["nif_exact"] or row["nin_exact"]
    ]
    institutional_data: list[InstitutionalSourceResult] = []
    if total == 1 and len(exact_identifier_hits) == 1 and request is not None:
        institutional_data = await _aggregate_institutional_records(
            request,
            user,
            exact_identifier_hits[0]["identity_id"],
        )

    logger.info(
        "identity_search_completed",
        actor_id=user.id,
        agency_scope=user.agency_id,
        criteria_fields=[
            name for name, value in (
                ("nom", filters.nom),
                ("prenom", filters.prenom),
                ("nif", filters.nif),
                ("nin", filters.nin),
                ("date_naissance", filters.date_naissance),
            ) if value is not None
        ],
        result_count=total,
    )

    return IdentitySearchResponse(
        items=items,
        total=total,
        page=filters.page,
        page_size=filters.page_size,
        total_pages=(total + filters.page_size - 1) // filters.page_size,
        institutional_data=institutional_data,
    )