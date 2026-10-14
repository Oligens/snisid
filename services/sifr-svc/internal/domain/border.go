package domain

import (
	"time"

	"github.com/google/uuid"
)

type BorderPost struct {
	PostID              uuid.UUID  `json:"post_id" db:"post_id"`
	PostCode            string     `json:"post_code" db:"post_code"`
	Name                string     `json:"name" db:"name"`
	DeptCode            string     `json:"dept_code" db:"dept_code"`
	BorderCountry       string     `json:"border_country" db:"border_country"`
	PostLat             float64    `json:"post_lat" db:"post_lat"`
	PostLng             float64    `json:"post_lng" db:"post_lng"`
	IsOfficial          bool       `json:"is_official" db:"is_official"`
	IsActive            bool       `json:"is_active" db:"is_active"`
	LanesCount          int        `json:"lanes_count" db:"lanes_count"`
	HasBiometricScanner bool       `json:"has_biometric_scanner" db:"has_biometric_scanner"`
	HasVehicleScanner   bool       `json:"has_vehicle_scanner" db:"has_vehicle_scanner"`
	OperatingHours      string     `json:"operating_hours" db:"operating_hours"`
	CommandingOfficer   *uuid.UUID `json:"commanding_officer,omitempty" db:"commanding_officer"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
}

type Crossing struct {
	CrossingID          uuid.UUID         `json:"crossing_id" db:"crossing_id"`
	PostID              uuid.UUID         `json:"post_id" db:"post_id"`
	Direction           CrossingDirection `json:"direction" db:"direction"`
	CrossingDatetime    time.Time         `json:"crossing_datetime" db:"crossing_datetime"`
	SNISIDPersonID      *uuid.UUID        `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	DocumentType        DocType           `json:"document_type" db:"document_type"`
	DocumentNumber      string            `json:"document_number" db:"document_number"`
	DocumentCountry     string            `json:"document_country" db:"document_country"`
	DocumentExpiry      *time.Time        `json:"document_expiry,omitempty" db:"document_expiry"`
	TravelerName        string            `json:"traveler_name" db:"traveler_name"`
	TravelerDob         *time.Time        `json:"traveler_dob,omitempty" db:"traveler_dob"`
	TravelerNationality string            `json:"traveler_nationality" db:"traveler_nationality"`
	VehiclePlate        string            `json:"vehicle_plate" db:"vehicle_plate"`
	LaneNumber          int               `json:"lane_number" db:"lane_number"`
	ProcessingOfficer   uuid.UUID         `json:"processing_officer" db:"processing_officer"`
	AlertTriggered      bool              `json:"alert_triggered" db:"alert_triggered"`
	AlertType           *AlertType        `json:"alert_type,omitempty" db:"alert_type"`
	AlertActionTaken    string            `json:"alert_action_taken" db:"alert_action_taken"`
	ProcessingTimeSec   int               `json:"processing_time_sec" db:"processing_time_sec"`
	CreatedAt           time.Time         `json:"created_at" db:"created_at"`
}

type AlertLog struct {
	AlertID        uuid.UUID  `json:"alert_id" db:"alert_id"`
	CrossingID     uuid.UUID  `json:"crossing_id" db:"crossing_id"`
	PostID         uuid.UUID  `json:"post_id" db:"post_id"`
	AlertType      AlertType  `json:"alert_type" db:"alert_type"`
	SNISIDPersonID *uuid.UUID `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	DocumentNumber string     `json:"document_number" db:"document_number"`
	VehiclePlate   string     `json:"vehicle_plate" db:"vehicle_plate"`
	AlertSource    string     `json:"alert_source" db:"alert_source"`
	SourceRecordID *uuid.UUID `json:"source_record_id,omitempty" db:"source_record_id"`
	NotifiedUnits  []string   `json:"notified_units" db:"notified_units"`
	ActionTaken    string     `json:"action_taken" db:"action_taken"`
	Resolved       bool       `json:"resolved" db:"resolved"`
	ResolvedBy     *uuid.UUID `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type ClandestineCrossing struct {
	ReportID         uuid.UUID  `json:"report_id" db:"report_id"`
	LocationDesc     string     `json:"location_desc" db:"location_desc"`
	DeptCode         string     `json:"dept_code" db:"dept_code"`
	Lat              float64    `json:"lat" db:"lat"`
	Lng              float64    `json:"lng" db:"lng"`
	ReportedDate     time.Time  `json:"reported_date" db:"reported_date"`
	CrossingType     string     `json:"crossing_type" db:"crossing_type"`
	EstimatedPersons int        `json:"estimated_persons" db:"estimated_persons"`
	GangRelated      bool       `json:"gang_related" db:"gang_related"`
	GangID           *uuid.UUID `json:"gang_id,omitempty" db:"gang_id"`
	TraffickingType  string     `json:"trafficking_type" db:"trafficking_type"`
	ReportedBy       uuid.UUID  `json:"reported_by" db:"reported_by"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

type CrossingRequest struct {
	PostID              uuid.UUID  `json:"post_id" binding:"required"`
	Direction           string     `json:"direction" binding:"required"`
	CrossingDatetime    *time.Time `json:"crossing_datetime"`
	SNISIDPersonID      *uuid.UUID `json:"snisid_person_id"`
	DocumentType        string     `json:"document_type" binding:"required"`
	DocumentNumber      string     `json:"document_number"`
	DocumentCountry     string     `json:"document_country"`
	DocumentExpiry      *time.Time `json:"document_expiry"`
	TravelerName        string     `json:"traveler_name" binding:"required"`
	TravelerDob         *time.Time `json:"traveler_dob"`
	TravelerNationality string     `json:"traveler_nationality"`
	VehiclePlate        string     `json:"vehicle_plate"`
	LaneNumber          int        `json:"lane_number"`
	ProcessingOfficer   uuid.UUID  `json:"processing_officer" binding:"required"`
}

type CrossingResult struct {
	Clearance   bool       `json:"clearance"`
	AlertType   *AlertType `json:"alert_type,omitempty"`
	AlertSource string     `json:"alert_source,omitempty"`
	IsDangerous bool       `json:"is_dangerous"`
}

type BorderRepository interface {
	CreateCrossing(crossing *Crossing) (*Crossing, error)
	FindCrossingsByPerson(personID uuid.UUID) ([]Crossing, error)
	FindCrossingsByPost(postID *uuid.UUID, limit, offset int) ([]Crossing, int, error)
	FindActiveAlerts() ([]AlertLog, error)
	GetBorderPosts() ([]BorderPost, error)
	CreateClandestineReport(report *ClandestineCrossing) (*ClandestineCrossing, error)
	GetDailyStats(postID *uuid.UUID) (map[string]interface{}, error)
}
