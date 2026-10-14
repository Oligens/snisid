package domain

import (
	"time"

	"github.com/google/uuid"
)

type DocType string

const (
	DocNationalID DocType = "NATIONAL_ID"
	DocPassport   DocType = "PASSPORT"
	DocResidence  DocType = "RESIDENCE_PERMIT"
	DocRefugee    DocType = "REFUGEE_DOC"
)

type CardStatus string

const (
	CardIssued  CardStatus = "ISSUED"
	CardActive  CardStatus = "ACTIVE"
	CardExpired CardStatus = "EXPIRED"
	CardRevoked CardStatus = "REVOKED"
	CardLost    CardStatus = "LOST"
	CardStolen  CardStatus = "STOLEN"
	CardRenewed CardStatus = "RENEWED"
)

type CardDocument struct {
	DocumentID              uuid.UUID  `json:"document_id"`
	DocumentNumber          string     `json:"document_number"`
	DocType                 DocType    `json:"doc_type"`
	CitizenID               uuid.UUID  `json:"citizen_id"`
	Status                  CardStatus `json:"status"`
	ChipSerial              *string    `json:"chip_serial,omitempty"`
	MRZLine1                *string    `json:"mrz_line1,omitempty"`
	MRZLine2                *string    `json:"mrz_line2,omitempty"`
	PublicKeyCertRef        *string    `json:"public_key_cert_ref,omitempty"`
	IssueDate               time.Time  `json:"issue_date"`
	ExpiryDate              time.Time  `json:"expiry_date"`
	IssuingOffice           *string    `json:"issuing_office,omitempty"`
	PersonalizationFacility string     `json:"personalization_facility"`
	PhotoRef                *string    `json:"photo_ref,omitempty"`
	SignatureRef            *string    `json:"signature_ref,omitempty"`
	SLTDReported            bool       `json:"sltd_reported"`
	CreatedBy               uuid.UUID  `json:"created_by"`
	CreatedAt               time.Time  `json:"created_at"`
}

type IssueRequest struct {
	DocType       string `json:"doc_type"`
	CitizenID     string `json:"citizen_id"`
	IssuingOffice string `json:"issuing_office,omitempty"`
	CreatedBy     string `json:"created_by"`
	PhotoRef      string `json:"photo_ref,omitempty"`
	SignatureRef  string `json:"signature_ref,omitempty"`
}
