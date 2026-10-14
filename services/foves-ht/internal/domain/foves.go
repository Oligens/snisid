package domain

import (
	"time"

	"github.com/google/uuid"
)

type VehicleCategory string

const (
	VehiclePrivateCar   VehicleCategory = "PRIVATE_CAR"
	VehicleMotorcycle   VehicleCategory = "MOTORCYCLE"
	VehicleTapTap       VehicleCategory = "TAP_TAP"
	VehicleBus          VehicleCategory = "BUS"
	VehicleTruck        VehicleCategory = "TRUCK"
	VehicleCommercial   VehicleCategory = "COMMERCIAL"
	VehicleGovernment   VehicleCategory = "GOVERNMENT"
	VehicleDiplomatic   VehicleCategory = "DIPLOMATIC"
	VehicleAgricultural VehicleCategory = "AGRICULTURAL"
)

type Vehicle struct {
	ID             uuid.UUID       `json:"id"`
	PlateNumber    string          `json:"plate_number"`
	VIN            string          `json:"vin"`
	Make           string          `json:"make"`
	Model          string          `json:"model"`
	Year           int             `json:"year"`
	Color          *string         `json:"color,omitempty"`
	Category       VehicleCategory `json:"category"`
	OwnerCitizenID uuid.UUID       `json:"owner_citizen_id"`
	IsStolen       bool            `json:"is_stolen"`
	IsActive       bool            `json:"is_active"`
	RegisteredAt   time.Time       `json:"registered_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type OwnershipTransfer struct {
	ID            uuid.UUID `json:"id"`
	VehicleID     uuid.UUID `json:"vehicle_id"`
	FromCitizenID uuid.UUID `json:"from_citizen_id"`
	ToCitizenID   uuid.UUID `json:"to_citizen_id"`
	TransferDate  time.Time `json:"transfer_date"`
	ContractRef   *string   `json:"contract_ref,omitempty"`
	ApprovedBy    *string   `json:"approved_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type DriverLicense struct {
	ID            uuid.UUID `json:"id"`
	CitizenID     uuid.UUID `json:"citizen_id"`
	LicenseNumber string    `json:"license_number"`
	CategoryA     bool      `json:"category_a"`
	CategoryB     bool      `json:"category_b"`
	CategoryC     bool      `json:"category_c"`
	CategoryD     bool      `json:"category_d"`
	CategoryE     bool      `json:"category_e"`
	CategoryF     bool      `json:"category_f"`
	IssuedDate    time.Time `json:"issued_date"`
	ExpiryDate    time.Time `json:"expiry_date"`
	PointsBalance int16     `json:"points_balance"`
	IsSuspended   bool      `json:"is_suspended"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
