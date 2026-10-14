package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/iam-ht/internal/domain"
	"github.com/snisid/iam-ht/internal/kafka"
	"github.com/snisid/iam-ht/internal/repository"
)

type IAMService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewIAMService(repo repository.Repository, producer *kafka.Producer) *IAMService {
	return &IAMService{repo: repo, producer: producer}
}

func (s *IAMService) Authorize(ctx context.Context, citizenID string) (*domain.IdentityAssurance, error) {
	cid, _ := uuid.Parse(citizenID)
	return s.repo.GetAssurance(ctx, cid)
}

func (s *IAMService) StepUpAssurance(ctx context.Context, citizenID string) (*domain.IdentityAssurance, error) {
	cid, _ := uuid.Parse(citizenID)
	existing, err := s.repo.GetAssurance(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("identity not found")
	}
	existing.AssuranceLevel = domain.IAL2BiometricVerified
	now := time.Now().UTC()
	existing.BiometricVerifiedAt = &now
	if err := s.repo.UpsertAssurance(ctx, existing); err != nil {
		return nil, fmt.Errorf("step-up failed: %w", err)
	}

	s.publishEvent(ctx, "iam.assurance.stepup", existing)
	return existing, nil
}

func (s *IAMService) publishEvent(ctx context.Context, eventType string, data any) {
	if s.producer == nil {
		return
	}
	s.producer.Publish(ctx, kafka.Event{EventType: eventType, Timestamp: time.Now().UTC(), Data: data})
	if err := recover(); err != nil {
		log.Printf("publish failed: %v", err)
	}
}
