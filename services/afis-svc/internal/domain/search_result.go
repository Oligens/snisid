package domain

import (
	"time"

	"github.com/google/uuid"
)

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

type SubjectProfile struct {
	SubjectID      uuid.UUID           `json:"subject_id"`
	SNISIDPersonID *uuid.UUID          `json:"snisid_person_id,omitempty"`
	FIRRecordID    *uuid.UUID          `json:"fir_record_id,omitempty"`
	SubjectType    SubjectType         `json:"subject_type"`
	NationalAFISID *string             `json:"national_afis_id,omitempty"`
	EnrollingUnit  string              `json:"enrolling_unit"`
	Fingerprints   []Fingerprint       `json:"fingerprints"`
	SearchHistory  []SearchTransaction `json:"search_history,omitempty"`
}
