package domain

import (
	"time"

	"github.com/google/uuid"
)

type MissingPerson struct {
	CaseID              uuid.UUID  `json:"case_id" db:"case_id"`
	NationalDipeID      string     `json:"national_dipe_id" db:"national_dipe_id"`
	CaseType            CaseType   `json:"case_type" db:"case_type"`
	Status              CaseStatus `json:"status" db:"status"`
	SnisidPersonID      *uuid.UUID `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	FullName            string     `json:"full_name" db:"full_name"`
	Aliases             []string   `json:"aliases" db:"aliases"`
	DOB                 *time.Time `json:"dob,omitempty" db:"dob"`
	Gender              *string    `json:"gender,omitempty" db:"gender"`
	Nationality         string     `json:"nationality" db:"nationality"`
	Occupation          *string    `json:"occupation,omitempty" db:"occupation"`
	PhotoRefs           []string   `json:"photo_refs" db:"photo_refs"`
	HeightCM            *int16     `json:"height_cm,omitempty" db:"height_cm"`
	WeightKG            *int16     `json:"weight_kg,omitempty" db:"weight_kg"`
	SkinTone            *string    `json:"skin_tone,omitempty" db:"skin_tone"`
	EyeColor            *string    `json:"eye_color,omitempty" db:"eye_color"`
	HairColor           *string    `json:"hair_color,omitempty" db:"hair_color"`
	DistinguishingMarks *string    `json:"distinguishing_marks,omitempty" db:"distinguishing_marks"`
	ClothingLastSeen    *string    `json:"clothing_last_seen,omitempty" db:"clothing_last_seen"`
	LastSeenDate        time.Time  `json:"last_seen_date" db:"last_seen_date"`
	LastSeenLocation    *string    `json:"last_seen_location,omitempty" db:"last_seen_location"`
	LastSeenDeptCode    *string    `json:"last_seen_dept_code,omitempty" db:"last_seen_dept_code"`
	LastSeenCommune     *string    `json:"last_seen_commune,omitempty" db:"last_seen_commune"`
	LastSeenLat         *float64   `json:"last_seen_lat,omitempty" db:"last_seen_lat"`
	LastSeenLng         *float64   `json:"last_seen_lng,omitempty" db:"last_seen_lng"`
	Circumstances       *string    `json:"circumstances,omitempty" db:"circumstances"`
	SivcAlertID         *uuid.UUID `json:"sivc_alert_id,omitempty" db:"sivc_alert_id"`
	GangID              *uuid.UUID `json:"gang_id,omitempty" db:"gang_id"`
	ExtorsCaseID        *uuid.UUID `json:"extors_case_id,omitempty" db:"extors_case_id"`
	ReportedByName      *string    `json:"reported_by_name,omitempty" db:"reported_by_name"`
	ReportedByPhone     *string    `json:"reported_by_phone,omitempty" db:"reported_by_phone"`
	ReportedBySnisid    *uuid.UUID `json:"reported_by_snisid,omitempty" db:"reported_by_snisid"`
	ReportDate          time.Time  `json:"report_date" db:"report_date"`
	ReportingUnit       *string    `json:"reporting_unit,omitempty" db:"reporting_unit"`
	AfisSubjectID       *uuid.UUID `json:"afis_subject_id,omitempty" db:"afis_subject_id"`
	DnaSampleRef        *string    `json:"dna_sample_ref,omitempty" db:"dna_sample_ref"`
	DnaProfileID        *uuid.UUID `json:"dna_profile_id,omitempty" db:"dna_profile_id"`
	InterpolNoticeRef   *string    `json:"interpol_notice_ref,omitempty" db:"interpol_notice_ref"`
	NcmecRef            *string    `json:"ncmec_ref,omitempty" db:"ncmec_ref"`
	ResolutionDate      *time.Time `json:"resolution_date,omitempty" db:"resolution_date"`
	ResolutionNotes     *string    `json:"resolution_notes,omitempty" db:"resolution_notes"`
	RvinCaseID          *uuid.UUID `json:"rvin_case_id,omitempty" db:"rvin_case_id"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

type Sighting struct {
	SightingID   uuid.UUID  `json:"sighting_id" db:"sighting_id"`
	CaseID       uuid.UUID  `json:"case_id" db:"case_id"`
	SightingDate time.Time  `json:"sighting_date" db:"sighting_date"`
	LocationDesc *string    `json:"location_desc,omitempty" db:"location_desc"`
	DeptCode     *string    `json:"dept_code,omitempty" db:"dept_code"`
	Lat          *float64   `json:"lat,omitempty" db:"lat"`
	Lng          *float64   `json:"lng,omitempty" db:"lng"`
	ReportedBy   *uuid.UUID `json:"reported_by,omitempty" db:"reported_by"`
	ReportMethod *string    `json:"report_method,omitempty" db:"report_method"`
	Confidence   *int16     `json:"confidence,omitempty" db:"confidence"`
	PhotoRef     *string    `json:"photo_ref,omitempty" db:"photo_ref"`
	Verified     bool       `json:"verified" db:"verified"`
	VerifiedBy   *uuid.UUID `json:"verified_by,omitempty" db:"verified_by"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type DisasterMissing struct {
	DisasterID       uuid.UUID `json:"disaster_id" db:"disaster_id"`
	CaseID           uuid.UUID `json:"case_id" db:"case_id"`
	DisasterType     string    `json:"disaster_type" db:"disaster_type"`
	DisasterName     *string   `json:"disaster_name,omitempty" db:"disaster_name"`
	DisasterDate     time.Time `json:"disaster_date" db:"disaster_date"`
	LastKnownAddress *string   `json:"last_known_address,omitempty" db:"last_known_address"`
	ShelterChecked   []string  `json:"shelter_checked" db:"shelter_checked"`
	HospitalChecked  []string  `json:"hospital_checked" db:"hospital_checked"`
	MorgueChecked    []string  `json:"morgue_checked" db:"morgue_checked"`
	RcHaitiRef       *string   `json:"rc_haiti_ref,omitempty" db:"rc_haiti_ref"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

type MatchResult struct {
	Person     *MissingPerson `json:"person,omitempty"`
	Sightings  []*Sighting    `json:"sightings,omitempty"`
	HasMatch   bool           `json:"has_match"`
	Confidence float64        `json:"confidence"`
}

type MorphQuery struct {
	HeightCM *int16  `json:"height_cm,omitempty"`
	WeightKG *int16  `json:"weight_kg,omitempty"`
	Gender   *string `json:"gender,omitempty"`
	SkinTone *string `json:"skin_tone,omitempty"`
	DeptCode *string `json:"dept_code,omitempty"`
}

type ReportDisappearanceRequest struct {
	CaseType            CaseType   `json:"case_type" binding:"required"`
	SnisidPersonID      *uuid.UUID `json:"snisid_person_id,omitempty"`
	FullName            string     `json:"full_name" binding:"required"`
	Aliases             []string   `json:"aliases,omitempty"`
	DOB                 *time.Time `json:"dob,omitempty"`
	Gender              *string    `json:"gender,omitempty"`
	Nationality         string     `json:"nationality"`
	Occupation          *string    `json:"occupation,omitempty"`
	PhotoRefs           []string   `json:"photo_refs,omitempty"`
	HeightCM            *int16     `json:"height_cm,omitempty"`
	WeightKG            *int16     `json:"weight_kg,omitempty"`
	SkinTone            *string    `json:"skin_tone,omitempty"`
	EyeColor            *string    `json:"eye_color,omitempty"`
	HairColor           *string    `json:"hair_color,omitempty"`
	DistinguishingMarks *string    `json:"distinguishing_marks,omitempty"`
	ClothingLastSeen    *string    `json:"clothing_last_seen,omitempty"`
	LastSeenDate        *time.Time `json:"last_seen_date,omitempty"`
	LastSeenLocation    *string    `json:"last_seen_location,omitempty"`
	LastSeenDeptCode    *string    `json:"last_seen_dept_code,omitempty"`
	LastSeenCommune     *string    `json:"last_seen_commune,omitempty"`
	LastSeenLat         *float64   `json:"last_seen_lat,omitempty"`
	LastSeenLng         *float64   `json:"last_seen_lng,omitempty"`
	Circumstances       *string    `json:"circumstances,omitempty"`
	SivcAlertID         *uuid.UUID `json:"sivc_alert_id,omitempty"`
	GangID              *uuid.UUID `json:"gang_id,omitempty"`
	ExtorsCaseID        *uuid.UUID `json:"extors_case_id,omitempty"`
	ReportedByName      *string    `json:"reported_by_name,omitempty"`
	ReportedByPhone     *string    `json:"reported_by_phone,omitempty"`
	ReportedBySnisid    *uuid.UUID `json:"reported_by_snisid,omitempty"`
	ReportingUnit       *string    `json:"reporting_unit,omitempty"`
}

type AddSightingRequest struct {
	SightingDate time.Time  `json:"sighting_date" binding:"required"`
	LocationDesc *string    `json:"location_desc,omitempty"`
	DeptCode     *string    `json:"dept_code,omitempty"`
	Lat          *float64   `json:"lat,omitempty"`
	Lng          *float64   `json:"lng,omitempty"`
	ReportedBy   *uuid.UUID `json:"reported_by,omitempty"`
	ReportMethod *string    `json:"report_method,omitempty"`
	Confidence   *int16     `json:"confidence,omitempty"`
	PhotoRef     *string    `json:"photo_ref,omitempty"`
}

type ResolveCaseRequest struct {
	Status          CaseStatus `json:"status" binding:"required"`
	ResolutionNotes *string    `json:"resolution_notes,omitempty"`
}

type MissingRepository interface {
	CreateCase(c *MissingPerson) error
	FindByID(id uuid.UUID) (*MissingPerson, error)
	GetOpenCases(limit, offset int) ([]*MissingPerson, int, error)
	AddSighting(s *Sighting) error
	GetSightingsByCase(caseID uuid.UUID) ([]*Sighting, error)
	ResolveCase(id uuid.UUID, status CaseStatus, notes *string) error
	GetStatsByType() (map[CaseType]int, error)
}
