"""
SNISID Identity Service — SQLAlchemy Models
=============================================
Core data models for national identity management.
PII fields are stored with encryption markers for at-rest protection.
"""
from __future__ import annotations

import enum
from datetime import date, datetime, timezone
from typing import Any

from sqlalchemy import (
    Date,
    DateTime,
    Enum,
    ForeignKey,
    Index,
    Integer,
    String,
    Text,
    Float,
)
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import Mapped, mapped_column, relationship

from shared.database import Base


class IdentityStatus(str, enum.Enum):
    """Lifecycle status of a national identity."""
    PENDING = "pending"
    ACTIVE = "active"
    SUSPENDED = "suspended"
    REVOKED = "revoked"
    DECEASED = "deceased"


class Gender(str, enum.Enum):
    MALE = "male"
    FEMALE = "female"
    OTHER = "other"


class DocumentType(str, enum.Enum):
    NATIONAL_ID = "national_id"
    PASSPORT = "passport"
    BIRTH_CERTIFICATE = "birth_certificate"
    DRIVING_LICENSE = "driving_license"
    RESIDENCE_PERMIT = "residence_permit"


class DocumentStatus(str, enum.Enum):
    ACTIVE = "active"
    EXPIRED = "expired"
    REVOKED = "revoked"
    LOST = "lost"
    STOLEN = "stolen"


class BiometricType(str, enum.Enum):
    FINGERPRINT = "fingerprint"
    IRIS = "iris"
    FACE = "face"
    VOICE = "voice"


class Citizen(Base):
    """
    Core national identity record.
    Represents a citizen in the SNISID system.
    """

    __tablename__ = "citizens"

    # National identification
    national_id: Mapped[str] = mapped_column(
        String(20), unique=True, nullable=False, index=True,
        doc="Unique national identification number",
    )
    nif: Mapped[str | None] = mapped_column(
        String(32), nullable=True, index=True,
        doc="Optional tax identification number",
    )

    # Personal information (PII — encrypted at rest)
    first_name: Mapped[str] = mapped_column(String(100), nullable=False)
    last_name: Mapped[str] = mapped_column(String(100), nullable=False)
    middle_name: Mapped[str | None] = mapped_column(String(100), nullable=True)
    date_of_birth: Mapped[date] = mapped_column(Date, nullable=False)
    place_of_birth: Mapped[str] = mapped_column(String(200), nullable=False)
    gender: Mapped[Gender] = mapped_column(
        Enum(Gender, name="gender_enum"), nullable=False
    )
    nationality: Mapped[str] = mapped_column(String(3), nullable=False, doc="ISO 3166-1 alpha-3")
    marital_status: Mapped[str | None] = mapped_column(String(20), nullable=True)

    # Contact
    email: Mapped[str | None] = mapped_column(String(255), nullable=True)
    phone: Mapped[str | None] = mapped_column(String(20), nullable=True)

    # Address (structured JSON)
    address: Mapped[dict[str, Any] | None] = mapped_column(JSONB, nullable=True)

    # Photo
    photo_url: Mapped[str | None] = mapped_column(String(500), nullable=True)

    # Biometric hashes (stored as references, not raw data)
    fingerprint_hash: Mapped[str | None] = mapped_column(String(128), nullable=True)
    iris_hash: Mapped[str | None] = mapped_column(String(128), nullable=True)
    face_encoding_hash: Mapped[str | None] = mapped_column(String(128), nullable=True)

    # Lifecycle
    status: Mapped[IdentityStatus] = mapped_column(
        Enum(IdentityStatus, name="identity_status_enum"),
        nullable=False,
        default=IdentityStatus.PENDING,
        index=True,
    )
    verified_at: Mapped[datetime | None] = mapped_column(
        DateTime(timezone=True), nullable=True
    )
    verified_by: Mapped[str | None] = mapped_column(String(36), nullable=True)
    suspension_reason: Mapped[str | None] = mapped_column(Text, nullable=True)
    revocation_reason: Mapped[str | None] = mapped_column(Text, nullable=True)

    # Agency / provenance
    agency_id: Mapped[str] = mapped_column(
        String(36), nullable=False, index=True,
        doc="Agency that registered this identity",
    )
    created_by: Mapped[str] = mapped_column(
        String(36), nullable=False,
        doc="User who created this record",
    )

    # Event sourcing version
    version: Mapped[int] = mapped_column(Integer, nullable=False, default=0)

    # Relationships
    documents: Mapped[list[IdentityDocument]] = relationship(
        "IdentityDocument", back_populates="citizen", lazy="selectin"
    )
    biometrics: Mapped[list[BiometricRecord]] = relationship(
        "BiometricRecord", back_populates="citizen", lazy="selectin"
    )

    __table_args__ = (
        Index("ix_citizens_name", "last_name", "first_name"),
        Index("ix_citizens_dob", "date_of_birth"),
        Index("ix_citizens_agency_status", "agency_id", "status"),
    )


