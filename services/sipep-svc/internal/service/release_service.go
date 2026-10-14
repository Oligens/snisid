package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/sipep-svc/internal/domain"
)

type ReleaseService struct {
	inmateRepo    domain.InmateRepository
	detentionRepo domain.DetentionRepository
	eventPub      domain.EventPublisher
}

func NewReleaseService(
	inmateRepo domain.InmateRepository,
	detentionRepo domain.DetentionRepository,
	eventPub domain.EventPublisher,
) *ReleaseService {
	return &ReleaseService{
		inmateRepo:    inmateRepo,
		detentionRepo: detentionRepo,
		eventPub:      eventPub,
	}
}

func (s *ReleaseService) ProcessRelease(
	ctx context.Context,
	inmateID uuid.UUID,
	req domain.ReleaseRequest,
	authorizedBy uuid.UUID,
) (*domain.Detention, error) {
	inmate, err := s.inmateRepo.FindByID(ctx, inmateID)
	if err != nil {
		return nil, fmt.Errorf("détenu introuvable: %w", err)
	}
	if !inmate.IsCurrentlyDetained {
		return nil, fmt.Errorf("détenu non actuellement incarcéré")
	}

	detention, err := s.detentionRepo.GetActiveDetention(ctx, inmateID)
	if err != nil {
		return nil, fmt.Errorf("dossier détention actif introuvable: %w", err)
	}

	now := time.Now()
	detention.ReleaseDate = &now
	detention.ReleaseType = req.ReleaseType
	detention.ReleasingAuthority = req.Authority

	if err := s.detentionRepo.Update(ctx, detention); err != nil {
		return nil, fmt.Errorf("mise à jour détention: %w", err)
	}

	inmate.IsCurrentlyDetained = false
	inmate.UpdatedAt = now
	_ = s.inmateRepo.Update(ctx, inmate)

	_ = s.eventPub.Publish("sipep.inmate.released", domain.InmateReleasedEvent{
		InmateID:     inmateID,
		PersonID:     inmate.SNISIDPersonID,
		FacilityCode: inmate.CurrentFacility,
		ReleaseType:  req.ReleaseType,
		ReleasedAt:   now,
		AuthorizedBy: authorizedBy,
	})

	if req.ReleaseType == domain.ReleaseTypeEscape {
		_ = s.eventPub.Publish("sipep.escape.alert", domain.EscapeAlertEvent{
			InmateID:     inmateID,
			PersonID:     inmate.SNISIDPersonID,
			FacilityCode: inmate.CurrentFacility,
			EscapedAt:    now,
		})
	}

	return detention, nil
}
