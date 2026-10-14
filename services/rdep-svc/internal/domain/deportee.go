package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Deportee struct {
	DeporteeID         uuid.UUID          `json:"deportee_id"`
	NationalRdepID     string             `json:"national_rdep_id"`
	SNISIDPersonID     uuid.UUID          `json:"snisid_person_id"`
	FIRRecordID        *uuid.UUID         `json:"fir_record_id,omitempty"`
	AFISSubjectID      *uuid.UUID         `json:"afis_subject_id,omitempty"`
	DeportationCountry DeportationCountry `json:"deportation_country"`
	DeportationDate    time.Time          `json:"deportation_date"`
	ArrivalPort        string             `json:"arrival_port"`
	ArrivalDeptCode    string             `json:"arrival_dept_code"`
	DeportingAgency    string             `json:"deporting_agency"`
	DeportationReason  string             `json:"deportation_reason"`
	FlightNumber       string             `json:"flight_number"`
	ForeignName        string             `json:"foreign_name"`
	ForeignAliases     []string           `json:"foreign_aliases"`
	ForeignIDNumber    string             `json:"foreign_id_number"`
	ForeignCountryID   string             `json:"foreign_country_id"`
	HasForeignRecord   bool               `json:"has_foreign_record"`
	CriminalRiskLevel  CriminalRisk       `json:"criminal_risk_level"`
	ConvictedOffenses  []string           `json:"convicted_offenses"`
	GangAffiliated     bool               `json:"gang_affiliated"`
	GangName           string             `json:"gang_name"`
	MonitoringRequired bool               `json:"monitoring_required"`
	MonitoringStatus   MonitoringStatus   `json:"monitoring_status"`
	MonitoringUnit     string             `json:"monitoring_unit"`
	MonitoringOfficer  *uuid.UUID         `json:"monitoring_officer,omitempty"`
	MonitoringEndDate  *time.Time         `json:"monitoring_end_date,omitempty"`
	CurrentAddress     string             `json:"current_address"`
	CurrentCommune     string             `json:"current_commune"`
	CurrentDeptCode    string             `json:"current_dept_code"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type DeporteeRepository interface {
	Create(ctx context.Context, deportee *Deportee) error
	FindByID(ctx context.Context, id uuid.UUID) (*Deportee, error)
	FindByPersonID(ctx context.Context, personID uuid.UUID) (*Deportee, error)
	Update(ctx context.Context, deportee *Deportee) error
	FindHighRisk(ctx context.Context) ([]*Deportee, error)
	FindGangAffiliated(ctx context.Context) ([]*Deportee, error)
	GetStatsByCountry(ctx context.Context) (map[string]int, error)
}

type EventPublisher interface {
	Publish(topic string, event interface{}) error
}

type FBIRecordClient interface {
	GetRecord(ctx context.Context, fbiNumber string) (*ForeignRecord, error)
}

type InterpolClient interface {
	CheckNotices(ctx context.Context, personID uuid.UUID) ([]string, error)
}

type AFISClient interface {
	CheckPrint(ctx context.Context, fingerprintData string) (*AFISHit, error)
}

type AFISHit struct {
	SubjectID uuid.UUID `json:"subject_id"`
	Score     float64   `json:"score"`
}

type ScreeningResult struct {
	PersonID         uuid.UUID       `json:"person_id"`
	HasLocalRecord   bool            `json:"has_local_record"`
	LocalRecordID    *uuid.UUID      `json:"local_record_id,omitempty"`
	HasForeignRecord bool            `json:"has_foreign_record"`
	ForeignRecords   []ForeignRecord `json:"foreign_records"`
	GangAffiliated   bool            `json:"gang_affiliated"`
	GangID           *uuid.UUID      `json:"gang_id,omitempty"`
	InterpolNotices  []string        `json:"interpol_notices"`
	RiskLevel        CriminalRisk    `json:"risk_level"`
}

type DeporteeIntakeRequest struct {
	SNISIDPersonID     string `json:"snisid_person_id" binding:"required"`
	DeportationCountry string `json:"deportation_country" binding:"required"`
	DeportationDate    string `json:"deportation_date" binding:"required"`
	ArrivalPort        string `json:"arrival_port" binding:"required"`
	DeportingAgency    string `json:"deporting_agency"`
	DeportationReason  string `json:"deportation_reason"`
	FlightNumber       string `json:"flight_number"`
	ForeignName        string `json:"foreign_name"`
	ForeignIDNumber    string `json:"foreign_id_number"`
	FBINumber          string `json:"fbi_number"`
	GangName           string `json:"gang_name"`
	CurrentAddress     string `json:"current_address"`
	CurrentCommune     string `json:"current_commune"`
}
