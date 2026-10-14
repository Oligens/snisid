package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/fir-svc/internal/domain"
)

type ExpungementService struct {
	recordRepo domain.CriminalRecordRepository
	eventPub   domain.EventPublisher
}

func NewExpungementService(recordRepo domain.CriminalRecordRepository, eventPub domain.EventPublisher) *ExpungementService {
	return &ExpungementService{recordRepo: recordRepo, eventPub: eventPub}
}

func (s *ExpungementService) RequestExpungement(
	ctx context.Context,
	recordID uuid.UUID,
	reason string,
) error {
	record, err := s.recordRepo.FindByID(ctx, recordID)
	if err != nil {
		return fmt.Errorf("casier introuvable: %w", err)
	}

	if record.IsExpunged {
		return fmt.Errorf("casier déjà réhabilité")
	}

	record.IsExpunged = true
	record.IsActive = false
	record.UpdatedAt = time.Now()

	if err := s.recordRepo.Update(ctx, record); err != nil {
		return fmt.Errorf("réhabilitation échouée: %w", err)
	}

	_ = s.eventPub.Publish("fir.record.expunged", map[string]interface{}{
		"record_id":   recordID,
		"reason":      reason,
		"expunged_at": time.Now(),
	})

	return nil
}
