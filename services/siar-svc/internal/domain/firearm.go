package domain

import (
	"time"

	"github.com/google/uuid"
)

type Firearm struct {
	ID                 uuid.UUID        `json:"firearm_id" db:"firearm_id"`
	NationalSiarID     string           `json:"national_siar_id" db:"national_siar_id"`
	SerialNumber       *string          `json:"serial_number,omitempty" db:"serial_number"`
	Make               string           `json:"make" db:"make"`
	Model              string           `json:"model" db:"model"`
	Caliber            string           `json:"caliber" db:"caliber"`
	WeaponType         WeaponType       `json:"weapon_type" db:"weapon_type"`
	ManufactureYear    *int16           `json:"manufacture_year,omitempty" db:"manufacture_year"`
	ManufactureCountry *string          `json:"manufacture_country,omitempty" db:"manufacture_country"`
	Status             Status           `json:"status" db:"status"`
	RegType            RegistrationType `json:"reg_type" db:"reg_type"`
	OwnerSnisidID      *uuid.UUID       `json:"owner_snisid_id,omitempty" db:"owner_snisid_id"`
	OwnerEntityName    *string          `json:"owner_entity_name,omitempty" db:"owner_entity_name"`
	LicenseNumber      *string          `json:"license_number,omitempty" db:"license_number"`
	LicenseExpiry      *time.Time       `json:"license_expiry,omitempty" db:"license_expiry"`
	ImportDate         *time.Time       `json:"import_date,omitempty" db:"import_date"`
	ImportCountry      *string          `json:"import_country,omitempty" db:"import_country"`
	ImportPermitRef    *string          `json:"import_permit_ref,omitempty" db:"import_permit_ref"`
	ImporterName       *string          `json:"importer_name,omitempty" db:"importer_name"`
	CustomsEntryRef    *string          `json:"customs_entry_ref,omitempty" db:"customs_entry_ref"`
	CurrentDeptCode    *string          `json:"current_dept_code,omitempty" db:"current_dept_code"`
	StorageLocation    *string          `json:"storage_location,omitempty" db:"storage_location"`
	FirRecordID        *uuid.UUID       `json:"fir_record_id,omitempty" db:"fir_record_id"`
	GangID             *uuid.UUID       `json:"gang_id,omitempty" db:"gang_id"`
	CaseReferences     []string         `json:"case_references,omitempty" db:"case_references"`
	IarmsRef           *string          `json:"iarms_ref,omitempty" db:"iarms_ref"`
	AtfEtraceRef       *string          `json:"atf_etrace_ref,omitempty" db:"atf_etrace_ref"`
	Notes              *string          `json:"notes,omitempty" db:"notes"`
	CreatedBy          uuid.UUID        `json:"created_by" db:"created_by"`
	CreatedAt          time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at" db:"updated_at"`
}

