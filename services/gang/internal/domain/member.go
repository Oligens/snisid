package domain

import (
	"time"

	"github.com/google/uuid"
)

type Member struct {
	MemberID         uuid.UUID  `json:"member_id"`
	GangID           uuid.UUID  `json:"gang_id"`
	NationalMemberID string     `json:"national_member_id"`
	FullName         string     `json:"full_name"`
	Aliases          []string   `json:"aliases"`
	Role             *string    `json:"role,omitempty"`
	DateOfBirth      *time.Time `json:"date_of_birth,omitempty"`
	PlaceOfBirth     *string    `json:"place_of_birth,omitempty"`
	Nationality      string     `json:"nationality"`
	IDType           *string    `json:"id_type,omitempty"`
	IDNumber         *string    `json:"id_number,omitempty"`
	PhotoRef         *string    `json:"photo_ref,omitempty"`
	FingerprintHash  *string    `json:"fingerprint_hash,omitempty"`
	LastKnownAddress *string    `json:"last_known_address,omitempty"`
	DeptCode         *string    `json:"dept_code,omitempty"`
	Commune          *string    `json:"commune,omitempty"`
	Lat              *float64   `json:"lat,omitempty"`
	Lng              *float64   `json:"lng,omitempty"`
	IsLeader         bool       `json:"is_leader"`
	IsArrested       bool       `json:"is_arrested"`
	ArrestDate       *time.Time `json:"arrest_date,omitempty"`
	ArrestRef        *string    `json:"arrest_ref,omitempty"`
	IsDeceased       bool       `json:"is_deceased"`
	DeathDate        *time.Time `json:"death_date,omitempty"`
	OFACDesignated   bool       `json:"ofac_designated"`
	OFACSDNRef       *string    `json:"ofac_sdn_ref,omitempty"`
	IntelConfidence  *int16     `json:"intel_confidence,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
	CreatedBy        uuid.UUID  `json:"created_by"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CreateMemberRequest struct {
	GangID          uuid.UUID  `json:"gang_id" validate:"required"`
	FullName        string     `json:"full_name" validate:"required"`
	Aliases         []string   `json:"aliases"`
	Role            *string    `json:"role"`
	DateOfBirth     *time.Time `json:"date_of_birth"`
	PlaceOfBirth    *string    `json:"place_of_birth"`
	IDType          *string    `json:"id_type"`
	IDNumber        *string    `json:"id_number"`
	IsLeader        bool       `json:"is_leader"`
	OFACDesignated  bool       `json:"ofac_designated"`
	IntelConfidence *int16     `json:"intel_confidence"`
}
