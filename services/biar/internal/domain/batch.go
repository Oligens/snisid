package domain

import (
	"time"

	"github.com/google/uuid"
)

type BatchSeizure struct {
	BatchID            uuid.UUID   `json:"batch_id"`
	BatchReference     string      `json:"batch_reference"`
	OperationName      *string     `json:"operation_name,omitempty"`
	SeizureDate        time.Time   `json:"seizure_date"`
	LocationDesc       *string     `json:"location_desc,omitempty"`
	DeptCode           *string     `json:"dept_code,omitempty"`
	TotalWeapons       int         `json:"total_weapons"`
	WeaponIDs          []uuid.UUID `json:"weapon_ids"`
	SeizingUnit        string      `json:"seizing_unit"`
	LeadOfficer        *uuid.UUID  `json:"lead_officer,omitempty"`
	PartneringAgencies []string    `json:"partnering_agencies"`
	Notes              *string     `json:"notes,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
}

type CreateBatchRequest struct {
	OperationName      *string     `json:"operation_name"`
	SeizureDate        time.Time   `json:"seizure_date" validate:"required"`
	LocationDesc       *string     `json:"location_desc"`
	DeptCode           *string     `json:"dept_code"`
	WeaponIDs          []uuid.UUID `json:"weapon_ids" validate:"required"`
	SeizingUnit        string      `json:"seizing_unit" validate:"required"`
	LeadOfficer        *uuid.UUID  `json:"lead_officer"`
	PartneringAgencies []string    `json:"partnering_agencies"`
	Notes              *string     `json:"notes"`
}
