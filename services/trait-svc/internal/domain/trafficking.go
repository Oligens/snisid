package domain

import (
	"time"

	"github.com/google/uuid"
)

type TraiffickingCase struct {
	ID                 uuid.UUID        `json:"id" db:"case_id"`
	NationalTraitID    string           `json:"national_trait_id" db:"national_trait_id"`
	TrafficType        TraiffickingType `json:"trait_type" db:"trait_type"`
	Status             string           `json:"status" db:"status"`
	VictimCount        int16            `json:"victim_count" db:"victim_count"`
	MinorCount         int16            `json:"minor_count" db:"minor_count"`
	OriginCountry      string           `json:"origin_country" db:"origin_country"`
	TransitCountries   []string         `json:"transit_countries" db:"transit_countries"`
	DestinationCountry *string          `json:"destination_country,omitempty" db:"destination_country"`
	RouteDescription   *string          `json:"route_description,omitempty" db:"route_description"`
	TransportMode      []string         `json:"transport_mode" db:"transport_mode"`
	MarIncidentID      *uuid.UUID       `json:"mar_incident_id,omitempty" db:"mar_incident_id"`
	SifrcrossingIDs    []uuid.UUID      `json:"sifr_crossing_ids" db:"sifr_crossing_ids"`
	GangID             *uuid.UUID       `json:"gang_id,omitempty" db:"gang_id"`
	RecruiterIDs       []uuid.UUID      `json:"recruiter_ids" db:"recruiter_ids"`
	TotalAmountPaid    *float64         `json:"total_amount_paid,omitempty" db:"total_amount_paid"`
	AmountPerPerson    *float64         `json:"amount_per_person,omitempty" db:"amount_per_person"`
	Currency           string           `json:"currency" db:"currency"`
	InvestigatingUnit  *string          `json:"investigating_unit,omitempty" db:"investigating_unit"`
	CaseReference      *string          `json:"case_reference,omitempty" db:"case_reference"`
	IomCaseRef         *string          `json:"iom_case_ref,omitempty" db:"iom_case_ref"`
	CreatedBy          uuid.UUID        `json:"created_by" db:"created_by"`
	CreatedAt          time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at" db:"updated_at"`
}

