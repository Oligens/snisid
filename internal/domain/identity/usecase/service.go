package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/snisid/platform/internal/domain/identity/entity"
	"github.com/snisid/platform/internal/domain/identity/repository"
	"github.com/snisid/platform/internal/platform/events"
)

type IdentityEvent struct {
	EventType  string           `json:"eventType"` // created, updated, flagged
	IdentityID string           `json:"identityId"`
	Identity   *entity.Identity `json:"identity"`
	Timestamp  time.Time        `json:"timestamp"`
}

type IdentityService interface {
	CreateIdentity(ctx context.Context, ident *entity.Identity, changedBy string) (*entity.Identity, error)
	UpdateIdentity(ctx context.Context, id string, updateFn func(*entity.Identity), reason, changedBy string) (*entity.Identity, error)
	GetIdentity(ctx context.Context, id string) (*entity.Identity, error)
	FlagIdentity(ctx context.Context, id, reason, changedBy string) error
	GetHistory(ctx context.Context, id string) ([]entity.IdentityHistory, error)
}

type service struct {
	repo     repository.IdentityRepository
	producer events.ProducerInterface
}

func NewIdentityService(repo repository.IdentityRepository, producer events.ProducerInterface) IdentityService {
	return &service{
		repo:     repo,
		producer: producer,
	}
}

func (s *service) CreateIdentity(ctx context.Context, ident *entity.Identity, changedBy string) (*entity.Identity, error) {
	ident.ID = fmt.Sprintf("ID-%d", time.Now().UnixNano())
	ident.CreatedAt = time.Now().UTC()
	ident.UpdatedAt = ident.CreatedAt
	ident.Status = entity.StatePending
	ident.Version = 1

	if err := s.repo.Create(ctx, ident, "Initial creation", changedBy); err != nil {
		return nil, err
	}

	s.publishEvent(ctx, "identity.created", ident)
	return ident, nil
}

func (s *service) UpdateIdentity(ctx context.Context, id string, updateFn func(*entity.Identity), reason, changedBy string) (*entity.Identity, error) {
	ident, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updateFn(ident)
	ident.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, ident, reason, changedBy); err != nil {
		return nil, err
	}

	s.publishEvent(ctx, "identity.updated", ident)
	return ident, nil
}

func (s *service) GetIdentity(ctx context.Context, id string) (*entity.Identity, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) FlagIdentity(ctx context.Context, id, reason, changedBy string) error {
	ident, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	ident.Status = entity.StateSuspended
	ident.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, ident, reason, changedBy); err != nil {
		return err
	}

	s.publishEvent(ctx, "identity.flagged", ident)
	return nil
}

func (s *service) GetHistory(ctx context.Context, id string) ([]entity.IdentityHistory, error) {
	return s.repo.GetHistory(ctx, id)
}

func (s *service) publishEvent(ctx context.Context, eventType string, ident *entity.Identity) {
	if s.producer == nil {
		return
	}

	evt := IdentityEvent{
		EventType:  eventType,
		IdentityID: ident.ID,
		Identity:   ident,
		Timestamp:  time.Now().UTC(),
	}

	// Topic routing can be done here or producer can just publish to the topic it was configured with.
	// For now, assuming producer is configured for an 'identity-events' topic,
	// we use eventType as the key or embedded in payload.
	_ = s.producer.Publish(ctx, ident.ID, evt)
}
