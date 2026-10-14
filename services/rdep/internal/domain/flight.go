package domain

import (
	"time"

	"github.com/google/uuid"
)

type Flight struct {
	FlightID         uuid.UUID          `json:"flight_id"`
	FlightNumber     string             `json:"flight_number"`
	FlightType       FlightType         `json:"flight_type"`
	OriginCountry    DeportationCountry `json:"origin_country"`
	DepartureAirport string             `json:"departure_airport"`
	ArrivalAirport   string             `json:"arrival_airport"`
	DepartureTime    time.Time          `json:"departure_time"`
	ArrivalTime      time.Time          `json:"arrival_time"`
	DeportingAgency  *string            `json:"deporting_agency,omitempty"`
	TotalPassengers  int                `json:"total_passengers"`
	ManifestRef      *string            `json:"manifest_ref,omitempty"`
	Notes            *string            `json:"notes,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

type CreateFlightRequest struct {
	FlightNumber     string             `json:"flight_number" binding:"required"`
	FlightType       FlightType         `json:"flight_type" binding:"required"`
	OriginCountry    DeportationCountry `json:"origin_country" binding:"required"`
	DepartureAirport string             `json:"departure_airport" binding:"required"`
	ArrivalAirport   string             `json:"arrival_airport" binding:"required"`
	DepartureTime    time.Time          `json:"departure_time" binding:"required"`
	ArrivalTime      time.Time          `json:"arrival_time" binding:"required"`
	DeportingAgency  *string            `json:"deporting_agency,omitempty"`
	TotalPassengers  int                `json:"total_passengers"`
	ManifestRef      *string            `json:"manifest_ref,omitempty"`
	Notes            *string            `json:"notes,omitempty"`
}
