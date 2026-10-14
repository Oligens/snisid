package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/snisid/certification-svc/internal/domain"
	"github.com/snisid/certification-svc/internal/kafka"
	"github.com/snisid/certification-svc/internal/repository"
)

type CertificationService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewCertificationService(repo repository.Repository, producer *kafka.Producer) *CertificationService {
	return &CertificationService{repo: repo, producer: producer}
}

func (s *CertificationService) CreateProfile(ctx context.Context, profile domain.AssuranceProfile) (*domain.AssuranceProfile, error) {
	profile.ProfileID = uuid.New()
	profile.IsActive = true
	profile.ValidFrom = time.Now().UTC()
	profile.LastAssessed = time.Now().UTC()
	profile.CreatedAt = time.Now().UTC()
	profile.UpdatedAt = time.Now().UTC()

	if profile.IAL == "" {
		profile.IAL = domain.IALNone
	}
	if profile.AAL == "" {
		profile.AAL = domain.AALNone
	}
	if profile.FAL == "" {
		profile.FAL = domain.FALNone
	}

	if err := s.repo.CreateProfile(ctx, &profile); err != nil {
		return nil, fmt.Errorf("create certification profile: %w", err)
	}

	s.logAudit(ctx, profile.IdentityID, "PROFILE_CREATED", "", "", "", profile.AssessorID)
	s.publishEvent(ctx, "certification.profile.created", &profile)
	return &profile, nil
}

func (s *CertificationService) GetProfile(ctx context.Context, identityID uuid.UUID) (*domain.AssuranceProfile, error) {
	return s.repo.FindProfileByIDentityID(ctx, identityID)
}

func (s *CertificationService) UpdateIAL(ctx context.Context, identityID uuid.UUID, ial domain.IALLevel, updatedBy string) (*domain.AssuranceProfile, error) {
	profile, err := s.repo.FindProfileByIDentityID(ctx, identityID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	oldIAL := string(profile.IAL)
	if err := s.repo.UpdateIAL(ctx, identityID, ial, updatedBy); err != nil {
		return nil, fmt.Errorf("update IAL: %w", err)
	}

	profile.IAL = ial
	profile.UpdatedAt = time.Now().UTC()

	s.logAudit(ctx, identityID, "IAL_UPDATED", "ial", oldIAL, string(ial), updatedBy)
	s.publishEvent(ctx, "certification.ial.updated", map[string]any{
		"identity_id": identityID.String(),
		"old_ial":     oldIAL,
		"new_ial":     string(ial),
	})

	return profile, nil
}

func (s *CertificationService) UpdateAAL(ctx context.Context, identityID uuid.UUID, aal domain.AALLevel, updatedBy string) (*domain.AssuranceProfile, error) {
	profile, err := s.repo.FindProfileByIDentityID(ctx, identityID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	oldAAL := string(profile.AAL)
	if err := s.repo.UpdateAAL(ctx, identityID, aal, updatedBy); err != nil {
		return nil, fmt.Errorf("update AAL: %w", err)
	}

	profile.AAL = aal
	profile.UpdatedAt = time.Now().UTC()

	s.logAudit(ctx, identityID, "AAL_UPDATED", "aal", oldAAL, string(aal), updatedBy)
	s.publishEvent(ctx, "certification.aal.updated", map[string]any{
		"identity_id": identityID.String(),
		"old_aal":     oldAAL,
		"new_aal":     string(aal),
	})

	return profile, nil
}

func (s *CertificationService) VerifyCompliance(ctx context.Context, identityID uuid.UUID, requiredIAL domain.IALLevel, requiredAAL domain.AALLevel) (map[string]any, error) {
	profile, err := s.repo.FindProfileByIDentityID(ctx, identityID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	ialMet := meetsIAL(profile.IAL, requiredIAL)
	aalMet := meetsAAL(profile.AAL, requiredAAL)
	overall := ialMet && aalMet

	result := map[string]any{
		"identity_id":  identityID.String(),
		"is_compliant": overall,
		"ial_required": string(requiredIAL),
		"ial_current":  string(profile.IAL),
		"ial_met":      ialMet,
		"aal_required": string(requiredAAL),
		"aal_current":  string(profile.AAL),
		"aal_met":      aalMet,
		"checked_at":   time.Now().UTC(),
	}

	s.publishEvent(ctx, "certification.compliance.checked", result)
	return result, nil
}

func (s *CertificationService) GetAudit(ctx context.Context, identityID uuid.UUID) ([]domain.CertificationAudit, error) {
	if identityID != uuid.Nil {
		return s.repo.FindAuditByIDentityID(ctx, identityID)
	}
	return s.repo.FindAllAudit(ctx)
}

func (s *CertificationService) logAudit(ctx context.Context, identityID uuid.UUID, action, field, oldValue, newValue, performedBy string) {
	audit := &domain.CertificationAudit{
		AuditID:     uuid.New(),
		IdentityID:  identityID,
		Action:      action,
		Field:       field,
		OldValue:    oldValue,
		NewValue:    newValue,
		PerformedBy: performedBy,
		PerformedAt: time.Now().UTC(),
	}
	if err := s.repo.CreateAuditEntry(ctx, audit); err != nil {
		log.Printf("failed to create audit entry: %v", err)
	}
}

func (s *CertificationService) publishEvent(ctx context.Context, eventType string, data any) {
	if s.producer == nil {
		return
	}
	evt := kafka.Event{
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
	if profile, ok := data.(*domain.AssuranceProfile); ok {
		evt.IdentityID = profile.IdentityID.String()
		evt.ProfileID = profile.ProfileID.String()
	}
	if m, ok := data.(map[string]any); ok {
		if rid, ok := m["identity_id"]; ok {
			evt.IdentityID = fmt.Sprintf("%v", rid)
		}
	}
	if err := s.producer.Publish(ctx, evt); err != nil {
		log.Printf("failed to publish event: %v", err)
	}
}

func meetsIAL(current, required domain.IALLevel) bool {
	levels := map[domain.IALLevel]int{
		domain.IALNone: 0,
		domain.IAL1:    1,
		domain.IAL2:    2,
		domain.IAL3:    3,
	}
	return levels[current] >= levels[required]
}

func meetsAAL(current, required domain.AALLevel) bool {
	levels := map[domain.AALLevel]int{
		domain.AALNone: 0,
		domain.AAL1:    1,
		domain.AAL2:    2,
		domain.AAL3:    3,
	}
	return levels[current] >= levels[required]
}
