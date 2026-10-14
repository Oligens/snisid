package domain

import (
	"time"

	"github.com/google/uuid"
)

type TerritoryZone struct {
	ZoneID               uuid.UUID    `json:"zone_id" db:"zone_id"`
	GangID               uuid.UUID    `json:"gang_id" db:"gang_id"`
	ZoneName             *string      `json:"zone_name,omitempty" db:"zone_name"`
	DeptCode             string       `json:"dept_code" db:"dept_code"`
	Commune              *string      `json:"commune,omitempty" db:"commune"`
	SectionCommunale     *string      `json:"section_communale,omitempty" db:"section_communale"`
	Geom                 string       `json:"geom" db:"geom"`
	AreaKm2              *float64     `json:"area_km2,omitempty" db:"area_km2"`
	CentroidLat          *float64     `json:"centroid_lat,omitempty" db:"centroid_lat"`
	CentroidLng          *float64     `json:"centroid_lng,omitempty" db:"centroid_lng"`
	ControlLevel         ControlLevel `json:"control_level" db:"control_level"`
	EstimatedPopulation  *int         `json:"estimated_population,omitempty" db:"estimated_population"`
	StrategicImportance  *int         `json:"strategic_importance,omitempty" db:"strategic_importance"`
	ControlsNationalRoad bool         `json:"controls_national_road" db:"controls_national_road"`
	RoadNumbers          []string     `json:"road_numbers" db:"road_numbers"`
	ControlsPort         bool         `json:"controls_port" db:"controls_port"`
	ControlsAirport      bool         `json:"controls_airport" db:"controls_airport"`
	ControlsMarket       bool         `json:"controls_market" db:"controls_market"`
	ValidFrom            time.Time    `json:"valid_from" db:"valid_from"`
	ValidTo              *time.Time   `json:"valid_to,omitempty" db:"valid_to"`
	IsCurrent            bool         `json:"is_current" db:"is_current"`
	IntelligenceSource   Source       `json:"intelligence_source" db:"intelligence_source"`
	ConfidenceLevel      *int         `json:"confidence_level,omitempty" db:"confidence_level"`
	AnalystNotes         *string      `json:"analyst_notes,omitempty" db:"analyst_notes"`
	CreatedBy            uuid.UUID    `json:"created_by" db:"created_by"`
	CreatedAt            time.Time    `json:"created_at" db:"created_at"`
}

type ZoneHistory struct {
	HistoryID       uuid.UUID     `json:"history_id" db:"history_id"`
	ZoneID          uuid.UUID     `json:"zone_id" db:"zone_id"`
	ChangeType      string        `json:"change_type" db:"change_type"`
	PreviousControl *ControlLevel `json:"previous_control,omitempty" db:"previous_control"`
	NewControl      *ControlLevel `json:"new_control,omitempty" db:"new_control"`
	ChangeDate      time.Time     `json:"change_date" db:"change_date"`
	TriggerEvent    *string       `json:"trigger_event,omitempty" db:"trigger_event"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
}

type Checkpoint struct {
	CheckpointID  uuid.UUID `json:"checkpoint_id" db:"checkpoint_id"`
	GangID        uuid.UUID `json:"gang_id" db:"gang_id"`
	Location      string    `json:"location" db:"location"`
	LocationDesc  *string   `json:"location_desc,omitempty" db:"location_desc"`
	DeptCode      *string   `json:"dept_code,omitempty" db:"dept_code"`
	RoadNumber    *string   `json:"road_number,omitempty" db:"road_number"`
	IsArmed       bool      `json:"is_armed" db:"is_armed"`
	ExtortionType *string   `json:"extortion_type,omitempty" db:"extortion_type"`
	ReportedAt    time.Time `json:"reported_at" db:"reported_at"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type SafetyCheckResult struct {
	IsSafe            bool         `json:"is_safe"`
	Lat               float64      `json:"lat"`
	Lng               float64      `json:"lng"`
	ContainingZone    *ZoneInfo    `json:"containing_zone,omitempty"`
	NearbyZones       []ZoneInfo   `json:"nearby_zones,omitempty"`
	IsInCheckpoint    bool         `json:"is_in_checkpoint"`
	NearbyCheckpoints []Checkpoint `json:"nearby_checkpoints,omitempty"`
}

type RouteSafetyResult struct {
	IsSafe       bool                `json:"is_safe"`
	TotalPoints  int                 `json:"total_points"`
	SafePoints   int                 `json:"safe_points"`
	UnsafePoints int                 `json:"unsafe_points"`
	Waypoints    []SafetyCheckResult `json:"waypoints"`
}

type ZoneInfo struct {
	ZoneID       uuid.UUID    `json:"zone_id"`
	ZoneName     *string      `json:"zone_name,omitempty"`
	GangID       uuid.UUID    `json:"gang_id"`
	ControlLevel ControlLevel `json:"control_level"`
	DeptCode     string       `json:"dept_code"`
	AreaKm2      *float64     `json:"area_km2,omitempty"`
}

type Point struct {
	Lat float64 `json:"lat" binding:"required"`
	Lng float64 `json:"lng" binding:"required"`
}

type SeizureRequest struct {
	GangID             uuid.UUID    `json:"gang_id" binding:"required"`
	ZoneName           *string      `json:"zone_name,omitempty"`
	DeptCode           string       `json:"dept_code" binding:"required"`
	Commune            *string      `json:"commune,omitempty"`
	SectionCommunale   *string      `json:"section_communale,omitempty"`
	Geom               string       `json:"geom" binding:"required"`
	ControlLevel       ControlLevel `json:"control_level" binding:"required"`
	IntelligenceSource Source       `json:"intelligence_source" binding:"required"`
	ConfidenceLevel    *int         `json:"confidence_level,omitempty"`
	AnalystNotes       *string      `json:"analyst_notes,omitempty"`
	CreatedBy          uuid.UUID    `json:"created_by" binding:"required"`
}

type SyncResult struct {
	ZonesSynced       int `json:"zones_synced"`
	CheckpointsSynced int `json:"checkpoints_synced"`
}

type Repository interface {
	FindAllZones() ([]TerritoryZone, error)
	FindZonesContainingPoint(lat, lng float64) ([]TerritoryZone, error)
	FindNearbyCheckpoints(lat, lng float64, radiusMeters float64) ([]Checkpoint, error)
	FindZonesByDept(deptCode string) ([]TerritoryZone, error)
	FindZonesByGang(gangID uuid.UUID) ([]TerritoryZone, error)
	CreateZone(zone *TerritoryZone) (*TerritoryZone, error)
	UpdateZone(zone *TerritoryZone) (*TerritoryZone, error)
	GetZoneHistory(zoneID uuid.UUID) ([]ZoneHistory, error)
	CreateCheckpoint(cp *Checkpoint) (*Checkpoint, error)
}
