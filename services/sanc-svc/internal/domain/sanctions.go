package domain

import (
	"time"

	"github.com/google/uuid"
)

type SanctionEntry struct {
	SancID            uuid.UUID  `json:"sanc_id" db:"sanc_id"`
	Source            Source     `json:"source" db:"source"`
	SourceRefID       string     `json:"source_ref_id" db:"source_ref_id"`
	EntityType        EntityType `json:"entity_type" db:"entity_type"`
	EntityName        string     `json:"entity_name" db:"entity_name"`
	Aliases           []string   `json:"aliases" db:"aliases"`
	Nationality       []string   `json:"nationality" db:"nationality"`
	DateOfBirth       *time.Time `json:"date_of_birth,omitempty" db:"date_of_birth"`
	PlaceOfBirth      *string    `json:"place_of_birth,omitempty" db:"place_of_birth"`
	PassportNumbers   []string   `json:"passport_numbers" db:"passport_numbers"`
	NationalIDNumbers []string   `json:"national_id_numbers" db:"national_id_numbers"`
	MeasureTypes      []Measure  `json:"measure_types" db:"measure_types"`
	ListingDate       time.Time  `json:"listing_date" db:"listing_date"`
	EndDate           *time.Time `json:"end_date,omitempty" db:"end_date"`
	IsActive          bool       `json:"is_active" db:"is_active"`
	ListingReason     *string    `json:"listing_reason,omitempty" db:"listing_reason"`
	CommitteeNotes    *string    `json:"committee_notes,omitempty" db:"committee_notes"`
	SNISIDPersonID    *uuid.UUID `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	GangID            *uuid.UUID `json:"gang_id,omitempty" db:"gang_id"`
	ChefMemberID      *uuid.UUID `json:"chef_member_id,omitempty" db:"chef_member_id"`
	MatchConfidence   *int       `json:"match_confidence,omitempty" db:"match_confidence"`
	SourceUpdatedAt   *time.Time `json:"source_updated_at,omitempty" db:"source_updated_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

type SyncLog struct {
	SyncID           uuid.UUID  `json:"sync_id" db:"sync_id"`
	Source           Source     `json:"source" db:"source"`
	StartedAt        time.Time  `json:"started_at" db:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	EntriesProcessed int        `json:"entries_processed" db:"entries_processed"`
	EntriesAdded     int        `json:"entries_added" db:"entries_added"`
	EntriesUpdated   int        `json:"entries_updated" db:"entries_updated"`
	EntriesRemoved   int        `json:"entries_removed" db:"entries_removed"`
	Errors           int        `json:"errors" db:"errors"`
	Status           string     `json:"status" db:"status"`
	ErrorDetails     *string    `json:"error_details,omitempty" db:"error_details"`
}

type IdentityMatch struct {
	MatchID         uuid.UUID  `json:"match_id" db:"match_id"`
	SancID          uuid.UUID  `json:"sanc_id" db:"sanc_id"`
	SNISIDPersonID  uuid.UUID  `json:"snisid_person_id" db:"snisid_person_id"`
	MatchScore      float64    `json:"match_score" db:"match_score"`
	MatchFields     []string   `json:"match_fields" db:"match_fields"`
	ConfirmedBy     *uuid.UUID `json:"confirmed_by,omitempty" db:"confirmed_by"`
	IsConfirmed     bool       `json:"is_confirmed" db:"is_confirmed"`
	IsFalsePositive bool       `json:"is_false_positive" db:"is_false_positive"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

type PersonSanctionsResult struct {
	PersonID     uuid.UUID       `json:"person_id"`
	IsSanctioned bool            `json:"is_sanctioned"`
	Matches      []IdentityMatch `json:"matches,omitempty"`
	Entries      []SanctionEntry `json:"entries,omitempty"`
	MaxScore     float64         `json:"max_score"`
}

type SyncResult struct {
	Log            *SyncLog `json:"log"`
	EntriesAdded   int      `json:"entries_added"`
	EntriesUpdated int      `json:"entries_updated"`
	EntriesRemoved int      `json:"entries_removed"`
	Errors         int      `json:"errors"`
}

type CheckNameRequest struct {
	Name        string     `json:"name" binding:"required"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Aliases     []string   `json:"aliases,omitempty"`
}

type ConfirmMatchRequest struct {
	ConfirmedBy uuid.UUID `json:"confirmed_by" binding:"required"`
}

type SanctionsRepository interface {
	UpsertEntry(entry *SanctionEntry) error
	SearchByNameAndDOB(name string, dob *time.Time) ([]SanctionEntry, error)
	GetActiveEntries(limit, offset int) ([]SanctionEntry, int, error)
	GetEntriesBySource(source Source, limit, offset int) ([]SanctionEntry, int, error)
	GetUnconfirmedMatches() ([]IdentityMatch, error)
	SaveMatch(match *IdentityMatch) error
	ConfirmMatch(matchID uuid.UUID, confirmedBy uuid.UUID) error
	SaveSyncLog(log *SyncLog) error
	UpdateSyncLog(log *SyncLog) error
	GetSyncStatus(limit int) ([]SyncLog, error)
}