type License struct {
	ID                 uuid.UUID  `json:"license_id" db:"license_id"`
	LicenseNumber      string     `json:"license_number" db:"license_number"`
	HolderSnisidID     uuid.UUID  `json:"holder_snisid_id" db:"holder_snisid_id"`
	LicenseType        string     `json:"license_type" db:"license_type"`
	FirearmsAuthorized int        `json:"firearms_authorized" db:"firearms_authorized"`
	IssueDate          time.Time  `json:"issue_date" db:"issue_date"`
	ExpiryDate         time.Time  `json:"expiry_date" db:"expiry_date"`
	IssuingAuthority   string     `json:"issuing_authority" db:"issuing_authority"`
	IsActive           bool       `json:"is_active" db:"is_active"`
	RevocationReason   *string    `json:"revocation_reason,omitempty" db:"revocation_reason"`
	RevokedAt          *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

type Transfer struct {
	ID           uuid.UUID  `json:"transfer_id" db:"transfer_id"`
	FirearmID    uuid.UUID  `json:"firearm_id" db:"firearm_id"`
	FromOwnerID  *uuid.UUID `json:"from_owner_id,omitempty" db:"from_owner_id"`
	ToOwnerID    *uuid.UUID `json:"to_owner_id,omitempty" db:"to_owner_id"`
	TransferType *string    `json:"transfer_type,omitempty" db:"transfer_type"`
	TransferDate time.Time  `json:"transfer_date" db:"transfer_date"`
	PermitRef    *string    `json:"permit_ref,omitempty" db:"permit_ref"`
	AuthorizedBy *uuid.UUID `json:"authorized_by,omitempty" db:"authorized_by"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type Seizure struct {
	ID             uuid.UUID  `json:"seizure_id" db:"seizure_id"`
	FirearmID      *uuid.UUID `json:"firearm_id,omitempty" db:"firearm_id"`
	SeizureDate    time.Time  `json:"seizure_date" db:"seizure_date"`
	SeizingUnit    string     `json:"seizing_unit" db:"seizing_unit"`
	SeizingOfficer *uuid.UUID `json:"seizing_officer,omitempty" db:"seizing_officer"`
	LocationDesc   *string    `json:"location_desc,omitempty" db:"location_desc"`
	DeptCode       *string    `json:"dept_code,omitempty" db:"dept_code"`
	Context        *string    `json:"context,omitempty" db:"context"`
	FromPersonID   *uuid.UUID `json:"from_person_id,omitempty" db:"from_person_id"`
	GangID         *uuid.UUID `json:"gang_id,omitempty" db:"gang_id"`
	CaseReference  *string    `json:"case_reference,omitempty" db:"case_reference"`
	DisposedOf     bool       `json:"disposed_of" db:"disposed_of"`
	DisposalMethod *string    `json:"disposal_method,omitempty" db:"disposal_method"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type RegisterFirearmRequest struct {
	SerialNumber       *string          `json:"serial_number"`
	Make               string           `json:"make" binding:"required"`
	Model              string           `json:"model" binding:"required"`
	Caliber            string           `json:"caliber" binding:"required"`
	WeaponType         WeaponType       `json:"weapon_type" binding:"required"`
	ManufactureYear    *int16           `json:"manufacture_year"`
	ManufactureCountry *string          `json:"manufacture_country"`
	RegType            RegistrationType `json:"reg_type" binding:"required"`
	OwnerSnisidID      *uuid.UUID       `json:"owner_snisid_id"`
	OwnerEntityName    *string          `json:"owner_entity_name"`
	LicenseNumber      *string          `json:"license_number"`
	LicenseExpiry      *time.Time       `json:"license_expiry"`
	ImportDate         *time.Time       `json:"import_date"`
	ImportCountry      *string          `json:"import_country"`
	ImportPermitRef    *string          `json:"import_permit_ref"`
	ImporterName       *string          `json:"importer_name"`
	CustomsEntryRef    *string          `json:"customs_entry_ref"`
	CurrentDeptCode    *string          `json:"current_dept_code"`
	StorageLocation    *string          `json:"storage_location"`
	GangID             *uuid.UUID       `json:"gang_id"`
	Notes              *string          `json:"notes"`
	CreatedBy          uuid.UUID        `json:"created_by" binding:"required"`
}

type SeizureRequest struct {
	SerialNumber   *string          `json:"serial_number"`
	Make           string           `json:"make" binding:"required"`
	Model          string           `json:"model" binding:"required"`
	Caliber        string           `json:"caliber" binding:"required"`
	WeaponType     WeaponType       `json:"weapon_type" binding:"required"`
	RegType        RegistrationType `json:"reg_type" binding:"required"`
	SeizingUnit    string           `json:"seizing_unit" binding:"required"`
	SeizingOfficer *uuid.UUID       `json:"seizing_officer"`
	LocationDesc   *string          `json:"location_desc"`
	DeptCode       *string          `json:"dept_code"`
	Context        *string          `json:"context"`
	FromPersonID   *uuid.UUID       `json:"from_person_id"`
	GangID         *uuid.UUID       `json:"gang_id"`
	CaseReference  *string          `json:"case_reference"`
	CreatedBy      uuid.UUID        `json:"created_by" binding:"required"`
}

type StolenRequest struct {
	FirearmID  uuid.UUID  `json:"firearm_id" binding:"required"`
	ReportedBy uuid.UUID  `json:"reported_by" binding:"required"`
	ReportDate *time.Time `json:"report_date"`
	Notes      *string    `json:"notes"`
}

type CreateLicenseRequest struct {
	LicenseNumber      string    `json:"license_number" binding:"required"`
	HolderSnisidID     uuid.UUID `json:"holder_snisid_id" binding:"required"`
	LicenseType        string    `json:"license_type" binding:"required"`
	FirearmsAuthorized int       `json:"firearms_authorized"`
	IssueDate          time.Time `json:"issue_date" binding:"required"`
	ExpiryDate         time.Time `json:"expiry_date" binding:"required"`
	IssuingAuthority   string    `json:"issuing_authority" binding:"required"`
}

type StatsByType struct {
	WeaponType WeaponType `json:"weapon_type" db:"weapon_type"`
	Count      int        `json:"count" db:"count"`
}

func NewSeizure(req *SeizureRequest, firearmID *uuid.UUID) *Seizure {
	now := time.Now()
	return &Seizure{
		FirearmID:      firearmID,
		SeizureDate:    now,
		SeizingUnit:    req.SeizingUnit,
		SeizingOfficer: req.SeizingOfficer,
		LocationDesc:   req.LocationDesc,
		DeptCode:       req.DeptCode,
		Context:        req.Context,
		FromPersonID:   req.FromPersonID,
		GangID:         req.GangID,
		CaseReference:  req.CaseReference,
		CreatedAt:      now,
	}
}
