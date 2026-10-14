package domain

import (
	"time"

	"github.com/google/uuid"
)

type VesselArrival struct {
	ID              uuid.UUID  `json:"arrival_id" db:"arrival_id"`
	PortCode        string     `json:"port_code" db:"port_code"`
	VesselIMO       *string    `json:"vessel_imo,omitempty" db:"vessel_imo"`
	VesselName      string     `json:"vessel_name" db:"vessel_name"`
	FlagCountry     *string    `json:"flag_country,omitempty" db:"flag_country"`
	ShippingCompany *string    `json:"shipping_company,omitempty" db:"shipping_company"`
	ArrivalDate     time.Time  `json:"arrival_date" db:"arrival_date"`
	OriginPort      *string    `json:"origin_port,omitempty" db:"origin_port"`
	OriginCountry   *string    `json:"origin_country,omitempty" db:"origin_country"`
	ContainerCount  int        `json:"container_count" db:"container_count"`
	ManifestRef     *string    `json:"manifest_ref,omitempty" db:"manifest_ref"`
	MARVesselID     *uuid.UUID `json:"mar_vessel_id,omitempty" db:"mar_vessel_id"`
	RiskScore       int16      `json:"risk_score" db:"risk_score"`
	RiskLevel       RiskLevel  `json:"risk_level" db:"risk_level"`
	CBPTargetingRef *string    `json:"cbp_targeting_ref,omitempty" db:"cbp_targeting_ref"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

type Container struct {
	ID                 uuid.UUID       `json:"container_id" db:"container_id"`
	ArrivalID          uuid.UUID       `json:"arrival_id" db:"arrival_id"`
	ContainerNumber    string          `json:"container_number" db:"container_number"`
	ContainerType      *string         `json:"container_type,omitempty" db:"container_type"`
	DeclaredContent    string          `json:"declared_content" db:"declared_content"`
	DeclaredWeightKg   *float64        `json:"declared_weight_kg,omitempty" db:"declared_weight_kg"`
	DeclaredValueUSD   *float64        `json:"declared_value_usd,omitempty" db:"declared_value_usd"`
	ShipperName        *string         `json:"shipper_name,omitempty" db:"shipper_name"`
	ShipperCountry     *string         `json:"shipper_country,omitempty" db:"shipper_country"`
	ConsigneeName      *string         `json:"consignee_name,omitempty" db:"consignee_name"`
	ConsigneeSNISIDID  *uuid.UUID      `json:"consignee_snisid_id,omitempty" db:"consignee_snisid_id"`
	Status             ContainerStatus `json:"status" db:"status"`
	RiskScore          int16           `json:"risk_score" db:"risk_score"`
	RiskLevel          RiskLevel       `json:"risk_level" db:"risk_level"`
	RiskFlags          []string        `json:"risk_flags" db:"risk_flags"`
	SelectedForScan    bool            `json:"selected_for_scan" db:"selected_for_scan"`
	ScanDate           *time.Time      `json:"scan_date,omitempty" db:"scan_date"`
	ScanResult         *string         `json:"scan_result,omitempty" db:"scan_result"`
	Seized             bool            `json:"seized" db:"seized"`
	SeizureDescription *string         `json:"seizure_description,omitempty" db:"seizure_description"`
	CaseReference      *string         `json:"case_reference,omitempty" db:"case_reference"`
	CBPTargetingMatch  bool            `json:"cbp_targeting_match" db:"cbp_targeting_match"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
}

type RiskFactor struct {
	ID          uuid.UUID `json:"factor_id" db:"factor_id"`
	ContainerID uuid.UUID `json:"container_id" db:"container_id"`
	FactorType  string    `json:"factor_type" db:"factor_type"`
	Description string    `json:"description" db:"description"`
	WeightScore int16     `json:"weight_score" db:"weight_score"`
	Source      *string   `json:"source,omitempty" db:"source"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type RiskAssessment struct {
	ContainerID    uuid.UUID    `json:"container_id" db:"container_id"`
	TotalScore     int16        `json:"total_score" db:"total_score"`
	FinalRiskLevel RiskLevel    `json:"final_risk_level" db:"final_risk_level"`
	Factors        []RiskFactor `json:"factors"`
}

func (ra *RiskAssessment) AddFlag(flag string, weight int16, source string) {
	ra.TotalScore += weight
	ra.Factors = append(ra.Factors, RiskFactor{
		ID:          uuid.New(),
		ContainerID: ra.ContainerID,
		FactorType:  flag,
		Description: flag,
		WeightScore: weight,
		Source:      &source,
	})
}

func (ra *RiskAssessment) ComputeFinalRisk() {
	if ra.TotalScore >= 80 {
		ra.FinalRiskLevel = CRITICAL
	} else if ra.TotalScore >= 50 {
		ra.FinalRiskLevel = HIGH
	} else if ra.TotalScore >= 20 {
		ra.FinalRiskLevel = MEDIUM
	} else {
		ra.FinalRiskLevel = LOW
	}
}

func (c *Container) IsWeightValueAnomalous() bool {
	if c.DeclaredWeightKg == nil || c.DeclaredValueUSD == nil || *c.DeclaredWeightKg <= 0 {
		return false
	}
	ratio := *c.DeclaredValueUSD / *c.DeclaredWeightKg
	return ratio > 5000 || ratio < 0.5
}

type RecordArrivalRequest struct {
	PortCode        string    `json:"port_code" binding:"required"`
	VesselIMO       *string   `json:"vessel_imo,omitempty"`
	VesselName      string    `json:"vessel_name" binding:"required"`
	FlagCountry     *string   `json:"flag_country,omitempty"`
	ShippingCompany *string   `json:"shipping_company,omitempty"`
	ArrivalDate     time.Time `json:"arrival_date" binding:"required"`
	OriginPort      *string   `json:"origin_port,omitempty"`
	OriginCountry   *string   `json:"origin_country,omitempty"`
	ContainerCount  int       `json:"container_count"`
	ManifestRef     *string   `json:"manifest_ref,omitempty"`
	CBPTargetingRef *string   `json:"cbp_targeting_ref,omitempty"`
}

type ScanRequest struct {
	ScanResult string `json:"scan_result" binding:"required"`
}

type SeizeRequest struct {
	SeizureDescription string `json:"seizure_description" binding:"required"`
	CaseReference      string `json:"case_reference" binding:"required"`
}

type SeizureStats struct {
	TotalSeized       int `json:"total_seized" db:"total_seized"`
	TotalScanned      int `json:"total_scanned" db:"total_scanned"`
	HighRiskCount     int `json:"high_risk_count" db:"high_risk_count"`
	CriticalRiskCount int `json:"critical_risk_count" db:"critical_risk_count"`
}

type ContainerRepository interface {
	CreateArrival(arrival *VesselArrival) (*VesselArrival, error)
	FindArrivalByID(id uuid.UUID) (*VesselArrival, error)
	GetHighRiskContainers() ([]Container, error)
	CreateContainer(container *Container) (*Container, error)
	ScanContainer(id uuid.UUID, scanResult string) (*Container, error)
	SeizeContainer(id uuid.UUID, description string, caseRef string) (*Container, error)
	GetSeizureStats() (*SeizureStats, error)
}
