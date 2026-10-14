package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type MissionStatus string

const (
	MissionStatusPlanned    MissionStatus = "PLANNED"
	MissionStatusInProgress MissionStatus = "IN_PROGRESS"
	MissionStatusCompleted  MissionStatus = "COMPLETED"
	MissionStatusCancelled  MissionStatus = "CANCELLED"
)

type MobileUnit struct {
	ID          uuid.UUID      `json:"id"`
	UnitCode    string         `json:"unit_code"`
	TeamMembers pq.StringArray `json:"team_members"`
	Equipment   []byte         `json:"equipment"`
	Location    *string        `json:"location,omitempty"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
}

type Mission struct {
	ID             uuid.UUID     `json:"id"`
	Title          string        `json:"title"`
	Description    *string       `json:"description,omitempty"`
	Status         MissionStatus `json:"status"`
	AssignedUnitID *uuid.UUID    `json:"assigned_unit_id,omitempty"`
	DeptCode       string        `json:"dept_code"`
	StartedAt      *time.Time    `json:"started_at,omitempty"`
	CompletedAt    *time.Time    `json:"completed_at,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
}

type MissionLog struct {
	ID        uuid.UUID `json:"id"`
	MissionID uuid.UUID `json:"mission_id"`
	LoggedBy  uuid.UUID `json:"logged_by"`
	Action    string    `json:"action"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Notes     *string   `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateMissionRequest struct {
	Title          string `json:"title"`
	Description    string `json:"description,omitempty"`
	AssignedUnitID string `json:"assigned_unit_id,omitempty"`
	DeptCode       string `json:"dept_code"`
}

type CreateMissionLogRequest struct {
	LoggedBy  string  `json:"logged_by"`
	Action    string  `json:"action"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Notes     string  `json:"notes,omitempty"`
}

type CoverageStats struct {
	TotalMissions    int     `json:"total_missions"`
	ActiveUnits      int     `json:"active_units"`
	CoverageLat      float64 `json:"coverage_lat"`
	CoverageLng      float64 `json:"coverage_lng"`
	CoverageRadiusKm float64 `json:"coverage_radius_km"`
}
