package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/gang-svc/internal/domain"
)

type IncidentService struct {
	incRepo  domain.IncidentRepository
	orgRepo  domain.OrganizationRepository
	eventPub domain.EventPublisher
}

func NewIncidentService(
	incRepo domain.IncidentRepository,
	orgRepo domain.OrganizationRepository,
	eventPub domain.EventPublisher,
) *IncidentService {
	return &IncidentService{
		incRepo:  incRepo,
		orgRepo:  orgRepo,
		eventPub: eventPub,
	}
}

type CreateIncidentRequest struct {
	GangID       string   `json:"gang_id" binding:"required"`
	IncidentType string   `json:"incident_type" binding:"required"`
	IncidentDate string   `json:"incident_date" binding:"required"`
	LocationDesc string   `json:"location_desc"`
	DeptCode     string   `json:"dept_code"`
	Commune      string   `json:"commune"`
	Lat          float64  `json:"lat"`
	Lng          float64  `json:"lng"`
	Casualties   int      `json:"casualties"`
	VictimIDs    []string `json:"victim_ids"`
	Description  string   `json:"description"`
	IntelSource  string   `json:"intel_source"`
	CreatedBy    string   `json:"created_by" binding:"required"`
}

func (s *IncidentService) CreateIncident(ctx context.Context, req CreateIncidentRequest) (*domain.Incident, error) {
	gangID, err := uuid.Parse(req.GangID)
	if err != nil {
		return nil, fmt.Errorf("UUID invalide: %w", err)
	}

	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("UUID invalide: %w", err)
	}

	incidentDate, err := time.Parse("2006-01-02T15:04:05Z", req.IncidentDate)
	if err != nil {
		incidentDate, err = time.Parse("2006-01-02", req.IncidentDate)
		if err != nil {
			return nil, fmt.Errorf("date invalide: %w", err)
		}
	}

	var victimIDs []uuid.UUID
	for _, vStr := range req.VictimIDs {
		if vID, err := uuid.Parse(vStr); err == nil {
			victimIDs = append(victimIDs, vID)
		}
	}

	incident := &domain.Incident{
		IncidentID:   uuid.New(),
		GangID:       gangID,
		IncidentType: req.IncidentType,
		IncidentDate: incidentDate,
		LocationDesc: req.LocationDesc,
		DeptCode:     req.DeptCode,
		Commune:      req.Commune,
		Lat:          req.Lat,
		Lng:          req.Lng,
		Casualties:   req.Casualties,
		VictimIDs:    victimIDs,
		Description:  req.Description,
		IntelSource:  req.IntelSource,
		CreatedBy:    createdBy,
		CreatedAt:    time.Now(),
	}

	if err := s.incRepo.Create(ctx, incident); err != nil {
		return nil, nil
	}

	_ = s.eventPub.Publish("gang.incident.reported", incident)
	return incident, nil
}

func (s *IncidentService) GetByGangID(ctx context.Context, gangID uuid.UUID) ([]*domain.Incident, error) {
	return s.incRepo.FindByGangID(ctx, gangID)
}

func (s *IncidentService) GetByDeptCode(ctx context.Context, deptCode string) ([]*domain.Incident, error) {
	return s.incRepo.FindByDeptCode(ctx, deptCode)
}
