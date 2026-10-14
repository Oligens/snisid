package domain

import (
	"time"

	"github.com/google/uuid"
)

type GangAssociation struct {
	AssocID               uuid.UUID  `json:"assoc_id" db:"assoc_id"`
	AlertID               uuid.UUID  `json:"alert_id" db:"alert_id"`
	GangIdentifier        *string    `json:"gang_identifier,omitempty" db:"gang_identifier"`
	GangTerritoryDept     *string    `json:"gang_territory_dept,omitempty" db:"gang_territory_dept"`
	GangTerritoryCommunes []string   `json:"gang_territory_communes" db:"gang_territory_communes"`
	GangSnisidID          *uuid.UUID `json:"gang_snisid_id,omitempty" db:"gang_snisid_id"`
	VehicleRole           *string    `json:"vehicle_role,omitempty" db:"vehicle_role"`
	AssociationConfidence *int16     `json:"association_confidence,omitempty" db:"association_confidence"`
	IntelligenceSource    *string    `json:"intelligence_source,omitempty" db:"intelligence_source"`
	SourceClassification  *string    `json:"source_classification,omitempty" db:"source_classification"`
	FirstSeenDate         *time.Time `json:"first_seen_date,omitempty" db:"first_seen_date"`
	LastConfirmedDate     *time.Time `json:"last_confirmed_date,omitempty" db:"last_confirmed_date"`
	Notes                 *string    `json:"notes,omitempty" db:"notes"`
	CreatedBy             uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}
