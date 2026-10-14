package domain

import (
	"time"

	"github.com/google/uuid"
)

type Deportee struct {
	DeporteeID     uuid.UUID  `json:"deportee_id"`
	NationalRDEPID string     `json:"national_rdep_id"`
	SNISIDPersonID uuid.UUID  `json:"snisid_person_id"`
	FIRRecordID    *uuid.UUID `json:"fir_record_id,omitempty"`
	AFISSubjectID  *uuid.UUID `json:"afis_subject_id,omitempty"`

	DeportationCountry DeportationCountry `json:"deportation_country"`
	DeportationDate    time.Time          `json:"deportation_date"`
	ArrivalPort        string             `json:"arrival_port"`
	ArrivalDeptCode    *string            `json:"arrival_dept_code,omitempty"`
	DeportingAgency    *string            `json:"deporting_agency,omitempty"`
	DeportationReason  *string            `json:"deportation_reason,omitempty"`
	FlightID           *uuid.UUID         `json:"flight_id,omitempty"`
	FlightNumber       *string            `json:"flight_number,omitempty"`

	ForeignName      *string  `json:"foreign_name,omitempty"`
	ForeignAliases   []string `json:"foreign_aliases,omitempty"`
	ForeignIDNumber  *string  `json:"foreign_id_number,omitempty"`
	ForeignCountryID *string  `json:"foreign_country_id,omitempty"`

	HasForeignRecord  bool         `json:"has_foreign_record"`
	CriminalRiskLevel CriminalRisk `json:"criminal_risk_level"`
	ConvictedOffenses []string     `json:"convicted_offenses,omitempty"`
	GangAffiliated    bool         `json:"gang_affiliated"`
	GangName          *string      `json:"gang_name,omitempty"`

	MonitoringRequired bool             `json:"monitoring_required"`
	MonitoringStatus   MonitoringStatus `json:"monitoring_status"`
	MonitoringUnit     *string          `json:"monitoring_unit,omitempty"`
	MonitoringOfficer  *uuid.UUID       `json:"monitoring_officer,omitempty"`
	MonitoringEndDate  *time.Time       `json:"monitoring_end_date,omitempty"`

	CurrentAddress  *string `json:"current_address,omitempty"`
	CurrentCommune  *string `json:"current_commune,omitempty"`
	CurrentDeptCode *string `json:"current_dept_code,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeporteeIntakeRequest struct {
	SNISIDPersonID     uuid.UUID          `json:"snisid_person_id" binding:"required"`
	AFISSubjectID      *uuid.UUID         `json:"afis_subject_id,omitempty"`
	DeportationCountry DeportationCountry `json:"deportation_country" binding:"required"`
	DeportationDate    time.Time          `json:"deportation_date" binding:"required"`
	ArrivalPort        string             `json:"arrival_port" binding:"required"`
	ArrivalDeptCode    *string            `json:"arrival_dept_code,omitempty"`
	DeportingAgency    *string            `json:"deporting_agency,omitempty"`
	DeportationReason  *string            `json:"deportation_reason,omitempty"`
	FlightNumber       *string            `json:"flight_number,omitempty"`

	ForeignName      *string  `json:"foreign_name,omitempty"`
	ForeignAliases   []string `json:"foreign_aliases,omitempty"`
	ForeignIDNumber  *string  `json:"foreign_id_number,omitempty"`
	ForeignCountryID *string  `json:"foreign_country_id,omitempty"`

	FingerprintData *string `json:"fingerprint_data,omitempty"`
	FBINumber       *string `json:"fbi_number,omitempty"`
	GangName        *string `json:"gang_name,omitempty"`
}

type ForeignRecord struct {
	ForeignRecordID    uuid.UUID          `json:"foreign_record_id"`
	DeporteeID         uuid.UUID          `json:"deportee_id"`
	Country            DeportationCountry `json:"country"`
	CourtName          *string            `json:"court_name,omitempty"`
	OffenseDescription string             `json:"offense_description"`
	OffenseDate        *time.Time         `json:"offense_date,omitempty"`
	ConvictionDate     *time.Time         `json:"conviction_date,omitempty"`
	Sentence           *string            `json:"sentence,omitempty"`
	PrisonServed       *string            `json:"prison_served,omitempty"`
	FBINumber          *string            `json:"fbi_number,omitempty"`
	InterpolRef        *string            `json:"interpol_ref,omitempty"`
	SourceDocument     *string            `json:"source_document,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
}

type MonitoringEvent struct {
	EventID     uuid.UUID           `json:"event_id"`
	DeporteeID  uuid.UUID           `json:"deportee_id"`
	EventType   MonitoringEventType `json:"event_type"`
	EventDate   time.Time           `json:"event_date"`
	LocationLat *float64            `json:"location_lat,omitempty"`
	LocationLng *float64            `json:"location_lng,omitempty"`
	Notes       *string             `json:"notes,omitempty"`
	ReportedBy  uuid.UUID           `json:"reported_by"`
	CreatedAt   time.Time           `json:"created_at"`
}

type ScreeningResult struct {
	PersonID         uuid.UUID       `json:"person_id"`
	RiskLevel        CriminalRisk    `json:"risk_level"`
	HasLocalRecord   bool            `json:"has_local_record"`
	LocalRecordID    *uuid.UUID      `json:"local_record_id,omitempty"`
	HasForeignRecord bool            `json:"has_foreign_record"`
	ForeignRecords   []ForeignRecord `json:"foreign_records,omitempty"`
	GangAffiliated   bool            `json:"gang_affiliated"`
	GangID           *uuid.UUID      `json:"gang_id,omitempty"`
	InterpolNotices  []string        `json:"interpol_notices,omitempty"`
}

type StatsByCountry struct {
	Country       DeportationCountry `json:"country"`
	TotalCount    int                `json:"total_count"`
	HighRiskCount int                `json:"high_risk_count"`
	GangCount     int                `json:"gang_count"`
}
