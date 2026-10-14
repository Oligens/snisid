package domain

import (
	"time"

	"github.com/google/uuid"
)

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

type LatentSubmission struct {
	CaseReference string         `json:"case_reference" validate:"required"`
	CrimeSceneID  *uuid.UUID     `json:"crime_scene_id,omitempty"`
	LocationDesc  *string        `json:"location_desc,omitempty"`
	DeptCode      *string        `json:"dept_code,omitempty"`
	FoundAt       time.Time      `json:"found_at" validate:"required"`
	ImageBase64   string         `json:"image_base64" validate:"required"`
	Position      FingerPosition `json:"position"`
	ExaminedBy    uuid.UUID      `json:"examined_by" validate:"required"`
}

type LatentMatchConfirm struct {
	MatchedSubjectID uuid.UUID `json:"matched_subject_id" validate:"required"`
	MatchScore       float64   `json:"match_score" validate:"required,min=0,max=100"`
	ExaminedBy       uuid.UUID `json:"examined_by" validate:"required"`
}
