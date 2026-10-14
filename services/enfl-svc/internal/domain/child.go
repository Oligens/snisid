package domain

import (
	"time"

	"github.com/google/uuid"
)

type RiskCategory string

const (
	MissingAbduction          RiskCategory = "MISSING_ABDUCTION"
	GangRecruitment           RiskCategory = "GANG_RECRUITMENT"
	DomesticServitudeRestavek RiskCategory = "DOMESTIC_SERVITUDE_RESTAVEK"
	SexualExploitation        RiskCategory = "SEXUAL_EXPLOITATION"
	Trafficking               RiskCategory = "TRAFFICKING"
	UnaccompaniedMigrant      RiskCategory = "UNACCOMPANIED_MIGRANT"
	SeparatedDisaster         RiskCategory = "SEPARATED_DISASTER"
	StreetChild               RiskCategory = "STREET_CHILD"
	Other                     RiskCategory = "OTHER"
)

type ChildStatus string

const (
	AtRisk        ChildStatus = "AT_RISK"
	Missing       ChildStatus = "MISSING"
	LocatedSafe   ChildStatus = "LOCATED_SAFE"
	LocatedAtRisk ChildStatus = "LOCATED_AT_RISK"
	InCare        ChildStatus = "IN_CARE"
	Repatriated   ChildStatus = "REPATRIATED"
	Deceased      ChildStatus = "DECEASED"
)

type Child struct {
	ID                  uuid.UUID    `json:"child_id" db:"child_id"`
	NationalEnflID      string       `json:"national_enfl_id" db:"national_enfl_id"`
	SNISIDPersonID      *uuid.UUID   `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	DipeCaseID          *uuid.UUID   `json:"dipe_case_id,omitempty" db:"dipe_case_id"`
	TraitCaseID         *uuid.UUID   `json:"trait_case_id,omitempty" db:"trait_case_id"`
	RiskCategory        RiskCategory `json:"risk_category" db:"risk_category"`
	Status              ChildStatus  `json:"status" db:"status"`
	FullName            string       `json:"full_name" db:"full_name"`
	DOB                 time.Time    `json:"dob" db:"dob"`
	AgeAtRegistration   *int         `json:"age_at_registration,omitempty" db:"age_at_registration"`
	Gender              *string      `json:"gender,omitempty" db:"gender"`
	Nationality         *string      `json:"nationality" db:"nationality"`
	PhotoRefs           []string     `json:"photo_refs" db:"photo_refs"`
	DistinguishingMarks *string      `json:"distinguishing_marks,omitempty" db:"distinguishing_marks"`
	HeightCm            *int         `json:"height_cm,omitempty" db:"height_cm"`
	SkinTone            *string      `json:"skin_tone,omitempty" db:"skin_tone"`
	GuardianName        *string      `json:"guardian_name,omitempty" db:"guardian_name"`
	GuardianPhone       *string      `json:"guardian_phone,omitempty" db:"guardian_phone"`
	GuardianSNISIDID    *uuid.UUID   `json:"guardian_snisid_id,omitempty" db:"guardian_snisid_id"`
	LastKnownLocation   *string      `json:"last_known_location,omitempty" db:"last_known_location"`
	DeptCode            *string      `json:"dept_code,omitempty" db:"dept_code"`
	Commune             *string      `json:"commune,omitempty" db:"commune"`
	DisappearanceDate   *time.Time   `json:"disappearance_date,omitempty" db:"disappearance_date"`
	GangID              *uuid.UUID   `json:"gang_id,omitempty" db:"gang_id"`
	RecruiterSNISIDID   *uuid.UUID   `json:"recruiter_snisid_id,omitempty" db:"recruiter_snisid_id"`
	AfisSubjectID       *uuid.UUID   `json:"afis_subject_id,omitempty" db:"afis_subject_id"`
	DNAProfileID        *uuid.UUID   `json:"dna_profile_id,omitempty" db:"dna_profile_id"`
	InterpolICSERef     *string      `json:"interpol_icse_ref,omitempty" db:"interpol_icse_ref"`
	NcmecRef            *string      `json:"ncmec_ref,omitempty" db:"ncmec_ref"`
	IbesrRef            *string      `json:"ibesr_ref,omitempty" db:"ibesr_ref"`
	AssistanceType      []string     `json:"assistance_type" db:"assistance_type"`
	CurrentShelter      *string      `json:"current_shelter,omitempty" db:"current_shelter"`
	AssignedCaseworker  *uuid.UUID   `json:"assigned_caseworker,omitempty" db:"assigned_caseworker"`
	CreatedBy           uuid.UUID    `json:"created_by" db:"created_by"`
	CreatedAt           time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at" db:"updated_at"`
}

type Restavek struct {
	RestavekID         uuid.UUID  `json:"restavek_id" db:"restavek_id"`
	ChildID            uuid.UUID  `json:"child_id" db:"child_id"`
	EmployingHousehold *string    `json:"employing_household,omitempty" db:"employing_household"`
	HouseholdDept      *string    `json:"household_dept,omitempty" db:"household_dept"`
	HouseholdCommune   *string    `json:"household_commune,omitempty" db:"household_commune"`
	EmployingPersonID  *uuid.UUID `json:"employing_person_id,omitempty" db:"employing_person_id"`
	ReportedConditions *string    `json:"reported_conditions,omitempty" db:"reported_conditions"`
	SchoolAttendance   *bool      `json:"school_attendance,omitempty" db:"school_attendance"`
	IbesrInspection    *bool      `json:"ibesr_inspection,omitempty" db:"ibesr_inspection"`
	LastInspectionDate *time.Time `json:"last_inspection_date,omitempty" db:"last_inspection_date"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

type RegisterChildRequest struct {
	RiskCategory        string `json:"risk_category" binding:"required"`
	FullName            string `json:"full_name" binding:"required"`
	DOB                 string `json:"dob" binding:"required"`
	Gender              string `json:"gender"`
	Nationality         string `json:"nationality"`
	DistinguishingMarks string `json:"distinguishing_marks"`
	HeightCm            *int   `json:"height_cm"`
	SkinTone            string `json:"skin_tone"`
	GuardianName        string `json:"guardian_name"`
	GuardianPhone       string `json:"guardian_phone"`
	DeptCode            string `json:"dept_code"`
	Commune             string `json:"commune"`
	DisappearanceDate   string `json:"disappearance_date"`
	GangID              string `json:"gang_id"`
}

type LocateChildRequest struct {
	Location string `json:"location" binding:"required"`
	Status   string `json:"status"`
}

type ChildRepository interface {
	Create(child *Child) (*Child, error)
	FindByID(id uuid.UUID) (*Child, error)
	FindMissing() ([]Child, error)
	FindRestaveks() ([]Restavek, error)
	UpdateStatus(id uuid.UUID, status ChildStatus, location string) error
	FindGangRecruited() ([]Child, error)
}
