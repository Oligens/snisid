package domain

import (
	"time"

	"github.com/google/uuid"
)

type AssetCategory string

const (
	Energy     AssetCategory = "ENERGY"
	Transport  AssetCategory = "TRANSPORT"
	Water      AssetCategory = "WATER"
	Telecoms   AssetCategory = "TELECOMS"
	Health     AssetCategory = "HEALTH"
	Finance    AssetCategory = "FINANCE"
	Government AssetCategory = "GOVERNMENT"
	Education  AssetCategory = "EDUCATION"
	FoodSupply AssetCategory = "FOOD_SUPPLY"
)

type ThreatLevel string

const (
	ThreatNormal   ThreatLevel = "NORMAL"
	ThreatElevated ThreatLevel = "ELEVATED"
	ThreatHigh     ThreatLevel = "HIGH"
	ThreatSevere   ThreatLevel = "SEVERE"
	ThreatCritical ThreatLevel = "CRITICAL"
)

type Asset struct {
	ID                 uuid.UUID     `json:"asset_id" db:"asset_id"`
	NationalSipciID    string        `json:"national_sipci_id" db:"national_sipci_id"`
	AssetName          string        `json:"asset_name" db:"asset_name"`
	AssetCategory      AssetCategory `json:"asset_category" db:"asset_category"`
	OwnerEntity        *string       `json:"owner_entity,omitempty" db:"owner_entity"`
	OperatingOrg       *string       `json:"operating_org,omitempty" db:"operating_org"`
	DeptCode           string        `json:"dept_code" db:"dept_code"`
	Commune            *string       `json:"commune,omitempty" db:"commune"`
	Lat                float64       `json:"lat" db:"lat"`
	Lng                float64       `json:"lng" db:"lng"`
	CriticalityScore   *int          `json:"criticality_score,omitempty" db:"criticality_score"`
	PopulationServed   *int          `json:"population_served,omitempty" db:"population_served"`
	SinglePointFailure *bool         `json:"single_point_failure,omitempty" db:"single_point_failure"`
	CurrentThreatLevel ThreatLevel   `json:"current_threat_level" db:"current_threat_level"`
	IsInGangZone       *bool         `json:"is_in_gang_zone,omitempty" db:"is_in_gang_zone"`
	UnderExtortion     *bool         `json:"under_extortion,omitempty" db:"under_extortion"`
	IncidentCount12m   *int          `json:"incident_count_12m,omitempty" db:"incident_count_12m"`
	ProtectionUnit     *string       `json:"protection_unit,omitempty" db:"protection_unit"`
	SiteManagerPhone   *string       `json:"site_manager_phone,omitempty" db:"site_manager_phone"`
	CreatedBy          uuid.UUID     `json:"created_by" db:"created_by"`
	CreatedAt          time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at" db:"updated_at"`
}

type AssetIncident struct {
	IncidentID         uuid.UUID  `json:"incident_id" db:"incident_id"`
	AssetID            uuid.UUID  `json:"asset_id" db:"asset_id"`
	IncidentType       string     `json:"incident_type" db:"incident_type"`
	IncidentDate       time.Time  `json:"incident_date" db:"incident_date"`
	PerpetratorType    *string    `json:"perpetrator_type,omitempty" db:"perpetrator_type"`
	GangID             *uuid.UUID `json:"gang_id,omitempty" db:"gang_id"`
	Description        string     `json:"description" db:"description"`
	ImpactSeverity     *int       `json:"impact_severity,omitempty" db:"impact_severity"`
	PopulationAffected *int       `json:"population_affected,omitempty" db:"population_affected"`
	EconomicLossUSD    *float64   `json:"economic_loss_usd,omitempty" db:"economic_loss_usd"`
	CaseReference      *string    `json:"case_reference,omitempty" db:"case_reference"`
	CreatedBy          uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

type RegisterAssetRequest struct {
	AssetName        string  `json:"asset_name" binding:"required"`
	AssetCategory    string  `json:"asset_category" binding:"required"`
	DeptCode         string  `json:"dept_code" binding:"required"`
	Commune          string  `json:"commune"`
	Lat              float64 `json:"lat" binding:"required"`
	Lng              float64 `json:"lng" binding:"required"`
	CriticalityScore *int    `json:"criticality_score"`
	PopulationServed *int    `json:"population_served"`
	OwnerEntity      string  `json:"owner_entity"`
	OperatingOrg     string  `json:"operating_org"`
	SiteManagerPhone string  `json:"site_manager_phone"`
}

type ReportIncidentRequest struct {
	AssetID         string   `json:"asset_id" binding:"required"`
	IncidentType    string   `json:"incident_type" binding:"required"`
	IncidentDate    string   `json:"incident_date" binding:"required"`
	PerpetratorType string   `json:"perpetrator_type"`
	GangID          string   `json:"gang_id"`
	Description     string   `json:"description" binding:"required"`
	ImpactSeverity  *int     `json:"impact_severity"`
	EconomicLossUSD *float64 `json:"economic_loss_usd"`
}

type RiskAssessment struct {
	AssetID       string      `json:"asset_id"`
	BaseScore     float64     `json:"base_score"`
	ThreatFactors []string    `json:"threat_factors"`
	FinalScore    float64     `json:"final_score"`
	ThreatLevel   ThreatLevel `json:"threat_level"`
}

type AssetRepository interface {
	Create(asset *Asset) (*Asset, error)
	FindByID(id uuid.UUID) (*Asset, error)
	FindAll() ([]Asset, error)
	FindCritical() ([]Asset, error)
	FindUnderThreat() ([]Asset, error)
	CreateIncident(incident *AssetIncident) (*AssetIncident, error)
	FindRecentIncidents() ([]AssetIncident, error)
	UpdateThreatLevel(id uuid.UUID, level ThreatLevel) error
	CountRecentIncidents(assetID uuid.UUID, months int) (int, error)
}
