package domain

import (
	"time"

	"github.com/google/uuid"
)

type Modality string

const (
	ModalityFingerprint Modality = "FINGERPRINT"
	ModalityFace        Modality = "FACE"
	ModalityIris        Modality = "IRIS"
	ModalityVoice       Modality = "VOICE"
)

type BioTemplate struct {
	TemplateID             uuid.UUID  `json:"template_id"`
	CitizenID              uuid.UUID  `json:"citizen_id"`
	Modality               Modality   `json:"modality"`
	MilvusVectorID         string     `json:"milvus_vector_id"`
	QualityScore           float64    `json:"quality_score"`
	CaptureDevice          *string    `json:"capture_device,omitempty"`
	CaptureLocation        *string    `json:"capture_location,omitempty"`
	CapturedBy             uuid.UUID  `json:"captured_by"`
	IsActive               bool       `json:"is_active"`
	SupersededByTemplateID *uuid.UUID `json:"superseded_by_template_id,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
}

type VerificationLog struct {
	VerificationID   uuid.UUID  `json:"verification_id"`
	CitizenID        *uuid.UUID `json:"citizen_id,omitempty"`
	Modality         Modality   `json:"modality"`
	RequestingModule string     `json:"requesting_module"`
	MatchScore       float64    `json:"match_score"`
	IsMatch          bool       `json:"is_match"`
	VerifiedAt       time.Time  `json:"verified_at"`
}

type EnrollRequest struct {
	CitizenID       string `json:"citizen_id"`
	Modality        string `json:"modality"`
	ImageData       []byte `json:"image_data"`
	CaptureDevice   string `json:"capture_device,omitempty"`
	CaptureLocation string `json:"capture_location,omitempty"`
	CapturedBy      string `json:"captured_by"`
}

type VerifyRequest struct {
	CitizenID  string `json:"citizen_id"`
	Modality   string `json:"modality"`
	SampleData []byte `json:"sample_data"`
}

type IdentifyRequest struct {
	Modality   string  `json:"modality"`
	SampleData []byte  `json:"sample_data"`
	Threshold  float64 `json:"threshold,omitempty"`
}

type VerifyResult struct {
	IsMatch    bool    `json:"is_match"`
	Score      float64 `json:"score"`
	TemplateID string  `json:"template_id,omitempty"`
}

type IdentifyResult struct {
	Candidates []IdentifyCandidate `json:"candidates"`
}

type IdentifyCandidate struct {
	CitizenID  string  `json:"citizen_id"`
	Score      float64 `json:"score"`
	TemplateID string  `json:"template_id"`
}
