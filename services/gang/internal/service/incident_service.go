package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/gang/internal/domain"
	"github.com/snisid/platform/services/gang/internal/repository"
)

type IncidentService struct {
	incidentRepo repository.IncidentRepository
	gangRepo     repository.GangRepository
}

func NewIncidentService(incidentRepo repository.IncidentRepository, gangRepo repository.GangRepository) *IncidentService {
	return &IncidentService{incidentRepo: incidentRepo, gangRepo: gangRepo}
}

func (s *IncidentService) CreateIncident(ctx context.Context, req domain.CreateIncidentRequest, createdBy uuid.UUID) (*domain.Incident, error) {
	if _, err := s.gangRepo.GetByID(ctx, req.GangID); err != nil {
		return nil, fmt.Errorf("gang introuvable: %w", err)
	}
	casualties := int16(0)
	if req.Casualties != nil {
		casualties = *req.Casualties
	}
	inc := &domain.Incident{
		IncidentID:         uuid.New(),
		GangID:             req.GangID,
		IncidentType:       req.IncidentType,
		IncidentDate:       req.IncidentDate,
		LocationDesc:       req.LocationDesc,
		DeptCode:           req.DeptCode,
		Commune:            req.Commune,
		Lat:                req.Lat,
		Lng:                req.Lng,
		Casualties:         casualties,
		Description:        req.Description,
		IntelligenceSource: req.IntelligenceSource,
		CreatedBy:          createdBy,
		CreatedAt:          time.Now(),
	}
	if err := s.incidentRepo.Create(ctx, inc); err != nil {
		return nil, fmt.Errorf("erreur création incident: %w", err)
	}
	return inc, nil
}

func (s *IncidentService) GetIncidents(ctx context.Context, gangID uuid.UUID) ([]*domain.Incident, error) {
	return s.incidentRepo.ByGangID(ctx, gangID)
}
