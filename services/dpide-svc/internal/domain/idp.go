package domain

import (
	"time"

	"github.com/google/uuid"
)

type DisplacementCause string

const (
	GangViolence      DisplacementCause = "GANG_VIOLENCE"
	Earthquake        DisplacementCause = "EARTHQUAKE"
	Hurricane         DisplacementCause = "HURRICANE"
	Flood             DisplacementCause = "FLOOD"
	Fire              DisplacementCause = "FIRE"
	PoliticalViolence DisplacementCause = "POLITICAL_VIOLENCE"
	Other             DisplacementCause = "OTHER"
)

type IDPStatus string

const (
	Displaced      IDPStatus = "DISPLACED"
	InCamp         IDPStatus = "IN_CAMP"
	WithHostFamily IDPStatus = "WITH_HOST_FAMILY"
	Relocated      IDPStatus = "RELOCATED"
	ReturnedHome   IDPStatus = "RETURNED_HOME"
	Emigrated      IDPStatus = "EMIGRATED"
	Deceased       IDPStatus = "DECEASED"
)

type IDP struct {
	ID                uuid.UUID         `json:"idp_id" db:"idp_id"`
	NationalDpideID   string            `json:"national_dpide_id" db:"national_dpide_id"`
	SNISIDPersonID    *uuid.UUID        `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	FullName          string            `json:"full_name" db:"full_name"`
	DOB               *time.Time        `json:"dob,omitempty" db:"dob"`
	Gender            *string           `json:"gender,omitempty" db:"gender"`
	HouseholdSize     *int              `json:"household_size,omitempty" db:"household_size"`
	MinorsCount       *int              `json:"minors_count,omitempty" db:"minors_count"`
	DisplacementCause DisplacementCause `json:"displacement_cause" db:"displacement_cause"`
	DisplacementDate  time.Time         `json:"displacement_date" db:"displacement_date"`
	OriginAddress     *string           `json:"origin_address,omitempty" db:"origin_address"`
	OriginDeptCode    string            `json:"origin_dept_code" db:"origin_dept_code"`
	OriginCommune     *string           `json:"origin_commune,omitempty" db:"origin_commune"`
	Status            IDPStatus         `json:"status" db:"status"`
	CurrentLocation   *string           `json:"current_location,omitempty" db:"current_location"`
	CurrentDeptCode   *string           `json:"current_dept_code,omitempty" db:"current_dept_code"`
	CurrentCommune    *string           `json:"current_commune,omitempty" db:"current_commune"`
	CurrentLat        *float64          `json:"current_lat,omitempty" db:"current_lat"`
	CurrentLng        *float64          `json:"current_lng,omitempty" db:"current_lng"`
	CampID            *uuid.UUID        `json:"camp_id,omitempty" db:"camp_id"`
	ShelterType       *string           `json:"shelter_type,omitempty" db:"shelter_type"`
	HasNFI            *bool             `json:"has_nfi,omitempty" db:"has_nfi"`
	ReceivesFoodAid   *bool             `json:"receives_food_aid,omitempty" db:"receives_food_aid"`
	MedicalNeeds      []string          `json:"medical_needs" db:"medical_needs"`
	IomDtmRef         *string           `json:"iom_dtm_ref,omitempty" db:"iom_dtm_ref"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
}

type Camp struct {
	ID                uuid.UUID          `json:"camp_id" db:"camp_id"`
	CampName          string             `json:"camp_name" db:"camp_name"`
	DeptCode          string             `json:"dept_code" db:"dept_code"`
	Commune           *string            `json:"commune,omitempty" db:"commune"`
	Lat               *float64           `json:"lat,omitempty" db:"lat"`
	Lng               *float64           `json:"lng,omitempty" db:"lng"`
	DisplacementCause *DisplacementCause `json:"displacement_cause,omitempty" db:"displacement_cause"`
	ManagingOrg       *string            `json:"managing_org,omitempty" db:"managing_org"`
	Capacity          *int               `json:"capacity,omitempty" db:"capacity"`
	CurrentPopulation *int               `json:"current_population,omitempty" db:"current_population"`
	IsActive          *bool              `json:"is_active,omitempty" db:"is_active"`
	HasMedicalPost    *bool              `json:"has_medical_post,omitempty" db:"has_medical_post"`
	HasSchool         *bool              `json:"has_school,omitempty" db:"has_school"`
	CreatedAt         time.Time          `json:"created_at" db:"created_at"`
}

type IDPStats struct {
	TotalIDPs      int `json:"total_idps" db:"total_idps"`
	DisplacedCount int `json:"displaced_count" db:"displaced_count"`
	InCampCount    int `json:"in_camp_count" db:"in_camp_count"`
	ReturnedCount  int `json:"returned_count" db:"returned_count"`
	CampCount      int `json:"camp_count" db:"camp_count"`
}

type RegisterIDPRequest struct {
	FullName          string   `json:"full_name" binding:"required"`
	DOB               string   `json:"dob"`
	Gender            string   `json:"gender"`
	HouseholdSize     *int     `json:"household_size"`
	MinorsCount       *int     `json:"minors_count"`
	DisplacementCause string   `json:"displacement_cause" binding:"required"`
	DisplacementDate  string   `json:"displacement_date" binding:"required"`
	OriginDeptCode    string   `json:"origin_dept_code" binding:"required"`
	OriginCommune     string   `json:"origin_commune"`
	CurrentLocation   string   `json:"current_location"`
	CurrentDeptCode   string   `json:"current_dept_code"`
	CurrentCommune    string   `json:"current_commune"`
	CurrentLat        *float64 `json:"current_lat"`
	CurrentLng        *float64 `json:"current_lng"`
	ShelterType       string   `json:"shelter_type"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type IDPRepository interface {
	Create(idp *IDP) (*IDP, error)
	FindByID(id uuid.UUID) (*IDP, error)
	FindCamps() ([]Camp, error)
	GetStats() (*IDPStats, error)
	UpdateStatus(id uuid.UUID, status IDPStatus) error
}
