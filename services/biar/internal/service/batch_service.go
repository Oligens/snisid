package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/biar/internal/domain"
	"github.com/snisid/platform/services/biar/internal/repository"
)

type BatchService struct {
	batchRepo  repository.BatchRepository
	weaponRepo repository.WeaponRepository
}

func NewBatchService(batchRepo repository.BatchRepository, weaponRepo repository.WeaponRepository) *BatchService {
	return &BatchService{batchRepo: batchRepo, weaponRepo: weaponRepo}
}

func (s *BatchService) CreateBatch(ctx context.Context, req domain.CreateBatchRequest) (*domain.BatchSeizure, error) {
	batchRef := fmt.Sprintf("BATCH-BIAR-%06d", time.Now().UnixMilli()%1000000)

	batch := &domain.BatchSeizure{
		BatchID:            uuid.New(),
		BatchReference:     batchRef,
		OperationName:      req.OperationName,
		SeizureDate:        req.SeizureDate,
		LocationDesc:       req.LocationDesc,
		DeptCode:           req.DeptCode,
		TotalWeapons:       len(req.WeaponIDs),
		WeaponIDs:          req.WeaponIDs,
		SeizingUnit:        req.SeizingUnit,
		LeadOfficer:        req.LeadOfficer,
		PartneringAgencies: req.PartneringAgencies,
		Notes:              req.Notes,
		CreatedAt:          time.Now(),
	}

	if err := s.batchRepo.Create(ctx, batch); err != nil {
		return nil, fmt.Errorf("erreur création lot saisie: %w", err)
	}
	return batch, nil
}

func (s *BatchService) GetBatch(ctx context.Context, id uuid.UUID) (*domain.BatchSeizure, error) {
	return s.batchRepo.GetByID(ctx, id)
}

func (s *BatchService) ListBatches(ctx context.Context) ([]*domain.BatchSeizure, error) {
	return s.batchRepo.List(ctx)
}
