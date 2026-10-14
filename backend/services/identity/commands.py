"""
SNISID Identity Service — CQRS Commands
==========================================
Command handlers for identity write operations.
Each handler: validates → loads aggregate → executes → persists events → publishes to Kafka.
"""
from __future__ import annotations

import uuid
from typing import Any

from pydantic import BaseModel, Field
from sqlalchemy.ext.asyncio import AsyncSession

from shared.database.event_store import EventStore
from shared.logging import get_logger
from services.identity.aggregate import IdentityAggregate, InvalidStateError
from services.identity.events import IdentityEvent

logger = get_logger(__name__)


# ── Command Definitions ──────────────────────────────────────────────


class CreateIdentityCommand(BaseModel):
    """Command to create a new national identity."""

    command_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    national_id: str = Field(..., min_length=5, max_length=20)
    nif: str | None = Field(default=None, min_length=3, max_length=32)
    first_name: str = Field(..., min_length=1, max_length=100)
    last_name: str = Field(..., min_length=1, max_length=100)
    date_of_birth: str = Field(..., pattern=r"^\d{4}-\d{2}-\d{2}$")
    place_of_birth: str = Field(..., min_length=1, max_length=200)
    gender: str = Field(..., pattern=r"^(male|female|other)$")
    nationality: str = Field(..., min_length=3, max_length=3)
    agency_id: str = Field(...)
    actor_id: str = Field(...)
    correlation_id: str | None = None


class UpdateIdentityCommand(BaseModel):
    """Command to update an existing identity."""

    command_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    identity_id: str = Field(...)
    changes: dict[str, Any] = Field(...)
    actor_id: str = Field(...)
    correlation_id: str | None = None


class VerifyIdentityCommand(BaseModel):
    """Command to verify an identity."""

    command_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    identity_id: str = Field(...)
    verification_method: str = Field(...)
    verifier_id: str = Field(...)
    actor_id: str = Field(...)
    correlation_id: str | None = None


class SuspendIdentityCommand(BaseModel):
    """Command to suspend an identity."""

    command_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    identity_id: str = Field(...)
    reason: str = Field(..., min_length=10)
    actor_id: str = Field(...)
    correlation_id: str | None = None


class RevokeIdentityCommand(BaseModel):
    """Command to permanently revoke an identity."""

    command_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    identity_id: str = Field(...)
    reason: str = Field(..., min_length=10)
    actor_id: str = Field(...)
    correlation_id: str | None = None


class EnrollBiometricCommand(BaseModel):
    """Command to enroll a biometric template."""

    command_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    identity_id: str = Field(...)
    biometric_type: str = Field(..., pattern=r"^(fingerprint|iris|face|voice)$")
    template_hash: str = Field(..., min_length=32)
    quality_score: float = Field(..., ge=0.0, le=1.0)
    actor_id: str = Field(...)
    correlation_id: str | None = None


class IssueDocumentCommand(BaseModel):
    """Command to issue an identity document."""

    command_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    identity_id: str = Field(...)
    document_type: str = Field(...)
    document_number: str = Field(..., min_length=1)
    issue_date: str = Field(..., pattern=r"^\d{4}-\d{2}-\d{2}$")
    expiry_date: str = Field(..., pattern=r"^\d{4}-\d{2}-\d{2}$")
    issuing_agency: str = Field(...)
    actor_id: str = Field(...)
    correlation_id: str | None = None


# ── Command Handlers ─────────────────────────────────────────────────


