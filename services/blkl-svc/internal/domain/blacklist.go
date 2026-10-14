package domain

import (
	"time"

	"github.com/google/uuid"
)

type BlklBlacklist struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	EntryID         string          `json:"entry_id" db:"entry_id"`
	NationalBlklID  *string         `json:"national_blkl_id,omitempty" db:"national_blkl_id"`
	SNISIDPersonID  uuid.UUID       `json:"snisid_person_id" db:"snisid_person_id"`
	RestrictionType RestrictionType `json:"restriction_type" db:"restriction_type"`
	Source          Source          `json:"source" db:"source"`
	SourceRecordID  *string         `json:"source_record_id,omitempty" db:"source_record_id"`
	Reason          string          `json:"reason" db:"reason"`
	CourtOrderRef   *string         `json:"court_order_ref,omitempty" db:"court_order_ref"`
	OrderedBy       *string         `json:"ordered_by,omitempty" db:"ordered_by"`
	EffectiveDate   time.Time       `json:"effective_date" db:"effective_date"`
	ExpiryDate      *time.Time      `json:"expiry_date,omitempty" db:"expiry_date"`
	IsPermanent     bool            `json:"is_permanent" db:"is_permanent"`
	IsActive        bool            `json:"is_active" db:"is_active"`
	AlertLevel      *string         `json:"alert_level,omitempty" db:"alert_level"`
	ArmedDangerous  bool            `json:"armed_dangerous" db:"armed_dangerous"`
	CreatedBy       string          `json:"created_by" db:"created_by"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

type BlklAlertsLog struct {
	ID             uuid.UUID `json:"id" db:"id"`
	BlacklistID    uuid.UUID `json:"blacklist_id" db:"blacklist_id"`
	PersonID       uuid.UUID `json:"person_id" db:"person_id"`
	AlertType      string    `json:"alert_type" db:"alert_type"`
	Message        string    `json:"message" db:"message"`
	Acknowledged   bool      `json:"acknowledged" db:"acknowledged"`
	AcknowledgedBy *string   `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type BlacklistCheckResult struct {
	IsBlacklisted  bool              `json:"is_blacklisted"`
	PersonID       uuid.UUID         `json:"person_id"`
	Restrictions   []RestrictionType `json:"restrictions,omitempty"`
	ArmedDangerous bool              `json:"armed_dangerous"`
}

type AddEntryRequest struct {
	SNISIDPersonID  uuid.UUID       `json:"snisid_person_id" binding:"required"`
	RestrictionType RestrictionType `json:"restriction_type" binding:"required"`
	Source          Source          `json:"source" binding:"required"`
	SourceRecordID  *string         `json:"source_record_id,omitempty"`
	Reason          string          `json:"reason" binding:"required"`
	CourtOrderRef   *string         `json:"court_order_ref,omitempty"`
	OrderedBy       *string         `json:"ordered_by,omitempty"`
	EffectiveDate   *time.Time      `json:"effective_date,omitempty"`
	ExpiryDate      *time.Time      `json:"expiry_date,omitempty"`
	IsPermanent     bool            `json:"is_permanent"`
	AlertLevel      *string         `json:"alert_level,omitempty"`
	ArmedDangerous  bool            `json:"armed_dangerous"`
	CreatedBy       string          `json:"created_by" binding:"required"`
}

type LiftEntryRequest struct {
	LiftedBy string `json:"lifted_by" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
}

type Repository interface {
	CheckPerson(personID uuid.UUID) (*BlacklistCheckResult, error)
	AddEntry(entry *BlklBlacklist) (*BlklBlacklist, error)
	LiftEntry(id uuid.UUID, liftedBy string) error
	GetActiveEntries() ([]BlklBlacklist, error)
	GetExpiringSoon(days int) ([]BlklBlacklist, error)
	GetByID(id uuid.UUID) (*BlklBlacklist, error)
	LogAlert(log *BlklAlertsLog) error
}
