package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/counterintel-ht/internal/domain"
	"github.com/snisid/counterintel-ht/internal/kafka"
	"github.com/snisid/counterintel-ht/internal/repository"
)

type CounterintelService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewCounterintelService(repo repository.Repository, producer *kafka.Producer) *CounterintelService {
	return &CounterintelService{repo: repo, producer: producer}
}

func (s *CounterintelService) CreateInvestigation(ctx context.Context, req domain.CreateInvestigationRequest) (*domain.BackgroundInvestigation, error) {
	now := time.Now().UTC()
	inv := &domain.BackgroundInvestigation{
		ID:                   uuid.New(),
		SubjectIdentityRef:   req.SubjectIdentityRef,
		InvestigationType:    domain.InvType(req.InvestigationType),
		Status:               domain.InvPending,
		CriminalRecordCheck:  req.CriminalRecordCheck,
		FinancialCheck:       req.FinancialCheck,
		ForeignContactsCheck: req.ForeignContactsCheck,
		SocialMediaCheck:     req.SocialMediaCheck,
		DrugTest:             req.DrugTest,
		PsychEval:            req.PsychEval,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := s.repo.CreateInvestigation(ctx, inv); err != nil {
		return nil, fmt.Errorf("create investigation: %w", err)
	}
	s.publishEvent(ctx, "counterintel.investigation.created", inv.ID.String(), "investigation", inv)
	return inv, nil
}

func (s *CounterintelService) GetInvestigation(ctx context.Context, id uuid.UUID) (*domain.BackgroundInvestigation, error) {
	return s.repo.GetInvestigation(ctx, id)
}

func (s *CounterintelService) GetPendingInvestigations(ctx context.Context) ([]domain.BackgroundInvestigation, error) {
	return s.repo.GetPendingInvestigations(ctx)
}

func (s *CounterintelService) AdjudicateInvestigation(ctx context.Context, id uuid.UUID, req domain.AdjudicateRequest) (*domain.BackgroundInvestigation, error) {
	inv, err := s.repo.GetInvestigation(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	cl := domain.ClearanceLevel(req.ClearanceLevelGranted)
	inv.Status = domain.InvFavorable
	inv.Adjudicator = &req.Adjudicator
	inv.AdjudicationNotes = strPtr(req.AdjudicationNotes)
	inv.CompletedAt = &now
	inv.ClearanceLevelGranted = &cl
	inv.UpdatedAt = now

	expiresAt := now.AddDate(5, 0, 0)
	inv.ExpiresAt = &expiresAt

	if err := s.repo.UpdateInvestigation(ctx, inv); err != nil {
		return nil, fmt.Errorf("adjudicate investigation: %w", err)
	}
	s.publishEvent(ctx, "counterintel.investigation.adjudicated", inv.ID.String(), "investigation", inv)
	return inv, nil
}

func (s *CounterintelService) ReportThreat(ctx context.Context, req domain.ReportThreatRequest) (*domain.InsiderThreatAlert, error) {
	alert := &domain.InsiderThreatAlert{
		ID:           uuid.New(),
		SubjectID:    req.SubjectID,
		AlertType:    domain.AlertType(req.AlertType),
		Severity:     domain.Severity(req.Severity),
		Description:  req.Description,
		EvidenceRefs: req.EvidenceRefs,
		DetectedBy:   req.DetectedBy,
		Status:       domain.ThreatOpen,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.repo.CreateThreatAlert(ctx, alert); err != nil {
		return nil, fmt.Errorf("report threat: %w", err)
	}
	s.publishEvent(ctx, "counterintel.threat.reported", alert.ID.String(), "threat", alert)
	return alert, nil
}

func (s *CounterintelService) GetActiveThreats(ctx context.Context) ([]domain.InsiderThreatAlert, error) {
	return s.repo.GetActiveThreats(ctx)
}

func (s *CounterintelService) ReportContact(ctx context.Context, req domain.ReportContactRequest) (*domain.ForeignContact, error) {
	now := time.Now().UTC()
	c := &domain.ForeignContact{
		ID:                uuid.New(),
		SubjectID:         req.SubjectID,
		ContactName:       req.ContactName,
		ForeignGovernment: req.ForeignGovernment,
		RelationshipType:  domain.RelationshipType(req.RelationshipType),
		Frequency:         strPtr(req.Frequency),
		Notes:             strPtr(req.Notes),
		CreatedAt:         now,
	}
	if req.LastContactAt != "" {
		t, err := time.Parse(time.RFC3339, req.LastContactAt)
		if err == nil {
			c.LastContactAt = &t
		}
	}
	if req.ApprovedBy != "" {
		uid, err := uuid.Parse(req.ApprovedBy)
		if err == nil {
			c.ApprovedBy = &uid
		}
	}
	if err := s.repo.CreateForeignContact(ctx, c); err != nil {
		return nil, fmt.Errorf("report contact: %w", err)
	}
	s.publishEvent(ctx, "counterintel.contact.reported", c.ID.String(), "contact", c)
	return c, nil
}

func (s *CounterintelService) GetContactsBySubject(ctx context.Context, subjectID string) ([]domain.ForeignContact, error) {
	return s.repo.GetContactsBySubject(ctx, subjectID)
}

func (s *CounterintelService) publishEvent(ctx context.Context, eventType, entityID, entityType string, data any) {
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
