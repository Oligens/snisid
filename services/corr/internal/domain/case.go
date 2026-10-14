package domain

import (
	"time"

	"github.com/google/uuid"
)

type IntegrityCase struct {
	CaseID          uuid.UUID      `json:"case_id"`
	NationalCorrID  string         `json:"national_corr_id"`
	OfficerSnisidID uuid.UUID      `json:"officer_snisid_id"`
	OfficerBadge    string         `json:"officer_badge,omitempty"`
	OfficerUnit     string         `json:"officer_unit,omitempty"`
	OfficerRank     string         `json:"officer_rank,omitempty"`
	AllegationType  AllegationType `json:"allegation_type"`
	Severity        Severity       `json:"severity"`
	Status          CaseStatus     `json:"status"`

	AllegationSummary string     `json:"allegation_summary"`
	IncidentDateFrom  *time.Time `json:"incident_date_from,omitempty"`
	IncidentDateTo    *time.Time `json:"incident_date_to,omitempty"`
	EvidenceRefs      []string   `json:"evidence_refs"`

	GangID           *uuid.UUID  `json:"gang_id,omitempty"`
	GangMemberIDs    []uuid.UUID `json:"gang_member_ids"`
	FinancialGainUSD *float64    `json:"financial_gain_usd,omitempty"`
	BlanCaseID       *uuid.UUID  `json:"blan_case_id,omitempty"`

	ReportedByType         string     `json:"reported_by_type,omitempty"`
	ReportedByID           *uuid.UUID `json:"reported_by_id,omitempty"`
	ReportingDate          time.Time  `json:"reporting_date"`
	IsWhistleblower        bool       `json:"is_whistleblower"`
	WhistleblowerProtected bool       `json:"whistleblower_protected"`

	IgpnInvestigator   *uuid.UUID `json:"igpnh_investigator,omitempty"`
	InvestigationStart *time.Time `json:"investigation_start,omitempty"`
	InvestigationEnd   *time.Time `json:"investigation_end,omitempty"`
	InvestigationNotes string     `json:"investigation_notes,omitempty"`

	SanctionsApplied  string `json:"sanctions_applied,omitempty"`
	ReferredToParquet bool   `json:"referred_to_parquet"`
	ParquetRef        string `json:"parquet_ref,omitempty"`
	ULCCRef           string `json:"ulcc_ref,omitempty"`

	CreatedBy uuid.UUID `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateCaseRequest struct {
	OfficerSnisidID uuid.UUID      `json:"officer_snisid_id" validate:"required"`
	OfficerBadge    string         `json:"officer_badge"`
	OfficerUnit     string         `json:"officer_unit"`
	OfficerRank     string         `json:"officer_rank"`
	AllegationType  AllegationType `json:"allegation_type" validate:"required"`
	Severity        Severity       `json:"severity" validate:"required"`

	AllegationSummary string     `json:"allegation_summary" validate:"required"`
	IncidentDateFrom  *time.Time `json:"incident_date_from"`
	IncidentDateTo    *time.Time `json:"incident_date_to"`

	GangID           *uuid.UUID `json:"gang_id"`
	FinancialGainUSD *float64   `json:"financial_gain_usd"`

	ReportedByType         string `json:"reported_by_type"`
	IsWhistleblower        bool   `json:"is_whistleblower"`
	WhistleblowerProtected bool   `json:"whistleblower_protected"`
}
