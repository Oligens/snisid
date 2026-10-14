package domain

import (
	"time"

	"github.com/google/uuid"
)

type Conviction struct {
	ConvictionID         uuid.UUID    `json:"conviction_id"`
	RecordID             uuid.UUID    `json:"record_id"`
	CaseReference        string       `json:"case_reference"`
	CourtName            string       `json:"court_name"`
	CourtDept            string       `json:"court_dept"`
	OffenseClass         OffenseClass `json:"offense_class"`
	OffenseDescription   string       `json:"offense_description"`
	IPCCode              string       `json:"ipc_code"`
	VerdictDate          time.Time    `json:"verdict_date"`
	CaseStatus           CaseStatus   `json:"case_status"`
	SentenceType         SentenceType `json:"sentence_type"`
	SentenceDurationDays *int         `json:"sentence_duration_days,omitempty"`
	FineAmountGDES       *float64     `json:"fine_amount_gdes,omitempty"`
	SentenceStart        *time.Time   `json:"sentence_start,omitempty"`
	SentenceEnd          *time.Time   `json:"sentence_end,omitempty"`
	IsForeignRecord      bool         `json:"is_foreign_record"`
	ForeignCountry       string       `json:"foreign_country,omitempty"`
	InterpolCCCRef       string       `json:"interpol_ccc_ref,omitempty"`
	JudgeName            string       `json:"judge_name"`
	Notes                string       `json:"notes,omitempty"`
	CreatedAt            time.Time    `json:"created_at"`
}
