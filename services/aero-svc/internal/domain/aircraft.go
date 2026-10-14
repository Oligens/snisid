package domain

import (
	"time"

	"github.com/google/uuid"
)

type Aircraft struct {
	AircraftID       uuid.UUID    `json:"aircraft_id" db:"aircraft_id"`
	RegistrationMark string       `json:"registration_mark" db:"registration_mark"`
	ICAOHexCode      *string      `json:"icao_hex_code,omitempty" db:"icao_hex_code"`
	AircraftType     AircraftType `json:"aircraft_type" db:"aircraft_type"`
	Make             *string      `json:"make,omitempty" db:"make"`
	Model            *string      `json:"model,omitempty" db:"model"`
	ManufactureYear  *int16       `json:"manufacture_year,omitempty" db:"manufacture_year"`
	FlagCountry      *string      `json:"flag_country,omitempty" db:"flag_country"`
	OwnerName        *string      `json:"owner_name,omitempty" db:"owner_name"`
	OwnerSnisidID    *uuid.UUID   `json:"owner_snisid_id,omitempty" db:"owner_snisid_id"`
	OperatorName     *string      `json:"operator_name,omitempty" db:"operator_name"`
	IsRegistered     bool         `json:"is_registered" db:"is_registered"`
	IsSuspected      bool         `json:"is_suspected" db:"is_suspected"`
	IsStolen         bool         `json:"is_stolen" db:"is_stolen"`
	GangID           *uuid.UUID   `json:"gang_id,omitempty" db:"gang_id"`
	DrugTrafficking  bool         `json:"drug_trafficking" db:"drug_trafficking"`
	InterpolRef      *string      `json:"interpol_ref,omitempty" db:"interpol_ref"`
	FAARegistryRef   *string      `json:"faa_registry_ref,omitempty" db:"faa_registry_ref"`
	Notes            *string      `json:"notes,omitempty" db:"notes"`
	CreatedAt        time.Time    `json:"created_at" db:"created_at"`
}

