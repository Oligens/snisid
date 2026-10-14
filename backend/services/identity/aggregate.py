"""
SNISID Identity Service — Aggregate Root
==========================================
Event-sourced aggregate implementing all business rules
for national identity lifecycle management.
"""
from __future__ import annotations

import uuid
from datetime import date, datetime, timezone
from typing import Any

from shared.logging import get_logger

from services.identity.events import (
    AddressChanged,
    BiometricEnrolled,
    DocumentIssued,
    IdentityCreated,
    IdentityEvent,
    IdentityReactivated,
    IdentityRevoked,
    IdentitySuspended,
    IdentityUpdated,
    IdentityVerified,
)

logger = get_logger(__name__)


class IdentityAggregate:
    """
    Event-sourced aggregate for national identity management.

    Business rules enforced:
    - National ID must be unique (checked externally)
    - Cannot verify a non-pending/non-active identity
    - Cannot suspend an already suspended/revoked identity
    - Cannot revoke an already revoked identity
    - Cannot enroll biometrics on a revoked identity
    - Age validation: DOB cannot be in the future
    - Biometric uniqueness: template_hash must be unique (checked externally)
    """

    def __init__(self) -> None:
        self._id: str = ""
        self._version: int = 0
        self._pending_events: list[IdentityEvent] = []

        # State
        self._national_id: str = ""
        self._nif: str | None = None
        self._first_name: str = ""
        self._last_name: str = ""
        self._middle_name: str | None = None
        self._date_of_birth: str = ""
        self._place_of_birth: str = ""
        self._gender: str = ""
        self._nationality: str = ""
        self._status: str = "pending"
        self._agency_id: str = ""
        self._created_by: str = ""
        self._address: dict[str, Any] | None = None
        self._photo_url: str | None = None
        self._biometrics: list[dict[str, Any]] = []
        self._documents: list[dict[str, Any]] = []
        self._verified_at: datetime | None = None
        self._verified_by: str | None = None
        self._suspension_reason: str | None = None
        self._revocation_reason: str | None = None

    # ── Properties ────────────────────────────────────────────────────

    @property
    def id(self) -> str:
        return self._id

    @property
    def version(self) -> int:
        return self._version

    @property
    def national_id(self) -> str:
        return self._national_id

    @property
    def nif(self) -> str | None:
        return self._nif

    @property
    def status(self) -> str:
        return self._status

    @property
    def pending_events(self) -> list[IdentityEvent]:
        return list(self._pending_events)

    # ── Commands ──────────────────────────────────────────────────────

    @classmethod
    def create(
        cls,
        national_id: str,
        first_name: str,
        last_name: str,
        date_of_birth: str,
        place_of_birth: str,
        gender: str,
        nationality: str,
        agency_id: str,
        actor_id: str,
        correlation_id: str | None = None,
        nif: str | None = None,
    ) -> IdentityAggregate:
        """
        Create a new identity. Factory method.

        Raises:
            ValueError: If validation fails.
        """
        # Validate DOB is not in the future
        dob = date.fromisoformat(date_of_birth)
        if dob > date.today():
            raise ValueError("Date of birth cannot be in the future")

        # Validate nationality is ISO 3166-1 alpha-3 (basic check)
        if len(nationality) != 3 or not nationality.isalpha():
            raise ValueError("Nationality must be a 3-letter ISO country code")

        aggregate = cls()
        aggregate._id = str(uuid.uuid4())

        event = IdentityCreated.create(
            aggregate_id=aggregate._id,
            national_id=national_id,
            first_name=first_name.strip(),
            last_name=last_name.strip(),
            date_of_birth=date_of_birth,
            place_of_birth=place_of_birth.strip(),
            gender=gender.lower(),
            nationality=nationality.upper(),
            agency_id=agency_id,
            actor_id=actor_id,
            correlation_id=correlation_id,
            nif=nif.strip() if nif else None,
        )
        aggregate._raise_event(event)
        return aggregate

    def update(
        self,
        changes: dict[str, Any],
        actor_id: str,
        correlation_id: str | None = None,
    ) -> None:
        """Update identity fields. Only allowed on PENDING or ACTIVE identities."""
        if self._status not in ("pending", "active"):
            raise InvalidStateError(
                f"Cannot update identity in '{self._status}' status"
            )

        # Filter out unchanged fields
        allowed_fields = {
            "first_name", "last_name", "middle_name", "phone",
            "email", "marital_status", "photo_url", "nif",
        }
        filtered = {k: v for k, v in changes.items() if k in allowed_fields}
        if not filtered:
            raise ValueError("No valid fields to update")

        event = IdentityUpdated.create(
            aggregate_id=self._id,
            changes=filtered,
            actor_id=actor_id,
            correlation_id=correlation_id,
        )
        self._raise_event(event)

    def verify(
        self,
        verification_method: str,
        verifier_id: str,
        actor_id: str,
        correlation_id: str | None = None,
    ) -> None:
        """Mark identity as verified."""
        if self._status not in ("pending", "active"):
            raise InvalidStateError(
                f"Cannot verify identity in '{self._status}' status"
            )

        event = IdentityVerified.create(
            aggregate_id=self._id,
            verification_method=verification_method,
            verifier_id=verifier_id,
            actor_id=actor_id,
            correlation_id=correlation_id,
        )
        self._raise_event(event)

    def suspend(
        self,
        reason: str,
        actor_id: str,
        correlation_id: str | None = None,
    ) -> None:
        """Suspend the identity."""
        if self._status in ("suspended", "revoked", "deceased"):
            raise InvalidStateError(
                f"Cannot suspend identity in '{self._status}' status"
            )

        if not reason.strip():
            raise ValueError("Suspension reason is required")

        event = IdentitySuspended.create(
            aggregate_id=self._id,
            reason=reason,
            actor_id=actor_id,
            correlation_id=correlation_id,
        )
        self._raise_event(event)

    def revoke(
        self,
        reason: str,
        actor_id: str,
        correlation_id: str | None = None,
    ) -> None:
        """Permanently revoke the identity."""
        if self._status == "revoked":
            raise InvalidStateError("Identity is already revoked")

        if not reason.strip():
            raise ValueError("Revocation reason is required")

        event = IdentityRevoked.create(
            aggregate_id=self._id,
            reason=reason,
            actor_id=actor_id,
            correlation_id=correlation_id,
        )
        self._raise_event(event)

    def reactivate(
        self,
        reason: str,
        actor_id: str,
        correlation_id: str | None = None,
    ) -> None:
        """Reactivate a suspended identity."""
        if self._status != "suspended":
            raise InvalidStateError(
                f"Can only reactivate suspended identities, current: '{self._status}'"
            )

        event = IdentityReactivated.create(
            aggregate_id=self._id,
            reason=reason,
            actor_id=actor_id,
            correlation_id=correlation_id,
        )
        self._raise_event(event)

    def enroll_biometric(
        self,
        biometric_type: str,
        template_hash: str,
        quality_score: float,
        actor_id: str,
        correlation_id: str | None = None,
    ) -> None:
        """Enroll a biometric template."""
        if self._status == "revoked":
            raise InvalidStateError("Cannot enroll biometrics on a revoked identity")

        if quality_score < 0.0 or quality_score > 1.0:
            raise ValueError("Quality score must be between 0.0 and 1.0")

        # Check if this biometric type already enrolled
        for bio in self._biometrics:
            if bio["biometric_type"] == biometric_type and bio.get("is_primary"):
                logger.info(
                    "biometric_replaced",
                    aggregate_id=self._id,
                    biometric_type=biometric_type,
                )

        event = BiometricEnrolled.create(
            aggregate_id=self._id,
            biometric_type=biometric_type,
            template_hash=template_hash,
            quality_score=quality_score,
            actor_id=actor_id,
            correlation_id=correlation_id,
        )
        self._raise_event(event)

    def issue_document(
        self,
        document_type: str,
        document_number: str,
        issue_date: str,
        expiry_date: str,
        issuing_agency: str,
        actor_id: str,
        correlation_id: str | None = None,
    ) -> None:
        """Issue an identity document."""
        if self._status in ("revoked", "deceased"):
            raise InvalidStateError(
                f"Cannot issue documents for identity in '{self._status}' status"
            )

        event = DocumentIssued.create(
            aggregate_id=self._id,
            document_type=document_type,
            document_number=document_number,
            issue_date=issue_date,
            expiry_date=expiry_date,
            issuing_agency=issuing_agency,
            actor_id=actor_id,
            correlation_id=correlation_id,
        )
        self._raise_event(event)

    def change_address(
        self,
        new_address: dict[str, Any],
        actor_id: str,
        correlation_id: str | None = None,
    ) -> None:
        """Change the citizen's address."""
        if self._status in ("revoked", "deceased"):
            raise InvalidStateError(
                f"Cannot change address for identity in '{self._status}' status"
            )

        event = AddressChanged.create(
            aggregate_id=self._id,
            old_address=self._address,
            new_address=new_address,
            actor_id=actor_id,
            correlation_id=correlation_id,
        )
        self._raise_event(event)

    # ── Event Application ─────────────────────────────────────────────

    def _raise_event(self, event: IdentityEvent) -> None:
        """Raise a domain event: apply + queue for persistence."""
        self._version += 1
        event.version = self._version
        self._apply(event)
        self._pending_events.append(event)

    def _apply(self, event: IdentityEvent) -> None:
        """Apply an event to mutate aggregate state."""
        handler_name = f"_when_{event.event_type.replace('.', '_')}"
        handler = getattr(self, handler_name, None)
        if handler:
            handler(event)
        else:
            logger.warning("unhandled_event_type", event_type=event.event_type)

    def _when_identity_created(self, event: IdentityCreated) -> None:
        data = event.data
        self._national_id = data["national_id"]
        self._nif = data.get("nif")
        self._first_name = data["first_name"]
        self._last_name = data["last_name"]
        self._date_of_birth = data["date_of_birth"]
        self._place_of_birth = data["place_of_birth"]
        self._gender = data["gender"]
        self._nationality = data["nationality"]
        self._agency_id = data["agency_id"]
        self._status = "pending"
        self._created_by = event.actor_id or ""

    def _when_identity_updated(self, event: IdentityUpdated) -> None:
        changes = event.data.get("changes", {})
        for field, value in changes.items():
            if hasattr(self, f"_{field}"):
                setattr(self, f"_{field}", value)

    def _when_identity_verified(self, event: IdentityVerified) -> None:
        self._status = "active"
        self._verified_at = event.timestamp
        self._verified_by = event.data.get("verifier_id")

    def _when_identity_suspended(self, event: IdentitySuspended) -> None:
        self._status = "suspended"
        self._suspension_reason = event.data.get("reason")

    def _when_identity_revoked(self, event: IdentityRevoked) -> None:
        self._status = "revoked"
        self._revocation_reason = event.data.get("reason")

    def _when_identity_reactivated(self, event: IdentityReactivated) -> None:
        self._status = "active"
        self._suspension_reason = None

    def _when_identity_biometric_enrolled(self, event: BiometricEnrolled) -> None:
        self._biometrics.append({
            "biometric_type": event.data["biometric_type"],
            "template_hash": event.data["template_hash"],
            "quality_score": event.data["quality_score"],
            "is_primary": True,
        })

    def _when_identity_document_issued(self, event: DocumentIssued) -> None:
        self._documents.append({
            "document_type": event.data["document_type"],
            "document_number": event.data["document_number"],
            "issue_date": event.data["issue_date"],
            "expiry_date": event.data["expiry_date"],
            "issuing_agency": event.data["issuing_agency"],
        })

    def _when_identity_address_changed(self, event: AddressChanged) -> None:
        self._address = event.data.get("new_address")

    # ── Reconstitution ────────────────────────────────────────────────

    @classmethod
    def from_events(cls, aggregate_id: str, events: list[IdentityEvent]) -> IdentityAggregate:
        """Reconstitute an aggregate from its event history."""
        aggregate = cls()
        aggregate._id = aggregate_id
        for event in events:
            aggregate._apply(event)
            aggregate._version = event.version
        return aggregate

    @classmethod
    def from_snapshot(
        cls, aggregate_id: str, snapshot: dict[str, Any], version: int
    ) -> IdentityAggregate:
        """Restore aggregate from a snapshot."""
        aggregate = cls()
        aggregate._id = aggregate_id
        aggregate._version = version
        for key, value in snapshot.items():
            if hasattr(aggregate, f"_{key}"):
                setattr(aggregate, f"_{key}", value)
        return aggregate

    def take_snapshot(self) -> dict[str, Any]:
        """Serialize current state as a snapshot."""
        return {
            "national_id": self._national_id,
            "nif": self._nif,
            "first_name": self._first_name,
            "last_name": self._last_name,
            "middle_name": self._middle_name,
            "date_of_birth": self._date_of_birth,
            "place_of_birth": self._place_of_birth,
            "gender": self._gender,
            "nationality": self._nationality,
            "status": self._status,
            "agency_id": self._agency_id,
            "created_by": self._created_by,
            "address": self._address,
            "photo_url": self._photo_url,
            "biometrics": self._biometrics,
            "documents": self._documents,
            "verified_at": self._verified_at.isoformat() if self._verified_at else None,
            "verified_by": self._verified_by,
            "suspension_reason": self._suspension_reason,
            "revocation_reason": self._revocation_reason,
        }

    def clear_pending_events(self) -> None:
        """Clear pending events after persistence."""
        self._pending_events.clear()


class InvalidStateError(Exception):
    """Raised when a command violates the aggregate's business rules."""
    pass
