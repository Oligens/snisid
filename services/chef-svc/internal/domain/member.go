package domain

import "time"

type CriminalMember struct {
	MemberID            string       `json:"member_id" db:"member_id"`
	NationalChefID      *string      `json:"national_chef_id" db:"national_chef_id"`
	SNISIDPersonID      *string      `json:"snisid_person_id" db:"snisid_person_id"`
	FIRRecordID         *string      `json:"fir_record_id" db:"fir_record_id"`
	AFISSubjectID       *string      `json:"afis_subject_id" db:"afis_subject_id"`
	RDEPDeporteeID      *string      `json:"rdep_deportee_id" db:"rdep_deportee_id"`
	PrimaryGangID       string       `json:"primary_gang_id" db:"primary_gang_id"`
	RoleInGang          RoleType     `json:"role_in_gang" db:"role_in_gang"`
	RoleDescription     *string      `json:"role_description" db:"role_description"`
	JoinedDate          *time.Time   `json:"joined_date" db:"joined_date"`
	RankLevel           *int         `json:"rank_level" db:"rank_level"`
	Aliases             []string     `json:"aliases" db:"aliases"`
	KnownLanguages      []string     `json:"known_languages" db:"known_languages"`
	TattooDescription   *string      `json:"tattoo_description" db:"tattoo_description"`
	PhysicalDescription *string      `json:"physical_description" db:"physical_description"`
	PhotoRefs           []string     `json:"photo_refs" db:"photo_refs"`
	TerritoryDept       *string      `json:"territory_dept" db:"territory_dept"`
	TerritoryCommunes   []string     `json:"territory_communes" db:"territory_communes"`
	KnownArmed          bool         `json:"known_armed" db:"known_armed"`
	WeaponTypes         []string     `json:"weapon_types" db:"weapon_types"`
	TrainedCombatant    bool         `json:"trained_combatant" db:"trained_combatant"`
	Status              MemberStatus `json:"status" db:"status"`
	UNDesignated        bool         `json:"un_designated" db:"un_designated"`
	OFACDesignated      bool         `json:"ofac_designated" db:"ofac_designated"`
	OFACSDNRef          *string      `json:"ofac_sdn_ref" db:"ofac_sdn_ref"`
	InterpolNoticeRef   *string      `json:"interpol_notice_ref" db:"interpol_notice_ref"`
	LastKnownAddress    *string      `json:"last_known_address" db:"last_known_address"`
	LastSeenAt          *time.Time   `json:"last_seen_at" db:"last_seen_at"`
	IntelConfidence     *int         `json:"intel_confidence" db:"intel_confidence"`
	CreatedBy           string       `json:"created_by" db:"created_by"`
	CreatedAt           time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at" db:"updated_at"`
}

type IntelligenceNote struct {
	NoteID      string     `json:"note_id" db:"note_id"`
	MemberID    string     `json:"member_id" db:"member_id"`
	SourceID    *string    `json:"source_id" db:"source_id"`
	NoteType    string     `json:"note_type" db:"note_type"`
	Content     string     `json:"content" db:"content"`
	Confidence  *int       `json:"confidence" db:"confidence"`
	CollectedAt *time.Time `json:"collected_at" db:"collected_at"`
	CreatedBy   string     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type Sighting struct {
	SightingID string    `json:"sighting_id" db:"sighting_id"`
	MemberID   string    `json:"member_id" db:"member_id"`
	SourceID   *string   `json:"source_id" db:"source_id"`
	Dept       string    `json:"dept" db:"dept"`
	Commune    *string   `json:"commune" db:"commune"`
	Latitude   *float64  `json:"latitude" db:"latitude"`
	Longitude  *float64  `json:"longitude" db:"longitude"`
	SpottedAt  time.Time `json:"spotted_at" db:"spotted_at"`
	Confidence *int      `json:"confidence" db:"confidence"`
	Notes      *string   `json:"notes" db:"notes"`
	CreatedBy  string    `json:"created_by" db:"created_by"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type MemberRepository interface {
	Create(m *CriminalMember) error
	GetByID(id string) (*CriminalMember, error)
	GetByGang(gangID string) ([]CriminalMember, error)
	GetSanctioned() ([]CriminalMember, error)
	GetLeaders() ([]CriminalMember, error)
	UpdateStatus(id string, status MemberStatus) error
	Update(m *CriminalMember) error
}

type IntelNoteRepository interface {
	Create(n *IntelligenceNote) error
	GetByMemberID(memberID string) ([]IntelligenceNote, error)
}

type SightingRepository interface {
	Create(s *Sighting) error
	GetByMemberID(memberID string) ([]Sighting, error)
}

type EventPublisher interface {
	PublishEvent(eventType string, payload interface{}) error
}
