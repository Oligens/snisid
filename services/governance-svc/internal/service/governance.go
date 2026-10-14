package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/snisid/governance-svc/internal/domain"
	"github.com/snisid/governance-svc/internal/kafka"
	"github.com/snisid/governance-svc/internal/repository"
)

type GovernanceService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewGovernanceService(repo repository.Repository, producer *kafka.Producer) *GovernanceService {
	return &GovernanceService{repo: repo, producer: producer}
}

func (s *GovernanceService) RegisterLicense(ctx context.Context, name, spdxID string, licenseType domain.LicenseType, version, publisher string, isOsiApproved bool, text string) (*domain.SoftwareLicense, error) {
	l := &domain.SoftwareLicense{
		LicenseID:     uuid.New(),
		Name:          name,
		SPDXID:        spdxID,
		LicenseType:   licenseType,
		Version:       version,
		Publisher:     publisher,
		IsOsiApproved: isOsiApproved,
		Text:          text,
		RegisteredAt:  time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if err := s.repo.CreateLicense(ctx, l); err != nil {
		return nil, fmt.Errorf("register license: %w", err)
	}
	s.publishEvent(ctx, "governance.license.registered", l.LicenseID.String())
	return l, nil
}

func (s *GovernanceService) ListLicenses(ctx context.Context) ([]domain.SoftwareLicense, error) {
	return s.repo.ListLicenses(ctx)
}

func (s *GovernanceService) CreatePolicy(ctx context.Context, name, description string) (*domain.GovernancePolicy, error) {
	p := &domain.GovernancePolicy{
		PolicyID:    uuid.New(),
		Name:        name,
		Description: description,
		IsActive:    true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := s.repo.CreatePolicy(ctx, p); err != nil {
		return nil, fmt.Errorf("create policy: %w", err)
	}
	s.publishEvent(ctx, "governance.policy.created", p.PolicyID.String())
	return p, nil
}

func (s *GovernanceService) ListPolicies(ctx context.Context) ([]domain.GovernancePolicy, error) {
	return s.repo.ListPolicies(ctx)
}

func (s *GovernanceService) CheckCompliance(ctx context.Context, licenseID, policyID uuid.UUID) (*domain.LicenseAudit, error) {
	license, err := s.repo.GetLicense(ctx, licenseID)
	if err != nil {
		return nil, err
	}
	rules, err := s.repo.GetPolicyRules(ctx, policyID)
	if err != nil {
		return nil, err
	}

	status := domain.ComplianceStatusCompliant
	findings := "All rules satisfied"
	for _, rule := range rules {
		if rule.RuleType == "LICENSE_TYPE" && string(license.LicenseType) != rule.Condition {
			status = domain.ComplianceStatusNonCompliant
			findings = fmt.Sprintf("License type %s does not match required %s", license.LicenseType, rule.Condition)
			break
		}
	}

	audit := &domain.LicenseAudit{
		AuditID:   uuid.New(),
		LicenseID: licenseID,
		PolicyID:  policyID,
		Status:    status,
		Findings:  findings,
		AuditedAt: time.Now().UTC(),
	}
	if err := s.repo.CreateAudit(ctx, audit); err != nil {
		return nil, fmt.Errorf("create audit: %w", err)
	}
	s.publishEvent(ctx, "governance.compliance.checked", licenseID.String())
	return audit, nil
}

func (s *GovernanceService) GenerateComplianceReport(ctx context.Context) ([]domain.LicenseAudit, error) {
	licenses, err := s.repo.ListAllLicenses(ctx)
	if err != nil {
		return nil, err
	}
	var allAudits []domain.LicenseAudit
	for _, l := range licenses {
		audits, err := s.repo.GetAudits(ctx, l.LicenseID)
		if err != nil {
			continue
		}
		allAudits = append(allAudits, audits...)
	}
	return allAudits, nil
}

func (s *GovernanceService) GenerateAttribution(ctx context.Context) (*domain.AttributionReport, error) {
	licenses, err := s.repo.ListAllLicenses(ctx)
	if err != nil {
		return nil, err
	}
	report := &domain.AttributionReport{
		ReportID:      uuid.New(),
		Components:    licenses,
		GeneratedAt:   time.Now().UTC(),
		TotalLicenses: len(licenses),
	}
	return report, nil
}

func (s *GovernanceService) publishEvent(ctx context.Context, eventType string, id string) {
	if s.producer == nil {
		return
	}
	evt := kafka.Event{
		EventType: eventType,
		LicenseID: id,
		Timestamp: time.Now().UTC(),
		Data:      map[string]string{"id": id},
	}
	if err := s.producer.Publish(ctx, evt); err != nil {
		log.Printf("failed to publish event: %v", err)
	}
}