class IdentityCommandHandler:
    """
    Handles all identity commands using Event Sourcing.
    
    Flow: Load aggregate from events → Execute command → Persist new events → Publish to Kafka
    """

    def __init__(
        self,
        session: AsyncSession,
        kafka_publisher: Any | None = None,
    ) -> None:
        self._session = session
        self._event_store = EventStore(session)
        self._kafka = kafka_publisher

    async def handle_create(self, cmd: CreateIdentityCommand) -> dict[str, Any]:
        """Handle CreateIdentityCommand."""
        logger.info(
            "handling_create_identity",
            national_id=cmd.national_id,
            nif=cmd.nif,
            command_id=cmd.command_id,
        )

        # Create aggregate (business rules validated inside)
        aggregate = IdentityAggregate.create(
            national_id=cmd.national_id,
            first_name=cmd.first_name,
            last_name=cmd.last_name,
            date_of_birth=cmd.date_of_birth,
            place_of_birth=cmd.place_of_birth,
            gender=cmd.gender,
            nationality=cmd.nationality,
            agency_id=cmd.agency_id,
            actor_id=cmd.actor_id,
            correlation_id=cmd.correlation_id,
        )

        # Persist events
        await self._persist_events(aggregate)

        logger.info(
            "identity_created",
            identity_id=aggregate.id,
            national_id=cmd.national_id,
        )

        return {"identity_id": aggregate.id, "status": "pending", "version": aggregate.version}

    async def handle_update(self, cmd: UpdateIdentityCommand) -> dict[str, Any]:
        """Handle UpdateIdentityCommand."""
        aggregate = await self._load_aggregate(cmd.identity_id)
        aggregate.update(
            changes=cmd.changes,
            actor_id=cmd.actor_id,
            correlation_id=cmd.correlation_id,
        )
        await self._persist_events(aggregate)
        return {"identity_id": aggregate.id, "version": aggregate.version}

    async def handle_verify(self, cmd: VerifyIdentityCommand) -> dict[str, Any]:
        """Handle VerifyIdentityCommand."""
        aggregate = await self._load_aggregate(cmd.identity_id)
        aggregate.verify(
            verification_method=cmd.verification_method,
            verifier_id=cmd.verifier_id,
            actor_id=cmd.actor_id,
            correlation_id=cmd.correlation_id,
        )
        await self._persist_events(aggregate)
        return {"identity_id": aggregate.id, "status": "active", "version": aggregate.version}

    async def handle_suspend(self, cmd: SuspendIdentityCommand) -> dict[str, Any]:
        """Handle SuspendIdentityCommand."""
        aggregate = await self._load_aggregate(cmd.identity_id)
        aggregate.suspend(
            reason=cmd.reason,
            actor_id=cmd.actor_id,
            correlation_id=cmd.correlation_id,
        )
        await self._persist_events(aggregate)
        return {"identity_id": aggregate.id, "status": "suspended", "version": aggregate.version}

    async def handle_revoke(self, cmd: RevokeIdentityCommand) -> dict[str, Any]:
        """Handle RevokeIdentityCommand."""
        aggregate = await self._load_aggregate(cmd.identity_id)
        aggregate.revoke(
            reason=cmd.reason,
            actor_id=cmd.actor_id,
            correlation_id=cmd.correlation_id,
        )
        await self._persist_events(aggregate)
        return {"identity_id": aggregate.id, "status": "revoked", "version": aggregate.version}

    async def handle_enroll_biometric(
        self, cmd: EnrollBiometricCommand
    ) -> dict[str, Any]:
        """Handle EnrollBiometricCommand."""
        aggregate = await self._load_aggregate(cmd.identity_id)
        aggregate.enroll_biometric(
            biometric_type=cmd.biometric_type,
            template_hash=cmd.template_hash,
            quality_score=cmd.quality_score,
            actor_id=cmd.actor_id,
            correlation_id=cmd.correlation_id,
        )
        await self._persist_events(aggregate)
        return {"identity_id": aggregate.id, "version": aggregate.version}

    async def handle_issue_document(
        self, cmd: IssueDocumentCommand
    ) -> dict[str, Any]:
        """Handle IssueDocumentCommand."""
        aggregate = await self._load_aggregate(cmd.identity_id)
        aggregate.issue_document(
            document_type=cmd.document_type,
            document_number=cmd.document_number,
            issue_date=cmd.issue_date,
            expiry_date=cmd.expiry_date,
            issuing_agency=cmd.issuing_agency,
            actor_id=cmd.actor_id,
            correlation_id=cmd.correlation_id,
        )
        await self._persist_events(aggregate)
        return {"identity_id": aggregate.id, "version": aggregate.version}

    # ── Private Helpers ───────────────────────────────────────────────

    async def _load_aggregate(self, aggregate_id: str) -> IdentityAggregate:
        """Load an aggregate from the event store, with snapshot optimization."""
        # Try to load from snapshot first
        snapshot = await self._event_store.get_snapshot(aggregate_id, "Identity")

        if snapshot:
            aggregate = IdentityAggregate.from_snapshot(
                aggregate_id, snapshot.state, snapshot.version
            )
            # Load events after snapshot
            events_data = await self._event_store.get_events(
                aggregate_id, after_version=snapshot.version
            )
        else:
            aggregate = IdentityAggregate()
            aggregate._id = aggregate_id
            events_data = await self._event_store.get_events(aggregate_id)

        if not events_data and not snapshot:
            raise IdentityNotFoundError(f"Identity {aggregate_id} not found")

        # Replay events
        for stored_event in events_data:
            event = IdentityEvent(
                event_id=stored_event.event_id,
                event_type=stored_event.event_type,
                aggregate_id=stored_event.aggregate_id,
                version=stored_event.version,
                data=stored_event.event_data,
                actor_id=stored_event.actor_id,
                timestamp=stored_event.timestamp,
            )
            aggregate._apply(event)
            aggregate._version = stored_event.version

        return aggregate

    async def _persist_events(self, aggregate: IdentityAggregate) -> None:
        """Persist pending events to the event store and publish to Kafka."""
        pending = aggregate.pending_events
        if not pending:
            return

        # Convert to event store format
        events_data = [
            {
                "event_type": e.event_type,
                "data": e.data,
                "metadata": {"actor_id": e.actor_id, "correlation_id": e.correlation_id},
                "causation_id": e.event_id,
            }
            for e in pending
        ]

        # Persist to event store
        expected_version = aggregate.version - len(pending)
        await self._event_store.append_events(
            aggregate_id=aggregate.id,
            aggregate_type="Identity",
            events=events_data,
            expected_version=expected_version,
            actor_id=pending[0].actor_id,
            correlation_id=pending[0].correlation_id,
        )

        # Check if snapshot needed
        if aggregate.version % EventStore.SNAPSHOT_INTERVAL == 0:
            await self._event_store.save_snapshot(
                aggregate_id=aggregate.id,
                aggregate_type="Identity",
                version=aggregate.version,
                state=aggregate.take_snapshot(),
            )

        # Publish to Kafka (fire-and-forget)
        if self._kafka:
            for event in pending:
                try:
                    await self._kafka.publish(
                        topic="snisid.identity.events",
                        key=aggregate.id,
                        event={
                            "event_id": event.event_id,
                            "event_type": event.event_type,
                            "aggregate_id": event.aggregate_id,
                            "data": event.data,
                            "timestamp": event.timestamp.isoformat(),
                            "version": event.version,
                            "actor_id": event.actor_id,
                        },
                    )
                except Exception as e:
                    logger.error(
                        "kafka_publish_failed",
                        event_id=event.event_id,
                        error=str(e),
                    )

        aggregate.clear_pending_events()


class IdentityNotFoundError(Exception):
    """Raised when an identity aggregate is not found."""
    pass
