package domain

import (
	"time"

	"github.com/google/uuid"
)

type IllicitWeapon struct {
	WeaponID                 uuid.UUID         `json:"weapon_id" db:"weapon_id"`
	NationalBIARID           string            `json:"national_biar_id" db:"national_biar_id"`
	SerialNumber             *string           `json:"serial_number,omitempty" db:"serial_number"`
	SerialObliterated        bool              `json:"serial_obliterated" db:"serial_obliterated"`
	Make                     *string           `json:"make,omitempty" db:"make"`
	Model                    *string           `json:"model,omitempty" db:"model"`
	Caliber                  *string           `json:"caliber,omitempty" db:"caliber"`
	WeaponType               string            `json:"weapon_type" db:"weapon_type"`
	ManufactureCountry       *string           `json:"manufacture_country,omitempty" db:"manufacture_country"`
	EstimatedManufactureYear *int16            `json:"estimated_manufacture_year,omitempty" db:"estimated_manufacture_year"`
	RecoveryDate             time.Time         `json:"recovery_date" db:"recovery_date"`
	RecoveryContext          RecoveryContext   `json:"recovery_context" db:"recovery_context"`
	RecoveryLocation         *string           `json:"recovery_location,omitempty" db:"recovery_location"`
	RecoveryDeptCode         *string           `json:"recovery_dept_code,omitempty" db:"recovery_dept_code"`
	RecoveryCommune          *string           `json:"recovery_commune,omitempty" db:"recovery_commune"`
	RecoveryLat              *float64          `json:"recovery_lat,omitempty" db:"recovery_lat"`
	RecoveryLng              *float64          `json:"recovery_lng,omitempty" db:"recovery_lng"`
	SeizingUnit              string            `json:"seizing_unit" db:"seizing_unit"`
	SeizingOfficer           *uuid.UUID        `json:"seizing_officer,omitempty" db:"seizing_officer"`
	CaseReference            *string           `json:"case_reference,omitempty" db:"case_reference"`
	FromPersonID             *uuid.UUID        `json:"from_person_id,omitempty" db:"from_person_id"`
	GangID                   *uuid.UUID        `json:"gang_id,omitempty" db:"gang_id"`
	CrimeCategory            *string           `json:"crime_category,omitempty" db:"crime_category"`
	AssociatedCases          []string          `json:"associated_cases" db:"associated_cases"`
	OriginCountry            *string           `json:"origin_country,omitempty" db:"origin_country"`
	TransitCountries         []string          `json:"transit_countries" db:"transit_countries"`
	TraffickingRoute         *string           `json:"trafficking_route,omitempty" db:"trafficking_route"`
	ImportMethod             *string           `json:"import_method,omitempty" db:"import_method"`
	IARMSRef                 *string           `json:"iarms_ref,omitempty" db:"iarms_ref"`
	ATFEtraceRef             *string           `json:"atf_etrace_ref,omitempty" db:"atf_etrace_ref"`
	ReportedToInterpol       bool              `json:"reported_to_interpol" db:"reported_to_interpol"`
	InterpolReportedAt       *time.Time        `json:"interpol_reported_at,omitempty" db:"interpol_reported_at"`
	Disposition              WeaponDisposition `json:"disposition" db:"disposition"`
	DisposalDate             *time.Time        `json:"disposal_date,omitempty" db:"disposal_date"`
	DisposalAuth             *uuid.UUID        `json:"disposal_auth,omitempty" db:"disposal_auth"`
	QuantityAmmunition       int               `json:"quantity_ammunition" db:"quantity_ammunition"`
	AmmunitionType           *string           `json:"ammunition_type,omitempty" db:"ammunition_type"`
	PhotosRefs               []string          `json:"photos_refs" db:"photos_refs"`
	Notes                    *string           `json:"notes,omitempty" db:"notes"`
	CreatedBy                uuid.UUID         `json:"created_by" db:"created_by"`
	CreatedAt                time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time         `json:"updated_at" db:"updated_at"`
}

