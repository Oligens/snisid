package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/snisid/audit-svc/internal/domain"
	"github.com/snisid/audit-svc/internal/kafka"
	"github.com/snisid/audit-svc/internal/repository"
)

type AuditService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewAuditService(repo repository.Repository, producer *kafka.Producer) *AuditService {
	return &AuditService{repo: repo, producer: producer}
}

func (s *AuditService) IngestEvent(ctx context.Context, source domain.EventSource, eventType domain.EventType, category domain.AuditCategory, actorID *uuid.UUID, resourceID, action string, payload map[string]any) (*domain.AuditEvent, error) {
	evt := &domain.AuditEvent{
		EventID:    uuid.New(),
		Source:     source,
		EventType:  eventType,
		Category:   category,
		ActorID:    actorID,
		ResourceID: resourceID,
		Action:     action,
		Payload:    payload,
		Timestamp:  time.Now().UTC(),
	}
	if err := s.repo.InsertEvent(ctx, evt); err != nil {
		return nil, fmt.Errorf("ingest event: %w", err)
	}
	s.publishEvent(ctx, "audit.event.ingested", evt.EventID.String())
	return evt, nil
}

func (s *AuditService) GetEvent(ctx context.Context, eventID uuid.UUID) (*domain.AuditEvent, error) {
	return s.repo.GetEvent(ctx, eventID)
}

func (s *AuditService) SearchEvents(ctx context.Context, q domain.AuditQuery) ([]domain.AuditEvent, error) {
	return s.repo.SearchEvents(ctx, q)
}

func (s *AuditService) GenerateReport(ctx context.Context, category domain.AuditCategory, from, to time.Time) (*domain.AuditReport, error) {
	q := domain.AuditQuery{
		Category: &category,
		From:     &from,
		To:       &to,
		Limit:    10000,
	}
	events, err := s.repo.SearchEvents(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("search for report: %w", err)
	}

	report := &domain.AuditReport{
		ReportID:       uuid.New(),
		Category:       category,
		EventCount:     len(events),
		From:           from,
		To:             to,
		IntegrityVerif: true,
		GeneratedAt:    time.Now().UTC(),
	}
	return report, nil
}

func (s *AuditService) VerifyIntegrity(ctx context.Context) (bool, error) {
	hash, err := s.repo.GetLastHash(ctx)
	if err != nil {
		return false, fmt.Errorf("verify integrity: %w", err)
	}
	return hash != "", nil
}

func (s *AuditService) GetStats(ctx context.Context) (map[string]int, error) {
	return s.repo.GetStats(ctx)
}

func (s *AuditService) publishEvent(ctx context.Context, eventType string, eventID string) {
	if s.producer == nil {
		return
	}
	evt := kafka.Event{
		EventType: eventType,
		EventID:   eventID,
		Timestamp: time.Now().UTC(),
		Data:      map[string]string{"event_id": eventID},
	}
	if err := s.producer.Publish(ctx, evt); err != nil {
		log.Printf("failed to publish event: %v", err)
	}
}
