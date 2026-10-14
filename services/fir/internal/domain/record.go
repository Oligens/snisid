package domain

import (
	"time"

	"github.com/google/uuid"
)

type CriminalRecord struct {
	RecordID          uuid.UUID  `json:"record_id"`
	NationalFIRID     string     `json:"national_fir_id"`
	SNISIDPersonID    uuid.UUID  `json:"snisid_person_id"`
	AFISSubjectID     *uuid.UUID `json:"afis_subject_id,omitempty"`
	IsHaitianNational bool       `json:"is_haitian_national"`
	IsActive          bool       `json:"is_active"`
	IsExpunged        bool       `json:"is_expunged"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Charge struct {
	ChargeID             uuid.UUID     `json:"charge_id"`
	RecordID             uuid.UUID     `json:"record_id"`
	IsArrest             bool          `json:"is_arrest"`
	ArrestDate           *time.Time    `json:"arrest_date,omitempty"`
	ArrestingUnit        *string       `json:"arresting_unit,omitempty"`
	ArrestingOfficer     *uuid.UUID    `json:"arresting_officer,omitempty"`
	ArrestLocation       *string       `json:"arrest_location,omitempty"`
	DeptCode             *string       `json:"dept_code,omitempty"`
	ChargesText          *string       `json:"charges_text,omitempty"`
	OffenseClass         OffenseClass  `json:"offense_class"`
	CaseReference        *string       `json:"case_reference,omitempty"`
	ReleaseDate          *time.Time    `json:"release_date,omitempty"`
	ReleaseReason        *string       `json:"release_reason,omitempty"`
	CourtName            *string       `json:"court_name,omitempty"`
	CourtDept            *string       `json:"court_dept,omitempty"`
	OffenseDescription   *string       `json:"offense_description,omitempty"`
	IPCCode              *string       `json:"ipc_code,omitempty"`
	VerdictDate          *time.Time    `json:"verdict_date,omitempty"`
	CaseStatus           CaseStatus    `json:"case_status"`
	SentenceType         *SentenceType `json:"sentence_type,omitempty"`
	SentenceDurationDays *int          `json:"sentence_duration_days,omitempty"`
	FineAmountGdes       *float64      `json:"fine_amount_gdes,omitempty"`
	SentenceStart        *time.Time    `json:"sentence_start,omitempty"`
	SentenceEnd          *time.Time    `json:"sentence_end,omitempty"`
	IsForeignRecord      bool          `json:"is_foreign_record"`
	ForeignCountry       *string       `json:"foreign_country,omitempty"`
	InterpolCCCRef       *string       `json:"interpol_ccc_ref,omitempty"`
	JudgeName            *string       `json:"judge_name,omitempty"`
	Notes                *string       `json:"notes,omitempty"`
	CreatedAt            time.Time     `json:"created_at"`
}

type Alias struct {
	AliasID    uuid.UUID `json:"alias_id"`
	RecordID   uuid.UUID `json:"record_id"`
	FirstName  *string   `json:"first_name,omitempty"`
	LastName   *string   `json:"last_name,omitempty"`
	BirthDate  *string   `json:"birth_date,omitempty"`
	IDDocument *string   `json:"id_document,omitempty"`
	Notes      *string   `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type Movement struct {
	MovementID   uuid.UUID    `json:"movement_id"`
	RecordID     uuid.UUID    `json:"record_id"`
	ChargeID     *uuid.UUID   `json:"charge_id,omitempty"`
	MovementType MovementType `json:"movement_type"`
	Description  *string      `json:"description,omitempty"`
	ChangedBy    *uuid.UUID   `json:"changed_by,omitempty"`
	Metadata     *string      `json:"metadata,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}

type Certificate struct {
	CertID            uuid.UUID         `json:"cert_id"`
	RecordID          *uuid.UUID        `json:"record_id,omitempty"`
	SNISIDPersonID    uuid.UUID         `json:"snisid_person_id"`
	CertificateNumber string            `json:"certificate_number"`
	IssuedFor         *string           `json:"issued_for,omitempty"`
	Result            CertificateResult `json:"result"`
	IssuedBy          uuid.UUID         `json:"issued_by"`
	IssuingOffice     *string           `json:"issuing_office,omitempty"`
	IssuedAt          time.Time         `json:"issued_at"`
	ExpiresAt         *time.Time        `json:"expires_at,omitempty"`
	QRCodeRef         *string           `json:"qr_code_ref,omitempty"`
}
