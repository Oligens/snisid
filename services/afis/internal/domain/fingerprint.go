package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Fingerprint struct {
	PrintID         uuid.UUID      `json:"print_id" db:"print_id"`
	SubjectID       uuid.UUID      `json:"subject_id" db:"subject_id"`
	FingerPosition  FingerPosition `json:"finger_position" db:"finger_position"`
	CaptureMethod   CaptureMethod  `json:"capture_method" db:"capture_method"`
	NFIQ2Score      int16          `json:"nfiq2_score" db:"nfiq2_score"`
	QualityAccepted bool           `json:"quality_accepted" db:"quality_accepted"`
	ImageRef        string         `json:"image_ref" db:"image_ref"`
	MinutiaeCount   *int16         `json:"minutiae_count,omitempty" db:"minutiae_count"`
	MilvusVectorID  *string        `json:"milvus_vector_id,omitempty" db:"milvus_vector_id"`
	CapturedAt      time.Time      `json:"captured_at" db:"captured_at"`
	CreatedBy       uuid.UUID      `json:"created_by" db:"created_by"`
}

var ErrQualityTooLow = errors.New("qualité empreinte insuffisante: NFIQ2 score < 60")
var ErrMissingRequiredFingers = errors.New("empreintes obligatoires manquantes (pouces + index requis)")

func (f *Fingerprint) IsHighQuality() bool { return f.NFIQ2Score >= 80 }

type SearchResult struct {
	CandidateID    uuid.UUID `json:"candidate_id"`
	SubjectID      uuid.UUID `json:"subject_id"`
	Score          float64   `json:"score"`
	Rank           int       `json:"rank"`
	NationalAFISID string    `json:"national_afis_id"`
}

type EnrollmentRequest struct {
	SubjectType    SubjectType          `json:"subject_type" validate:"required"`
	SNISIDPersonID *uuid.UUID           `json:"snisid_person_id,omitempty"`
	FIRRecordID    *uuid.UUID           `json:"fir_record_id,omitempty"`
	EnrollingUnit  string               `json:"enrolling_unit" validate:"required"`
	Fingerprints   []FingerprintCapture `json:"fingerprints" validate:"required,min=2"`
}

type FingerprintCapture struct {
	Position    FingerPosition `json:"position" validate:"required"`
	Method      CaptureMethod  `json:"method"`
	ImageBase64 string         `json:"image_base64" validate:"required"`
	NFIQ2Score  int16          `json:"nfiq2_score" validate:"required,min=0,max=100"`
}

type Subject struct {
	SubjectID        uuid.UUID   `json:"subject_id" db:"subject_id"`
	SNISIDPersonID   *uuid.UUID  `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	FIRRecordID      *uuid.UUID  `json:"fir_record_id,omitempty" db:"fir_record_id"`
	SubjectType      SubjectType `json:"subject_type" db:"subject_type"`
	NationalAFISID   *string     `json:"national_afis_id,omitempty" db:"national_afis_id"`
	AliasIDs         []uuid.UUID `json:"alias_ids,omitempty" db:"alias_ids"`
	EnrolmentDate    time.Time   `json:"enrolment_date" db:"enrolment_date"`
	EnrollingUnit    string      `json:"enrolling_unit" db:"enrolling_unit"`
	EnrollingOfficer uuid.UUID   `json:"enrolling_officer" db:"enrolling_officer"`
	CreatedAt        time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at" db:"updated_at"`
}

type LatentPrint struct {
	LatentID         uuid.UUID      `json:"latent_id" db:"latent_id"`
	CaseReference    string         `json:"case_reference" db:"case_reference"`
	CrimeSceneID     *uuid.UUID     `json:"crime_scene_id,omitempty" db:"crime_scene_id"`
	LocationDesc     *string        `json:"location_desc,omitempty" db:"location_desc"`
	DeptCode         *string        `json:"dept_code,omitempty" db:"dept_code"`
	FoundAt          time.Time      `json:"found_at" db:"found_at"`
	ImageRef         string         `json:"image_ref" db:"image_ref"`
	NFIQ2Score       *int16         `json:"nfiq2_score,omitempty" db:"nfiq2_score"`
	FingerPosition   FingerPosition `json:"finger_position" db:"finger_position"`
	IsIdentified     bool           `json:"is_identified" db:"is_identified"`
	MatchedSubjectID *uuid.UUID     `json:"matched_subject_id,omitempty" db:"matched_subject_id"`
	MatchScore       *float64       `json:"match_score,omitempty" db:"match_score"`
	ExaminedBy       *uuid.UUID     `json:"examined_by,omitempty" db:"examined_by"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
}

type SearchTransaction struct {
	TransactionID    uuid.UUID       `json:"transaction_id" db:"transaction_id"`
	TransactionType  TransactionType `json:"transaction_type" db:"transaction_type"`
	QuerySubjectID   *uuid.UUID      `json:"query_subject_id,omitempty" db:"query_subject_id"`
	QueryLatentID    *uuid.UUID      `json:"query_latent_id,omitempty" db:"query_latent_id"`
	HitsCount        int16           `json:"hits_count" db:"hits_count"`
	TopScore         *float64        `json:"top_score,omitempty" db:"top_score"`
	TopMatchID       *uuid.UUID      `json:"top_match_id,omitempty" db:"top_match_id"`
	SearchDurationMs *int32          `json:"search_duration_ms,omitempty" db:"search_duration_ms"`
	RequestedBy      uuid.UUID       `json:"requested_by" db:"requested_by"`
	RequestingUnit   *string         `json:"requesting_unit,omitempty" db:"requesting_unit"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
}
