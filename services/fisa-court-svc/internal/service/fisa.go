package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/fisa-court-svc/internal/domain"
	"github.com/snisid/fisa-court-svc/internal/kafka"
	"github.com/snisid/fisa-court-svc/internal/repository"
)

type FISAService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewFISAService(repo repository.Repository, producer *kafka.Producer) *FISAService {
	return &FISAService{repo: repo, producer: producer}
}

func (s *FISAService) FileWarrant(ctx context.Context, req domain.FileWarrantRequest) (*domain.SurveillanceWarrant, error) {
	now := time.Now().UTC()
	officerID, _ := uuid.Parse(req.ApplicantOfficer)
	w := &domain.SurveillanceWarrant{
		ID:                   uuid.New(),
		WarrantID:            fmt.Sprintf("FISA-%s", uuid.New().String()[:8]),
		WarrantType:          domain.WarrantType(req.WarrantType),
		TargetIdentity:       req.TargetIdentity,
		TargetDetails:        strPtr(req.TargetDetails),
		IssuingCourt:         req.IssuingCourt,
		JudgeName:            req.JudgeName,
		ApplicantAgency:      req.ApplicantAgency,
		ApplicantOfficer:     officerID,
		ProbableCauseSummary: strPtr(req.ProbableCauseSummary),
		DurationDays:         req.DurationDays,
		Renewals:             0,
		Status:               domain.WarrantPending,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := s.repo.CreateWarrant(ctx, w); err != nil {
		return nil, fmt.Errorf("file warrant: %w", err)
	}
	s.publishEvent(ctx, "fisa.warrant.filed", w.ID.String(), "warrant", w)
	return w, nil
}

func (s *FISAService) ApproveWarrant(ctx context.Context, id uuid.UUID, req domain.ApproveWarrantRequest) (*domain.SurveillanceWarrant, error) {
	w, err := s.repo.GetWarrant(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	start := now
	end := now.AddDate(0, 0, w.DurationDays)
	reviewAt := now.AddDate(0, 3, 0)

	w.Status = domain.WarrantActive
	w.AuthorizedStart = &start
	w.AuthorizedEnd = &end
	w.ReviewRequiredAt = &reviewAt
	w.UpdatedAt = now

	if err := s.repo.UpdateWarrant(ctx, w); err != nil {
		return nil, fmt.Errorf("approve warrant: %w", err)
	}
	s.publishEvent(ctx, "fisa.warrant.approved", w.ID.String(), "warrant", w)
	return w, nil
}

func (s *FISAService) GetActiveWarrants(ctx context.Context) ([]domain.SurveillanceWarrant, error) {
	return s.repo.GetActiveWarrants(ctx)
}

func (s *FISAService) RenewWarrant(ctx context.Context, id uuid.UUID, req domain.RenewWarrantRequest) (*domain.SurveillanceWarrant, error) {
	w, err := s.repo.GetWarrant(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	w.Renewals++
	w.DurationDays = req.DurationDays
	w.AuthorizedEnd = nil
	w.UpdatedAt = now

	if err := s.repo.UpdateWarrant(ctx, w); err != nil {
		return nil, fmt.Errorf("renew warrant: %w", err)
	}
	s.publishEvent(ctx, "fisa.warrant.renewed", w.ID.String(), "warrant", w)
	return w, nil
}

func (s *FISAService) FileReport(ctx context.Context, req domain.FileReportRequest) (*domain.SurveillanceReport, error) {
	warrantID, _ := uuid.Parse(req.WarrantID)
	submittedBy, _ := uuid.Parse(req.SubmittedBy)
	periodStart, _ := time.Parse(time.RFC3339, req.ReportingPeriodStart)
	periodEnd, _ := time.Parse(time.RFC3339, req.ReportingPeriodEnd)

	rep := &domain.SurveillanceReport{
		ID:                        uuid.New(),
		WarrantID:                 warrantID,
		ReportingPeriodStart:      periodStart,
		ReportingPeriodEnd:        periodEnd,
		CommunicationsIntercepted: req.CommunicationsIntercepted,
		MinimizationApplied:       req.MinimizationApplied,
		IncidentalCollection:      req.IncidentalCollection,
		USPersonIdentities:        req.USPersonIdentities,
		ResultsSummary:            strPtr(req.ResultsSummary),
		SubmittedBy:               submittedBy,
		SubmittedAt:               time.Now().UTC(),
	}
	if err := s.repo.CreateReport(ctx, rep); err != nil {
		return nil, fmt.Errorf("file report: %w", err)
	}
	s.publishEvent(ctx, "fisa.report.filed", rep.ID.String(), "report", rep)
	return rep, nil
}

func (s *FISAService) GetDocketByTerm(ctx context.Context, term string) (*domain.FISADocket, error) {
	return s.repo.GetDocketByTerm(ctx, term)
}

func (s *FISAService) EmergencyAuthorization(ctx context.Context, req domain.EmergencyAuthorizationRequest) (*domain.SurveillanceWarrant, error) {
	now := time.Now().UTC()
	officerID, _ := uuid.Parse(req.ApplicantOfficer)
	approvedBy, _ := uuid.Parse(req.ApprovedBy)
	end := now.AddDate(0, 0, 3)

	w := &domain.SurveillanceWarrant{
		ID:                   uuid.New(),
		WarrantID:            fmt.Sprintf("EMERGENCY-%s", uuid.New().String()[:8]),
		WarrantType:          domain.WarrantType(req.WarrantType),
		TargetIdentity:       req.TargetIdentity,
		TargetDetails:        strPtr(req.TargetDetails),
		IssuingCourt:         "FISA Court",
		JudgeName:            "Emergency Docket",
		ApplicantAgency:      req.ApplicantAgency,
		ApplicantOfficer:     officerID,
		ProbableCauseSummary: strPtr(req.ProbableCause),
		DurationDays:         3,
		AuthorizedStart:      &now,
		AuthorizedEnd:        &end,
		Renewals:             0,
		Status:               domain.WarrantActive,
		ReviewRequiredAt:     &end,
		EmergencyAuthorized:  true,
		EmergencyApprovedBy:  &approvedBy,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := s.repo.CreateWarrant(ctx, w); err != nil {
		return nil, fmt.Errorf("emergency authorization: %w", err)
	}
	s.publishEvent(ctx, "fisa.emergency.authorized", w.ID.String(), "warrant", w)
	return w, nil
}

func (s *FISAService) publishEvent(ctx context.Context, eventType, entityID, entityType string, data any) {
	if s.producer == nil {
		return
	}
	evt := kafka.Event{
		EventType:  eventType,
		EntityID:   entityID,
		EntityType: entityType,
		Timestamp:  time.Now().UTC(),
		Data:       data,
	}
	if err := s.producer.Publish(ctx, evt); err != nil {
		log.Printf("failed to publish event: %v", err)
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
