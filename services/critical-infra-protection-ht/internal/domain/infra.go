package domain

import (
	"time"

	"github.com/google/uuid"
)

type Sector string

const (
	SectorEnergy     Sector = "ENERGY"
	SectorTelecom    Sector = "TELECOM"
	SectorWater      Sector = "WATER"
	SectorBanking    Sector = "BANKING"
	SectorTransport  Sector = "TRANSPORT"
	SectorHealth     Sector = "HEALTH"
	SectorGovernment Sector = "GOVERNMENT"
	SectorFood       Sector = "FOOD"
)

type Criticality string

const (
	CritCritical Criticality = "CRITICAL"
	CritHigh     Criticality = "HIGH"
	CritMedium   Criticality = "MEDIUM"
	CritLow      Criticality = "LOW"
)

type IncidentType string

const (
	IncidentCyberAttack     IncidentType = "CYBER_ATTACK"
	IncidentPhysicalBreach  IncidentType = "PHYSICAL_BREACH"
	IncidentNaturalDisaster IncidentType = "NATURAL_DISASTER"
	IncidentSabotage        IncidentType = "SABOTAGE"
	IncidentOutage          IncidentType = "OUTAGE"
)

type IncidentStatus string

const (
	IncStatusReported   IncidentStatus = "REPORTED"
	IncStatusResponding IncidentStatus = "RESPONDING"
	IncStatusContained  IncidentStatus = "CONTAINED"
	IncStatusResolved   IncidentStatus = "RESOLVED"
)

type CriticalAsset struct {
	ID                    uuid.UUID   `json:"id"`
	AssetName             string      `json:"asset_name"`
	Sector                Sector      `json:"sector"`
	OwnerEntity           string      `json:"owner_entity"`
	LocationLat           float64     `json:"location_lat"`
	LocationLng           float64     `json:"location_lng"`
	Region                string      `json:"region"`
	DeptCode              string      `json:"dept_code"`
	Criticality           Criticality `json:"criticality"`
	CyberMaturityScore    float64     `json:"cyber_maturity_score"`
	PhysicalSecurityScore float64     `json:"physical_security_score"`
	LastCISAAssessmentAt  *time.Time  `json:"last_cisa_assessment_at,omitempty"`
	ContactName           string      `json:"contact_name"`
	ContactPhone          string      `json:"contact_phone"`
	HasBackupGenerator    bool        `json:"has_backup_generator"`
	HasCyberInsurance     bool        `json:"has_cyber_insurance"`
	CreatedAt             time.Time   `json:"created_at"`
}

type InfrastructureIncident struct {
	ID               uuid.UUID      `json:"id"`
	AssetID          uuid.UUID      `json:"asset_id"`
	IncidentType     IncidentType   `json:"incident_type"`
	Severity         string         `json:"severity"`
	Description      string         `json:"description"`
	ImpactAssessment *string        `json:"impact_assessment,omitempty"`
	DowntimeHours    *float64       `json:"downtime_hours,omitempty"`
	EstimatedLossUSD *float64       `json:"estimated_loss_usd,omitempty"`
	RespondedBy      *uuid.UUID     `json:"responded_by,omitempty"`
	Status           IncidentStatus `json:"status"`
	CreatedAt        time.Time      `json:"created_at"`
}

type SectorRiskAssessment struct {
	ID                uuid.UUID  `json:"id"`
	Sector            Sector     `json:"sector"`
	AssessmentDate    time.Time  `json:"assessment_date"`
	OverallRiskScore  int        `json:"overall_risk_score"`
	TopThreats        []string   `json:"top_threats"`
	Vulnerabilities   []string   `json:"vulnerabilities"`
	Recommendations   []string   `json:"recommendations"`
	AssessorAgency    string     `json:"assessor_agency"`
	NextAssessmentDue *time.Time `json:"next_assessment_due,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type CreateAssetRequest struct {
	AssetName             string  `json:"asset_name" binding:"required"`
	Sector                string  `json:"sector" binding:"required"`
	OwnerEntity           string  `json:"owner_entity" binding:"required"`
	LocationLat           float64 `json:"location_lat" binding:"required"`
	LocationLng           float64 `json:"location_lng" binding:"required"`
	Region                string  `json:"region" binding:"required"`
	DeptCode              string  `json:"dept_code" binding:"required"`
	Criticality           string  `json:"criticality" binding:"required"`
	CyberMaturityScore    float64 `json:"cyber_maturity_score"`
	PhysicalSecurityScore float64 `json:"physical_security_score"`
	LastCISAAssessmentAt  string  `json:"last_cisa_assessment_at"`
	ContactName           string  `json:"contact_name" binding:"required"`
	ContactPhone          string  `json:"contact_phone" binding:"required"`
	HasBackupGenerator    bool    `json:"has_backup_generator"`
	HasCyberInsurance     bool    `json:"has_cyber_insurance"`
}

type ReportIncidentRequest struct {
	AssetID      string `json:"asset_id" binding:"required"`
	IncidentType string `json:"incident_type" binding:"required"`
	Severity     string `json:"severity" binding:"required"`
	Description  string `json:"description" binding:"required"`
}

type CreateAssessmentRequest struct {
	Sector           string   `json:"sector" binding:"required"`
	OverallRiskScore int      `json:"overall_risk_score" binding:"required"`
	TopThreats       []string `json:"top_threats"`
	Vulnerabilities  []string `json:"vulnerabilities"`
	Recommendations  []string `json:"recommendations"`
	AssessorAgency   string   `json:"assessor_agency" binding:"required"`
}
