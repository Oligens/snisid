package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ProtectionOrder struct {
	OrderID              uuid.UUID   `json:"order_id"`
	OrderNumber          string      `json:"order_number"`
	OrderType            OrderType   `json:"order_type"`
	Status               OrderStatus `json:"status"`
	ProtectedPersonID    uuid.UUID   `json:"protected_person_id"`
	SubjectPersonID      uuid.UUID   `json:"subject_person_id"`
	SubjectFIRID         *uuid.UUID  `json:"subject_fir_id,omitempty"`
	ExclusionRadiusM     int         `json:"exclusion_radius_m"`
	ExclusionAddresses   []string    `json:"exclusion_addresses"`
	NoContactModes       []string    `json:"no_contact_modes"`
	GeographicBanGeoJSON string      `json:"geographic_ban_geojson"`
	IssuingCourt         string      `json:"issuing_court"`
	IssuingJudge         string      `json:"issuing_judge"`
	IssueDate            time.Time   `json:"issue_date"`
	ExpiryDate           time.Time   `json:"expiry_date"`
	IsRenewable          bool        `json:"is_renewable"`
	ViolationCount       int         `json:"violation_count"`
	LastViolationAt      *time.Time  `json:"last_violation_at,omitempty"`
	CreatedBy            uuid.UUID   `json:"created_by"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
}

type ProtectionOrderRepository interface {
	Create(ctx context.Context, order *ProtectionOrder) error
	FindByID(ctx context.Context, id uuid.UUID) (*ProtectionOrder, error)
	FindActiveBySubject(ctx context.Context, personID uuid.UUID) ([]*ProtectionOrder, error)
	FindExpiringSoon(ctx context.Context, days int) ([]*ProtectionOrder, error)
	FindByGangID(ctx context.Context, gangID uuid.UUID) ([]*ProtectionOrder, error)
	Update(ctx context.Context, order *ProtectionOrder) error
}

type EventPublisher interface {
	Publish(topic string, event interface{}) error
}

type OPRCheckResult struct {
	HasActiveOrder bool               `json:"has_active_order"`
	Orders         []*ProtectionOrder `json:"orders,omitempty"`
	HighestType    OrderType          `json:"highest_type,omitempty"`
}

type ViolationRequest struct {
	OrderID       uuid.UUID `json:"order_id" binding:"required"`
	ViolationType string    `json:"violation_type" binding:"required"`
	LocationDesc  string    `json:"location_desc"`
	DeptCode      string    `json:"dept_code"`
}

type ViolationEvent struct {
	OrderID    uuid.UUID `json:"order_id"`
	PersonID   uuid.UUID `json:"person_id"`
	ViolType   string    `json:"viol_type"`
	ReportedBy uuid.UUID `json:"reported_by"`
}

type WarrantRequestEvent struct {
	PersonID uuid.UUID `json:"person_id"`
	Reason   string    `json:"reason"`
}
