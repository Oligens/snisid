package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/gang/internal/domain"
	"github.com/snisid/platform/services/gang/internal/repository"
)

type GangService struct {
	repo repository.GangRepository
}

func NewGangService(repo repository.GangRepository) *GangService {
	return &GangService{repo: repo}
}

func (s *GangService) CreateOrganization(ctx context.Context, req domain.CreateOrganizationRequest, createdBy uuid.UUID) (*domain.Organization, error) {
	nationalID := fmt.Sprintf("GANG-HT-%06d", time.Now().UnixMilli()%1000000)
	org := &domain.Organization{
		GangID:            uuid.New(),
		NationalGangID:    nationalID,
		Name:              req.Name,
		Aliases:           req.Aliases,
		StructureType:     req.StructureType,
		PrimaryActivity:   req.PrimaryActivity,
		ActivityLevel:     req.ActivityLevel,
		EstimatedMembers:  req.EstimatedMembers,
		ArmedMembersPct:   req.ArmedMembersPct,
		HeavyWeapons:      req.HeavyWeapons,
		PrimaryDeptCode:   req.PrimaryDeptCode,
		TerritoryCommunes: req.TerritoryCommunes,
		OFACDesignation:   req.OFACDesignation,
		OFACSDNRef:        req.OFACSDNRef,
		EstablishedDate:   req.EstablishedDate,
		IntelConfidence:   req.IntelConfidence,
		IsActive:          true,
		CreatedBy:         createdBy,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if org.ActivityLevel == "" {
		org.ActivityLevel = domain.GangActivityHigh
	}
	if err := s.repo.Create(ctx, org); err != nil {
		return nil, fmt.Errorf("erreur création organisation: %w", err)
	}
	return org, nil
}

func (s *GangService) GetOrganization(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *GangService) ListOrganizations(ctx context.Context) ([]*domain.Organization, error) {
	return s.repo.List(ctx)
}

func (s *GangService) ByDeptCode(ctx context.Context, code string) ([]*domain.Organization, error) {
	return s.repo.ByDeptCode(ctx, code)
}

func (s *GangService) Sanctioned(ctx context.Context) ([]*domain.Organization, error) {
	return s.repo.Sanctioned(ctx)
}
