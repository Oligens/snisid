package domain

import (
	"time"

	"github.com/google/uuid"
)

type WarrantType string

const (
	WarrantTitleIIIPhone    WarrantType = "TITLE_III_PHONE"
	WarrantTitleIIIInternet WarrantType = "TITLE_III_INTERNET"
	WarrantFISAElectronic   WarrantType = "FISA_ELECTRONIC"
	WarrantFISAPhysical     WarrantType = "FISA_PHYSICAL"
	WarrantPenRegister      WarrantType = "PEN_REGISTER"
	WarrantTrapTrace        WarrantType = "TRAP_TRACE"
)

type WarrantStatus string

const (
	WarrantDraft    WarrantStatus = "DRAFT"
	WarrantPending  WarrantStatus = "PENDING"
	WarrantApproved WarrantStatus = "APPROVED"
	WarrantActive   WarrantStatus = "ACTIVE"
	WarrantExpired  WarrantStatus = "EXPIRED"
	WarrantRevoked  WarrantStatus = "REVOKED"
)

type SurveillanceWarrant struct {
	ID                   uuid.UUID     `json:"id"`
	WarrantID            string        `json:"warrant_id"`
	WarrantType          WarrantType   `json:"warrant_type"`
	TargetIdentity       string        `json:"target_identity"`
	TargetDetails        *string       `json:"target_details,omitempty"`
	IssuingCourt         string        `json:"issuing_court"`
	JudgeName            string        `json:"judge_name"`
	ApplicantAgency      string        `json:"applicant_agency"`
	ApplicantOfficer     uuid.UUID     `json:"applicant_officer"`
	ProbableCauseSummary *string       `json:"probable_cause_summary,omitempty"`
	DurationDays         int           `json:"duration_days"`
	AuthorizedStart      *time.Time    `json:"authorized_start,omitempty"`
	AuthorizedEnd        *time.Time    `json:"authorized_end,omitempty"`
	Renewals             int           `json:"renewals"`
	Status               WarrantStatus `json:"status"`
	ReviewRequiredAt     *time.Time    `json:"review_required_at,omitempty"`
	EmergencyAuthorized  bool          `json:"emergency_authorized"`
	EmergencyApprovedBy  *uuid.UUID    `json:"emergency_approved_by,omitempty"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
}

type SurveillanceReport struct {
	ID                        uuid.UUID `json:"id"`
	WarrantID                 uuid.UUID `json:"warrant_id"`
	ReportingPeriodStart      time.Time `json:"reporting_period_start"`
	ReportingPeriodEnd        time.Time `json:"reporting_period_end"`
	CommunicationsIntercepted int       `json:"communications_intercepted"`
	MinimizationApplied       bool      `json:"minimization_applied"`
	IncidentalCollection      int       `json:"incidental_collection"`
	USPersonIdentities        int       `json:"us_person_identities"`
	ResultsSummary            *string   `json:"results_summary,omitempty"`
	SubmittedBy               uuid.UUID `json:"submitted_by"`
	SubmittedAt               time.Time `json:"submitted_at"`
}

type FISADocket struct {
	ID                   uuid.UUID  `json:"id"`
	DocketNumber         string     `json:"docket_number"`
	CourtTerm            string     `json:"court_term"`
	JudgePresiding       string     `json:"judge_presiding"`
	ApplicationsFiled    int        `json:"applications_filed"`
	ApplicationsApproved int        `json:"applications_approved"`
	ApplicationsModified int        `json:"applications_modified"`
	ApplicationsDenied   int        `json:"applications_denied"`
	TotalTargets         int        `json:"total_targets"`
	ForeignTargets       int        `json:"foreign_targets"`
	USPersonTargets      int        `json:"us_person_targets"`
	SealedUntil          *time.Time `json:"sealed_until,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

type FileWarrantRequest struct {
	WarrantType          string `json:"warrant_type" binding:"required"`
	TargetIdentity       string `json:"target_identity" binding:"required"`
	TargetDetails        string `json:"target_details"`
	IssuingCourt         string `json:"issuing_court" binding:"required"`
	JudgeName            string `json:"judge_name" binding:"required"`
	ApplicantAgency      string `json:"applicant_agency" binding:"required"`
	ApplicantOfficer     string `json:"applicant_officer" binding:"required"`
	ProbableCauseSummary string `json:"probable_cause_summary"`
	DurationDays         int    `json:"duration_days" binding:"required"`
}

type ApproveWarrantRequest struct {
	JudgeName string `json:"judge_name" binding:"required"`
}

type RenewWarrantRequest struct {
	DurationDays int `json:"duration_days" binding:"required"`
}

type FileReportRequest struct {
	WarrantID                 string `json:"warrant_id" binding:"required"`
	ReportingPeriodStart      string `json:"reporting_period_start" binding:"required"`
	ReportingPeriodEnd        string `json:"reporting_period_end" binding:"required"`
	CommunicationsIntercepted int    `json:"communications_intercepted"`
	MinimizationApplied       bool   `json:"minimization_applied"`
	IncidentalCollection      int    `json:"incidental_collection"`
	USPersonIdentities        int    `json:"us_person_identities"`
	ResultsSummary            string `json:"results_summary"`
	SubmittedBy               string `json:"submitted_by" binding:"required"`
}

type EmergencyAuthorizationRequest struct {
	WarrantType      string `json:"warrant_type" binding:"required"`
	TargetIdentity   string `json:"target_identity" binding:"required"`
	TargetDetails    string `json:"target_details"`
	ApplicantAgency  string `json:"applicant_agency" binding:"required"`
	ApplicantOfficer string `json:"applicant_officer" binding:"required"`
	ProbableCause    string `json:"probable_cause" binding:"required"`
	ApprovedBy       string `json:"approved_by" binding:"required"`
}
