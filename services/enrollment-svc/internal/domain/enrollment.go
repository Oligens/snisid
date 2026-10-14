package domain

import (
	"time"

	"github.com/google/uuid"
)

type IdentityProofingLevel string

const (
	IALNone IdentityProofingLevel = "IAL_NONE"
	IAL1    IdentityProofingLevel = "IAL1"
	IAL2    IdentityProofingLevel = "IAL2"
	IAL3    IdentityProofingLevel = "IAL3"
)

type VerificationMethod string

const (
	VerificationDocument  VerificationMethod = "DOCUMENT"
	VerificationBiometric VerificationMethod = "BIOMETRIC"
	VerificationInPerson  VerificationMethod = "IN_PERSON"
	VerificationRemote    VerificationMethod = "REMOTE"
	VerificationKBA       VerificationMethod = "KNOWLEDGE_BASED"
)

type DocumentType string

const (
	DocPassport         DocumentType = "PASSPORT"
	DocNationalID       DocumentType = "NATIONAL_ID"
	DocDriversLicense   DocumentType = "DRIVERS_LICENSE"
	DocBirthCertificate DocumentType = "BIRTH_CERTIFICATE"
	DocResidencePermit  DocumentType = "RESIDENCE_PERMIT"
)

type EnrollmentStatus string

const (
	StatusDraft              EnrollmentStatus = "DRAFT"
	StatusPendingDocuments   EnrollmentStatus = "PENDING_DOCUMENTS"
	StatusDocumentsReceived  EnrollmentStatus = "DOCUMENTS_RECEIVED"
	StatusPendingBiometrics  EnrollmentStatus = "PENDING_BIOMETRICS"
	StatusBiometricsCaptured EnrollmentStatus = "BIOMETRICS_CAPTURED"
	StatusPendingReview      EnrollmentStatus = "PENDING_REVIEW"
	StatusApproved           EnrollmentStatus = "APPROVED"
	StatusRejected           EnrollmentStatus = "REJECTED"
	StatusExpired            EnrollmentStatus = "EXPIRED"
)

type EnrollmentRequest struct {
	RequestID       uuid.UUID             `json:"request_id"`
	CitizenID       *uuid.UUID            `json:"citizen_id,omitempty"`
	FullName        string                `json:"full_name"`
	DateOfBirth     string                `json:"date_of_birth"`
	Nationality     string                `json:"nationality"`
	Email           string                `json:"email"`
	Phone           string                `json:"phone"`
	ProofingLevel   IdentityProofingLevel `json:"proofing_level"`
	Status          EnrollmentStatus      `json:"status"`
	SubmittedAt     time.Time             `json:"submitted_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	AssignedOfficer *string               `json:"assigned_officer,omitempty"`
	Remarks         string                `json:"remarks,omitempty"`
}

type IdentityDocument struct {
	DocID       uuid.UUID    `json:"doc_id"`
	RequestID   uuid.UUID    `json:"request_id"`
	DocType     DocumentType `json:"doc_type"`
	DocNumber   string       `json:"doc_number"`
	IssuingAuth string       `json:"issuing_authority"`
	IssueDate   string       `json:"issue_date"`
	ExpiryDate  string       `json:"expiry_date"`
	FrontImage  string       `json:"front_image,omitempty"`
	BackImage   string       `json:"back_image,omitempty"`
	IsVerified  bool         `json:"is_verified"`
	VerifiedAt  *time.Time   `json:"verified_at,omitempty"`
	UploadedAt  time.Time    `json:"uploaded_at"`
}

type BiometricSample struct {
	SampleID   uuid.UUID `json:"sample_id"`
	RequestID  uuid.UUID `json:"request_id"`
	SampleType string    `json:"sample_type"`
	Format     string    `json:"format"`
	Data       string    `json:"data,omitempty"`
	Quality    float64   `json:"quality"`
	CapturedAt time.Time `json:"captured_at"`
	DeviceID   string    `json:"device_id,omitempty"`
	OperatorID string    `json:"operator_id,omitempty"`
}

type EnrollmentReview struct {
	ReviewID      uuid.UUID             `json:"review_id"`
	RequestID     uuid.UUID             `json:"request_id"`
	OfficerID     string                `json:"officer_id"`
	OfficerName   string                `json:"officer_name"`
	Decision      string                `json:"decision"`
	Reason        string                `json:"reason,omitempty"`
	VerifiedLevel IdentityProofingLevel `json:"verified_level"`
	ReviewedAt    time.Time             `json:"reviewed_at"`
}
