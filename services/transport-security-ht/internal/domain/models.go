package domain

import (
	"time"

	"github.com/google/uuid"
)

type ScreeningResult string

const (
	ScreeningResultClear    ScreeningResult = "CLEAR"
	ScreeningResultReferral ScreeningResult = "REFERRAL"
	ScreeningResultDenied   ScreeningResult = "DENIED"
	ScreeningResultArrest   ScreeningResult = "ARREST"
)

type DocumentType string

const (
	DocumentTypePassport   DocumentType = "PASSPORT"
	DocumentTypeNationalID DocumentType = "NATIONAL_ID"
	DocumentTypeVisa       DocumentType = "VISA"
)

type TravelMode string

const (
	TravelModeAir    TravelMode = "AIR"
	TravelModeMarine TravelMode = "MARINE"
	TravelModeLand   TravelMode = "LAND"
)

type ScreeningPointType string

const (
	ScreeningPointTypeAirport        ScreeningPointType = "AIRPORT"
	ScreeningPointTypeSeaport        ScreeningPointType = "SEAPORT"
	ScreeningPointTypeBorderCrossing ScreeningPointType = "BORDER_CROSSING"
)

type WatchlistType string

const (
	WatchlistTypeNoFly               WatchlistType = "NO_FLY"
	WatchlistTypeAdditionalScreening WatchlistType = "ADDITIONAL_SCREENING"
)

type ZoneType string

const (
	ZoneTypeSterile     ZoneType = "STERILE"
	ZoneTypeSecure      ZoneType = "SECURE"
	ZoneTypePublic      ZoneType = "PUBLIC"
	ZoneTypeOperational ZoneType = "OPERATIONAL"
)

type AccessLevel string

const (
	AccessLevelPublic         AccessLevel = "PUBLIC"
	AccessLevelStaff          AccessLevel = "STAFF"
	AccessLevelCrew           AccessLevel = "CREW"
	AccessLevelAuthorizedOnly AccessLevel = "AUTHORIZED_ONLY"
)

type ZoneStatus string

const (
	ZoneStatusSecure      ZoneStatus = "SECURE"
	ZoneStatusBreached    ZoneStatus = "BREACHED"
	ZoneStatusMaintenance ZoneStatus = "MAINTENANCE"
)

type PassengerScreening struct {
	ScreeningID         uuid.UUID          `json:"screening_id"`
	TravelerIdentityRef string             `json:"traveler_identity_ref"`
	DocumentType        DocumentType       `json:"document_type"`
	DocumentNumber      string             `json:"document_number"`
	Nationality         string             `json:"nationality"`
	TravelMode          TravelMode         `json:"travel_mode"`
	ScreeningPointType  ScreeningPointType `json:"screening_point_type"`
	ScreeningPointName  string             `json:"screening_point_name"`
	FlightNumber        *string            `json:"flight_number,omitempty"`
	VesselName          *string            `json:"vessel_name,omitempty"`
	DepartureAt         time.Time          `json:"departure_at"`
	ArrivalAt           time.Time          `json:"arrival_at"`
	WatchlistMatch      bool               `json:"watchlist_match"`
	WatchlistRef        *uuid.UUID         `json:"watchlist_ref,omitempty"`
	ScreeningResult     ScreeningResult    `json:"screening_result"`
	ScreeningOfficer    string             `json:"screening_officer"`
	ScreenedAt          time.Time          `json:"screened_at"`
}

type NoFlyPassenger struct {
	IdentityRef   string        `json:"identity_ref"`
	ListType      WatchlistType `json:"list_type"`
	AddedBy       uuid.UUID     `json:"added_by"`
	Reason        string        `json:"reason"`
	CourtOrderRef *string       `json:"court_order_ref,omitempty"`
	ExpiresAt     *time.Time    `json:"expires_at,omitempty"`
	InterpolRef   *string       `json:"interpol_ref,omitempty"`
}

type AirportSecurityZone struct {
	ZoneID          uuid.UUID   `json:"zone_id"`
	AirportCode     string      `json:"airport_code"`
	ZoneName        string      `json:"zone_name"`
	ZoneType        ZoneType    `json:"zone_type"`
	AccessLevel     AccessLevel `json:"access_level"`
	CameraCount     int         `json:"camera_count"`
	LastInspectedAt time.Time   `json:"last_inspected_at"`
	Status          ZoneStatus  `json:"status"`
}
