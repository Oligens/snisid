package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/snisid/sla-svc/internal/domain"
	"github.com/snisid/sla-svc/internal/kafka"
	"github.com/snisid/sla-svc/internal/repository"
)

type SLAService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewSLAService(repo repository.Repository, producer *kafka.Producer) *SLAService {
	return &SLAService{repo: repo, producer: producer}
}

func (s *SLAService) DefineSLA(ctx context.Context, name, description, owner string) (*domain.SLA, error) {
	sla := &domain.SLA{
		SLAID:       uuid.New(),
		Name:        name,
		Description: description,
		Owner:       owner,
		IsActive:    true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := s.repo.CreateSLA(ctx, sla); err != nil {
		return nil, fmt.Errorf("define sla: %w", err)
	}
	s.publishEvent(ctx, "sla.defined", sla.SLAID.String())
	return sla, nil
}

func (s *SLAService) ListSLAs(ctx context.Context) ([]domain.SLA, error) {
	return s.repo.ListSLAs(ctx)
}

func (s *SLAService) RecordSLI(ctx context.Context, slaID uuid.UUID, sloID uuid.UUID, name string, value float64) (*domain.ServiceLevelIndicator, error) {
	sli := &domain.ServiceLevelIndicator{
		SLIID:      uuid.New(),
		SLOID:      sloID,
		SLAID:      slaID,
		Name:       name,
		Value:      value,
		RecordedAt: time.Now().UTC(),
	}
	if err := s.repo.RecordSLI(ctx, sli); err != nil {
		return nil, fmt.Errorf("record sli: %w", err)
	}

	sloList, err := s.repo.GetSLOs(ctx, slaID)
	if err != nil {
		return sli, nil
	}

	for _, slo := range sloList {
		if slo.SLOID == sloID && value > slo.Threshold {
			breach := &domain.BreachRecord{
				BreachID:   uuid.New(),
				SLAID:      slaID,
				SLOID:      sloID,
				SLIValue:   value,
				Threshold:  slo.Threshold,
				DetectedAt: time.Now().UTC(),
				IsActive:   true,
			}
			if err := s.repo.CreateBreach(ctx, breach); err != nil {
				log.Printf("failed to record breach: %v", err)
			}
			s.publishEvent(ctx, "sla.breach.detected", slaID.String())
		}
	}

	s.publishEvent(ctx, "sla.sli.recorded", slaID.String())
	return sli, nil
}

func (s *SLAService) GetSLAStatus(ctx context.Context, slaID uuid.UUID) (*domain.SLAReport, error) {
	sla, err := s.repo.GetSLA(ctx, slaID)
	if err != nil {
		return nil, err
	}
	slos, err := s.repo.GetSLOs(ctx, slaID)
	if err != nil {
		return nil, err
	}
	breaches, err := s.repo.GetBreaches(ctx, slaID)
	if err != nil {
		return nil, err
	}

	compliance := make(map[string]float64)
	var totalScore float64
	for _, slo := range slos {
		compliance[slo.Name] = 100.0
		totalScore += 100.0
	}
	if len(slos) > 0 {
		totalScore /= float64(len(slos))
	}

	report := &domain.SLAReport{
		ReportID:      uuid.New(),
		SLAID:         sla.SLAID,
		SLOCompliance: compliance,
		OverallScore:  totalScore,
		BreachCount:   len(breaches),
		From:          sla.CreatedAt,
		To:            time.Now().UTC(),
		GeneratedAt:   time.Now().UTC(),
	}
	return report, nil
}

func (s *SLAService) GetBreaches(ctx context.Context, slaID uuid.UUID) ([]domain.BreachRecord, error) {
	return s.repo.GetBreaches(ctx, slaID)
}

func (s *SLAService) GetDashboard(ctx context.Context) ([]domain.SLAReport, error) {
	slas, err := s.repo.ListSLAs(ctx)
	if err != nil {
		return nil, err
	}
	var reports []domain.SLAReport
	for _, sla := range slas {
		report, err := s.GetSLAStatus(ctx, sla.SLAID)
		if err != nil {
			continue
		}
		reports = append(reports, *report)
	}
	return reports, nil
}

func (s *SLAService) TriggerEscalation(ctx context.Context, slaID uuid.UUID) error {
	s.publishEvent(ctx, "sla.escalation.triggered", slaID.String())
	return nil
}

func (s *SLAService) publishEvent(ctx context.Context, eventType string, slaID string) {
	if s.producer == nil {
		return
	}
	evt := kafka.Event{
		EventType: eventType,
		SLAID:     slaID,
		Timestamp: time.Now().UTC(),
		Data:      map[string]string{"sla_id": slaID},
	}
	if err := s.producer.Publish(ctx, evt); err != nil {
		log.Printf("failed to publish event: %v", err)
	}
}
