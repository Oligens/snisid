package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/rdep/internal/domain"
)

var ErrFlightNotFound = errors.New("vol non trouvé")

type FlightService struct {
	mu       sync.RWMutex
	flights  map[uuid.UUID]*domain.Flight
	byNumber map[string]uuid.UUID
}

func NewFlightService() *FlightService {
	return &FlightService{
		flights:  make(map[uuid.UUID]*domain.Flight),
		byNumber: make(map[string]uuid.UUID),
	}
}

func (s *FlightService) Create(ctx context.Context, req domain.CreateFlightRequest) (*domain.Flight, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	flight := &domain.Flight{
		FlightID:         uuid.New(),
		FlightNumber:     req.FlightNumber,
		FlightType:       req.FlightType,
		OriginCountry:    req.OriginCountry,
		DepartureAirport: req.DepartureAirport,
		ArrivalAirport:   req.ArrivalAirport,
		DepartureTime:    req.DepartureTime,
		ArrivalTime:      req.ArrivalTime,
		DeportingAgency:  req.DeportingAgency,
		TotalPassengers:  req.TotalPassengers,
		ManifestRef:      req.ManifestRef,
		Notes:            req.Notes,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	s.flights[flight.FlightID] = flight
	s.byNumber[req.FlightNumber] = flight.FlightID

	return flight, nil
}

func (s *FlightService) GetByID(ctx context.Context, flightID uuid.UUID) (*domain.Flight, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flight, ok := s.flights[flightID]
	if !ok {
		return nil, ErrFlightNotFound
	}
	return flight, nil
}

func (s *FlightService) GetByNumber(ctx context.Context, flightNumber string) (*domain.Flight, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flightID, ok := s.byNumber[flightNumber]
	if !ok {
		return nil, ErrFlightNotFound
	}
	flight, ok := s.flights[flightID]
	if !ok {
		return nil, ErrFlightNotFound
	}
	return flight, nil
}

func (s *FlightService) List(ctx context.Context) ([]domain.Flight, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]domain.Flight, 0, len(s.flights))
	for _, f := range s.flights {
		results = append(results, *f)
	}
	return results, nil
}
