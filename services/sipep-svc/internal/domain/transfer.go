package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Transfer struct {
	TransferID     uuid.UUID `json:"transfer_id"`
	InmateID       uuid.UUID `json:"inmate_id"`
	FromFacility   string    `json:"from_facility"`
	ToFacility     string    `json:"to_facility"`
	TransferDate   time.Time `json:"transfer_date"`
	TransferReason string    `json:"transfer_reason"`
	AuthorizedBy   uuid.UUID `json:"authorized_by"`
	TransportUnit  string    `json:"transport_unit"`
	CreatedAt      time.Time `json:"created_at"`
}

type TransferRepository interface {
	Create(ctx context.Context, transfer *Transfer) error
	FindByID(ctx context.Context, id uuid.UUID) (*Transfer, error)
	FindByInmateID(ctx context.Context, inmateID uuid.UUID) ([]*Transfer, error)
}
