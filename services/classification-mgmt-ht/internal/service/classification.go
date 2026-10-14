package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/classification-mgmt-ht/internal/domain"
	"github.com/snisid/classification-mgmt-ht/internal/kafka"
	"github.com/snisid/classification-mgmt-ht/internal/repository"
)

type ClassificationService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewClassificationService(repo repository.Repository, producer *kafka.Producer) *ClassificationService {
	return &ClassificationService{repo: repo, producer: producer}
}

func (s *ClassificationService) CreateRule(ctx context.Context, req domain.CreateRuleRequest) (*domain.ClassificationRule, error) {
	now := time.Now().UTC()
	createdBy, _ := uuid.Parse(req.CreatedBy)
	rule := &domain.ClassificationRule{
		ID:                  uuid.New(),
		DataType:            req.DataType,
		SensitivityLevel:    domain.SensitivityLevel(req.SensitivityLevel),
		HandlingCaveats:     req.HandlingCaveats,
		DisseminationLimit:  strPtr(req.DisseminationLimit),
		EncryptionRequired:  req.EncryptionRequired,
		AccessControlMFA:    req.AccessControlMFA,
		AuditLogging:        req.AuditLogging,
		RetentionDays:       req.RetentionDays,
		DestructionRequired: req.DestructionRequired,
		CreatedBy:           createdBy,
		Active:              true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := s.repo.CreateRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("create rule: %w", err)
	}
	s.publishEvent(ctx, "classification.rule.created", rule.ID.String(), "rule", rule)
	return rule, nil
}

func (s *ClassificationService) GetRulesByDataType(ctx context.Context, dataType string) ([]domain.ClassificationRule, error) {
	return s.repo.GetRulesByDataType(ctx, dataType)
}

func (s *ClassificationService) TagResource(ctx context.Context, req domain.TagResourceRequest) (*domain.DataTag, error) {
	now := time.Now().UTC()
	taggedBy, _ := uuid.Parse(req.TaggedBy)
	tag := &domain.DataTag{
		ID:                   uuid.New(),
		ResourceURI:          req.ResourceURI,
		ClassificationTop:    domain.SensitivityLevel(req.ClassificationTop),
		ClassificationAtomic: req.ClassificationAtomic,
		HandlingCaveats:      req.HandlingCaveats,
		OwnerAgency:          req.OwnerAgency,
		TaggedBy:             taggedBy,
		TaggedAt:             now,
	}
	if err := s.repo.CreateTag(ctx, tag); err != nil {
		return nil, fmt.Errorf("tag resource: %w", err)
	}
	s.publishEvent(ctx, "classification.tag.created", tag.ID.String(), "tag", tag)
	return tag, nil
}

func (s *ClassificationService) GetClassificationByURI(ctx context.Context, uri string) (*domain.DataTag, error) {
	return s.repo.GetTagByURI(ctx, uri)
}

func (s *ClassificationService) LogAudit(ctx context.Context, req domain.LogAuditRequest) (*domain.ClassificationAudit, error) {
	authorizedBy, _ := uuid.Parse(req.AuthorizedBy)
	entry := &domain.ClassificationAudit{
		ID:                      uuid.New(),
		ResourceURI:             req.ResourceURI,
		Action:                  domain.AuditAction(req.Action),
		FromLevel:               strPtr(req.FromLevel),
		ToLevel:                 strPtr(req.ToLevel),
		Rationale:               strPtr(req.Rationale),
		AuthorizedBy:            authorizedBy,
		ClassificationAuthority: req.ClassificationAuthority,
		Timestamp:               time.Now().UTC(),
		IPAddress:               req.IPAddress,
	}
	if err := s.repo.CreateAuditLog(ctx, entry); err != nil {
		return nil, fmt.Errorf("log audit: %w", err)
	}
	s.publishEvent(ctx, "classification.audit.logged", entry.ID.String(), "audit", entry)
	return entry, nil
}

func (s *ClassificationService) GetRecentAuditLogs(ctx context.Context) ([]domain.ClassificationAudit, error) {
	return s.repo.GetRecentAuditLogs(ctx)
}

func (s *ClassificationService) GetDashboard(ctx context.Context) (*domain.DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx)
}

func (s *ClassificationService) publishEvent(ctx context.Context, eventType, entityID, entityType string, data any) {
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
