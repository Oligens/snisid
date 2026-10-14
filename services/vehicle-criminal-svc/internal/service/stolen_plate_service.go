package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/vehicle-criminal-svc/internal/domain"
	"github.com/snisid/vehicle-criminal-svc/internal/repository"
	"go.uber.org/zap"
)

type StolenPlateService struct {
	repo    repository.StolenPlateRepository
	hotlist repository.HotlistCache
	kafka   repository.EventPublisher
	logger  *zap.Logger
}

func NewStolenPlateService(
	repo repository.StolenPlateRepository,
	hotlist repository.HotlistCache,
	kafka repository.EventPublisher,
	logger *zap.Logger,
) *StolenPlateService {
	return &StolenPlateService{
		repo:    repo,
		hotlist: hotlist,
		kafka:   kafka,
		logger:  logger,
	}
}

func (s *StolenPlateService) DeclareStolen(
	ctx context.Context,
	req domain.DeclareStolenPlateRequest,
	createdBy uuid.UUID,
) (*domain.StolenPlate, error) {
	if err := domain.ValidatePlateNumber(req.PlateNumber); err != nil {
		return nil, fmt.Errorf("format plaque invalide: %w", err)
	}

	existing, _ := s.repo.FindStolenByPlate(ctx, req.PlateNumber)
	if existing != nil {
		return nil, fmt.Errorf("cette plaque est déjà déclarée volée (ID: %s)", existing.PlateID)
	}

	plate := domain.NewStolenPlate(req, createdBy)

	if err := s.repo.Create(ctx, plate); err != nil {
		return nil, fmt.Errorf("erreur déclaration plaque volée: %w", err)
	}

	event := domain.AlertCreatedEvent{
		AlertID:       plate.PlateID,
		PlateNumber:   plate.PlateNumber,
		CrimeCategory: domain.CrimeCategoryPlatTheft,
		AlertLevel:    domain.AlertLevelWanted,
		ReportingUnit: plate.ReportingUnit,
		Timestamp:     time.Now(),
	}
	_ = s.kafka.Publish(ctx, "sivc.plates.declared_stolen", event)

	s.logger.Info("Plaque déclarée volée",
		zap.String("plate_id", plate.PlateID.String()),
		zap.String("plate", plate.PlateNumber),
		zap.Bool("state_plate_clone", plate.IsStatePlateClone),
	)

	return plate, nil
}

func (s *StolenPlateService) CheckPlate(ctx context.Context, plateNumber string) (*domain.StolenPlate, error) {
	return s.repo.FindStolenByPlate(ctx, plateNumber)
}

func (s *StolenPlateService) MarkRecovered(
	ctx context.Context,
	id uuid.UUID,
	location string,
	deptCode string,
	updatedBy uuid.UUID,
) error {
	if err := s.repo.MarkRecovered(ctx, id, location, deptCode); err != nil {
		return err
	}

	plate, err := s.repo.FindByID(ctx, id)
	if err == nil && plate != nil {
		_ = s.hotlist.DeletePlateAlert(ctx, plate.PlateNumber)
	}

	s.logger.Info("Plaque récupérée",
		zap.String("plate_id", id.String()),
		zap.String("location", location),
		zap.String("dept", deptCode),
	)

	return nil
}
