package domain

import (
	"time"

	"github.com/google/uuid"
)

type Territory struct {
	TerritoryID     uuid.UUID       `json:"territory_id"`
	GangID          uuid.UUID       `json:"gang_id"`
	DeptCode        string          `json:"dept_code"`
	Commune         string          `json:"commune"`
	Locality        *string         `json:"locality,omitempty"`
	GeoJSON         *map[string]any `json:"geojson,omitempty"`
	IsClaimed       bool            `json:"is_claimed"`
	IsContested     bool            `json:"is_contested"`
	ContestedWith   []uuid.UUID     `json:"contested_with"`
	ControlledSince *time.Time      `json:"controlled_since,omitempty"`
	Notes           *string         `json:"notes,omitempty"`
	CreatedBy       uuid.UUID       `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type CreateTerritoryRequest struct {
	GangID          uuid.UUID   `json:"gang_id" validate:"required"`
	DeptCode        string      `json:"dept_code" validate:"required,len=2"`
	Commune         string      `json:"commune" validate:"required"`
	Locality        *string     `json:"locality"`
	IsClaimed       *bool       `json:"is_claimed"`
	IsContested     *bool       `json:"is_contested"`
	ContestedWith   []uuid.UUID `json:"contested_with"`
	ControlledSince *time.Time  `json:"controlled_since"`
	Notes           *string     `json:"notes"`
}