type ClandestineStrip struct {
	StripID           uuid.UUID   `json:"strip_id" db:"strip_id"`
	StripName         string      `json:"strip_name" db:"strip_name"`
	DeptCode          string      `json:"dept_code" db:"dept_code"`
	Commune           *string     `json:"commune,omitempty" db:"commune"`
	Lat               float64     `json:"lat" db:"lat"`
	Lng               float64     `json:"lng" db:"lng"`
	LengthM           *int        `json:"length_m,omitempty" db:"length_m"`
	SurfaceType       *string     `json:"surface_type,omitempty" db:"surface_type"`
	Status            StripStatus `json:"status" db:"status"`
	CapableAircraft   []string    `json:"capable_aircraft" db:"capable_aircraft"`
	GangID            *uuid.UUID  `json:"gang_id,omitempty" db:"gang_id"`
	FirstDetected     *time.Time  `json:"first_detected,omitempty" db:"first_detected"`
	LastActivityDate  *time.Time  `json:"last_activity_date,omitempty" db:"last_activity_date"`
	SourceIntel       *string     `json:"source_intel,omitempty" db:"source_intel"`
	SatelliteImageRef *string     `json:"satellite_image_ref,omitempty" db:"satellite_image_ref"`
	CreatedBy         uuid.UUID   `json:"created_by" db:"created_by"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
}

type SuspiciousFlight struct {
	FlightID           uuid.UUID  `json:"flight_id" db:"flight_id"`
	AircraftID         *uuid.UUID `json:"aircraft_id,omitempty" db:"aircraft_id"`
	RegistrationMark   *string    `json:"registration_mark,omitempty" db:"registration_mark"`
	FlightDate         time.Time  `json:"flight_date" db:"flight_date"`
	OriginAirport      *string    `json:"origin_airport,omitempty" db:"origin_airport"`
	DestinationAirport *string    `json:"destination_airport,omitempty" db:"destination_airport"`
	OriginCountry      *string    `json:"origin_country,omitempty" db:"origin_country"`
	DestinationCountry *string    `json:"destination_country,omitempty" db:"destination_country"`
	LandingStripID     *uuid.UUID `json:"landing_strip_id,omitempty" db:"landing_strip_id"`
	LandingLocation    *string    `json:"landing_location,omitempty" db:"landing_location"`
	FlightType         *string    `json:"flight_type,omitempty" db:"flight_type"`
	CargoSuspected     *string    `json:"cargo_suspected,omitempty" db:"cargo_suspected"`
	SourceRadar        *string    `json:"source_radar,omitempty" db:"source_radar"`
	SourceInformant    bool       `json:"source_informant" db:"source_informant"`
	CaseReference      *string    `json:"case_reference,omitempty" db:"case_reference"`
	CreatedBy          uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

type RegistrationCheckResult struct {
	IsRegistered bool      `json:"is_registered"`
	IsSuspected  bool      `json:"is_suspected"`
	IsStolen     bool      `json:"is_stolen"`
	Aircraft     *Aircraft `json:"aircraft,omitempty"`
}

type StripStats struct {
	TotalStrips  int            `json:"total_strips"`
	ActiveStrips int            `json:"active_strips"`
	ByDepartment map[string]int `json:"by_department"`
	ByStatus     map[string]int `json:"by_status"`
}

type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type GeoJSONGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type ReportStripRequest struct {
	StripName         string   `json:"strip_name" binding:"required"`
	DeptCode          string   `json:"dept_code" binding:"required"`
	Commune           *string  `json:"commune,omitempty"`
	Lat               float64  `json:"lat" binding:"required"`
	Lng               float64  `json:"lng" binding:"required"`
	LengthM           *int     `json:"length_m,omitempty"`
	SurfaceType       *string  `json:"surface_type,omitempty"`
	Status            string   `json:"status"`
	CapableAircraft   []string `json:"capable_aircraft"`
	GangID            *string  `json:"gang_id,omitempty"`
	FirstDetected     *string  `json:"first_detected,omitempty"`
	LastActivityDate  *string  `json:"last_activity_date,omitempty"`
	SourceIntel       *string  `json:"source_intel,omitempty"`
	SatelliteImageRef *string  `json:"satellite_image_ref,omitempty"`
	CreatedBy         string   `json:"created_by" binding:"required"`
}

type ReportFlightRequest struct {
	AircraftID         string `json:"aircraft_id"`
	RegistrationMark   string `json:"registration_mark"`
	FlightDate         string `json:"flight_date" binding:"required"`
	OriginAirport      string `json:"origin_airport"`
	DestinationAirport string `json:"destination_airport"`
	OriginCountry      string `json:"origin_country"`
	DestinationCountry string `json:"destination_country"`
	LandingStripID     string `json:"landing_strip_id"`
	LandingLocation    string `json:"landing_location"`
	FlightType         string `json:"flight_type"`
	CargoSuspected     string `json:"cargo_suspected"`
	SourceRadar        string `json:"source_radar"`
	SourceInformant    bool   `json:"source_informant"`
	CaseReference      string `json:"case_reference"`
	CreatedBy          string `json:"created_by" binding:"required"`
}

type Repository interface {
	CreateAircraft(aircraft *Aircraft) (*Aircraft, error)
	FindByRegistration(mark string) (*Aircraft, error)
	CreateStrip(strip *ClandestineStrip) (*ClandestineStrip, error)
	FindActiveStrips() ([]ClandestineStrip, error)
	GetStripsMap() ([]ClandestineStrip, error)
	CreateFlight(flight *SuspiciousFlight) (*SuspiciousFlight, error)
	GetFlightsByDate(from, to time.Time) ([]SuspiciousFlight, error)
	GetStripStats() (*StripStats, error)
}
