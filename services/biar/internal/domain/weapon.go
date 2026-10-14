package domain

import (
	"time"

	"github.com/google/uuid"
)

type IllicitWeapon struct {
	WeaponID                 uuid.UUID `json:"weapon_id"`
	NationalBIARID           string    `json:"national_biar_id"`
	SerialNumber             *string   `json:"serial_number,omitempty"`
	SerialObliterated        bool      `json:"serial_obliterated"`
	Make                     *string   `json:"make,omitempty"`
	Model                    *string   `json:"model,omitempty"`
	Caliber                  *string   `json:"caliber,omitempty"`
	WeaponType               string    `json:"weapon_type"`
	ManufactureCountry       *string   `json:"manufacture_country,omitempty"`
	EstimatedManufactureYear *int16    `json:"estimated_manufacture_year,omitempty"`

	RecoveryDate     time.Time       `json:"recovery_date"`
	RecoveryContext  RecoveryContext `json:"recovery_context"`
	RecoveryLocation *string         `json:"recovery_location,omitempty"`
	RecoveryDeptCode *string         `json:"recovery_dept_code,omitempty"`
	RecoveryCommune  *string         `json:"recovery_commune,omitempty"`
	RecoveryLat      *float64        `json:"recovery_lat,omitempty"`
	RecoveryLng      *float64        `json:"recovery_lng,omitempty"`
	SeizingUnit      string          `json:"seizing_unit"`
	SeizingOfficer   *uuid.UUID      `json:"seizing_officer,omitempty"`
	CaseReference    *string         `json:"case_reference,omitempty"`

	FromPersonID    *uuid.UUID `json:"from_person_id,omitempty"`
	GangID          *uuid.UUID `json:"gang_id,omitempty"`
	CrimeCategory   *string    `json:"crime_category,omitempty"`
	AssociatedCases []string   `json:"associated_cases"`

	OriginCountry    *string  `json:"origin_country,omitempty"`
	TransitCountries []string `json:"transit_countries"`
	TraffickingRoute *string  `json:"trafficking_route,omitempty"`
	ImportMethod     *string  `json:"import_method,omitempty"`

	IARMSRef           *string    `json:"iarms_ref,omitempty"`
	ATFETraceRef       *string    `json:"atf_etrace_ref,omitempty"`
	ReportedToInterpol bool       `json:"reported_to_interpol"`
	InterpolReportedAt *time.Time `json:"interpol_reported_at,omitempty"`

	Disposition  WeaponDisposition `json:"disposition"`
	DisposalDate *time.Time        `json:"disposal_date,omitempty"`
	DisposalAuth *uuid.UUID        `json:"disposal_auth,omitempty"`

	QuantityAmmunition int      `json:"quantity_ammunition"`
	AmmunitionType     *string  `json:"ammunition_type,omitempty"`
	PhotosRefs         []string `json:"photos_refs"`
	Notes              *string  `json:"notes,omitempty"`

	CreatedBy uuid.UUID `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateWeaponRequest struct {
	SerialNumber             *string `json:"serial_number"`
	SerialObliterated        bool    `json:"serial_obliterated"`
	Make                     *string `json:"make"`
	Model                    *string `json:"model"`
	Caliber                  *string `json:"caliber"`
	WeaponType               string  `json:"weapon_type" validate:"required"`
	ManufactureCountry       *string `json:"manufacture_country"`
	EstimatedManufactureYear *int16  `json:"estimated_manufacture_year"`

	RecoveryDate     time.Time       `json:"recovery_date" validate:"required"`
	RecoveryContext  RecoveryContext `json:"recovery_context" validate:"required"`
	RecoveryLocation *string         `json:"recovery_location"`
	RecoveryDeptCode *string         `json:"recovery_dept_code"`
	RecoveryCommune  *string         `json:"recovery_commune"`
	RecoveryLat      *float64        `json:"recovery_lat"`
	RecoveryLng      *float64        `json:"recovery_lng"`
	SeizingUnit      string          `json:"seizing_unit" validate:"required"`
	SeizingOfficer   *uuid.UUID      `json:"seizing_officer"`
	CaseReference    *string         `json:"case_reference"`

	FromPersonID    *uuid.UUID `json:"from_person_id"`
	GangID          *uuid.UUID `json:"gang_id"`
	CrimeCategory   *string    `json:"crime_category"`
	AssociatedCases []string   `json:"associated_cases"`

	OriginCountry    *string  `json:"origin_country"`
	TransitCountries []string `json:"transit_countries"`
	TraffickingRoute *string  `json:"trafficking_route"`
	ImportMethod     *string  `json:"import_method"`

	QuantityAmmunition int      `json:"quantity_ammunition"`
	AmmunitionType     *string  `json:"ammunition_type"`
	PhotosRefs         []string `json:"photos_refs"`
	Notes              *string  `json:"notes"`
}
