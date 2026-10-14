package domain

import (
	"time"

	"github.com/google/uuid"
)

type DrugIncident struct {
	IncidentID        uuid.UUID  `json:"incident_id" db:"incident_id"`
	AlertID           uuid.UUID  `json:"alert_id" db:"alert_id"`
	DrugTypes         []string   `json:"drug_types" db:"drug_types"`
	SeizureWeightKg   *float64   `json:"seizure_weight_kg,omitempty" db:"seizure_weight_kg"`
	EstimatedValueUSD *float64   `json:"estimated_value_usd,omitempty" db:"estimated_value_usd"`
	SeizureDate       *time.Time `json:"seizure_date,omitempty" db:"seizure_date"`
	SeizureLocation   *string    `json:"seizure_location,omitempty" db:"seizure_location"`
	SeizureDeptCode   *string    `json:"seizure_dept_code,omitempty" db:"seizure_dept_code"`
	SeizureCommune    *string    `json:"seizure_commune,omitempty" db:"seizure_commune"`
	RouteType         *RouteType `json:"route_type,omitempty" db:"route_type"`
	OriginCountry     *string    `json:"origin_country,omitempty" db:"origin_country"`
	TransitPoints     []string   `json:"transit_points" db:"transit_points"`
	Destination       *string    `json:"destination,omitempty" db:"destination"`
	SuspectedCartel   *string    `json:"suspected_cartel,omitempty" db:"suspected_cartel"`
	BltsCaseNumber    *string    `json:"blts_case_number,omitempty" db:"blts_case_number"`
	InterpolRef       *string    `json:"interpol_ref,omitempty" db:"interpol_ref"`
	ConcealmentMethod *string    `json:"concealment_method,omitempty" db:"concealment_method"`
	Notes             *string    `json:"notes,omitempty" db:"notes"`
	CreatedBy         uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}
