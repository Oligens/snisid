package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/sipep-svc/internal/domain"
)

type TransferService struct {
	inmateRepo    domain.InmateRepository
	detentionRepo domain.DetentionRepository
	transferRepo  domain.TransferRepository
	eventPub      domain.EventPublisher
}

func NewTransferService(
	inmateRepo domain.InmateRepository,
	detentionRepo domain.DetentionRepository,
	transferRepo domain.TransferRepository,
	eventPub domain.EventPublisher,
) *TransferService {
	return &TransferService{
		inmateRepo:    inmateRepo,
		detentionRepo: detentionRepo,
		transferRepo:  transferRepo,
		eventPub:      eventPub,
	}
}

type TransferRequest struct {
	InmateID       string `json:"inmate_id" binding:"required"`
	ToFacility     string `json:"to_facility" binding:"required"`
	TransferReason string `json:"transfer_reason"`
	AuthorizedBy   string `json:"authorized_by" binding:"required"`
	TransportUnit  string `json:"transport_unit"`
}

func (s *TransferService) ProcessTransfer(ctx context.Context, req TransferRequest) (*domain.Transfer, error) {
	inmateID, err := uuid.Parse(req.InmateID)
	if err != nil {
		return nil, fmt.Errorf("UUID invalide: %w", err)
	}

	authorizedBy, err := uuid.Parse(req.AuthorizedBy)
	if err != nil {
		return nil, fmt.Errorf("UUID autorisateur invalide: %w", err)
	}

	inmate, err := s.inmateRepo.FindByID(ctx, inmateID)
	if err != nil {
		return nil, fmt.Errorf("détenu introuvable: %w", err)
	}

	if !inmate.IsCurrentlyDetained {
		return nil, fmt.Errorf("détenu non actuellement incarcéré")
	}

	fromFacility := inmate.CurrentFacility

	transfer := &domain.Transfer{
		TransferID:     uuid.New(),
		InmateID:       inmateID,
		FromFacility:   fromFacility,
		ToFacility:     req.ToFacility,
		TransferDate:   time.Now(),
		TransferReason: req.TransferReason,
		AuthorizedBy:   authorizedBy,
		TransportUnit:  req.TransportUnit,
		CreatedAt:      time.Now(),
	}

	if err := s.transferRepo.Create(ctx, transfer); err != nil {
		return nil, fmt.Errorf("création transfert: %w", err)
	}

	inmate.CurrentFacility = req.ToFacility
	inmate.UpdatedAt = time.Now()
	_ = s.inmateRepo.Update(ctx, inmate)

	_ = s.eventPub.Publish("sipep.inmate.transferred", map[string]interface{}{
		"transfer_id":   transfer.TransferID,
		"inmate_id":     inmateID,
		"from_facility": fromFacility,
		"to_facility":   req.ToFacility,
		"transfer_at":   time.Now(),
	})

	return transfer, nil
}

func (s *TransferService) GetTransfers(ctx context.Context, inmateID uuid.UUID) ([]*domain.Transfer, error) {
	return s.transferRepo.FindByInmateID(ctx, inmateID)
}
