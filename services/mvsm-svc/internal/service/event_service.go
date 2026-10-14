package service

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/mvsm-svc/internal/domain"
)

type EventService struct {
	repo domain.EventRepository
	log  *zap.Logger
}

func NewEventService(repo domain.EventRepository, log *zap.Logger) *EventService {
	return &EventService{repo: repo, log: log}
}

func (s *EventService) CreateEvent(req *domain.CreateEventRequest) (*domain.Event, error) {
	sd, err := time.Parse(time.RFC3339, req.ScheduledDate)
	if err != nil {
		return nil, err
	}

	event := &domain.Event{
		EventType:     domain.EventType(req.EventType),
		RiskLevel:     domain.RiskLow,
		ScheduledDate: sd,
		CreatedBy:     uuid.New(),
	}

	if req.EventName != "" {
		event.EventName = &req.EventName
	}
	if req.LocationDesc != "" {
		event.LocationDesc = &req.LocationDesc
	}
	if req.DeptCode != "" {
		event.DeptCode = &req.DeptCode
	}
	if req.Commune != "" {
		event.Commune = &req.Commune
	}
	if req.Lat != nil {
		event.Lat = req.Lat
	}
	if req.Lng != nil {
		event.Lng = req.Lng
	}
	if req.EstimatedCrowd != nil {
		event.EstimatedCrowd = req.EstimatedCrowd
	}
	if req.OrganizerName != "" {
		event.OrganizerName = &req.OrganizerName
	}

	return s.repo.Create(event)
}

func (s *EventService) ListUpcoming() ([]domain.Event, error) {
	return s.repo.FindUpcoming()
}

func (s *EventService) ListActive() ([]domain.Event, error) {
	return s.repo.FindActive()
}

func (s *EventService) AddUpdate(eventID uuid.UUID, req *domain.AddUpdateRequest) error {
	update := &domain.RealTimeUpdate{
		EventID:    eventID,
		UpdateTime: time.Now(),
		Situation:  req.Situation,
		ReportedBy: uuid.New(),
	}

	if req.CurrentCrowdEst != nil {
		update.CurrentCrowdEst = req.CurrentCrowdEst
	}
	if req.RiskChange != "" {
		level := domain.RiskLevel(req.RiskChange)
		update.RiskChange = &level
	}
	if req.ActionTaken != "" {
		update.ActionTaken = &req.ActionTaken
	}
	if req.Lat != nil {
		update.Lat = req.Lat
	}
	if req.Lng != nil {
		update.Lng = req.Lng
	}

	return s.repo.AddUpdate(update)
}

func (s *EventService) UpdateRiskLevel(id uuid.UUID, req *domain.UpdateRiskRequest) error {
	return s.repo.UpdateRiskLevel(id, domain.RiskLevel(req.RiskLevel))
}
