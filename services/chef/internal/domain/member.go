package domain

import (
	"time"

	"github.com/google/uuid"
)

type CriminalMember struct {
	MemberID       uuid.UUID  `json:"member_id"`
	NationalChefID string     `json:"national_chef_id"`
	SNISIDPersonID uuid.UUID  `json:"snisid_person_id"`
	FIRRecordID    *uuid.UUID `json:"fir_record_id,omitempty"`
	AFISSubjectID  *uuid.UUID `json:"afis_subject_id,omitempty"`
	RDePDeporteeID *uuid.UUID `json:"rdep_deportee_id,omitempty"`

	PrimaryGangID   uuid.UUID    `json:"primary_gang_id"`
	RoleInGang      ChefRoleType `json:"role_in_gang"`
	RoleDescription string       `json:"role_description,omitempty"`
	JoinedDate      *time.Time   `json:"joined_date,omitempty"`
	RankLevel       *int16       `json:"rank_level,omitempty"`

	Aliases             []string `json:"aliases,omitempty"`
	KnownLanguages      []string `json:"known_languages,omitempty"`
	TattooDescription   string   `json:"tattoo_description,omitempty"`
	PhysicalDescription string   `json:"physical_description,omitempty"`
	PhotoRefs           []string `json:"photo_refs,omitempty"`

	TerritoryDept     string   `json:"territory_dept,omitempty"`
	TerritoryCommunes []string `json:"territory_communes,omitempty"`

	KnownArmed       bool     `json:"known_armed"`
	WeaponTypes      []string `json:"weapon_types,omitempty"`
	TrainedCombatant bool     `json:"trained_combatant"`

	Status            ChefStatus `json:"status"`
	UNDesignated      bool       `json:"un_designated"`
	UNDesignationDate *time.Time `json:"un_designation_date,omitempty"`
	OFACDesignated    bool       `json:"ofac_designated"`
	OFACSDNRef        string     `json:"ofac_sdn_ref,omitempty"`
	InterpolNoticeRef string     `json:"interpol_notice_ref,omitempty"`

	LastKnownAddress string     `json:"last_known_address,omitempty"`
	LastKnownDept    string     `json:"last_known_dept,omitempty"`
	LastSeenAt       *time.Time `json:"last_seen_at,omitempty"`
	LastSeenLocation string     `json:"last_seen_location,omitempty"`

	EstimatedWealthUSD *float64 `json:"estimated_wealth_usd,omitempty"`
	KnownAssets        []string `json:"known_assets,omitempty"`

	IntelClassification string    `json:"intel_classification"`
	IntelConfidence     *int16    `json:"intel_confidence,omitempty"`
	CreatedBy           uuid.UUID `json:"created_by"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type IntelNote struct {
	NoteID        uuid.UUID `json:"note_id"`
	MemberID      uuid.UUID `json:"member_id"`
	NoteDate      time.Time `json:"note_date"`
	IntelType     string    `json:"intel_type,omitempty"`
	Content       string    `json:"content"`
	SourceClassif string    `json:"source_classif,omitempty"`
	CreatedBy     uuid.UUID `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

type CrossGangLink struct {
	LinkID     uuid.UUID `json:"link_id"`
	MemberAID  uuid.UUID `json:"member_a_id"`
	MemberBID  uuid.UUID `json:"member_b_id"`
	LinkType   string    `json:"link_type,omitempty"`
	Confidence *int16    `json:"confidence,omitempty"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type Sighting struct {
	SightingID   uuid.UUID `json:"sighting_id"`
	MemberID     uuid.UUID `json:"member_id"`
	SightedAt    time.Time `json:"sighted_at"`
	LocationDesc string    `json:"location_desc,omitempty"`
	DeptCode     string    `json:"dept_code,omitempty"`
	Commune      string    `json:"commune,omitempty"`
	Lat          *float64  `json:"lat,omitempty"`
	Lng          *float64  `json:"lng,omitempty"`
	SourceType   string    `json:"source_type,omitempty"`
	Confidence   *int16    `json:"confidence,omitempty"`
	PhotoRef     string    `json:"photo_ref,omitempty"`
	ReportedBy   uuid.UUID `json:"reported_by"`
	CreatedAt    time.Time `json:"created_at"`
}

type StatusChangedEvent struct {
	MemberID  uuid.UUID  `json:"member_id"`
	GangID    uuid.UUID  `json:"gang_id"`
	OldStatus ChefStatus `json:"old_status"`
	NewStatus ChefStatus `json:"new_status"`
	ChangedBy uuid.UUID  `json:"changed_by"`
	Notes     string     `json:"notes"`
}

type MemberArrestedEvent struct {
	MemberID   uuid.UUID    `json:"member_id"`
	PersonID   uuid.UUID    `json:"person_id"`
	GangID     uuid.UUID    `json:"gang_id"`
	RoleInGang ChefRoleType `json:"role_in_gang"`
}

type CreateMemberRequest struct {
	SNISIDPersonID  uuid.UUID    `json:"snisid_person_id" binding:"required"`
	FIRRecordID     *uuid.UUID   `json:"fir_record_id,omitempty"`
	AFISSubjectID   *uuid.UUID   `json:"afis_subject_id,omitempty"`
	RDePDeporteeID  *uuid.UUID   `json:"rdep_deportee_id,omitempty"`
	PrimaryGangID   uuid.UUID    `json:"primary_gang_id" binding:"required"`
	RoleInGang      ChefRoleType `json:"role_in_gang" binding:"required"`
	RoleDescription string       `json:"role_description,omitempty"`
	Aliases         []string     `json:"aliases,omitempty"`
	TerritoryDept   string       `json:"territory_dept,omitempty"`
	KnownArmed      bool         `json:"known_armed"`
	Status          ChefStatus   `json:"status"`
	CreatedBy       uuid.UUID    `json:"created_by" binding:"required"`
}

type UpdateStatusRequest struct {
	Status ChefStatus `json:"status" binding:"required"`
	Notes  string     `json:"notes,omitempty"`
}

type CreateIntelNoteRequest struct {
	IntelType     string `json:"intel_type" binding:"required"`
	Content       string `json:"content" binding:"required"`
	SourceClassif string `json:"source_classif,omitempty"`
}

type CreateSightingRequest struct {
	SightedAt    time.Time `json:"sighted_at" binding:"required"`
	LocationDesc string    `json:"location_desc,omitempty"`
	DeptCode     string    `json:"dept_code,omitempty"`
	Commune      string    `json:"commune,omitempty"`
	Lat          *float64  `json:"lat,omitempty"`
	Lng          *float64  `json:"lng,omitempty"`
	SourceType   string    `json:"source_type" binding:"required"`
	Confidence   *int16    `json:"confidence,omitempty"`
	PhotoRef     string    `json:"photo_ref,omitempty"`
}
