package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	GangID                     uuid.UUID       `json:"gang_id"`
	NationalGangID             string          `json:"national_gang_id"`
	Name                       string          `json:"name"`
	Aliases                    []string        `json:"aliases"`
	StructureType              StructureType   `json:"structure_type"`
	PrimaryActivity            PrimaryActivity `json:"primary_activity"`
	ActivityLevel              ActivityLevel   `json:"activity_level"`
	EstimatedMembers           int             `json:"estimated_members"`
	ArmedMembersPct            int             `json:"armed_members_pct"`
	HeavyWeapons               bool            `json:"heavy_weapons"`
	PrimaryDeptCode            string          `json:"primary_dept_code"`
	TerritoryCommunes          []string        `json:"territory_communes"`
	TerritoryGeoJSON           string          `json:"territory_geojson"`
	EstimatedRevenueUSDMonthly float64         `json:"estimated_revenue_usd_monthly"`
	PrimaryIncomeSources       []string        `json:"primary_income_sources"`
	UNDesignationDate          *time.Time      `json:"un_designation_date,omitempty"`
	OFACDesignation            bool            `json:"ofac_designation"`
	OFACSDNRef                 string          `json:"ofac_sdn_ref"`
	AlliedGangIDs              []uuid.UUID     `json:"allied_gang_ids"`
	RivalGangIDs               []uuid.UUID     `json:"rival_gang_ids"`
	EstablishedDate            *time.Time      `json:"established_date,omitempty"`
	CurrentLeaderID            *uuid.UUID      `json:"current_leader_id,omitempty"`
	IntelConfidence            int             `json:"intel_confidence"`
	LastIntelUpdate            *time.Time      `json:"last_intel_update,omitempty"`
	IsActive                   bool            `json:"is_active"`
	CreatedBy                  uuid.UUID       `json:"created_by"`
	CreatedAt                  time.Time       `json:"created_at"`
	UpdatedAt                  time.Time       `json:"updated_at"`
}

type OrganizationRepository interface {
	Create(ctx context.Context, org *Organization) error
	FindByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	FindAll(ctx context.Context) ([]*Organization, error)
	FindByDeptCode(ctx context.Context, deptCode string) ([]*Organization, error)
	FindSanctioned(ctx context.Context) ([]*Organization, error)
	Update(ctx context.Context, org *Organization) error
}

type EventPublisher interface {
	Publish(topic string, event interface{}) error
}
