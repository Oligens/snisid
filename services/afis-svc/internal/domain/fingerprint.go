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
