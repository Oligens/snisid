package domain

import (
	"time"

	"github.com/google/uuid"
)

type WhistleblowerReport struct {
	ReportID            uuid.UUID      `json:"report_id"`
	ReportToken         string         `json:"report_token"`
	AllegationType      AllegationType `json:"allegation_type"`
	SeverityEstimate    *Severity      `json:"severity_estimate,omitempty"`
	OfficerUnitHint     string         `json:"officer_unit_hint,omitempty"`
	OfficerRankHint     string         `json:"officer_rank_hint,omitempty"`
	Description         string         `json:"description"`
	EvidenceDescription string         `json:"evidence_description,omitempty"`
	SubmissionDate      time.Time      `json:"submission_date"`
	IPHash              string         `json:"ip_hash,omitempty"`
	Processed           bool           `json:"processed"`
	ProcessedBy         *uuid.UUID     `json:"processed_by,omitempty"`
	IntegrityCaseID     *uuid.UUID     `json:"integrity_case_id,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
}

type CreateWhistleblowerRequest struct {
	AllegationType      AllegationType `json:"allegation_type" validate:"required"`
	SeverityEstimate    *Severity      `json:"severity_estimate"`
	OfficerUnitHint     string         `json:"officer_unit_hint"`
	OfficerRankHint     string         `json:"officer_rank_hint"`
	Description         string         `json:"description" validate:"required"`
	EvidenceDescription string         `json:"evidence_description"`
}
