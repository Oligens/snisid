package domain

import (
	"time"

	"github.com/google/uuid"
)

type CrimeType string

const (
	Homicide             CrimeType = "HOMICIDE"
	MassKilling          CrimeType = "MASS_KILLING"
	Rape                 CrimeType = "RAPE"
	GangRape             CrimeType = "GANG_RAPE"
	Torture              CrimeType = "TORTURE"
	ForcedDisappearance  CrimeType = "FORCED_DISAPPEARANCE"
	ExtrajudicialKilling CrimeType = "EXTRAJUDICIAL_KILLING"
	KidnappingVictim     CrimeType = "KIDNAPPING_VICTIM"
	Mutilation           CrimeType = "MUTILATION"
	OtherGrave           CrimeType = "OTHER_GRAVE"
)

type VictimStatus string

const (
	AliveSurvivor        VictimStatus = "ALIVE_SURVIVOR"
	DeceasedIdentified   VictimStatus = "DECEASED_IDENTIFIED"
	DeceasedUnidentified VictimStatus = "DECEASED_UNIDENTIFIED"
	MissingPresumedDead  VictimStatus = "MISSING_PRESUMED_DEAD"
)

type Victim struct {
	ID               uuid.UUID    `json:"victim_id" db:"victim_id"`
	NationalVictID   string       `json:"national_vict_id" db:"national_vict_id"`
	SNISIDPersonID   *uuid.UUID   `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	CrimeType        CrimeType    `json:"crime_type" db:"crime_type"`
	VictimStatus     VictimStatus `json:"victim_status" db:"victim_status"`
	FullName         *string      `json:"full_name,omitempty" db:"full_name"`
	DOB              *time.Time   `json:"dob,omitempty" db:"dob"`
	Gender           *string      `json:"gender,omitempty" db:"gender"`
	Nationality      *string      `json:"nationality" db:"nationality"`
	Occupation       *string      `json:"occupation,omitempty" db:"occupation"`
	IncidentDate     time.Time    `json:"incident_date" db:"incident_date"`
	IncidentLocation *string      `json:"incident_location,omitempty" db:"incident_location"`
	DeptCode         *string      `json:"dept_code,omitempty" db:"dept_code"`
	Commune          *string      `json:"commune,omitempty" db:"commune"`
	Lat              *float64     `json:"lat,omitempty" db:"lat"`
	Lng              *float64     `json:"lng,omitempty" db:"lng"`
	PerpetratorIDs   []uuid.UUID  `json:"perpetrator_ids" db:"perpetrator_ids"`
	GangID           *uuid.UUID   `json:"gang_id,omitempty" db:"gang_id"`
	CaseReference    *string      `json:"case_reference,omitempty" db:"case_reference"`
	ParquetRef       *string      `json:"parquet_ref,omitempty" db:"parquet_ref"`
	NeedsReparation  *bool        `json:"needs_reparation,omitempty" db:"needs_reparation"`
	CreatedBy        uuid.UUID    `json:"created_by" db:"created_by"`
	CreatedAt        time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at" db:"updated_at"`
}

type MassIncident struct {
	ID                uuid.UUID   `json:"mass_id" db:"mass_id"`
	IncidentName      string      `json:"incident_name" db:"incident_name"`
	CrimeType         CrimeType   `json:"crime_type" db:"crime_type"`
	IncidentDate      time.Time   `json:"incident_date" db:"incident_date"`
	DeptCode          *string     `json:"dept_code,omitempty" db:"dept_code"`
	Commune           *string     `json:"commune,omitempty" db:"commune"`
	Lat               *float64    `json:"lat,omitempty" db:"lat"`
	Lng               *float64    `json:"lng,omitempty" db:"lng"`
	VictimCount       int         `json:"victim_count" db:"victim_count"`
	SurvivorCount     *int        `json:"survivor_count,omitempty" db:"survivor_count"`
	PerpetratorGangID *uuid.UUID  `json:"perpetrator_gang_id,omitempty" db:"perpetrator_gang_id"`
	Description       *string     `json:"description,omitempty" db:"description"`
	DocumentedBy      []string    `json:"documented_by" db:"documented_by"`
	LinkedVictimIDs   []uuid.UUID `json:"linked_victim_ids" db:"linked_victim_ids"`
	CreatedBy         uuid.UUID   `json:"created_by" db:"created_by"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
}

type RegisterVictimRequest struct {
	CrimeType        string `json:"crime_type" binding:"required"`
	VictimStatus     string `json:"victim_status" binding:"required"`
	FullName         string `json:"full_name"`
	DOB              string `json:"dob"`
	Gender           string `json:"gender"`
	IncidentDate     string `json:"incident_date" binding:"required"`
	IncidentLocation string `json:"incident_location"`
	DeptCode         string `json:"dept_code"`
	Commune          string `json:"commune"`
	GangID           string `json:"gang_id"`
}

type CreateMassIncidentRequest struct {
	IncidentName string `json:"incident_name" binding:"required"`
	CrimeType    string `json:"crime_type" binding:"required"`
	IncidentDate string `json:"incident_date" binding:"required"`
	DeptCode     string `json:"dept_code"`
	Commune      string `json:"commune"`
	VictimCount  int    `json:"victim_count" binding:"required"`
	Description  string `json:"description"`
}

type CrimeStats struct {
	CrimeType CrimeType `json:"crime_type" db:"crime_type"`
	Count     int       `json:"count" db:"count"`
}

type VictimRepository interface {
	Create(victim *Victim) (*Victim, error)
	FindByID(id uuid.UUID) (*Victim, error)
	CreateMassIncident(mi *MassIncident) (*MassIncident, error)
	FindMassIncidents() ([]MassIncident, error)
	FindByGang(gangID uuid.UUID) ([]Victim, error)
	GetStatsByType() ([]CrimeStats, error)
	GetReparationList() ([]Victim, error)
}