type TraiffickingVictim struct {
	ID                 uuid.UUID    `json:"id" db:"victim_id"`
	CaseID             uuid.UUID    `json:"case_id" db:"case_id"`
	SnisidPersonID     *uuid.UUID   `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	VictimStatus       VictimStatus `json:"victim_status" db:"victim_status"`
	FullName           *string      `json:"full_name,omitempty" db:"full_name"`
	Nationality        string       `json:"nationality" db:"nationality"`
	Dob                *time.Time   `json:"dob,omitempty" db:"dob"`
	Gender             *string      `json:"gender,omitempty" db:"gender"`
	IsMinor            bool         `json:"is_minor" db:"is_minor"`
	ExploitationType   *string      `json:"exploitation_type,omitempty" db:"exploitation_type"`
	RescueDate         *time.Time   `json:"rescue_date,omitempty" db:"rescue_date"`
	RescueLocation     *string      `json:"rescue_location,omitempty" db:"rescue_location"`
	CurrentLocation    *string      `json:"current_location,omitempty" db:"current_location"`
	AssistanceProvided []string     `json:"assistance_provided" db:"assistance_provided"`
	DipeCaseID         *uuid.UUID   `json:"dipe_case_id,omitempty" db:"dipe_case_id"`
	AfisSubjectID      *uuid.UUID   `json:"afis_subject_id,omitempty" db:"afis_subject_id"`
	CreatedAt          time.Time    `json:"created_at" db:"created_at"`
}

type TraiffickingNetwork struct {
	ID               uuid.UUID   `json:"id" db:"network_id"`
	NetworkName      string      `json:"network_name" db:"network_name"`
	PrimaryRoute     *string     `json:"primary_route,omitempty" db:"primary_route"`
	OriginDept       *string     `json:"origin_dept,omitempty" db:"origin_dept"`
	KnownMembers     []uuid.UUID `json:"known_members" db:"known_members"`
	GangAffiliations []uuid.UUID `json:"gang_affiliations" db:"gang_affiliations"`
	MonthlyVolumeEst *int        `json:"monthly_volume_est,omitempty" db:"monthly_volume_est"`
	FeePerPersonUsd  *float64    `json:"fee_per_person_usd,omitempty" db:"fee_per_person_usd"`
	IsActive         bool        `json:"is_active" db:"is_active"`
	IntelConfidence  *int        `json:"intel_confidence,omitempty" db:"intel_confidence"`
	LinkedCases      []uuid.UUID `json:"linked_cases" db:"linked_cases"`
	CreatedBy        uuid.UUID   `json:"created_by" db:"created_by"`
	CreatedAt        time.Time   `json:"created_at" db:"created_at"`
}

type OpenCaseRequest struct {
	TrafficType        TraiffickingType `json:"trait_type" binding:"required"`
	Status             string           `json:"status"`
	OriginCountry      string           `json:"origin_country"`
	TransitCountries   []string         `json:"transit_countries"`
	DestinationCountry *string          `json:"destination_country,omitempty"`
	RouteDescription   *string          `json:"route_description,omitempty"`
	TransportMode      []string         `json:"transport_mode"`
	MarIncidentID      *uuid.UUID       `json:"mar_incident_id,omitempty"`
	GangID             *uuid.UUID       `json:"gang_id,omitempty"`
	Currency           string           `json:"currency"`
	InvestigatingUnit  *string          `json:"investigating_unit,omitempty"`
	CaseReference      *string          `json:"case_reference,omitempty"`
	IomCaseRef         *string          `json:"iom_case_ref,omitempty"`
	CreatedBy          uuid.UUID        `json:"created_by" binding:"required"`
}

type AddVictimRequest struct {
	SnisidPersonID     *uuid.UUID   `json:"snisid_person_id,omitempty"`
	VictimStatus       VictimStatus `json:"victim_status" binding:"required"`
	FullName           *string      `json:"full_name,omitempty"`
	Nationality        string       `json:"nationality"`
	Dob                *time.Time   `json:"dob,omitempty"`
	Gender             *string      `json:"gender,omitempty"`
	IsMinor            bool         `json:"is_minor"`
	ExploitationType   *string      `json:"exploitation_type,omitempty"`
	RescueDate         *time.Time   `json:"rescue_date,omitempty"`
	RescueLocation     *string      `json:"rescue_location,omitempty"`
	CurrentLocation    *string      `json:"current_location,omitempty"`
	AssistanceProvided []string     `json:"assistance_provided"`
	DipeCaseID         *uuid.UUID   `json:"dipe_case_id,omitempty"`
	AfisSubjectID      *uuid.UUID   `json:"afis_subject_id,omitempty"`
}

type DocumentNetworkRequest struct {
	NetworkName      string      `json:"network_name" binding:"required"`
	PrimaryRoute     *string     `json:"primary_route,omitempty"`
	OriginDept       *string     `json:"origin_dept,omitempty"`
	KnownMembers     []uuid.UUID `json:"known_members"`
	GangAffiliations []uuid.UUID `json:"gang_affiliations"`
	MonthlyVolumeEst *int        `json:"monthly_volume_est,omitempty"`
	FeePerPersonUsd  *float64    `json:"fee_per_person_usd,omitempty"`
	IntelConfidence  *int        `json:"intel_confidence,omitempty"`
	CreatedBy        uuid.UUID   `json:"created_by" binding:"required"`
}

type TypeStats struct {
	TrafficType TraiffickingType `json:"trait_type"`
	Count       int              `json:"count"`
	MinorCount  int              `json:"minor_count"`
}

type TraiffickingRepository interface {
	CreateCase(c *TraiffickingCase) (*TraiffickingCase, error)
	FindByID(id uuid.UUID) (*TraiffickingCase, error)
	FindCaseByNationalID(nationalID string) (*TraiffickingCase, error)
	AddVictim(v *TraiffickingVictim) (*TraiffickingVictim, error)
	GetVictimsByCase(caseID uuid.UUID) ([]TraiffickingVictim, error)
	GetMinorVictims() ([]TraiffickingVictim, error)
	CreateNetwork(n *TraiffickingNetwork) (*TraiffickingNetwork, error)
	GetActiveNetworks() ([]TraiffickingNetwork, error)
	GetStatsByType() ([]TypeStats, error)
	GetMaritimeCases() ([]TraiffickingCase, error)
	CountCasesByTypeAndYear(traitType TraiffickingType, year int) (int, error)
}
