package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Violation struct {
	ViolationID   uuid.UUID `json:"violation_id"`
	OrderID       uuid.UUID `json:"order_id"`
	ViolationDate time.Time `json:"violation_date"`
	ViolationType string    `json:"violation_type"`
	LocationDesc  string    `json:"location_desc"`
	DeptCode      string    `json:"dept_code"`
	ReportedBy    uuid.UUID `json:"reported_by"`
	ArrestMade    bool      `json:"arrest_made"`
	ArrestCaseRef string    `json:"arrest_case_ref"`
	CreatedAt     time.Time `json:"created_at"`
}

type ViolationRepository interface {
	Create(ctx context.Context, violation *Violation) error
	FindByID(ctx context.Context, id uuid.UUID) (*Violation, error)
	FindByOrderID(ctx context.Context, orderID uuid.UUID) ([]*Violation, error)
}

type WitnessProtection struct {
	ProtectionID      uuid.UUID  `json:"protection_id"`
	ProtectedPersonID uuid.UUID  `json:"protected_person_id"`
	ThreatLevel       string     `json:"threat_level"`
	GangID            *uuid.UUID `json:"gang_id,omitempty"`
	AliasAssigned     string     `json:"alias_assigned"`
	AssignedUnit      string     `json:"assigned_unit"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         time.Time  `json:"created_at"`
}

type WitnessProtectionRepository interface {
	Create(ctx context.Context, wp *WitnessProtection) error
	FindByID(ctx context.Context, id uuid.UUID) (*WitnessProtection, error)
	FindByPersonID(ctx context.Context, personID uuid.UUID) (*WitnessProtection, error)
}