class IdentityDocument(Base):
    """Identity documents issued to a citizen."""

    __tablename__ = "identity_documents"

    citizen_id: Mapped[str] = mapped_column(
        String(36), ForeignKey("citizens.id", ondelete="CASCADE"),
        nullable=False, index=True,
    )
    document_type: Mapped[DocumentType] = mapped_column(
        Enum(DocumentType, name="document_type_enum"), nullable=False
    )
    document_number: Mapped[str] = mapped_column(String(50), nullable=False)
    issue_date: Mapped[date] = mapped_column(Date, nullable=False)
    expiry_date: Mapped[date | None] = mapped_column(Date, nullable=True)
    issuing_agency: Mapped[str] = mapped_column(String(100), nullable=False)
    issuing_location: Mapped[str | None] = mapped_column(String(200), nullable=True)
    status: Mapped[DocumentStatus] = mapped_column(
        Enum(DocumentStatus, name="document_status_enum"),
        nullable=False,
        default=DocumentStatus.ACTIVE,
    )
    metadata_json: Mapped[dict[str, Any] | None] = mapped_column(JSONB, nullable=True)

    # Relationships
    citizen: Mapped[Citizen] = relationship("Citizen", back_populates="documents")

    __table_args__ = (
        Index(
            "uq_document_type_number",
            "document_type",
            "document_number",
            unique=True,
        ),
    )


class BiometricRecord(Base):
    """Biometric data records for a citizen."""

    __tablename__ = "biometric_records"

    citizen_id: Mapped[str] = mapped_column(
        String(36), ForeignKey("citizens.id", ondelete="CASCADE"),
        nullable=False, index=True,
    )
    biometric_type: Mapped[BiometricType] = mapped_column(
        Enum(BiometricType, name="biometric_type_enum"), nullable=False
    )
    template_hash: Mapped[str] = mapped_column(
        String(256), nullable=False, unique=True,
        doc="Hash of the biometric template for deduplication",
    )
    quality_score: Mapped[float] = mapped_column(
        Float, nullable=False,
        doc="Quality score 0.0 to 1.0",
    )
    capture_device: Mapped[str | None] = mapped_column(String(100), nullable=True)
    captured_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        nullable=False,
        default=lambda: datetime.now(timezone.utc),
    )
    captured_by: Mapped[str] = mapped_column(String(36), nullable=False)
    is_primary: Mapped[bool] = mapped_column(default=False, nullable=False)

    # Relationships
    citizen: Mapped[Citizen] = relationship("Citizen", back_populates="biometrics")

    __table_args__ = (
        Index("ix_biometric_type_citizen", "biometric_type", "citizen_id"),
    )


class CitizenReadModel(Base):
    """
    Denormalized read model for fast citizen queries.
    Populated from event projections — not written to directly by API.
    """

    __tablename__ = "citizens_read"

    national_id: Mapped[str] = mapped_column(
        String(20), unique=True, nullable=False, index=True
    )
    nif: Mapped[str | None] = mapped_column(String(32), nullable=True, index=True)
    full_name: Mapped[str] = mapped_column(String(300), nullable=False)
    first_name: Mapped[str] = mapped_column(String(100), nullable=False)
    last_name: Mapped[str] = mapped_column(String(100), nullable=False)
    date_of_birth: Mapped[date] = mapped_column(Date, nullable=False)
    gender: Mapped[str] = mapped_column(String(10), nullable=False)
    nationality: Mapped[str] = mapped_column(String(3), nullable=False)
    status: Mapped[str] = mapped_column(String(20), nullable=False, index=True)
    agency_id: Mapped[str] = mapped_column(String(36), nullable=False, index=True)
    document_count: Mapped[int] = mapped_column(Integer, default=0)
    has_biometrics: Mapped[bool] = mapped_column(default=False)
    verified: Mapped[bool] = mapped_column(default=False)
    photo_url: Mapped[str | None] = mapped_column(String(500), nullable=True)
    address_summary: Mapped[str | None] = mapped_column(String(500), nullable=True)
    last_event_at: Mapped[datetime | None] = mapped_column(
        DateTime(timezone=True), nullable=True
    )

    __table_args__ = (
        Index("ix_read_fullname", "full_name"),
        Index("ix_read_name_search", "last_name", "first_name"),
    )
