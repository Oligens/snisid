package domain

import (
	"time"

	"github.com/google/uuid"
)

type Inmate struct {
	InmateID            uuid.UUID  `json:"inmate_id"`
	NationalInmateID    string     `json:"national_inmate_id"`
	SNISIDPersonID      uuid.UUID  `json:"snisid_person_id"`
	FIRRecordID         *uuid.UUID `json:"fir_record_id,omitempty"`
	AFISSubjectID       *uuid.UUID `json:"afis_subject_id,omitempty"`
	CurrentFacility     string     `json:"current_facility"`
	CurrentDeptCode     string     `json:"current_dept_code,omitempty"`
	CellBlock           string     `json:"cell_block,omitempty"`
	IsCurrentlyDetained bool       `json:"is_currently_detained"`
	IsMinor             bool       `json:"is_minor"`
	IsFemale            bool       `json:"is_female"`
	HasSpecialNeeds     bool       `json:"has_special_needs"`
	SpecialNeedsNotes   string     `json:"special_needs_notes,omitempty"`
	IntakeDate          time.Time  `json:"intake_date"`
	ExpectedReleaseDate *time.Time `json:"expected_release_date,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type Detention struct {
	DetentionID          uuid.UUID      `json:"detention_id"`
	InmateID             uuid.UUID      `json:"inmate_id"`
	Facility             string         `json:"facility"`
	DetentionBasis       DetentionBasis `json:"detention_basis"`
	LegalStatus          LegalStatus    `json:"legal_status"`
	CaseReference        string         `json:"case_reference,omitempty"`
	CourtName            string         `json:"court_name,omitempty"`
	ArrestingAuthority   string         `json:"arresting_authority,omitempty"`
	WarrantNumber        string         `json:"warrant_number,omitempty"`
	IntakeDate           time.Time      `json:"intake_date"`
	IntakeOfficer        uuid.UUID      `json:"intake_officer"`
	SentenceDurationDays *int           `json:"sentence_duration_days,omitempty"`
	ReleaseDate          *time.Time     `json:"release_date,omitempty"`
	ReleaseType          *ReleaseType   `json:"release_type,omitempty"`
	ReleasingAuthority   string         `json:"releasing_authority,omitempty"`
	Notes                string         `json:"notes,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
}

type Facility struct {
	FacilityID       uuid.UUID    `json:"facility_id"`
	Code             string       `json:"code"`
	Name             string       `json:"name"`
	Department       string       `json:"department"`
	DeptCode         string       `json:"dept_code"`
	FacilityType     FacilityType `json:"facility_type"`
	Capacity         int          `json:"capacity"`
	CurrentOccupancy int          `json:"current_occupancy"`
	Address          string       `json:"address,omitempty"`
	Phone            string       `json:"phone,omitempty"`
	IsActive         bool         `json:"is_active"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

type Movement struct {
	MovementID   uuid.UUID    `json:"movement_id"`
	InmateID     uuid.UUID    `json:"inmate_id"`
	FromBlock    string       `json:"from_block,omitempty"`
	ToBlock      string       `json:"to_block,omitempty"`
	FromFacility string       `json:"from_facility,omitempty"`
	ToFacility   string       `json:"to_facility,omitempty"`
	MovementType MovementType `json:"movement_type"`
	Reason       string       `json:"reason,omitempty"`
	AuthorizedBy uuid.UUID    `json:"authorized_by"`
	MovedAt      time.Time    `json:"moved_at"`
	CreatedAt    time.Time    `json:"created_at"`
}

type Visit struct {
	VisitID      uuid.UUID  `json:"visit_id"`
	InmateID     uuid.UUID  `json:"inmate_id"`
	VisitorName  string     `json:"visitor_name"`
	VisitorID    string     `json:"visitor_id,omitempty"`
	Relationship string     `json:"relationship,omitempty"`
	VisitDate    time.Time  `json:"visit_date"`
	CheckIn      *time.Time `json:"check_in,omitempty"`
	CheckOut     *time.Time `json:"check_out,omitempty"`
	AuthorizedBy *uuid.UUID `json:"authorized_by,omitempty"`
	Notes        string     `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type HealthEvent struct {
	EventID          uuid.UUID  `json:"event_id"`
	InmateID         uuid.UUID  `json:"inmate_id"`
	EventType        string     `json:"event_type"`
	EventDate        time.Time  `json:"event_date"`
	Description      string     `json:"description,omitempty"`
	TreatingFacility string     `json:"treating_facility,omitempty"`
	Outcome          string     `json:"outcome,omitempty"`
	ReportedBy       *uuid.UUID `json:"reported_by,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type ReleaseRequest struct {
	ReleaseType ReleaseType `json:"release_type" binding:"required"`
	Authority   string      `json:"authority" binding:"required"`
	Notes       string      `json:"notes,omitempty"`
}

type IntakeRequest struct {
	SNISIDPersonID       uuid.UUID      `json:"snisid_person_id" binding:"required"`
	Facility             string         `json:"facility" binding:"required"`
	CellBlock            string         `json:"cell_block,omitempty"`
	DetentionBasis       DetentionBasis `json:"detention_basis" binding:"required"`
	LegalStatus          LegalStatus    `json:"legal_status"`
	CaseReference        string         `json:"case_reference,omitempty"`
	CourtName            string         `json:"court_name,omitempty"`
	ArrestingAuthority   string         `json:"arresting_authority,omitempty"`
	WarrantNumber        string         `json:"warrant_number,omitempty"`
	IntakeOfficer        uuid.UUID      `json:"intake_officer" binding:"required"`
	IsMinor              bool           `json:"is_minor"`
	IsFemale             bool           `json:"is_female"`
	HasSpecialNeeds      bool           `json:"has_special_needs"`
	SpecialNeedsNotes    string         `json:"special_needs_notes,omitempty"`
	SentenceDurationDays *int           `json:"sentence_duration_days,omitempty"`
	ExpectedReleaseDate  *time.Time     `json:"expected_release_date,omitempty"`
}

type TransferRequest struct {
	InmateID      uuid.UUID `json:"inmate_id" binding:"required"`
	ToFacility    string    `json:"to_facility" binding:"required"`
	ToBlock       string    `json:"to_block,omitempty"`
	Reason        string    `json:"reason,omitempty"`
	AuthorizedBy  uuid.UUID `json:"authorized_by" binding:"required"`
	TransportUnit string    `json:"transport_unit,omitempty"`
}

type OccupancyReport struct {
	Facility       string  `json:"facility"`
	DepartmentCode string  `json:"department_code"`
	CurrentCount   int     `json:"current_count"`
	Capacity       int     `json:"capacity"`
	OccupancyRate  float64 `json:"occupancy_rate"`
}