type BatchSeizure struct {
	BatchID            uuid.UUID  `json:"batch_id" db:"batch_id"`
	BatchReference     string     `json:"batch_reference" db:"batch_reference"`
	OperationName      *string    `json:"operation_name,omitempty" db:"operation_name"`
	SeizureDate        time.Time  `json:"seizure_date" db:"seizure_date"`
	LocationDesc       *string    `json:"location_desc,omitempty" db:"location_desc"`
	DeptCode           *string    `json:"dept_code,omitempty" db:"dept_code"`
	TotalWeapons       int        `json:"total_weapons" db:"total_weapons"`
	WeaponIDs          []string   `json:"weapon_ids" db:"weapon_ids"`
	SeizingUnit        string     `json:"seizing_unit" db:"seizing_unit"`
	LeadOfficer        *uuid.UUID `json:"lead_officer,omitempty" db:"lead_officer"`
	PartneringAgencies []string   `json:"partnering_agencies" db:"partnering_agencies"`
	Notes              *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

type IARMSyncLog struct {
	SyncID       uuid.UUID  `json:"sync_id" db:"sync_id"`
	WeaponID     uuid.UUID  `json:"weapon_id" db:"weapon_id"`
	Direction    string     `json:"direction" db:"direction"`
	IARMSRef     *string    `json:"iarms_ref,omitempty" db:"iarms_ref"`
	SyncStatus   string     `json:"sync_status" db:"sync_status"`
	SyncedAt     *time.Time `json:"synced_at,omitempty" db:"synced_at"`
	ErrorMessage *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type SyncResult struct {
	SyncID    uuid.UUID `json:"sync_id"`
	Direction string    `json:"direction"`
	Count     int       `json:"count"`
	Status    string    `json:"status"`
}

type ReportWeaponRequest struct {
	SerialNumber             *string         `json:"serial_number,omitempty"`
	SerialObliterated        bool            `json:"serial_obliterated"`
	Make                     *string         `json:"make,omitempty"`
	Model                    *string         `json:"model,omitempty"`
	Caliber                  *string         `json:"caliber,omitempty"`
	WeaponType               string          `json:"weapon_type" binding:"required"`
	ManufactureCountry       *string         `json:"manufacture_country,omitempty"`
	EstimatedManufactureYear *int16          `json:"estimated_manufacture_year,omitempty"`
	RecoveryDate             *time.Time      `json:"recovery_date,omitempty"`
	RecoveryContext          RecoveryContext `json:"recovery_context" binding:"required"`
	RecoveryLocation         *string         `json:"recovery_location,omitempty"`
	RecoveryDeptCode         *string         `json:"recovery_dept_code,omitempty"`
	RecoveryCommune          *string         `json:"recovery_commune,omitempty"`
	RecoveryLat              *float64        `json:"recovery_lat,omitempty"`
	RecoveryLng              *float64        `json:"recovery_lng,omitempty"`
	SeizingUnit              string          `json:"seizing_unit" binding:"required"`
	SeizingOfficer           *uuid.UUID      `json:"seizing_officer,omitempty"`
	CaseReference            *string         `json:"case_reference,omitempty"`
	FromPersonID             *uuid.UUID      `json:"from_person_id,omitempty"`
	GangID                   *uuid.UUID      `json:"gang_id,omitempty"`
	CrimeCategory            *string         `json:"crime_category,omitempty"`
	AssociatedCases          []string        `json:"associated_cases,omitempty"`
	OriginCountry            *string         `json:"origin_country,omitempty"`
	TransitCountries         []string        `json:"transit_countries,omitempty"`
	TraffickingRoute         *string         `json:"trafficking_route,omitempty"`
	ImportMethod             *string         `json:"import_method,omitempty"`
	IARMSRef                 *string         `json:"iarms_ref,omitempty"`
	ATFEtraceRef             *string         `json:"atf_etrace_ref,omitempty"`
	QuantityAmmunition       int             `json:"quantity_ammunition"`
	AmmunitionType           *string         `json:"ammunition_type,omitempty"`
	PhotosRefs               []string        `json:"photos_refs,omitempty"`
	Notes                    *string         `json:"notes,omitempty"`
	CreatedBy                uuid.UUID       `json:"created_by" binding:"required"`
}

type ReportBatchRequest struct {
	OperationName      *string               `json:"operation_name,omitempty"`
	SeizureDate        *time.Time            `json:"seizure_date,omitempty"`
	LocationDesc       *string               `json:"location_desc,omitempty"`
	DeptCode           *string               `json:"dept_code,omitempty"`
	SeizingUnit        string                `json:"seizing_unit" binding:"required"`
	LeadOfficer        *uuid.UUID            `json:"lead_officer,omitempty"`
	PartneringAgencies []string              `json:"partnering_agencies,omitempty"`
	Notes              *string               `json:"notes,omitempty"`
	Weapons            []ReportWeaponRequest `json:"weapons" binding:"required,min=1"`
}

type WeaponRepository interface {
	CreateWeapon(w *IllicitWeapon) (*IllicitWeapon, error)
	FindByID(id uuid.UUID) (*IllicitWeapon, error)
	FindBySerial(sn string) ([]IllicitWeapon, error)
	UpdateWeapon(w *IllicitWeapon) (*IllicitWeapon, error)
	CreateBatch(b *BatchSeizure) (*BatchSeizure, error)
	GetWeaponsByGang(gangID uuid.UUID) ([]IllicitWeapon, error)
	GetWeaponsByOrigin(origin string) ([]IllicitWeapon, error)
	GetStatsByGang() ([]map[string]interface{}, error)
	GetStatsByOrigin() ([]map[string]interface{}, error)
	GetRoutes() ([]map[string]interface{}, error)
	UpsertFromIARMS(w *IllicitWeapon) (*IllicitWeapon, error)
	CreateSyncLog(log *IARMSyncLog) error
}
