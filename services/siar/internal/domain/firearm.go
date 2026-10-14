package domain

import (
	"time"

	"github.com/google/uuid"
)

type Firearm struct {
	FirearmID          uuid.UUID        `json:"firearm_id"`
	NationalSiarID     string           `json:"national_siar_id"`
	SerialNumber       string           `json:"serial_number,omitempty"`
	Make               string           `json:"make"`
	Model              string           `json:"model"`
	Caliber            string           `json:"caliber"`
	WeaponType         WeaponType       `json:"weapon_type"`
	ManufactureYear    *int16           `json:"manufacture_year,omitempty"`
	ManufactureCountry string           `json:"manufacture_country,omitempty"`
	Status             FirearmStatus    `json:"status"`
	RegType            RegistrationType `json:"reg_type"`

	OwnerSnisidID *uuid.UUID `json:"owner_snisid_id,omitempty"`
	OwnerEntity   string     `json:"owner_entity_name,omitempty"`
	LicenseNumber string     `json:"license_number,omitempty"`
	LicenseExpiry *time.Time `json:"license_expiry,omitempty"`

	ImportDate      *time.Time `json:"import_date,omitempty"`
	ImportCountry   string     `json:"import_country,omitempty"`
	ImportPermitRef string     `json:"import_permit_ref,omitempty"`
	ImporterName    string     `json:"importer_name,omitempty"`
	CustomsEntryRef string     `json:"customs_entry_ref,omitempty"`

	CurrentDeptCode string `json:"current_dept_code,omitempty"`
	StorageLocation string `json:"storage_location,omitempty"`

	FIRRecordID *uuid.UUID `json:"fir_record_id,omitempty"`
	GangID      *uuid.UUID `json:"gang_id,omitempty"`
	CaseRefs    []string   `json:"case_references,omitempty"`

	IARMSRef     string `json:"iarms_ref,omitempty"`
	ATFETraceRef string `json:"atf_etrace_ref,omitempty"`

	Notes     string    `json:"notes,omitempty"`
	CreatedBy uuid.UUID `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateFirearmRequest struct {
	SerialNumber       string           `json:"serial_number"`
	Make               string           `json:"make" binding:"required"`
	Model              string           `json:"model" binding:"required"`
	Caliber            string           `json:"caliber" binding:"required"`
	WeaponType         WeaponType       `json:"weapon_type" binding:"required"`
	ManufactureYear    *int16           `json:"manufacture_year"`
	ManufactureCountry string           `json:"manufacture_country"`
	RegType            RegistrationType `json:"reg_type" binding:"required"`

	OwnerSnisidID *uuid.UUID `json:"owner_snisid_id"`
	OwnerEntity   string     `json:"owner_entity_name"`
	LicenseNumber string     `json:"license_number"`

	ImportDate      *time.Time `json:"import_date"`
	ImportCountry   string     `json:"import_country"`
	ImportPermitRef string     `json:"import_permit_ref"`
	ImporterName    string     `json:"importer_name"`
	CustomsEntryRef string     `json:"customs_entry_ref"`

	CurrentDeptCode string `json:"current_dept_code"`
	StorageLocation string `json:"storage_location"`

	Notes string `json:"notes"`
}

type FirearmStatsByType struct {
	WeaponType WeaponType `json:"weapon_type"`
	Count      int        `json:"count"`
}
