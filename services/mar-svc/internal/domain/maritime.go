package domain

import (
	"time"

	"github.com/google/uuid"
)

type Vessel struct {
	VesselID           uuid.UUID    `json:"vessel_id" db:"vessel_id"`
	NationalMarID      string       `json:"national_mar_id" db:"national_mar_id"`
	VesselName         string       `json:"vessel_name" db:"vessel_name"`
	IMONumber          string       `json:"imo_number" db:"imo_number"`
	MMSI               string       `json:"mmsi" db:"mmsi"`
	CallSign           string       `json:"call_sign" db:"call_sign"`
	VesselType         VesselType   `json:"vessel_type" db:"vessel_type"`
	FlagCountry        string       `json:"flag_country" db:"flag_country"`
	HullColor          string       `json:"hull_color" db:"hull_color"`
	LengthM            float64      `json:"length_m" db:"length_m"`
	TonnageGT          int          `json:"tonnage_gt" db:"tonnage_gt"`
	EngineCount        int          `json:"engine_count" db:"engine_count"`
	Horsepower         int          `json:"horsepower" db:"horsepower"`
	OwnerName          string       `json:"owner_name" db:"owner_name"`
	OwnerSnisidID      *uuid.UUID   `json:"owner_snisid_id,omitempty" db:"owner_snisid_id"`
	RegistrationNumber string       `json:"registration_number" db:"registration_number"`
	RegistrationPort   string       `json:"registration_port" db:"registration_port"`
	Status             VesselStatus `json:"status" db:"status"`
	GangID             *uuid.UUID   `json:"gang_id,omitempty" db:"gang_id"`
	InterpolSVDRef     string       `json:"interpol_svd_ref" db:"interpol_svd_ref"`
	Notes              string       `json:"notes" db:"notes"`
	CreatedAt          time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at" db:"updated_at"`
}

type AISSighting struct {
	SightingID        uuid.UUID  `json:"sighting_id" db:"sighting_id"`
	VesselID          *uuid.UUID `json:"vessel_id,omitempty" db:"vessel_id"`
	MMSI              string     `json:"mmsi" db:"mmsi"`
	VesselName        string     `json:"vessel_name" db:"vessel_name"`
	SightingTimestamp time.Time  `json:"sighting_timestamp" db:"sighting_timestamp"`
	Lat               float64    `json:"lat" db:"lat"`
	Lng               float64    `json:"lng" db:"lng"`
	SpeedKnots        float64    `json:"speed_knots" db:"speed_knots"`
	HeadingDegrees    int        `json:"heading_degrees" db:"heading_degrees"`
	Destination       string     `json:"destination" db:"destination"`
	SourceType        string     `json:"source_type" db:"source_type"`
	ZoneCode          string     `json:"zone_code" db:"zone_code"`
	AlertTriggered    bool       `json:"alert_triggered" db:"alert_triggered"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

type Incident struct {
	IncidentID      uuid.UUID    `json:"incident_id" db:"incident_id"`
	VesselID        *uuid.UUID   `json:"vessel_id,omitempty" db:"vessel_id"`
	IncidentType    IncidentType `json:"incident_type" db:"incident_type"`
	IncidentDate    time.Time    `json:"incident_date" db:"incident_date"`
	Lat             float64      `json:"lat" db:"lat"`
	Lng             float64      `json:"lng" db:"lng"`
	ZoneDesc        string       `json:"zone_desc" db:"zone_desc"`
	RespondingUnit  string       `json:"responding_unit" db:"responding_unit"`
	Outcome         string       `json:"outcome" db:"outcome"`
	PersonsInvolved int          `json:"persons_involved" db:"persons_involved"`
	SnisidPersonIds []uuid.UUID  `json:"snisid_person_ids" db:"snisid_person_ids"`
	DrugTypes       []string     `json:"drug_types" db:"drug_types"`
	DrugWeightKg    float64      `json:"drug_weight_kg" db:"drug_weight_kg"`
	WeaponsFound    bool         `json:"weapons_found" db:"weapons_found"`
	WeaponsCount    int          `json:"weapons_count" db:"weapons_count"`
	MigrantsCount   int          `json:"migrants_count" db:"migrants_count"`
	BiarRefs        []uuid.UUID  `json:"biar_refs" db:"biar_refs"`
	CaseReference   string       `json:"case_reference" db:"case_reference"`
	PhotoRefs       []string     `json:"photo_refs" db:"photo_refs"`
	CreatedBy       uuid.UUID    `json:"created_by" db:"created_by"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
}

type WatchVessel struct {
	WatchID        uuid.UUID  `json:"watch_id" db:"watch_id"`
	VesselID       *uuid.UUID `json:"vessel_id,omitempty" db:"vessel_id"`
	MMSI           string     `json:"mmsi" db:"mmsi"`
	VesselName     string     `json:"vessel_name" db:"vessel_name"`
	WatchReason    string     `json:"watch_reason" db:"watch_reason"`
	AlertLevel     string     `json:"alert_level" db:"alert_level"`
	RequestingUnit string     `json:"requesting_unit" db:"requesting_unit"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	ExpiryDate     *time.Time `json:"expiry_date,omitempty" db:"expiry_date"`
	CreatedBy      uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type AISMessage struct {
	MMSI           string  `json:"mmsi" binding:"required"`
	VesselName     string  `json:"vessel_name"`
	Lat            float64 `json:"lat" binding:"required"`
	Lng            float64 `json:"lng" binding:"required"`
	SpeedKnots     float64 `json:"speed_knots"`
	HeadingDegrees int     `json:"heading_degrees"`
	Destination    string  `json:"destination"`
	SourceType     string  `json:"source_type"`
}

type MaritimeAlert struct {
	AlertID   uuid.UUID `json:"alert_id"`
	VesselID  uuid.UUID `json:"vessel_id"`
	AlertType string    `json:"alert_type"`
	Message   string    `json:"message"`
	ZoneCode  string    `json:"zone_code"`
	CreatedAt time.Time `json:"created_at"`
}

type MaritimeRepository interface {
	CreateVessel(v *Vessel) error
	FindVesselByID(id uuid.UUID) (*Vessel, error)
	CreateAISSighting(s *AISSighting) error
	GetLastSighting(mmsi string) (*AISSighting, error)
	GetLiveAIS(limit int) ([]AISSighting, error)
	CreateIncident(i *Incident) error
	GetRecentIncidents(limit int) ([]Incident, error)
	CreateWatch(w *WatchVessel) error
	GetActiveWatches() ([]WatchVessel, error)
	GetIncidentsByZone(zone string, limit int) ([]Incident, error)
	GetIncidentStats() (map[string]int64, error)
}
