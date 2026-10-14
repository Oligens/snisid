package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/rdep-svc/internal/domain"
)

type MonitoringService struct {
	deporteeRepo domain.DeporteeRepository
	eventRepo    domain.MonitoringEventRepository
	eventPub     domain.EventPublisher
}

func NewMonitoringService(
	deporteeRepo domain.DeporteeRepository,
	eventRepo domain.MonitoringEventRepository,
	eventPub domain.EventPublisher,
) *MonitoringService {
	return &MonitoringService{
		deporteeRepo: deporteeRepo,
		eventRepo:    eventRepo,
		eventPub:     eventPub,
	}
}

type MonitoringEventRequest struct {
	DeporteeID  string  `json:"deportee_id" binding:"required"`
	EventType   string  `json:"event_type" binding:"required"`
	LocationLat float64 `json:"location_lat"`
	LocationLng float64 `json:"location_lng"`
	Notes       string  `json:"notes"`
	ReportedBy  string  `json:"reported_by" binding:"required"`
}

func (s *MonitoringService) RecordEvent(ctx context.Context, req MonitoringEventRequest) (*domain.MonitoringEvent, error) {
	deporteeID, err := uuid.Parse(req.DeporteeID)
	if err != nil {
		return nil, fmt.Errorf("UUID déporté invalide: %w", err)
	}

	reportedBy, err := uuid.Parse(req.ReportedBy)
	if err != nil {
		return nil, fmt.Errorf("UUID rapporteur invalide: %w", err)
	}

	deportee, err := s.deporteeRepo.FindByID(ctx, deporteeID)
	if err != nil {
		return nil, fmt.Errorf("déporté introuvable: %w", err)
	}

	if deportee.MonitoringStatus != domain.MonitoringActive {
		return nil, fmt.Errorf("surveillance non active pour ce déporté")
	}

	event := &domain.MonitoringEvent{
		EventID:     uuid.New(),
		DeporteeID:  deporteeID,
		EventType:   req.EventType,
		EventDate:   time.Now(),
		LocationLat: req.LocationLat,
		LocationLng: req.LocationLng,
		Notes:       req.Notes,
		ReportedBy:  reportedBy,
		CreatedAt:   time.Now(),
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("création événement: %w", err)
	}

	if req.EventType == "VIOLATION" {
		_ = s.eventPub.Publish("rdep.monitoring.violation", map[string]interface{}{
			"deportee_id":  deporteeID,
			"person_id":    deportee.SNISIDPersonID,
			"event_id":     event.EventID,
			"violation_at": time.Now(),
		})
	}

	return event, nil
}

func (s *MonitoringService) GetEvents(ctx context.Context, deporteeID uuid.UUID) ([]*domain.MonitoringEvent, error) {
	return s.eventRepo.FindByDeporteeID(ctx, deporteeID)
}

func (s *MonitoringService) UpdateAddress(ctx context.Context, deporteeID uuid.UUID, address, commune, deptCode string) error {
	deportee, err := s.deporteeRepo.FindByID(ctx, deporteeID)
	if err != nil {
		return fmt.Errorf("déporté introuvable: %w", err)
	}

	deportee.CurrentAddress = address
	deportee.CurrentCommune = commune
	deportee.CurrentDeptCode = deptCode
	deportee.UpdatedAt = time.Now()

	return s.deporteeRepo.Update(ctx, deportee)
}
