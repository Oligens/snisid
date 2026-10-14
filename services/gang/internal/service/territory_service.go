package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/gang/internal/domain"
	"github.com/snisid/platform/services/gang/internal/repository"
)

type TerritoryService struct {
	territoryRepo repository.TerritoryRepository
	gangRepo      repository.GangRepository
}

func NewTerritoryService(territoryRepo repository.TerritoryRepository, gangRepo repository.GangRepository) *TerritoryService {
	return &TerritoryService{territoryRepo: territoryRepo, gangRepo: gangRepo}
}

func (s *TerritoryService) CreateTerritory(ctx context.Context, req domain.CreateTerritoryRequest, createdBy uuid.UUID) (*domain.Territory, error) {
	if _, err := s.gangRepo.GetByID(ctx, req.GangID); err != nil {
		return nil, fmt.Errorf("gang introuvable: %w", err)
	}
	claimed := true
	if req.IsClaimed != nil {
		claimed = *req.IsClaimed
	}
	contested := false
	if req.IsContested != nil {
		contested = *req.IsContested
	}
	t := &domain.Territory{
		TerritoryID:     uuid.New(),
		GangID:          req.GangID,
		DeptCode:        req.DeptCode,
		Commune:         req.Commune,
		Locality:        req.Locality,
		IsClaimed:       claimed,
		IsContested:     contested,
		ContestedWith:   req.ContestedWith,
		ControlledSince: req.ControlledSince,
		Notes:           req.Notes,
		CreatedBy:       createdBy,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := s.territoryRepo.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("erreur création territoire: %w", err)
	}
	return t, nil
}

func (s *TerritoryService) GetTerritories(ctx context.Context, gangID uuid.UUID) ([]*domain.Territory, error) {
	return s.territoryRepo.ByGangID(ctx, gangID)
}
