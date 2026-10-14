package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type STRReport struct {
	StrID                uuid.UUID  `json:"str_id" db:"str_id"`
	NationalStrID        string     `json:"national_str_id" db:"national_str_id"`
	ReportType           ReportType `json:"report_type" db:"report_type"`
	Status               STRStatus  `json:"status" db:"status"`
	ReportingInstitution string     `json:"reporting_institution" db:"reporting_institution"`
	InstitutionType      string     `json:"institution_type,omitempty" db:"institution_type"`
	ReportDate           time.Time  `json:"report_date" db:"report_date"`
	TransactionDate      *time.Time `json:"transaction_date,omitempty" db:"transaction_date"`
	TransactionAmount    float64    `json:"transaction_amount" db:"transaction_amount"`
	TransactionCurrency  string     `json:"transaction_currency" db:"transaction_currency"`
	TransactionAmountUSD float64    `json:"transaction_amount_usd" db:"transaction_amount_usd"`
	SubjectSnisidIDs     []string   `json:"subject_snisid_ids" db:"subject_snisid_ids"`
	SubjectNames         []string   `json:"subject_names" db:"subject_names"`
	SubjectAccounts      []string   `json:"subject_accounts" db:"subject_accounts"`
	SuspiciousActivity   string     `json:"suspicious_activity" db:"suspicious_activity"`
	MLTypology           string     `json:"ml_typology,omitempty" db:"ml_typology"`
	PredicateCrime       string     `json:"predicate_crime,omitempty" db:"predicate_crime"`
	GangID               *uuid.UUID `json:"gang_id,omitempty" db:"gang_id"`
	FPRPersonIDs         []string   `json:"fpr_person_ids" db:"fpr_person_ids"`
	SancMatchIDs         []string   `json:"sanc_match_ids" db:"sanc_match_ids"`
	AnalystID            *uuid.UUID `json:"analyst_id,omitempty" db:"analyst_id"`
	AnalysisNotes        string     `json:"analysis_notes,omitempty" db:"analysis_notes"`
	DisseminatedTo       []string   `json:"disseminated_to" db:"disseminated_to"`
	DisseminatedAt       *time.Time `json:"disseminated_at,omitempty" db:"disseminated_at"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

type FinancialProfile struct {
	ProfileID                 uuid.UUID       `json:"profile_id" db:"profile_id"`
	SNISIDPersonID            uuid.UUID       `json:"snisid_person_id" db:"snisid_person_id"`
	TotalSTRCount             int             `json:"total_str_count" db:"total_str_count"`
	TotalCTRCount             int             `json:"total_ctr_count" db:"total_ctr_count"`
	EstimatedIllegalAssetsUSD float64         `json:"estimated_illegal_assets_usd" db:"estimated_illegal_assets_usd"`
	KnownAccounts             json.RawMessage `json:"known_accounts" db:"known_accounts"`
	KnownProperties           json.RawMessage `json:"known_properties" db:"known_properties"`
	KnownBusinesses           []string        `json:"known_businesses" db:"known_businesses"`
	MLRiskScore               int16           `json:"ml_risk_score" db:"ml_risk_score"`
	IsPEP                     bool            `json:"is_pep" db:"is_pep"`
	LastUpdated               time.Time       `json:"last_updated" db:"last_updated"`
}

type MonCashPattern struct {
	PatternID        uuid.UUID  `json:"pattern_id" db:"pattern_id"`
	STRID            uuid.UUID  `json:"str_id" db:"str_id"`
	PhoneNumber      string     `json:"phone_number" db:"phone_number"`
	SNISIDPersonID   *uuid.UUID `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	PatternType      string     `json:"pattern_type,omitempty" db:"pattern_type"`
	TransactionCount int        `json:"transaction_count" db:"transaction_count"`
	TotalAmountHTG   float64    `json:"total_amount_htg" db:"total_amount_htg"`
	PeriodStart      *time.Time `json:"period_start,omitempty" db:"period_start"`
	PeriodEnd        *time.Time `json:"period_end,omitempty" db:"period_end"`
	Notes            string     `json:"notes,omitempty" db:"notes"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

type SubmitSTRRequest struct {
	ReportType           ReportType `json:"report_type" binding:"required"`
	ReportingInstitution string     `json:"reporting_institution" binding:"required"`
	InstitutionType      string     `json:"institution_type,omitempty"`
	ReportDate           *time.Time `json:"report_date,omitempty"`
	TransactionDate      *time.Time `json:"transaction_date,omitempty"`
	TransactionAmount    float64    `json:"transaction_amount"`
	TransactionCurrency  string     `json:"transaction_currency,omitempty"`
	TransactionAmountUSD float64    `json:"transaction_amount_usd"`
	SubjectSnisidIDs     []string   `json:"subject_snisid_ids,omitempty"`
	SubjectNames         []string   `json:"subject_names,omitempty"`
	SubjectAccounts      []string   `json:"subject_accounts,omitempty"`
	SuspiciousActivity   string     `json:"suspicious_activity" binding:"required"`
	MLTypology           string     `json:"ml_typology,omitempty"`
	PredicateCrime       string     `json:"predicate_crime,omitempty"`
	GangID               *uuid.UUID `json:"gang_id,omitempty"`
	FPRPersonIDs         []string   `json:"fpr_person_ids,omitempty"`
	SancMatchIDs         []string   `json:"sanc_match_ids,omitempty"`
}

type DisseminateRequest struct {
	DisseminatedTo []string `json:"disseminated_to" binding:"required"`
}

type Repository interface {
	CreateSTR(report *STRReport) error
	FindByID(id uuid.UUID) (*STRReport, error)
	GetFinancialProfile(personID uuid.UUID) (*FinancialProfile, error)
	CreateMonCashPattern(pattern *MonCashPattern) error
	GetUnanalyzedSTRs() ([]STRReport, error)
	DisseminateSTR(id uuid.UUID, disseminatedTo []string) error
	GetGangFinances(gangID uuid.UUID) ([]FinancialProfile, error)
	GetNextSequence(year string) (int, error)
}
