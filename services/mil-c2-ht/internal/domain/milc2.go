package domain

import (
	"time"

	"github.com/google/uuid"
)

type Branch string

const (
	BranchArmy           Branch = "ARMY"
	BranchNavy           Branch = "NAVY"
	BranchAirForce       Branch = "AIR_FORCE"
	BranchSpecialForces  Branch = "SPECIAL_FORCES"
	BranchNationalPolice Branch = "NATIONAL_POLICE"
)

type OperationalStatus string

const (
	OpStatusStandby  OperationalStatus = "STANDBY"
	OpStatusActive   OperationalStatus = "ACTIVE"
	OpStatusDeployed OperationalStatus = "DEPLOYED"
	OpStatusRest     OperationalStatus = "REST"
	OpStatusReserve  OperationalStatus = "RESERVE"
)

type OperationType string

const (
	OpTypeSecurity         OperationType = "SECURITY"
	OpTypeAntiGang         OperationType = "ANTI_GANG"
	OpTypeSearchRescue     OperationType = "SEARCH_RESCUE"
	OpTypeBorderPatrol     OperationType = "BORDER_PATROL"
	OpTypeDisasterResponse OperationType = "DISASTER_RESPONSE"
)

type OperationStatus string

const (
	OpStatusPlanning  OperationStatus = "PLANNING"
	OpStatusActiveOp  OperationStatus = "ACTIVE"
	OpStatusCompleted OperationStatus = "COMPLETED"
	OpStatusCancelled OperationStatus = "CANCELLED"
)

type ReportType string

const (
	ReportSITREP  ReportType = "SITREP"
	ReportINTREP  ReportType = "INTREP"
	ReportSPOTREP ReportType = "SPOTREP"
	ReportPATROL  ReportType = "PATROL"
)

type MilitaryUnit struct {
	UnitID            uuid.UUID         `json:"unit_id" db:"unit_id"`
	UnitName          string            `json:"unit_name" db:"unit_name"`
	Branch            Branch            `json:"branch" db:"branch"`
	ParentUnitID      *uuid.UUID        `json:"parent_unit_id" db:"parent_unit_id"`
	CommanderName     string            `json:"commander_name" db:"commander_name"`
	PersonnelCount    int               `json:"personnel_count" db:"personnel_count"`
	LocationLat       float64           `json:"location_lat" db:"location_lat"`
	LocationLng       float64           `json:"location_lng" db:"location_lng"`
	OperationalStatus OperationalStatus `json:"operational_status" db:"operational_status"`
	EquipmentSummary  string            `json:"equipment_summary" db:"equipment_summary"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
}

type Operation struct {
	OperationID       uuid.UUID       `json:"operation_id" db:"operation_id"`
	OperationName     string          `json:"operation_name" db:"operation_name"`
	OperationType     OperationType   `json:"operation_type" db:"operation_type"`
	Status            OperationStatus `json:"status" db:"status"`
	CommanderID       uuid.UUID       `json:"commander_id" db:"commander_id"`
	StartDate         time.Time       `json:"start_date" db:"start_date"`
	ExpectedEndDate   *time.Time      `json:"expected_end_date" db:"expected_end_date"`
	OperationalArea   string          `json:"operational_area" db:"operational_area"`
	RulesOfEngagement string          `json:"rules_of_engagement" db:"rules_of_engagement"`
	MissionObjective  string          `json:"mission_objective" db:"mission_objective"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

type TacticalReport struct {
	ReportID             uuid.UUID  `json:"report_id" db:"report_id"`
	OperationID          uuid.UUID  `json:"operation_id" db:"operation_id"`
	ReportingUnitID      uuid.UUID  `json:"reporting_unit_id" db:"reporting_unit_id"`
	ReportType           ReportType `json:"report_type" db:"report_type"`
	PositionLat          float64    `json:"position_lat" db:"position_lat"`
	PositionLng          float64    `json:"position_lng" db:"position_lng"`
	EnemyActivity        string     `json:"enemy_activity" db:"enemy_activity"`
	CivilianInteractions string     `json:"civilian_interactions" db:"civilian_interactions"`
	Casualties           int        `json:"casualties" db:"casualties"`
	Detainees            int        `json:"detainees" db:"detainees"`
	EquipmentStatus      string     `json:"equipment_status" db:"equipment_status"`
	SubmittedAt          time.Time  `json:"submitted_at" db:"submitted_at"`
}

type CommonOperatingPicture struct {
	Units      []MilitaryUnit   `json:"units"`
	Operations []Operation      `json:"operations"`
	Reports    []TacticalReport `json:"reports"`
}
