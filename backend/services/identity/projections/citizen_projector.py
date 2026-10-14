from __future__ import annotations

from datetime import datetime, timezone
from typing import Any

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from services.identity.models import Citizen, CitizenReadModel
from shared.logging import get_logger

logger = get_logger(__name__)

EVENT_TYPE_MAP = {
    "IdentityCreated": "created",
    "IdentityUpdated": "updated",
    "IdentityVerified": "verified",
    "IdentitySuspended": "suspended",
    "IdentityRevoked": "revoked",
    "BiometricEnrolled": "biometric_enrolled",
    "DocumentIssued": "document_issued",
}


class CitizenProjector:
    def __init__(self, session: AsyncSession) -> None:
        self._session = session

    async def project(self, event_type: str, aggregate_id: str, data: dict[str, Any]) -> None:
        mapping = EVENT_TYPE_MAP.get(event_type)
        if mapping is None:
            return
        handler = getattr(self, f"_on_{mapping}", None)
        if handler:
            await handler(aggregate_id, data)

    async def _on_created(self, aggregate_id: str, data: dict[str, Any]) -> None:
        read_model = CitizenReadModel(
            id=aggregate_id,
            national_id=data["national_id"],
            nif=data.get("nif"),
            full_name=f"{data['first_name']} {data['last_name']}",
            first_name=data["first_name"],
            last_name=data["last_name"],
            date_of_birth=datetime.fromisoformat(data["date_of_birth"]).date(),
            gender=data.get("gender", ""),
            nationality=data.get("nationality", ""),
            status="pending",
            agency_id=data.get("agency_id", ""),
            photo_url=data.get("photo_url"),
            last_event_at=datetime.now(timezone.utc),
        )
        self._session.add(read_model)
        await self._session.flush()

    async def _on_updated(self, aggregate_id: str, data: dict[str, Any]) -> None:
        result = await self._session.execute(
            select(CitizenReadModel).where(CitizenReadModel.id == aggregate_id)
        )
        model = result.scalar_one_or_none()
        if not model:
            return
        for key, value in (data.get("changes") or {}).items():
            if hasattr(model, key):
                setattr(model, key, value)
        if "first_name" in data.get("changes", {}) or "last_name" in data.get("changes", {}):
            model.full_name = f"{model.first_name} {model.last_name}"
        model.last_event_at = datetime.now(timezone.utc)
        await self._session.flush()

    async def _on_verified(self, aggregate_id: str, data: dict[str, Any]) -> None:
        result = await self._session.execute(
            select(CitizenReadModel).where(CitizenReadModel.id == aggregate_id)
        )
        model = result.scalar_one_or_none()
        if model:
            model.status = "active"
            model.verified = True
            model.last_event_at = datetime.now(timezone.utc)
            await self._session.flush()

    async def _on_suspended(self, aggregate_id: str, data: dict[str, Any]) -> None:
        result = await self._session.execute(
            select(CitizenReadModel).where(CitizenReadModel.id == aggregate_id)
        )
        model = result.scalar_one_or_none()
        if model:
            model.status = "suspended"
            model.last_event_at = datetime.now(timezone.utc)
            await self._session.flush()

    async def _on_revoked(self, aggregate_id: str, data: dict[str, Any]) -> None:
        result = await self._session.execute(
            select(CitizenReadModel).where(CitizenReadModel.id == aggregate_id)
        )
        model = result.scalar_one_or_none()
        if model:
            model.status = "revoked"
            model.last_event_at = datetime.now(timezone.utc)
            await self._session.delete(model)
            await self._session.flush()

    async def _on_biometric_enrolled(self, aggregate_id: str, data: dict[str, Any]) -> None:
        result = await self._session.execute(
            select(CitizenReadModel).where(CitizenReadModel.id == aggregate_id)
        )
        model = result.scalar_one_or_none()
        if model:
            model.has_biometrics = True
            model.last_event_at = datetime.now(timezone.utc)
            await self._session.flush()

    async def _on_document_issued(self, aggregate_id: str, data: dict[str, Any]) -> None:
        result = await self._session.execute(
            select(CitizenReadModel).where(CitizenReadModel.id == aggregate_id)
        )
        model = result.scalar_one_or_none()
        if model:
            model.document_count = (model.document_count or 0) + 1
            model.last_event_at = datetime.now(timezone.utc)
            await self._session.flush()
