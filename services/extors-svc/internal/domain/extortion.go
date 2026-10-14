package domain

import (
	"time"

	"github.com/google/uuid"
)

type ExtorsCase struct {
	CaseID            uuid.UUID       `json:"case_id" db:"case_id"`
	NationalExtorsID  string          `json:"national_extors_id" db:"national_extors_id"`
	ExtorsType        ExtorsType      `json:"extors_type" db:"extors_type"`
	Status            ExtorsStatus    `json:"status" db:"status"`
	GangID            *uuid.UUID      `json:"gang_id,omitempty" db:"gang_id"`
	GangName          *string         `json:"gang_name,omitempty" db:"gang_name"`
	PerpetratorIDs    []uuid.UUID     `json:"perpetrator_ids" db:"perpetrator_ids"`
	ChefMemberIDs     []uuid.UUID     `json:"chef_member_ids" db:"chef_member_ids"`
	VictimCount       int16           `json:"victim_count" db:"victim_count"`
	VictimSNISIDs     []uuid.UUID     `json:"victim_snisid_ids" db:"victim_snisid_ids"`
	VictimTypes       []string        `json:"victim_types" db:"victim_types"`
	VictimNationality []string        `json:"victim_nationality" db:"victim_nationality"`
	IsForeignerVictim bool            `json:"is_foreigner_victim" db:"is_foreigner_victim"`
	IncidentLocation  *string         `json:"incident_location,omitempty" db:"incident_location"`
	DeptCode          *string         `json:"dept_code,omitempty" db:"dept_code"`
	Commune           *string         `json:"commune,omitempty" db:"commune"`
	Lat               *float64        `json:"lat,omitempty" db:"lat"`
	Lng               *float64        `json:"lng,omitempty" db:"lng"`
	RouteNumber       *string         `json:"route_number,omitempty" db:"route_number"`
	DemandedAmount    *float64        `json:"demanded_amount,omitempty" db:"demanded_amount"`
	DemandedCurrency  *string         `json:"demanded_currency,omitempty" db:"demanded_currency"`
	PaidAmount        *float64        `json:"paid_amount,omitempty" db:"paid_amount"`
	PaidCurrency      *string         `json:"paid_currency,omitempty" db:"paid_currency"`
	PaymentChannel    *PaymentChannel `json:"payment_channel,omitempty" db:"payment_channel"`
	PaymentRef        *string         `json:"payment_ref,omitempty" db:"payment_ref"`
	PaymentDate       *time.Time      `json:"payment_date,omitempty" db:"payment_date"`
	FirstContactDate  time.Time       `json:"first_contact_date" db:"first_contact_date"`
	ResolutionDate    *time.Time      `json:"resolution_date,omitempty" db:"resolution_date"`
	CaseReference     *string         `json:"case_reference,omitempty" db:"case_reference"`
	InvestigatingUnit *string         `json:"investigating_unit,omitempty" db:"investigating_unit"`
	UcrefStrID        *uuid.UUID      `json:"ucref_str_id,omitempty" db:"ucref_str_id"`
	BlanCaseID        *uuid.UUID      `json:"blan_case_id,omitempty" db:"blan_case_id"`
	Notes             *string         `json:"notes,omitempty" db:"notes"`
	CreatedBy         uuid.UUID       `json:"created_by" db:"created_by"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

type RoadTollPoint struct {
	TollID            uuid.UUID  `json:"toll_id" db:"toll_id"`
	GangID            uuid.UUID  `json:"gang_id" db:"gang_id"`
	LocationDesc      string     `json:"location_desc" db:"location_desc"`
	RouteNumber       *string    `json:"route_number,omitempty" db:"route_number"`
	DeptCode          string     `json:"dept_code" db:"dept_code"`
	Commune           *string    `json:"commune,omitempty" db:"commune"`
	Lat               *float64   `json:"lat,omitempty" db:"lat"`
	Lng               *float64   `json:"lng,omitempty" db:"lng"`
	DailyRevenueUSD   *float64   `json:"daily_revenue_usd,omitempty" db:"daily_revenue_usd"`
	VehicleTypesTaxed []string   `json:"vehicle_types_taxed" db:"vehicle_types_taxed"`
	TollRates         *string    `json:"toll_rates,omitempty" db:"toll_rates"`
	ActiveSince       *time.Time `json:"active_since,omitempty" db:"active_since"`
	IsActive          bool       `json:"is_active" db:"is_active"`
	SourceIntel       *string    `json:"source_intel,omitempty" db:"source_intel"`
	LastConfirmedAt   *time.Time `json:"last_confirmed_at,omitempty" db:"last_confirmed_at"`
	CreatedBy         uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

type Negotiation struct {
	NegID           uuid.UUID `json:"neg_id" db:"neg_id"`
	CaseID          uuid.UUID `json:"case_id" db:"case_id"`
	NegotiationDate time.Time `json:"negotiation_date" db:"negotiation_date"`
	ContactMethod   *string   `json:"contact_method,omitempty" db:"contact_method"`
	ContactNumber   *string   `json:"contact_number,omitempty" db:"contact_number"`
	DemandUpdated   *float64  `json:"demand_updated,omitempty" db:"demand_updated"`
	DemandCurrency  *string   `json:"demand_currency,omitempty" db:"demand_currency"`
	PositionUpdate  *string   `json:"position_update,omitempty" db:"position_update"`
	RecordedBy      uuid.UUID `json:"recorded_by" db:"recorded_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type GangRevenueReport struct {
	GangID        uuid.UUID `json:"gang_id"`
	GangName      string    `json:"gang_name"`
	TollRevenue   float64   `json:"toll_revenue"`
	RansomRevenue float64   `json:"ransom_revenue"`
	TotalRevenue  float64   `json:"total_revenue"`
	ActiveTolls   int       `json:"active_tolls"`
	PaidRansoms   int       `json:"paid_ransoms"`
	ActiveRackets int       `json:"active_rackets"`
}

