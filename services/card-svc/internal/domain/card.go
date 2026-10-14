package domain

import (
	"time"

	"github.com/google/uuid"
)

type CardType string

const (
	CardTypeNationalID      CardType = "NATIONAL_ID"
	CardTypeResidencePermit CardType = "RESIDENCE_PERMIT"
	CardTypeDriversLicense  CardType = "DRIVERS_LICENSE"
	CardTypePassport        CardType = "PASSPORT"
	CardTypeEmployeeBadge   CardType = "EMPLOYEE_BADGE"
)

type CardStatus string

const (
	CardStatusOrdered      CardStatus = "ORDERED"
	CardStatusPersonalized CardStatus = "PERSONALIZED"
	CardStatusIssued       CardStatus = "ISSUED"
	CardStatusActive       CardStatus = "ACTIVE"
	CardStatusBlocked      CardStatus = "BLOCKED"
	CardStatusExpired      CardStatus = "EXPIRED"
	CardStatusDestroyed    CardStatus = "DESTROYED"
	CardStatusLost         CardStatus = "LOST"
	CardStatusStolen       CardStatus = "STOLEN"
)

type CardProfile struct {
	ProfileID   uuid.UUID `json:"profile_id"`
	CardType    CardType  `json:"card_type"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	FormFactor  string    `json:"form_factor"`
	Material    string    `json:"material"`
	HasChip     bool      `json:"has_chip"`
	HasMRZ      bool      `json:"has_mrz"`
	ValidDays   int       `json:"valid_days"`
	CreatedAt   time.Time `json:"created_at"`
}

type CardStock struct {
	StockID      uuid.UUID `json:"stock_id"`
	ProfileID    uuid.UUID `json:"profile_id"`
	SerialFrom   string    `json:"serial_from"`
	SerialTo     string    `json:"serial_to"`
	Quantity     int       `json:"quantity"`
	AvailableQty int       `json:"available_qty"`
	Location     string    `json:"location,omitempty"`
	ReceivedAt   time.Time `json:"received_at"`
}

type PersonalizationRequest struct {
	OrderID        uuid.UUID  `json:"order_id"`
	ProfileID      uuid.UUID  `json:"profile_id"`
	CardSerial     string     `json:"card_serial"`
	CitizenID      uuid.UUID  `json:"citizen_id"`
	FullName       string     `json:"full_name"`
	DateOfBirth    string     `json:"date_of_birth"`
	Nationality    string     `json:"nationality"`
	PhotoData      string     `json:"photo_data,omitempty"`
	SignatureData  string     `json:"signature_data,omitempty"`
	Status         CardStatus `json:"status"`
	OrderedAt      time.Time  `json:"ordered_at"`
	PersonalizedAt *time.Time `json:"personalized_at,omitempty"`
	IssuedAt       *time.Time `json:"issued_at,omitempty"`
	ActivatedAt    *time.Time `json:"activated_at,omitempty"`
	BlockedAt      *time.Time `json:"blocked_at,omitempty"`
	BlockReason    string     `json:"block_reason,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CardInventory struct {
	ProfileID    uuid.UUID `json:"profile_id"`
	ProfileName  string    `json:"profile_name"`
	CardType     CardType  `json:"card_type"`
	TotalStock   int       `json:"total_stock"`
	Available    int       `json:"available"`
	Personalized int       `json:"personalized"`
	Issued       int       `json:"issued"`
	Blocked      int       `json:"blocked"`
	Defective    int       `json:"defective"`
}

type Shipment struct {
	ShipmentID  uuid.UUID  `json:"shipment_id"`
	ProfileID   uuid.UUID  `json:"profile_id"`
	SerialFrom  string     `json:"serial_from"`
	SerialTo    string     `json:"serial_to"`
	Quantity    int        `json:"quantity"`
	TrackingRef string     `json:"tracking_ref,omitempty"`
	Vendor      string     `json:"vendor"`
	ReceivedBy  string     `json:"received_by,omitempty"`
	ReceivedAt  *time.Time `json:"received_at,omitempty"`
	Notes       string     `json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
