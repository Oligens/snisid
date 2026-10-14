"""
SNISID Identity Service — Domain Events
==========================================
All domain events for the Identity bounded context.
Events are immutable records of state changes.
"""
from __future__ import annotations

from datetime import datetime, timezone
from typing import Any

from pydantic import BaseModel, Field

import uuid


class IdentityEvent(BaseModel):
    """Base class for all identity domain events."""

    event_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    event_type: str = ""
    aggregate_id: str = ""
    aggregate_type: str = "Identity"
    timestamp: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    version: int = 0
    actor_id: str | None = None
    correlation_id: str | None = None
    data: dict[str, Any] = Field(default_factory=dict)


class IdentityCreated(IdentityEvent):
    event_type: str = "identity.created"
    data: dict[str, Any] = Field(default_factory=dict)

    @classmethod
    def create(
        cls,
        aggregate_id: str,
        national_id: str,
        first_name: str,
        last_name: str,
        date_of_birth: str,
        place_of_birth: str,
        gender: str,
        nationality: str,
        agency_id: str,
        actor_id: str | None = None,
        correlation_id: str | None = None,
        nif: str | None = None,
    ) -> IdentityCreated:
        return cls(
            aggregate_id=aggregate_id,
            actor_id=actor_id,
            correlation_id=correlation_id,
            data={
                "national_id": national_id,
                "nif": nif,
                "first_name": first_name,
                "last_name": last_name,
                "date_of_birth": date_of_birth,
                "place_of_birth": place_of_birth,
                "gender": gender,
                "nationality": nationality,
                "agency_id": agency_id,
            },
        )


class IdentityUpdated(IdentityEvent):
    event_type: str = "identity.updated"

    @classmethod
    def create(
        cls,
        aggregate_id: str,
        changes: dict[str, Any],
        actor_id: str | None = None,
        correlation_id: str | None = None,
    ) -> IdentityUpdated:
        return cls(
            aggregate_id=aggregate_id,
            actor_id=actor_id,
            correlation_id=correlation_id,
            data={"changes": changes},
        )


class IdentityVerified(IdentityEvent):
    event_type: str = "identity.verified"

    @classmethod
    def create(
        cls,
        aggregate_id: str,
        verification_method: str,
        verifier_id: str,
        actor_id: str | None = None,
        correlation_id: str | None = None,
    ) -> IdentityVerified:
        return cls(
            aggregate_id=aggregate_id,
            actor_id=actor_id,
            correlation_id=correlation_id,
            data={
                "verification_method": verification_method,
                "verifier_id": verifier_id,
            },
        )


class IdentitySuspended(IdentityEvent):
    event_type: str = "identity.suspended"

    @classmethod
    def create(
        cls,
        aggregate_id: str,
        reason: str,
        actor_id: str | None = None,
        correlation_id: str | None = None,
    ) -> IdentitySuspended:
        return cls(
            aggregate_id=aggregate_id,
            actor_id=actor_id,
            correlation_id=correlation_id,
            data={"reason": reason},
        )


class IdentityRevoked(IdentityEvent):
    event_type: str = "identity.revoked"

    @classmethod
    def create(
        cls,
        aggregate_id: str,
        reason: str,
        actor_id: str | None = None,
        correlation_id: str | None = None,
    ) -> IdentityRevoked:
        return cls(
            aggregate_id=aggregate_id,
            actor_id=actor_id,
            correlation_id=correlation_id,
            data={"reason": reason},
        )


class IdentityReactivated(IdentityEvent):
    event_type: str = "identity.reactivated"

    @classmethod
    def create(
        cls,
        aggregate_id: str,
        reason: str,
        actor_id: str | None = None,
        correlation_id: str | None = None,
    ) -> IdentityReactivated:
        return cls(
            aggregate_id=aggregate_id,
            actor_id=actor_id,
            correlation_id=correlation_id,
            data={"reason": reason},
        )


class BiometricEnrolled(IdentityEvent):
    event_type: str = "identity.biometric_enrolled"

    @classmethod
    def create(
        cls,
        aggregate_id: str,
        biometric_type: str,
        template_hash: str,
        quality_score: float,
        actor_id: str | None = None,
        correlation_id: str | None = None,
    ) -> BiometricEnrolled:
        return cls(
            aggregate_id=aggregate_id,
            actor_id=actor_id,
            correlation_id=correlation_id,
            data={
                "biometric_type": biometric_type,
                "template_hash": template_hash,
                "quality_score": quality_score,
            },
        )


class DocumentIssued(IdentityEvent):
    event_type: str = "identity.document_issued"

    @classmethod
    def create(
        cls,
        aggregate_id: str,
        document_type: str,
        document_number: str,
        issue_date: str,
        expiry_date: str,
        issuing_agency: str,
        actor_id: str | None = None,
        correlation_id: str | None = None,
    ) -> DocumentIssued:
        return cls(
            aggregate_id=aggregate_id,
            actor_id=actor_id,
            correlation_id=correlation_id,
            data={
                "document_type": document_type,
                "document_number": document_number,
                "issue_date": issue_date,
                "expiry_date": expiry_date,
                "issuing_agency": issuing_agency,
            },
        )


class AddressChanged(IdentityEvent):
    event_type: str = "identity.address_changed"

    @classmethod
    def create(
        cls,
        aggregate_id: str,
        old_address: dict[str, Any] | None,
        new_address: dict[str, Any],
        actor_id: str | None = None,
        correlation_id: str | None = None,
    ) -> AddressChanged:
        return cls(
            aggregate_id=aggregate_id,
            actor_id=actor_id,
            correlation_id=correlation_id,
            data={"old_address": old_address, "new_address": new_address},
        )


# Event type registry for deserialization
EVENT_TYPE_MAP: dict[str, type[IdentityEvent]] = {
    "identity.created": IdentityCreated,
    "identity.updated": IdentityUpdated,
    "identity.verified": IdentityVerified,
    "identity.suspended": IdentitySuspended,
    "identity.revoked": IdentityRevoked,
    "identity.reactivated": IdentityReactivated,
    "identity.biometric_enrolled": BiometricEnrolled,
    "identity.document_issued": DocumentIssued,
    "identity.address_changed": AddressChanged,
}


def deserialize_event(event_type: str, data: dict[str, Any]) -> IdentityEvent:
    """Deserialize an event from stored data."""
    cls = EVENT_TYPE_MAP.get(event_type, IdentityEvent)
    return cls(**data)
