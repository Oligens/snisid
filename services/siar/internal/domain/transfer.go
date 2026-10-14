package domain

import (
	"time"

	"github.com/google/uuid"
)

type Transfer struct {
	TransferID    uuid.UUID    `json:"transfer_id"`
	FirearmID     uuid.UUID    `json:"firearm_id"`
	FromOwnerID   *uuid.UUID   `json:"from_owner_id,omitempty"`
	FromOwnerName string       `json:"from_owner_name,omitempty"`
	ToOwnerID     *uuid.UUID   `json:"to_owner_id,omitempty"`
	ToOwnerName   string       `json:"to_owner_name,omitempty"`
	TransferType  TransferType `json:"transfer_type"`
	TransferDate  time.Time    `json:"transfer_date"`
	PermitRef     string       `json:"permit_ref,omitempty"`
	AuthorizedBy  *uuid.UUID   `json:"authorized_by,omitempty"`
	Notes         string       `json:"notes,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
}

type CreateTransferRequest struct {
	FirearmID     uuid.UUID    `json:"firearm_id" binding:"required"`
	FromOwnerID   *uuid.UUID   `json:"from_owner_id"`
	FromOwnerName string       `json:"from_owner_name"`
	ToOwnerID     *uuid.UUID   `json:"to_owner_id"`
	ToOwnerName   string       `json:"to_owner_name"`
	TransferType  TransferType `json:"transfer_type" binding:"required"`
	TransferDate  time.Time    `json:"transfer_date" binding:"required"`
	PermitRef     string       `json:"permit_ref"`
	Notes         string       `json:"notes"`
}

type Seizure struct {
	SeizureID      uuid.UUID  `json:"seizure_id"`
	FirearmID      *uuid.UUID `json:"firearm_id,omitempty"`
	SerialNumber   string     `json:"serial_number,omitempty"`
	Make           string     `json:"make,omitempty"`
	Model          string     `json:"model,omitempty"`
	SeizureDate    time.Time  `json:"seizure_date"`
	SeizingUnit    string     `json:"seizing_unit"`
	SeizingOfficer *uuid.UUID `json:"seizing_officer,omitempty"`
	LocationDesc   string     `json:"location_desc,omitempty"`
	DeptCode       string     `json:"dept_code,omitempty"`
	Context        string     `json:"context,omitempty"`
	FromPersonID   *uuid.UUID `json:"from_person_id,omitempty"`
	FromPersonName string     `json:"from_person_name,omitempty"`
	GangID         *uuid.UUID `json:"gang_id,omitempty"`
	CaseReference  string     `json:"case_reference,omitempty"`
	DisposedOf     bool       `json:"disposed_of"`
	DisposalMethod string     `json:"disposal_method,omitempty"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CreateSeizureRequest struct {
	FirearmID      *uuid.UUID `json:"firearm_id"`
	SerialNumber   string     `json:"serial_number"`
	Make           string     `json:"make"`
	Model          string     `json:"model"`
	SeizureDate    time.Time  `json:"seizure_date" binding:"required"`
	SeizingUnit    string     `json:"seizing_unit" binding:"required"`
	SeizingOfficer *uuid.UUID `json:"seizing_officer"`
	LocationDesc   string     `json:"location_desc"`
	DeptCode       string     `json:"dept_code"`
	Context        string     `json:"context"`
	FromPersonID   *uuid.UUID `json:"from_person_id"`
	FromPersonName string     `json:"from_person_name"`
	GangID         *uuid.UUID `json:"gang_id"`
	CaseReference  string     `json:"case_reference"`
}

type ReportStolenRequest struct {
	FirearmID    *uuid.UUID `json:"firearm_id"`
	SerialNumber string     `json:"serial_number"`
	Make         string     `json:"make" binding:"required"`
	Model        string     `json:"model" binding:"required"`
	Caliber      string     `json:"caliber"`
	OwnerID      *uuid.UUID `json:"owner_id"`
	OwnerName    string     `json:"owner_name"`
	DeptCode     string     `json:"dept_code"`
	IncidentDate time.Time  `json:"incident_date" binding:"required"`
	Notes        string     `json:"notes"`
}
