package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type HealthEvent struct {
	EventID          uuid.UUID  `json:"event_id"`
	InmateID         uuid.UUID  `json:"inmate_id"`
	EventType        string     `json:"event_type"` // INJURY, ILLNESS, DEATH, PSYCHIATRIC
	EventDate        time.Time  `json:"event_date"`
	Description      string     `json:"description"`
	TreatingFacility string     `json:"treating_facility"`
	Outcome          string     `json:"outcome"`
	ReportedBy       *uuid.UUID `json:"reported_by,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type HealthEventRepository interface {
	Create(ctx context.Context, event *HealthEvent) error
	FindByID(ctx context.Context, id uuid.UUID) (*HealthEvent, error)
	FindByInmateID(ctx context.Context, inmateID uuid.UUID) ([]*HealthEvent, error)
}

type ReleaseRequest struct {
	ReleaseType ReleaseType `json:"release_type" binding:"required"`
	Authority   string      `json:"authority" binding:"required"`
}

type InmateReleasedEvent struct {
	InmateID     uuid.UUID   `json:"inmate_id"`
	PersonID     uuid.UUID   `json:"person_id"`
	FacilityCode string      `json:"facility_code"`
	ReleaseType  ReleaseType `json:"release_type"`
	ReleasedAt   time.Time   `json:"released_at"`
	AuthorizedBy uuid.UUID   `json:"authorized_by"`
}

type EscapeAlertEvent struct {
	InmateID     uuid.UUID `json:"inmate_id"`
	PersonID     uuid.UUID `json:"person_id"`
	FacilityCode string    `json:"facility_code"`
	EscapedAt    time.Time `json:"escaped_at"`
}