type TypeStats struct {
	ExtorsType ExtorsType `json:"extors_type"`
	Count      int64      `json:"count"`
	PaidTotal  float64    `json:"paid_total"`
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   map[string]interface{} `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type GeoJSONCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type OpenCaseRequest struct {
	ExtorsType        ExtorsType  `json:"extors_type" binding:"required"`
	GangID            *uuid.UUID  `json:"gang_id,omitempty"`
	GangName          *string     `json:"gang_name,omitempty"`
	PerpetratorIDs    []uuid.UUID `json:"perpetrator_ids,omitempty"`
	ChefMemberIDs     []uuid.UUID `json:"chef_member_ids,omitempty"`
	VictimCount       *int16      `json:"victim_count,omitempty"`
	VictimSNISIDs     []uuid.UUID `json:"victim_snisid_ids,omitempty"`
	VictimTypes       []string    `json:"victim_types,omitempty"`
	VictimNationality []string    `json:"victim_nationality,omitempty"`
	IsForeignerVictim bool        `json:"is_foreigner_victim"`
	IncidentLocation  *string     `json:"incident_location,omitempty"`
	DeptCode          *string     `json:"dept_code,omitempty"`
	Commune           *string     `json:"commune,omitempty"`
	Lat               *float64    `json:"lat,omitempty"`
	Lng               *float64    `json:"lng,omitempty"`
	RouteNumber       *string     `json:"route_number,omitempty"`
	DemandedAmount    *float64    `json:"demanded_amount,omitempty"`
	DemandedCurrency  *string     `json:"demanded_currency,omitempty"`
	FirstContactDate  *time.Time  `json:"first_contact_date"`
	CaseReference     *string     `json:"case_reference,omitempty"`
	InvestigatingUnit *string     `json:"investigating_unit,omitempty"`
	Notes             *string     `json:"notes,omitempty"`
	CreatedBy         uuid.UUID   `json:"created_by" binding:"required"`
}

type AddNegotiationRequest struct {
	NegotiationDate *time.Time `json:"negotiation_date" binding:"required"`
	ContactMethod   *string    `json:"contact_method,omitempty"`
	ContactNumber   *string    `json:"contact_number,omitempty"`
	DemandUpdated   *float64   `json:"demand_updated,omitempty"`
	DemandCurrency  *string    `json:"demand_currency,omitempty"`
	PositionUpdate  *string    `json:"position_update,omitempty"`
	RecordedBy      uuid.UUID  `json:"recorded_by" binding:"required"`
}

type CreateTollPointRequest struct {
	GangID            uuid.UUID `json:"gang_id" binding:"required"`
	LocationDesc      string    `json:"location_desc" binding:"required"`
	RouteNumber       *string   `json:"route_number,omitempty"`
	DeptCode          string    `json:"dept_code" binding:"required"`
	Commune           *string   `json:"commune,omitempty"`
	Lat               *float64  `json:"lat,omitempty"`
	Lng               *float64  `json:"lng,omitempty"`
	DailyRevenueUSD   *float64  `json:"daily_revenue_usd,omitempty"`
	VehicleTypesTaxed []string  `json:"vehicle_types_taxed,omitempty"`
	TollRates         *string   `json:"toll_rates,omitempty"`
	SourceIntel       *string   `json:"source_intel,omitempty"`
	CreatedBy         uuid.UUID `json:"created_by" binding:"required"`
}

type Repository interface {
	CreateCase(c *ExtorsCase) (*ExtorsCase, error)
	FindByID(id uuid.UUID) (*ExtorsCase, error)
	AddNegotiation(n *Negotiation) (*Negotiation, error)
	CreateTollPoint(t *RoadTollPoint) (*RoadTollPoint, error)
	FindActiveTollsByGang(gangID uuid.UUID) ([]RoadTollPoint, error)
	FindPaidRansomsByGang(gangID uuid.UUID) ([]ExtorsCase, error)
	FindActiveRacketsByGang(gangID uuid.UUID) ([]ExtorsCase, error)
	GetTollsMap() ([]RoadTollPoint, error)
	GetStatsByType() ([]TypeStats, error)
}
