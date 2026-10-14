package domain

import (
	"time"

	"github.com/google/uuid"
)

type InvType string

const (
	InvStandard      InvType = "STANDARD"
	InvEnhanced      InvType = "ENHANCED"
	InvTopSecret     InvType = "TOP_SECRET"
	InvReinvestigate InvType = "REINVESTIGATION"
)

type InvStatus string

const (
	InvPending     InvStatus = "PENDING"
	InvInProgress  InvStatus = "IN_PROGRESS"
	InvFavorable   InvStatus = "FAVORABLE"
	InvUnfavorable InvStatus = "UNFAVORABLE"
)

type ClearanceLevel string

const (
	ClearUnclassified ClearanceLevel = "UNCLASSIFIED"
	ClearConfidential ClearanceLevel = "CONFIDENTIAL"
	ClearSecret       ClearanceLevel = "SECRET"
	ClearTopSecret    ClearanceLevel = "TOP_SECRET"
)

type AlertType string

const (
	AlertUnauthorizedAccess AlertType = "UNAUTHORIZED_ACCESS"
	AlertDataExfil          AlertType = "DATA_EXFIL"
	AlertPrivEscalation     AlertType = "PRIVILEGE_ESCALATION"
	AlertBehavioral         AlertType = "BEHAVIORAL"
	AlertCollusion          AlertType = "COLLUSION"
)

type Severity string

const (
	SevLow      Severity = "LOW"
	SevMedium   Severity = "MEDIUM"
	SevHigh     Severity = "HIGH"
	SevCritical Severity = "CRITICAL"
)

type ThreatStatus string

const (
	ThreatOpen          ThreatStatus = "OPEN"
	ThreatInvestigating ThreatStatus = "INVESTIGATING"
	ThreatConfirmed     ThreatStatus = "CONFIRMED"
	ThreatFalsePositive ThreatStatus = "FALSE_POSITIVE"
	ThreatMitigated     ThreatStatus = "MITIGATED"
)

type RelationshipType string

const (
	RelDiplomatic RelationshipType = "DIPLOMATIC"
	RelBusiness   RelationshipType = "BUSINESS"
	RelAcademic   RelationshipType = "ACADEMIC"
	RelFamily     RelationshipType = "FAMILY"
	RelPersonal   RelationshipType = "PERSONAL"
)

type BackgroundInvestigation struct {
	ID                    uuid.UUID       `json:"id"`
	SubjectIdentityRef    string          `json:"subject_identity_ref"`
	InvestigationType     InvType         `json:"investigation_type"`
	Status                InvStatus       `json:"status"`
	CriminalRecordCheck   bool            `json:"criminal_record_check"`
	FinancialCheck        bool            `json:"financial_check"`
	ForeignContactsCheck  bool            `json:"foreign_contacts_check"`
	SocialMediaCheck      bool            `json:"social_media_check"`
	DrugTest              bool            `json:"drug_test"`
	PsychEval             bool            `json:"psych_eval"`
	Adjudicator           *uuid.UUID      `json:"adjudicator,omitempty"`
	AdjudicationNotes     *string         `json:"adjudication_notes,omitempty"`
	CompletedAt           *time.Time      `json:"completed_at,omitempty"`
	ClearanceLevelGranted *ClearanceLevel `json:"clearance_level_granted,omitempty"`
	ExpiresAt             *time.Time      `json:"expires_at,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type InsiderThreatAlert struct {
	ID               uuid.UUID    `json:"id"`
	SubjectID        string       `json:"subject_id"`
	AlertType        AlertType    `json:"alert_type"`
	Severity         Severity     `json:"severity"`
	Description      string       `json:"description"`
	EvidenceRefs     []string     `json:"evidence_refs"`
	DetectedBy       string       `json:"detected_by"`
	Status           ThreatStatus `json:"status"`
	InvestigationRef *uuid.UUID   `json:"investigation_ref,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
}

type ForeignContact struct {
	ID                uuid.UUID        `json:"id"`
	SubjectID         string           `json:"subject_id"`
	ContactName       string           `json:"contact_name"`
	ForeignGovernment string           `json:"foreign_government"`
	RelationshipType  RelationshipType `json:"relationship_type"`
	LastContactAt     *time.Time       `json:"last_contact_at,omitempty"`
	Frequency         *string          `json:"frequency,omitempty"`
	ApprovedBy        *uuid.UUID       `json:"approved_by,omitempty"`
	Notes             *string          `json:"notes,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
}

type CreateInvestigationRequest struct {
	SubjectIdentityRef   string `json:"subject_identity_ref" binding:"required"`
	InvestigationType    string `json:"investigation_type" binding:"required"`
	CriminalRecordCheck  bool   `json:"criminal_record_check"`
	FinancialCheck       bool   `json:"financial_check"`
	ForeignContactsCheck bool   `json:"foreign_contacts_check"`
	SocialMediaCheck     bool   `json:"social_media_check"`
	DrugTest             bool   `json:"drug_test"`
	PsychEval            bool   `json:"psych_eval"`
}

type AdjudicateRequest struct {
	Adjudicator           uuid.UUID `json:"adjudicator" binding:"required"`
	AdjudicationNotes     string    `json:"adjudication_notes"`
	ClearanceLevelGranted string    `json:"clearance_level_granted" binding:"required"`
}

type ReportThreatRequest struct {
	SubjectID    string   `json:"subject_id" binding:"required"`
	AlertType    string   `json:"alert_type" binding:"required"`
	Severity     string   `json:"severity" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	EvidenceRefs []string `json:"evidence_refs"`
	DetectedBy   string   `json:"detected_by" binding:"required"`
}

type ReportContactRequest struct {
	SubjectID         string `json:"subject_id" binding:"required"`
	ContactName       string `json:"contact_name" binding:"required"`
	ForeignGovernment string `json:"foreign_government" binding:"required"`
	RelationshipType  string `json:"relationship_type" binding:"required"`
	LastContactAt     string `json:"last_contact_at"`
	Frequency         string `json:"frequency"`
	ApprovedBy        string `json:"approved_by"`
	Notes             string `json:"notes"`
}
