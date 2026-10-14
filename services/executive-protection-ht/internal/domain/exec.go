package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ProtectionLevel string

const (
	ProtectionPresident       ProtectionLevel = "PRESIDENT"
	ProtectionPrimeMinister   ProtectionLevel = "PRIME_MINISTER"
	ProtectionCabinetMinister ProtectionLevel = "CABINET_MINISTER"
	ProtectionJudge           ProtectionLevel = "JUDGE"
	ProtectionDiplomat        ProtectionLevel = "DIPLOMAT"
	ProtectionWitness         ProtectionLevel = "WITNESS"
)

type RiskAssessment string

const (
	RiskLow      RiskAssessment = "LOW"
	RiskMedium   RiskAssessment = "MEDIUM"
	RiskHigh     RiskAssessment = "HIGH"
	RiskCritical RiskAssessment = "CRITICAL"
)

type TransportMode string

const (
	TransportMotorcade        TransportMode = "MOTORCADE"
	TransportHelicopter       TransportMode = "HELICOPTER"
	TransportCommercialFlight TransportMode = "COMMERCIAL_FLIGHT"
)

type MovementStatus string

const (
	MovementDraft     MovementStatus = "DRAFT"
	MovementApproved  MovementStatus = "APPROVED"
	MovementActive    MovementStatus = "ACTIVE"
	MovementCompleted MovementStatus = "COMPLETED"
	MovementCancelled MovementStatus = "CANCELLED"
)

type ThreatType string

const (
	ThreatDirect      ThreatType = "DIRECT_THREAT"
	ThreatSocialMedia ThreatType = "SOCIAL_MEDIA"
	ThreatKnownGroup  ThreatType = "KNOWN_GROUP"
	ThreatStalker     ThreatType = "STALKER"
)

type ThreatStatus string

const (
	ThreatPending    ThreatStatus = "PENDING"
	ThreatActive     ThreatStatus = "ACTIVE"
	ThreatMitigated  ThreatStatus = "MITIGATED"
	ThreatFalseAlarm ThreatStatus = "FALSE_ALARM"
)

type Protectee struct {
	ID                 uuid.UUID       `json:"id"`
	FullName           string          `json:"full_name"`
	OfficialTitle      string          `json:"official_title"`
	ProtectionLevel    ProtectionLevel `json:"protection_level"`
	RiskAssessment     RiskAssessment  `json:"risk_assessment"`
	ActiveThreats      int             `json:"active_threats"`
	PrimaryAgentID     uuid.UUID       `json:"primary_agent_id"`
	SecondaryAgents    pq.StringArray  `json:"secondary_agents"`
	SecureVehiclePlate string          `json:"secure_vehicle_plate"`
	ResidenceLocation  *string         `json:"residence_location,omitempty"`
	WorkplaceLocation  *string         `json:"workplace_location,omitempty"`
	DailyScheduleRefs  pq.StringArray  `json:"daily_schedule_refs"`
	CreatedAt          time.Time       `json:"created_at"`
}

type MovementPlan struct {
	ID                uuid.UUID      `json:"id"`
	ProtecteeID       uuid.UUID      `json:"protectee_id"`
	EventName         string         `json:"event_name"`
	Date              time.Time      `json:"date"`
	DepartureLocation string         `json:"departure_location"`
	ArrivalLocation   string         `json:"arrival_location"`
	TransportMode     TransportMode  `json:"transport_mode"`
	RoutePlan         *string        `json:"route_plan,omitempty"`
	AdvanceDone       bool           `json:"advance_done"`
	ClearedBy         *uuid.UUID     `json:"cleared_by,omitempty"`
	Status            MovementStatus `json:"status"`
	CreatedAt         time.Time      `json:"created_at"`
}

type ThreatAssessment struct {
	ID           uuid.UUID      `json:"id"`
	ProtecteeID  uuid.UUID      `json:"protectee_id"`
	ThreatType   ThreatType     `json:"threat_type"`
	ThreatLevel  RiskAssessment `json:"threat_level"`
	ThreatDetail *string        `json:"threat_detail,omitempty"`
	SourceInfo   *string        `json:"source_info,omitempty"`
	AssessedBy   uuid.UUID      `json:"assessed_by"`
	Mitigation   *string        `json:"mitigation,omitempty"`
	Status       ThreatStatus   `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
}

type CreateProtecteeRequest struct {
	FullName           string   `json:"full_name"`
	OfficialTitle      string   `json:"official_title"`
	ProtectionLevel    string   `json:"protection_level"`
	RiskAssessment     string   `json:"risk_assessment"`
	PrimaryAgentID     string   `json:"primary_agent_id"`
	SecondaryAgents    []string `json:"secondary_agents,omitempty"`
	SecureVehiclePlate string   `json:"secure_vehicle_plate"`
	ResidenceLocation  string   `json:"residence_location,omitempty"`
	WorkplaceLocation  string   `json:"workplace_location,omitempty"`
	DailyScheduleRefs  []string `json:"daily_schedule_refs,omitempty"`
}

type CreateMovementPlanRequest struct {
	ProtecteeID       string `json:"protectee_id"`
	EventName         string `json:"event_name"`
	Date              string `json:"date"`
	DepartureLocation string `json:"departure_location"`
	ArrivalLocation   string `json:"arrival_location"`
	TransportMode     string `json:"transport_mode"`
	RoutePlan         string `json:"route_plan,omitempty"`
	AdvanceDone       bool   `json:"advance_done"`
}

type CreateThreatAssessmentRequest struct {
	ProtecteeID  string `json:"protectee_id"`
	ThreatType   string `json:"threat_type"`
	ThreatLevel  string `json:"threat_level"`
	ThreatDetail string `json:"threat_detail,omitempty"`
	SourceInfo   string `json:"source_info,omitempty"`
	AssessedBy   string `json:"assessed_by"`
	Mitigation   string `json:"mitigation,omitempty"`
}

type DashboardProtection struct {
	TotalProtectees   int     `json:"total_protectees"`
	ActiveProtectees  int     `json:"active_protectees"`
	UpcomingMovements int     `json:"upcoming_movements"`
	ActiveThreats     int     `json:"active_threats"`
	CriticalRiskCount int     `json:"critical_risk_count"`
	HighRiskCount     int     `json:"high_risk_count"`
	MediumRiskCount   int     `json:"medium_risk_count"`
	AvgActiveThreats  float64 `json:"avg_active_threats"`
}
