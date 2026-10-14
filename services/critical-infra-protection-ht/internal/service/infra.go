package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/critical-infra-protection-ht/internal/domain"
	"github.com/snisid/critical-infra-protection-ht/internal/kafka"
	"github.com/snisid/critical-infra-protection-ht/internal/repository"
)

type InfraProtService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewInfraProtService(repo repository.Repository, producer *kafka.Producer) *InfraProtService {
	return &InfraProtService{repo: repo, producer: producer}
}

func (s *InfraProtService) CreateAsset(ctx context.Context, req domain.CreateAssetRequest) (*domain.CriticalAsset, error) {
	now := time.Now().UTC()
	a := &domain.CriticalAsset{
		ID:                    uuid.New(),
		AssetName:             req.AssetName,
		Sector:                domain.Sector(req.Sector),
		OwnerEntity:           req.OwnerEntity,
		LocationLat:           req.LocationLat,
		LocationLng:           req.LocationLng,
		Region:                req.Region,
		DeptCode:              req.DeptCode,
		Criticality:           domain.Criticality(req.Criticality),
		CyberMaturityScore:    req.CyberMaturityScore,
		PhysicalSecurityScore: req.PhysicalSecurityScore,
		ContactName:           req.ContactName,
		ContactPhone:          req.ContactPhone,
		HasBackupGenerator:    req.HasBackupGenerator,
		HasCyberInsurance:     req.HasCyberInsurance,
		CreatedAt:             now,
	}
	if req.LastCISAAssessmentAt != "" {
		t, err := time.Parse(time.RFC3339, req.LastCISAAssessmentAt)
		if err == nil {
			a.LastCISAAssessmentAt = &t
		}
	}
	if err := s.repo.CreateAsset(ctx, a); err != nil {
		return nil, fmt.Errorf("create asset: %w", err)
	}
	s.publishEvent(ctx, "infraprot.asset.created", a.ID.String(), "asset", a)
	return a, nil
}

func (s *InfraProtService) GetAssetsBySector(ctx context.Context, sector string) ([]domain.CriticalAsset, error) {
	return s.repo.GetAssetsBySector(ctx, sector)
}

func (s *InfraProtService) ReportIncident(ctx context.Context, req domain.ReportIncidentRequest) (*domain.InfrastructureIncident, error) {
	assetID, err := uuid.Parse(req.AssetID)
	if err != nil {
		return nil, fmt.Errorf("invalid asset_id: %w", err)
	}
	inc := &domain.InfrastructureIncident{
		ID:           uuid.New(),
		AssetID:      assetID,
		IncidentType: domain.IncidentType(req.IncidentType),
		Severity:     req.Severity,
		Description:  req.Description,
		Status:       domain.IncStatusReported,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.repo.CreateIncident(ctx, inc); err != nil {
		return nil, fmt.Errorf("report incident: %w", err)
	}
	s.publishEvent(ctx, "infraprot.incident.reported", inc.ID.String(), "incident", inc)
	return inc, nil
}

func (s *InfraProtService) GetActiveIncidents(ctx context.Context) ([]domain.InfrastructureIncident, error) {
	return s.repo.GetActiveIncidents(ctx)
}

func (s *InfraProtService) GetIncidentsByAsset(ctx context.Context, assetID uuid.UUID) ([]domain.InfrastructureIncident, error) {
	return s.repo.GetIncidentsByAsset(ctx, assetID)
}

func (s *InfraProtService) CreateAssessment(ctx context.Context, req domain.CreateAssessmentRequest) (*domain.SectorRiskAssessment, error) {
	now := time.Now().UTC()
	a := &domain.SectorRiskAssessment{
		ID:               uuid.New(),
		Sector:           domain.Sector(req.Sector),
		AssessmentDate:   now,
		OverallRiskScore: req.OverallRiskScore,
		TopThreats:       req.TopThreats,
		Vulnerabilities:  req.Vulnerabilities,
		Recommendations:  req.Recommendations,
		AssessorAgency:   req.AssessorAgency,
		CreatedAt:        now,
	}
	nextDue := now.AddDate(0, 6, 0)
	a.NextAssessmentDue = &nextDue

	if err := s.repo.CreateAssessment(ctx, a); err != nil {
		return nil, fmt.Errorf("create assessment: %w", err)
	}
	s.publishEvent(ctx, "infraprot.assessment.created", a.ID.String(), "assessment", a)
	return a, nil
}

func (s *InfraProtService) GetNationalDashboard(ctx context.Context) ([]domain.SectorRiskAssessment, error) {
	return s.repo.GetNationalDashboard(ctx)
}

func (s *InfraProtService) publishEvent(ctx context.Context, eventType, entityID, entityType string, data any) {
	if s.producer == nil {
		return
	}
	evt := kafka.Event{
		EventType:  eventType,
		EntityID:   entityID,
		EntityType: entityType,
		Timestamp:  time.Now().UTC(),
		Data:       data,
	}
	if err := s.producer.Publish(ctx, evt); err != nil {
		log.Printf("failed to publish event: %v", err)
	}
}
