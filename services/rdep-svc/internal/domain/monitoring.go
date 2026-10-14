package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MonitoringEvent struct {
	EventID     uuid.UUID `json:"event_id"`
	DeporteeID  uuid.UUID `json:"deportee_id"`
	EventType   string    `json:"event_type"` // CHECK_IN, VIOLATION, ADDRESS_CHANGE
	EventDate   time.Time `json:"event_date"`
	LocationLat float64   `json:"location_lat"`
	LocationLng float64   `json:"location_lng"`
	Notes       string    `json:"notes"`
	ReportedBy  uuid.UUID `json:"reported_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type MonitoringEventRepository interface {
	Create(ctx context.Context, event *MonitoringEvent) error
	FindByID(ctx context.Context, id uuid.UUID) (*MonitoringEvent, error)
	FindByDeporteeID(ctx context.Context, deporteeID uuid.UUID) ([]*MonitoringEvent, error)
}
