package domain

import (
	"time"

	"github.com/google/uuid"
)

type DiscoverySource string

const (
	CrimeScene   DiscoverySource = "CRIME_SCENE"
	DisasterSite DiscoverySource = "DISASTER_SITE"
	MassGrave    DiscoverySource = "MASS_GRAVE"
	River        DiscoverySource = "RIVER"
	Street       DiscoverySource = "STREET"
	HospitalDOA  DiscoverySource = "HOSPITAL_DOA"
	OtherSource  DiscoverySource = "OTHER"
)

type RemainsStatus string

const (
	Unidentified        RemainsStatus = "UNIDENTIFIED"
	TentativeMatch      RemainsStatus = "TENTATIVE_MATCH"
	ConfirmedIdentified RemainsStatus = "CONFIRMED_IDENTIFIED"
	Claimed             RemainsStatus = "CLAIMED"
)

type UnidentifiedRemains struct {
	ID                  uuid.UUID       `json:"remains_id" db:"remains_id"`
	NationalRvinID      string          `json:"national_rvin_id" db:"national_rvin_id"`
	DiscoveryDate       time.Time       `json:"discovery_date" db:"discovery_date"`
	DiscoveryLocation   string          `json:"discovery_location" db:"discovery_location"`
	DeptCode            string          `json:"dept_code" db:"dept_code"`
	Commune             *string         `json:"commune,omitempty" db:"commune"`
	Lat                 *float64        `json:"lat,omitempty" db:"lat"`
	Lng                 *float64        `json:"lng,omitempty" db:"lng"`
	DiscoverySource     DiscoverySource `json:"discovery_source" db:"discovery_source"`
	Status              RemainsStatus   `json:"status" db:"status"`
	EstimatedSex        *string         `json:"estimated_sex,omitempty" db:"estimated_sex"`
	EstimatedAgeMin     *int            `json:"estimated_age_min,omitempty" db:"estimated_age_min"`
	EstimatedAgeMax     *int            `json:"estimated_age_max,omitempty" db:"estimated_age_max"`
	EstimatedHeightCm   *int            `json:"estimated_height_cm,omitempty" db:"estimated_height_cm"`
	SkinTone            *string         `json:"skin_tone,omitempty" db:"skin_tone"`
	DistinguishingMarks *string         `json:"distinguishing_marks,omitempty" db:"distinguishing_marks"`
	DecompositionLevel  *int            `json:"decomposition_level,omitempty" db:"decomposition_level"`
	DNASampleTaken      *bool           `json:"dna_sample_taken,omitempty" db:"dna_sample_taken"`
	DNASampleRef        *string         `json:"dna_sample_ref,omitempty" db:"dna_sample_ref"`
	MorgueLocation      *string         `json:"morgue_location,omitempty" db:"morgue_location"`
	CaseReference       *string         `json:"case_reference,omitempty" db:"case_reference"`
	ExaminerID          uuid.UUID       `json:"examiner_id" db:"examiner_id"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

type DNAResult struct {
	ComparisonID     uuid.UUID `json:"comparison_id" db:"comparison_id"`
	RemainsID        uuid.UUID `json:"remains_id" db:"remains_id"`
	ReferenceDNARef  string    `json:"reference_dna_ref" db:"reference_dna_ref"`
	MatchProbability float64   `json:"match_probability" db:"match_probability"`
	IsMatch          bool      `json:"is_match" db:"is_match"`
	LabReference     *string   `json:"lab_reference,omitempty" db:"lab_reference"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

type SourceStats struct {
	Source DiscoverySource `json:"discovery_source" db:"discovery_source"`
	Count  int             `json:"count" db:"count"`
}

type RegisterRemainsRequest struct {
	DiscoveryDate       string   `json:"discovery_date" binding:"required"`
	DiscoveryLocation   string   `json:"discovery_location" binding:"required"`
	DeptCode            string   `json:"dept_code" binding:"required"`
	Commune             string   `json:"commune"`
	Lat                 *float64 `json:"lat"`
	Lng                 *float64 `json:"lng"`
	DiscoverySource     string   `json:"discovery_source" binding:"required"`
	EstimatedSex        string   `json:"estimated_sex"`
	EstimatedAgeMin     *int     `json:"estimated_age_min"`
	EstimatedAgeMax     *int     `json:"estimated_age_max"`
	EstimatedHeightCm   *int     `json:"estimated_height_cm"`
	DistinguishingMarks string   `json:"distinguishing_marks"`
	MorgueLocation      string   `json:"morgue_location"`
}

type SubmitDNARequest struct {
	ReferenceDNARef  string  `json:"reference_dna_ref" binding:"required"`
	MatchProbability float64 `json:"match_probability"`
	IsMatch          bool    `json:"is_match"`
	LabReference     string  `json:"lab_reference"`
}

type RemainsRepository interface {
	Create(remains *UnidentifiedRemains) (*UnidentifiedRemains, error)
	FindByID(id uuid.UUID) (*UnidentifiedRemains, error)
	SubmitDNA(remainsID uuid.UUID, dna *DNAResult) error
	FindUnidentified() ([]UnidentifiedRemains, error)
	GetStatsBySource() ([]SourceStats, error)
}
