package domain

import (
	"time"

	"github.com/google/uuid"
)

type IDStatus string

const (
	StatusActive          IDStatus = "ACTIVE"
	StatusSuspended       IDStatus = "SUSPENDED"
	StatusDeceased        IDStatus = "DECEASED"
	StatusCancelled       IDStatus = "CANCELLED"
	StatusMergedDuplicate IDStatus = "MERGED_DUPLICATE"
)

type EnrollmentType string

const (
	EnrollmentBirth              EnrollmentType = "BIRTH"
	EnrollmentAdultFirst         EnrollmentType = "ADULT_FIRST_ENROLLMENT"
	EnrollmentNaturalization     EnrollmentType = "NATURALIZATION"
	EnrollmentRegularization     EnrollmentType = "REGULARIZATION"
	EnrollmentRefugeeStatus      EnrollmentType = "REFUGEE_STATUS"
	EnrollmentReconstructionLost EnrollmentType = "RECONSTRUCTION_LOST_RECORDS"
)

type Citizen struct {
	CitizenID           uuid.UUID      `json:"citizen_id"`
	NIN                 string         `json:"nin"`
	Status              IDStatus       `json:"status"`
	EnrollmentType      EnrollmentType `json:"enrollment_type"`
	FullNameLegal       string         `json:"full_name_legal"`
	FirstName           string         `json:"first_name"`
	MiddleNames         *string        `json:"middle_names,omitempty"`
	LastName            string         `json:"last_name"`
	MaidenName          *string        `json:"maiden_name,omitempty"`
	DOB                 time.Time      `json:"dob"`
	PobCommune          *string        `json:"pob_commune,omitempty"`
	PobDeptCode         *string        `json:"pob_dept_code,omitempty"`
	Gender              *string        `json:"gender,omitempty"`
	Nationality         string         `json:"nationality"`
	DeptCode            string         `json:"dept_code"`
	CurrentAddress      *string        `json:"current_address,omitempty"`
	CurrentCommune      *string        `json:"current_commune,omitempty"`
	BiometricTemplateID *uuid.UUID     `json:"biometric_template_id,omitempty"`
	PhotoRef            *string        `json:"photo_ref,omitempty"`
	MotherNIN           *string        `json:"mother_nin,omitempty"`
	FatherNIN           *string        `json:"father_nin,omitempty"`
	DateOfDeath         *time.Time     `json:"date_of_death,omitempty"`
	DeathCertificateRef *string        `json:"death_certificate_ref,omitempty"`
	IsMerged            bool           `json:"is_merged"`
	MergedIntoCitizenID *uuid.UUID     `json:"merged_into_citizen_id,omitempty"`
	CreatedBy           uuid.UUID      `json:"created_by"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

type EnrollmentRequest struct {
	Age             int            `json:"age"`
	EnrollmentType  EnrollmentType `json:"enrollment_type"`
	FullNameLegal   string         `json:"full_name_legal"`
	FirstName       string         `json:"first_name"`
	MiddleNames     *string        `json:"middle_names,omitempty"`
	LastName        string         `json:"last_name"`
	MaidenName      *string        `json:"maiden_name,omitempty"`
	DOB             time.Time      `json:"dob"`
	PobCommune      *string        `json:"pob_commune,omitempty"`
	PobDeptCode     *string        `json:"pob_dept_code,omitempty"`
	Gender          *string        `json:"gender,omitempty"`
	Nationality     string         `json:"nationality"`
	DeptCode        string         `json:"dept_code"`
	CurrentAddress  *string        `json:"current_address,omitempty"`
	CurrentCommune  *string        `json:"current_commune,omitempty"`
	PhotoRef        *string        `json:"photo_ref,omitempty"`
	MotherNIN       *string        `json:"mother_nin,omitempty"`
	FatherNIN       *string        `json:"father_nin,omitempty"`
	BiometricSample []byte         `json:"biometric_sample,omitempty"`
	CreatedBy       string         `json:"created_by"`
}

type EnrollmentResult struct {
	Citizen *Citizen `json:"citizen"`
	NIN     string   `json:"nin"`
}

type DedupCandidate struct {
	CitizenIDA       uuid.UUID
	CitizenIDB       uuid.UUID
	BiometricScore   float64
	DemographicScore float64
	CompositeScore   float64
}

type BiometricCheckResult struct {
	HasMatch         bool
	MatchedCitizenID uuid.UUID
	Confidence       float64
}

type DemographicMatch struct {
	CitizenID uuid.UUID
	Score     float64
}

var ErrDuplicateDetected = &DomainError{"duplicate biometric detected, enrollment blocked"}

type DomainError struct {
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}
