package domain

import (
	"time"

	"github.com/google/uuid"
)

type DisasterType string

const (
	Earthquake           DisasterType = "EARTHQUAKE"
	Hurricane            DisasterType = "HURRICANE"
	Tsunami              DisasterType = "TSUNAMI"
	Flood                DisasterType = "FLOOD"
	Landslide            DisasterType = "LANDSLIDE"
	FireMass             DisasterType = "FIRE_MASS"
	IndustrialAccident   DisasterType = "INDUSTRIAL_ACCIDENT"
	Epidemic             DisasterType = "EPIDEMIC"
	SecurityMassCasualty DisasterType = "SECURITY_MASS_CASUALTY"
)

type AlertLevel string

const (
	Watch       AlertLevel = "WATCH"
	Warning     AlertLevel = "WARNING"
	Emergency   AlertLevel = "EMERGENCY"
	Catastrophe AlertLevel = "CATASTROPHE"
)

type Disaster struct {
	ID                uuid.UUID    `json:"disaster_id" db:"disaster_id"`
	NationalSigdcID   string       `json:"national_sigdc_id" db:"national_sigdc_id"`
	DisasterType      DisasterType `json:"disaster_type" db:"disaster_type"`
	DisasterName      *string      `json:"disaster_name,omitempty" db:"disaster_name"`
	AlertLevel        AlertLevel   `json:"alert_level" db:"alert_level"`
	Status            *string      `json:"status,omitempty" db:"status"`
	OnsetDate         time.Time    `json:"onset_date" db:"onset_date"`
	AffectedDepts     []string     `json:"affected_depts" db:"affected_depts"`
	EpicenterLat      *float64     `json:"epicenter_lat,omitempty" db:"epicenter_lat"`
	EpicenterLng      *float64     `json:"epicenter_lng,omitempty" db:"epicenter_lng"`
	Magnitude         *float64     `json:"magnitude,omitempty" db:"magnitude"`
	EstimatedAffected *int         `json:"estimated_affected,omitempty" db:"estimated_affected"`
	ConfirmedDead     *int         `json:"confirmed_dead,omitempty" db:"confirmed_dead"`
	ConfirmedInjured  *int         `json:"confirmed_injured,omitempty" db:"confirmed_injured"`
	ConfirmedMissing  *int         `json:"confirmed_missing,omitempty" db:"confirmed_missing"`
	ResponseAgencies  []string     `json:"response_agencies" db:"response_agencies"`
	CreatedAt         time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at" db:"updated_at"`
}

type EarlyWarning struct {
	WarningID     uuid.UUID    `json:"warning_id" db:"warning_id"`
	DisasterType  DisasterType `json:"disaster_type" db:"disaster_type"`
	AlertLevel    AlertLevel   `json:"alert_level" db:"alert_level"`
	SourceAgency  *string      `json:"source_agency,omitempty" db:"source_agency"`
	MessageText   string       `json:"message_text" db:"message_text"`
	AffectedDepts []string     `json:"affected_depts" db:"affected_depts"`
	IssuedAt      time.Time    `json:"issued_at" db:"issued_at"`
	ExpiresAt     *time.Time   `json:"expires_at,omitempty" db:"expires_at"`
	ChannelsSent  []string     `json:"channels_sent" db:"channels_sent"`
}

type VictimRegistration struct {
	RegistrationID   uuid.UUID `json:"registration_id" db:"registration_id"`
	DisasterID       uuid.UUID `json:"disaster_id" db:"disaster_id"`
	FullName         *string   `json:"full_name,omitempty" db:"full_name"`
	Status           string    `json:"status" db:"status"`
	LocationFound    *string   `json:"location_found,omitempty" db:"location_found"`
	DeptCode         *string   `json:"dept_code,omitempty" db:"dept_code"`
	RegistrationDate time.Time `json:"registration_date" db:"registration_date"`
	RegisteredBy     uuid.UUID `json:"registered_by" db:"registered_by"`
}

type Resource struct {
	ResourceID   uuid.UUID `json:"resource_id" db:"resource_id"`
	DisasterID   uuid.UUID `json:"disaster_id" db:"disaster_id"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ProviderOrg  *string   `json:"provider_org,omitempty" db:"provider_org"`
	Quantity     *int      `json:"quantity,omitempty" db:"quantity"`
	DeptCode     *string   `json:"dept_code,omitempty" db:"dept_code"`
	Status       *string   `json:"status,omitempty" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type DeclareDisasterRequest struct {
	DisasterType  string   `json:"disaster_type" binding:"required"`
	DisasterName  string   `json:"disaster_name"`
	AlertLevel    string   `json:"alert_level" binding:"required"`
	OnsetDate     string   `json:"onset_date" binding:"required"`
	AffectedDepts []string `json:"affected_depts"`
	EpicenterLat  *float64 `json:"epicenter_lat"`
	EpicenterLng  *float64 `json:"epicenter_lng"`
	Magnitude     *float64 `json:"magnitude"`
}

type IssueWarningRequest struct {
	DisasterType  string   `json:"disaster_type" binding:"required"`
	AlertLevel    string   `json:"alert_level" binding:"required"`
	SourceAgency  string   `json:"source_agency"`
	MessageText   string   `json:"message_text" binding:"required"`
	AffectedDepts []string `json:"affected_depts"`
}

type RegisterVictimRequest struct {
	DisasterID    string `json:"disaster_id" binding:"required"`
	FullName      string `json:"full_name"`
	Status        string `json:"status" binding:"required"`
	LocationFound string `json:"location_found"`
	DeptCode      string `json:"dept_code"`
}

type DisasterRepository interface {
	CreateDisaster(d *Disaster) (*Disaster, error)
	FindActiveDisasters() ([]Disaster, error)
	SaveWarning(w *EarlyWarning) error
	CreateVictimRegistration(vr *VictimRegistration) (*VictimRegistration, error)
	FindResources(disasterID uuid.UUID) ([]Resource, error)
}
