package domain

import (
	"time"

	"github.com/google/uuid"
)

type SourceType string

const (
	SourceTypeMedical    SourceType = "MEDICAL"
	SourceTypeIndustrial SourceType = "INDUSTRIAL"
	SourceTypeScrapMetal SourceType = "SCRAP_METAL"
	SourceTypeResearch   SourceType = "RESEARCH"
	SourceTypeOrphan     SourceType = "ORPHAN"
)

type SourceStatus string

const (
	SourceStatusRegistered SourceStatus = "REGISTERED"
	SourceStatusMoving     SourceStatus = "MOVING"
	SourceStatusLost       SourceStatus = "LOST"
	SourceStatusStolen     SourceStatus = "STOLEN"
	SourceStatusRecovered  SourceStatus = "RECOVERED"
	SourceStatusDisposed   SourceStatus = "DISPOSED"
)

type ChemicalCategory string

const (
	ChemicalCategoryPrecursor       ChemicalCategory = "PRECURSOR"
	ChemicalCategoryDualUse         ChemicalCategory = "DUAL_USE"
	ChemicalCategoryToxicIndustrial ChemicalCategory = "TOXIC_INDUSTRIAL"
)

type AlertLevel string

const (
	AlertLevelYellow AlertLevel = "YELLOW"
	AlertLevelOrange AlertLevel = "ORANGE"
	AlertLevelRed    AlertLevel = "RED"
)

type RadioactiveSource struct {
	SourceID         uuid.UUID    `json:"source_id"`
	SourceType       SourceType   `json:"source_type"`
	Isotope          string       `json:"isotope"`
	ActivityCurie    float64      `json:"activity_curie"`
	LocationBuilding string       `json:"location_building"`
	LocationLat      float64      `json:"location_lat"`
	LocationLng      float64      `json:"location_lng"`
	CustodianOrg     string       `json:"custodian_org"`
	LicenseRef       string       `json:"license_ref"`
	Status           SourceStatus `json:"status"`
	LastVerifiedAt   time.Time    `json:"last_verified_at"`
	LastInventoryAt  time.Time    `json:"last_inventory_at"`
}

type ChemicalPrecursor struct {
	SubstanceName      string           `json:"substance_name"`
	CASNumber          string           `json:"cas_number"`
	Category           ChemicalCategory `json:"category"`
	QuantityKg         float64          `json:"quantity_kg"`
	StorageLocation    string           `json:"storage_location"`
	ImporterEntity     string           `json:"importer_entity"`
	EndUser            string           `json:"end_user"`
	EndUse             string           `json:"end_use"`
	PermitRef          string           `json:"permit_ref"`
	ReportedSuspicious bool             `json:"reported_suspicious"`
	FlaggedAt          *time.Time       `json:"flagged_at,omitempty"`
}

type RadiationAlert struct {
	DetectorID       string     `json:"detector_id"`
	DetectorLocation string     `json:"detector_location"`
	DetectedIsotope  string     `json:"detected_isotope"`
	DoseRateUSv      float64    `json:"dose_rate_usv"`
	AlertLevel       AlertLevel `json:"alert_level"`
	RespondedBy      *uuid.UUID `json:"responded_by,omitempty"`
	ResponseNotes    string     `json:"response_notes"`
	ClearedAt        *time.Time `json:"cleared_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}
