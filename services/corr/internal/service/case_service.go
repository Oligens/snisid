package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/corr/internal/domain"
	"github.com/snisid/platform/services/corr/internal/repository"
)

type CaseService struct {
	caseRepo    repository.CaseRepository
	officerRepo repository.OfficerRepository
}

func NewCaseService(caseRepo repository.CaseRepository, officerRepo repository.OfficerRepository) *CaseService {
	return &CaseService{caseRepo: caseRepo, officerRepo: officerRepo}
}

func (s *CaseService) CreateCase(ctx context.Context, req domain.CreateCaseRequest, createdBy uuid.UUID) (*domain.IntegrityCase, error) {
	nationalID := fmt.Sprintf("CORR-HT-%s-%06d", time.Now().Format("2006"), time.Now().UnixMilli()%1000000)
	now := time.Now()
	c := &domain.IntegrityCase{
		CaseID:          uuid.New(),
		NationalCorrID:  nationalID,
		OfficerSnisidID: req.OfficerSnisidID,
		OfficerBadge:    req.OfficerBadge,
		OfficerUnit:     req.OfficerUnit,
		OfficerRank:     req.OfficerRank,
		AllegationType:  req.AllegationType,
		Severity:        req.Severity,
		Status:          domain.StatusReported,

		AllegationSummary: req.AllegationSummary,
		IncidentDateFrom:  req.IncidentDateFrom,
		IncidentDateTo:    req.IncidentDateTo,

		GangID:           req.GangID,
		FinancialGainUSD: req.FinancialGainUSD,

		ReportedByType:         req.ReportedByType,
		ReportingDate:          now,
		IsWhistleblower:        req.IsWhistleblower,
		WhistleblowerProtected: req.WhistleblowerProtected,

		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.caseRepo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("erreur création cas: %w", err)
	}

	officer, err := s.officerRepo.GetBySnisidID(ctx, req.OfficerSnisidID)
	if err == nil {
		officer.UnderInvestigation = true
		officer.ActiveCaseID = &c.CaseID
		officer.InvestigationCount++
		officer.UpdatedAt = now
		s.officerRepo.Upsert(ctx, officer)
	}

	return c, nil
}

func (s *CaseService) GetCase(ctx context.Context, id uuid.UUID) (*domain.IntegrityCase, error) {
	return s.caseRepo.GetByID(ctx, id)
}

func (s *CaseService) ListActive(ctx context.Context) ([]*domain.IntegrityCase, error) {
	return s.caseRepo.ListActive(ctx)
}
