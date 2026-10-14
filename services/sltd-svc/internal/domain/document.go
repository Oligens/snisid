package domain

import (
	"time"

	"github.com/google/uuid"
)

type SltdDocument struct {
	DocID              uuid.UUID  `json:"doc_id" db:"doc_id"`
	NationalSltdID     string     `json:"national_sltd_id" db:"national_sltd_id"`
	DocType            DocType    `json:"doc_type" db:"doc_type"`
	DocumentNumber     string     `json:"document_number" db:"document_number"`
	IssuingCountry     string     `json:"issuing_country" db:"issuing_country"`
	HolderName         string     `json:"holder_name" db:"holder_name"`
	HolderSnisidID     *uuid.UUID `json:"holder_snisid_id,omitempty" db:"holder_snisid_id"`
	HolderDOB          *time.Time `json:"holder_dob,omitempty" db:"holder_dob"`
	HolderNationality  string     `json:"holder_nationality" db:"holder_nationality"`
	IssueDate          *time.Time `json:"issue_date,omitempty" db:"issue_date"`
	ExpiryDate         *time.Time `json:"expiry_date,omitempty" db:"expiry_date"`
	Status             DocStatus  `json:"status" db:"status"`
	ReportedDate       time.Time  `json:"reported_date" db:"reported_date"`
	ReportedBy         uuid.UUID  `json:"reported_by" db:"reported_by"`
	ReportingDeptCode  string     `json:"reporting_dept_code" db:"reporting_dept_code"`
	TheftContext       string     `json:"theft_context,omitempty" db:"theft_context"`
	FoundDate          *time.Time `json:"found_date,omitempty" db:"found_date"`
	FoundLocation      string     `json:"found_location,omitempty" db:"found_location"`
	InterpolSltdRef    string     `json:"interpol_sltd_ref,omitempty" db:"interpol_sltd_ref"`
	ReportedToInterpol bool       `json:"reported_to_interpol" db:"reported_to_interpol"`
	InterpolReportedAt *time.Time `json:"interpol_reported_at,omitempty" db:"interpol_reported_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

type SltdCheckLog struct {
	CheckID        uuid.UUID  `json:"check_id" db:"check_id"`
	DocumentNumber string     `json:"document_number" db:"document_number"`
	DocType        DocType    `json:"doc_type,omitempty" db:"doc_type"`
	CheckedBy      uuid.UUID  `json:"checked_by" db:"checked_by"`
	CheckLocation  string     `json:"check_location,omitempty" db:"check_location"`
	PostID         *uuid.UUID `json:"post_id,omitempty" db:"post_id"`
	Result         string     `json:"result" db:"result"`
	Source         string     `json:"source" db:"source"`
	SltdDocID      *uuid.UUID `json:"sltd_doc_id,omitempty" db:"sltd_doc_id"`
	CheckedAt      time.Time  `json:"checked_at" db:"checked_at"`
}

type CheckResult struct {
	IsBlacklisted bool          `json:"is_blacklisted"`
	IsStolen      bool          `json:"is_stolen"`
	IsLost        bool          `json:"is_lost"`
	Document      *SltdDocument `json:"document,omitempty"`
	Message       string        `json:"message"`
}

type ReportLostRequest struct {
	DocumentNumber    string `json:"document_number" binding:"required"`
	DocType           string `json:"doc_type" binding:"required"`
	IssuingCountry    string `json:"issuing_country" binding:"required"`
	HolderName        string `json:"holder_name" binding:"required"`
	HolderSnisidID    string `json:"holder_snisid_id"`
	HolderDOB         string `json:"holder_dob"`
	HolderNationality string `json:"holder_nationality"`
	IssueDate         string `json:"issue_date"`
	ExpiryDate        string `json:"expiry_date"`
	ReportedBy        string `json:"reported_by" binding:"required"`
	ReportingDeptCode string `json:"reporting_dept_code"`
	TheftContext      string `json:"theft_context"`
}

type ReportStolenRequest struct {
	DocumentNumber    string `json:"document_number" binding:"required"`
	DocType           string `json:"doc_type" binding:"required"`
	IssuingCountry    string `json:"issuing_country" binding:"required"`
	HolderName        string `json:"holder_name" binding:"required"`
	HolderSnisidID    string `json:"holder_snisid_id"`
	HolderDOB         string `json:"holder_dob"`
	HolderNationality string `json:"holder_nationality"`
	IssueDate         string `json:"issue_date"`
	ExpiryDate        string `json:"expiry_date"`
	ReportedBy        string `json:"reported_by" binding:"required"`
	ReportingDeptCode string `json:"reporting_dept_code"`
	TheftContext      string `json:"theft_context"`
}

type MarkFoundRequest struct {
	FoundLocation string `json:"found_location" binding:"required"`
	ReportedBy    string `json:"reported_by" binding:"required"`
}

type SLTDStats struct {
	TotalDocuments int            `json:"total_documents"`
	ByStatus       map[string]int `json:"by_status"`
	ByDocType      map[string]int `json:"by_doc_type"`
}

type Repository interface {
	CreateDocument(doc *SltdDocument) error
	FindByNumber(docNumber string, issuingCountry string) (*SltdDocument, error)
	FindByID(docID uuid.UUID) (*SltdDocument, error)
	UpdateDocument(doc *SltdDocument) error
	CreateCheckLog(log *SltdCheckLog) error
	GetStatsByType() (*SLTDStats, error)
}
