package domain

import (
	"time"

	"github.com/google/uuid"
)

type Evidence struct {
	EvidenceID     uuid.UUID  `json:"evidence_id"`
	CaseID         uuid.UUID  `json:"case_id"`
	EvidenceType   string     `json:"evidence_type"`
	Description    string     `json:"description"`
	FileHash       string     `json:"file_hash,omitempty"`
	StorageRef     string     `json:"storage_ref,omitempty"`
	CollectedBy    *uuid.UUID `json:"collected_by,omitempty"`
	CollectedAt    time.Time  `json:"collected_at"`
	ChainOfCustody []string   `json:"chain_of_custody"`
	IsVerified     bool       `json:"is_verified"`
	VerifiedBy     *uuid.UUID `json:"verified_by,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CreateEvidenceRequest struct {
	CaseID       uuid.UUID `json:"case_id" validate:"required"`
	EvidenceType string    `json:"evidence_type" validate:"required"`
	Description  string    `json:"description" validate:"required"`
	FileHash     string    `json:"file_hash"`
	StorageRef   string    `json:"storage_ref"`
}
