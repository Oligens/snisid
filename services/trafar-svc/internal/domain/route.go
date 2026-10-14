package domain

import (
	"time"

	"github.com/google/uuid"
)

type TrafarRoute struct {
	RouteID                uuid.UUID         `json:"route_id" db:"route_id"`
	RouteName              string            `json:"route_name" db:"route_name"`
	RouteType              RouteType         `json:"route_type" db:"route_type"`
	TraffickingMethod      TraffickingMethod `json:"trafficking_method" db:"trafficking_method"`
	OriginCountry          string            `json:"origin_country" db:"origin_country"`
	OriginCity             *string           `json:"origin_city,omitempty" db:"origin_city"`
	TransitPoints          []string          `json:"transit_points" db:"transit_points"`
	EntryPointHaiti        *string           `json:"entry_point_haiti,omitempty" db:"entry_point_haiti"`
	EntryDeptCode          *string           `json:"entry_dept_code,omitempty" db:"entry_dept_code"`
	AssociatedGangIDs      []uuid.UUID       `json:"associated_gang_ids" db:"associated_gang_ids"`
	KnownSuppliers         []string          `json:"known_suppliers" db:"known_suppliers"`
	ActivityLevel          string            `json:"activity_level" db:"activity_level"`
	EstimatedVolumeMonthly *int              `json:"estimated_volume_monthly,omitempty" db:"estimated_volume_monthly"`
	WeaponTypes            []string          `json:"weapon_types" db:"weapon_types"`
	IntelConfidence        *int              `json:"intel_confidence,omitempty" db:"intel_confidence"`
	FirstDetected          *time.Time        `json:"first_detected,omitempty" db:"first_detected"`
	LastConfirmed          *time.Time        `json:"last_confirmed,omitempty" db:"last_confirmed"`
	LinkedCaseRefs         []string          `json:"linked_case_refs" db:"linked_case_refs"`
	BIARWeaponIDs          []uuid.UUID       `json:"biar_weapon_ids" db:"biar_weapon_ids"`
	ATFCaseRefs            []string          `json:"atf_case_refs" db:"atf_case_refs"`
	UNODCRef               *string           `json:"unodc_ref,omitempty" db:"unodc_ref"`
	AnalystNotes           *string           `json:"analyst_notes,omitempty" db:"analyst_notes"`
	CreatedBy              uuid.UUID         `json:"created_by" db:"created_by"`
	CreatedAt              time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time         `json:"updated_at" db:"updated_at"`
}

type TrafarShipment struct {
	ShipmentID           uuid.UUID   `json:"shipment_id" db:"shipment_id"`
	RouteID              uuid.UUID   `json:"route_id" db:"route_id"`
	ShipmentDate         time.Time   `json:"shipment_date" db:"shipment_date"`
	Intercepted          bool        `json:"intercepted" db:"intercepted"`
	InterceptionDate     *time.Time  `json:"interception_date,omitempty" db:"interception_date"`
	InterceptionLocation *string     `json:"interception_location,omitempty" db:"interception_location"`
	InterceptionUnit     *string     `json:"interception_unit,omitempty" db:"interception_unit"`
	WeaponsCount         *int        `json:"weapons_count,omitempty" db:"weapons_count"`
	WeaponsTypes         []string    `json:"weapon_types" db:"weapon_types"`
	EstimatedValueUSD    *float64    `json:"estimated_value_usd,omitempty" db:"estimated_value_usd"`
	LinkedPersons        []uuid.UUID `json:"linked_persons" db:"linked_persons"`
	PortHTRef            *uuid.UUID  `json:"port_ht_ref,omitempty" db:"port_ht_ref"`
	MARHTRef             *uuid.UUID  `json:"mar_ht_ref,omitempty" db:"mar_ht_ref"`
	CaseReference        *string     `json:"case_reference,omitempty" db:"case_reference"`
	Notes                *string     `json:"notes,omitempty" db:"notes"`
	CreatedAt            time.Time   `json:"created_at" db:"created_at"`
}

type TrafarSupplier struct {
	SupplierID        uuid.UUID   `json:"supplier_id" db:"supplier_id"`
	SupplierName      string      `json:"supplier_name" db:"supplier_name"`
	SupplierType      *string     `json:"supplier_type,omitempty" db:"supplier_type"`
	Country           string      `json:"country" db:"country"`
	City              *string     `json:"city,omitempty" db:"city"`
	SNISIDPersonID    *uuid.UUID  `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	LinkedRoutes      []uuid.UUID `json:"linked_routes" db:"linked_routes"`
	ATFSubjectRef     *string     `json:"atf_subject_ref,omitempty" db:"atf_subject_ref"`
	InterpolNoticeRef *string     `json:"interpol_notice_ref,omitempty" db:"interpol_notice_ref"`
	IsActive          bool        `json:"is_active" db:"is_active"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
}

type RouteRepository interface {
	CreateRoute(route *TrafarRoute) error
	FindByID(id uuid.UUID) (*TrafarRoute, error)
	FindAll() ([]TrafarRoute, error)
	CreateShipment(shipment *TrafarShipment) error
	GetShipmentsByRoute(routeID uuid.UUID) ([]TrafarShipment, error)
	GetStatsByOrigin() ([]map[string]interface{}, error)
	GetSuppliers() ([]TrafarSupplier, error)
}

type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   map[string]interface{} `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}
