package domain

import (
	"time"

	"github.com/google/uuid"
)

type Dealer struct {
	DealerID           uuid.UUID    `json:"dealer_id"`
	DealerLicenseNo    string       `json:"dealer_license_no"`
	BusinessName       string       `json:"business_name"`
	BusinessRegNo      string       `json:"business_reg_no,omitempty"`
	OwnerSnisidID      uuid.UUID    `json:"owner_snisid_id"`
	OwnerName          string       `json:"owner_name"`
	Address            string       `json:"address,omitempty"`
	DeptCode           string       `json:"dept_code,omitempty"`
	Commune            string       `json:"commune,omitempty"`
	Phone              string       `json:"phone,omitempty"`
	Email              string       `json:"email,omitempty"`
	LicenseType        LicenseType  `json:"license_type"`
	Status             DealerStatus `json:"status"`
	LicenseIssueDate   time.Time    `json:"license_issue_date"`
	LicenseExpiryDate  time.Time    `json:"license_expiry_date"`
	PremisesInspected  bool         `json:"premises_inspected"`
	LastInspectionDate *time.Time   `json:"last_inspection_date,omitempty"`
	Notes              string       `json:"notes,omitempty"`
	CreatedBy          uuid.UUID    `json:"created_by"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
}

type CreateDealerRequest struct {
	BusinessName      string      `json:"business_name" binding:"required"`
	BusinessRegNo     string      `json:"business_reg_no"`
	OwnerSnisidID     uuid.UUID   `json:"owner_snisid_id" binding:"required"`
	OwnerName         string      `json:"owner_name" binding:"required"`
	Address           string      `json:"address"`
	DeptCode          string      `json:"dept_code"`
	Commune           string      `json:"commune"`
	Phone             string      `json:"phone"`
	Email             string      `json:"email"`
	LicenseType       LicenseType `json:"license_type"`
	LicenseIssueDate  time.Time   `json:"license_issue_date" binding:"required"`
	LicenseExpiryDate time.Time   `json:"license_expiry_date" binding:"required"`
	PremisesInspected bool        `json:"premises_inspected"`
	Notes             string      `json:"notes"`
}
