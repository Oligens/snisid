package domain

import (
	"time"

	"github.com/google/uuid"
)

type CivilActType string

const (
	ActBirth                CivilActType = "BIRTH"
	ActDeath                CivilActType = "DEATH"
	ActMarriage             CivilActType = "MARRIAGE"
	ActDivorce              CivilActType = "DIVORCE"
	ActAdoption             CivilActType = "ADOPTION"
	ActRecognitionPaternity CivilActType = "RECOGNITION_PATERNITY"
)

type CivilAct struct {
	ActID             uuid.UUID    `json:"act_id"`
	ActNumber         string       `json:"act_number"`
	ActType           CivilActType `json:"act_type"`
	CitizenID         *uuid.UUID   `json:"citizen_id,omitempty"`
	RegisteringOffice string       `json:"registering_office"`
	DeptCode          string       `json:"dept_code"`
	Commune           string       `json:"commune"`
	EventDate         time.Time    `json:"event_date"`
	DeclaredDate      time.Time    `json:"declared_date"`
	OfficerName       *string      `json:"officer_name,omitempty"`
	OfficerID         *uuid.UUID   `json:"officer_id,omitempty"`
	IsLateDeclaration bool         `json:"is_late_declaration"`
	IsReconstructed   bool         `json:"is_reconstructed"`
	CreatedAt         time.Time    `json:"created_at"`
}

type BirthDeclaration struct {
	ActID                 uuid.UUID  `json:"act_id"`
	ChildFullName         string     `json:"child_full_name"`
	ChildGender           *string    `json:"child_gender,omitempty"`
	MotherCitizenID       *uuid.UUID `json:"mother_citizen_id,omitempty"`
	FatherCitizenID       *uuid.UUID `json:"father_citizen_id,omitempty"`
	BirthWeightG          *int       `json:"birth_weight_g,omitempty"`
	BirthFacility         *string    `json:"birth_facility,omitempty"`
	AttendingProfessional *string    `json:"attending_professional,omitempty"`
}

type MarriageDeclaration struct {
	ActID               uuid.UUID `json:"act_id"`
	SpouseACitizenID    uuid.UUID `json:"spouse_a_citizen_id"`
	SpouseBCitizenID    uuid.UUID `json:"spouse_b_citizen_id"`
	MarriageRegime      *string   `json:"marriage_regime,omitempty"`
	PrenuptialAgreement bool      `json:"prenuptial_agreement"`
}

type DeathDeclaration struct {
	ActID             uuid.UUID `json:"act_id"`
	DeceasedCitizenID uuid.UUID `json:"deceased_citizen_id"`
	CauseOfDeath      *string   `json:"cause_of_death,omitempty"`
	DeathLocation     *string   `json:"death_location,omitempty"`
	MedicalCertifier  *string   `json:"medical_certifier,omitempty"`
	IsViolentDeath    bool      `json:"is_violent_death"`
	FIRCaseReference  *string   `json:"fir_case_reference,omitempty"`
}
