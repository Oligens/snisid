package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	PoliticalProtest   EventType = "POLITICAL_PROTEST"
	LaborStrike        EventType = "LABOR_STRIKE"
	CommunityAction    EventType = "COMMUNITY_ACTION"
	ReligiousGathering EventType = "RELIGIOUS_GATHERING"
	CulturalEvent      EventType = "CULTURAL_EVENT"
	PeyiLokBarricade   EventType = "PEYI_LOK_BARRICADE"
	GangMobilization   EventType = "GANG_MOBILIZATION"
	SpontaneousUnrest  EventType = "SPONTANEOUS_UNREST"
	OtherEvent         EventType = "OTHER"
)

type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskModerate RiskLevel = "MODERATE"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

type Event struct {
	ID              uuid.UUID  `json:"event_id" db:"event_id"`
	NationalMvsmID  string     `json:"national_mvsm_id" db:"national_mvsm_id"`
	EventType       EventType  `json:"event_type" db:"event_type"`
	EventName       *string    `json:"event_name,omitempty" db:"event_name"`
	RiskLevel       RiskLevel  `json:"risk_level" db:"risk_level"`
	Status          *string    `json:"status,omitempty" db:"status"`
	OrganizerName   *string    `json:"organizer_name,omitempty" db:"organizer_name"`
	GangID          *uuid.UUID `json:"gang_id,omitempty" db:"gang_id"`
	ScheduledDate   time.Time  `json:"scheduled_date" db:"scheduled_date"`
	ActualStart     *time.Time `json:"actual_start,omitempty" db:"actual_start"`
	ActualEnd       *time.Time `json:"actual_end,omitempty" db:"actual_end"`
	LocationDesc    *string    `json:"location_desc,omitempty" db:"location_desc"`
	DeptCode        *string    `json:"dept_code,omitempty" db:"dept_code"`
	Commune         *string    `json:"commune,omitempty" db:"commune"`
	Lat             *float64   `json:"lat,omitempty" db:"lat"`
	Lng             *float64   `json:"lng,omitempty" db:"lng"`
	EstimatedCrowd  *int       `json:"estimated_crowd,omitempty" db:"estimated_crowd"`
	PeakCrowd       *int       `json:"peak_crowd,omitempty" db:"peak_crowd"`
	IncidentsDuring *int       `json:"incidents_during,omitempty" db:"incidents_during"`
	Casualties      *int       `json:"casualties,omitempty" db:"casualties"`
	ArrestsMade     *int       `json:"arrests_made,omitempty" db:"arrests_made"`
	WeaponsFound    *int       `json:"weapons_found,omitempty" db:"weapons_found"`
	CreatedBy       uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type RealTimeUpdate struct {
	UpdateID        uuid.UUID  `json:"update_id" db:"update_id"`
	EventID         uuid.UUID  `json:"event_id" db:"event_id"`
	UpdateTime      time.Time  `json:"update_time" db:"update_time"`
	CurrentCrowdEst *int       `json:"current_crowd_est,omitempty" db:"current_crowd_est"`
	Situation       string     `json:"situation" db:"situation"`
	RiskChange      *RiskLevel `json:"risk_change,omitempty" db:"risk_change"`
	ActionTaken     *string    `json:"action_taken,omitempty" db:"action_taken"`
	ReportedBy      uuid.UUID  `json:"reported_by" db:"reported_by"`
	Lat             *float64   `json:"lat,omitempty" db:"lat"`
	Lng             *float64   `json:"lng,omitempty" db:"lng"`
}

type CreateEventRequest struct {
	EventType      string   `json:"event_type" binding:"required"`
	EventName      string   `json:"event_name"`
	ScheduledDate  string   `json:"scheduled_date" binding:"required"`
	LocationDesc   string   `json:"location_desc"`
	DeptCode       string   `json:"dept_code"`
	Commune        string   `json:"commune"`
	Lat            *float64 `json:"lat"`
	Lng            *float64 `json:"lng"`
	EstimatedCrowd *int     `json:"estimated_crowd"`
	OrganizerName  string   `json:"organizer_name"`
}

type AddUpdateRequest struct {
	Situation       string   `json:"situation" binding:"required"`
	CurrentCrowdEst *int     `json:"current_crowd_est"`
	RiskChange      string   `json:"risk_change"`
	ActionTaken     string   `json:"action_taken"`
	Lat             *float64 `json:"lat"`
	Lng             *float64 `json:"lng"`
}

type UpdateRiskRequest struct {
	RiskLevel string `json:"risk_level" binding:"required"`
}

type EventRepository interface {
	Create(event *Event) (*Event, error)
	FindUpcoming() ([]Event, error)
	FindActive() ([]Event, error)
	AddUpdate(update *RealTimeUpdate) error
	UpdateRiskLevel(id uuid.UUID, level RiskLevel) error
}
