package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/snisid/fips-cert-svc/internal/domain"
	"github.com/snisid/fips-cert-svc/internal/kafka"
	"github.com/snisid/fips-cert-svc/internal/repository"
)

type FIPSService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewFIPSService(repo repository.Repository, producer *kafka.Producer) *FIPSService {
	return &FIPSService{repo: repo, producer: producer}
}

func (s *FIPSService) RegisterModule(ctx context.Context, mod domain.CryptoModule) (*domain.CryptoModule, error) {
	mod.ModuleID = uuid.New()
	mod.Status = domain.StatusPending
	mod.CreatedAt = time.Now().UTC()
	mod.UpdatedAt = time.Now().UTC()

	if err := s.repo.CreateModule(ctx, &mod); err != nil {
		return nil, fmt.Errorf("register module: %w", err)
	}

	s.publishEvent(ctx, "fips.module.registered", mod.ModuleID.String(), &mod)
	return &mod, nil
}

func (s *FIPSService) ListModules(ctx context.Context) ([]domain.CryptoModule, error) {
	return s.repo.ListModules(ctx)
}

func (s *FIPSService) GetModule(ctx context.Context, id uuid.UUID) (*domain.CryptoModule, error) {
	return s.repo.FindModuleByID(ctx, id)
}

func (s *FIPSService) SubmitValidation(ctx context.Context, moduleID uuid.UUID, certNumber string, validationDate time.Time) (*domain.CryptoModule, error) {
	if err := s.repo.UpdateValidation(ctx, moduleID, domain.StatusValidated, certNumber, validationDate); err != nil {
		return nil, fmt.Errorf("submit validation: %w", err)
	}

	mod, err := s.repo.FindModuleByID(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	s.publishEvent(ctx, "fips.module.validated", moduleID.String(), mod)
	return mod, nil
}

func (s *FIPSService) ReportCVE(ctx context.Context, moduleID uuid.UUID, cveID, severity string, notes *string) (*domain.CVEScanResult, error) {
	scan := &domain.CVEScanResult{
		ScanID:     uuid.New(),
		ModuleID:   moduleID,
		CVEID:      cveID,
		Severity:   severity,
		Discovered: time.Now().UTC(),
		Patched:    boolPtr(false),
		Notes:      notes,
	}

	if err := s.repo.CreateCVEResult(ctx, scan); err != nil {
		return nil, fmt.Errorf("report cve: %w", err)
	}

	s.publishEvent(ctx, "fips.cve.reported", moduleID.String(), scan)
	return scan, nil
}

func (s *FIPSService) GetComplianceByService(ctx context.Context, service string) (*domain.ComplianceReport, error) {
	return s.repo.GetComplianceByService(ctx, service)
}

func (s *FIPSService) GetDashboard(ctx context.Context) ([]domain.ComplianceReport, error) {
	return s.repo.GetDashboard(ctx)
}

func (s *FIPSService) publishEvent(ctx context.Context, eventType, moduleID string, data any) {
	if s.producer == nil {
		return
	}
	evt := kafka.Event{
		EventType: eventType,
		ModuleID:  moduleID,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
	if err := s.producer.Publish(ctx, evt); err != nil {
		log.Printf("failed to publish event: %v", err)
	}
}

func boolPtr(b bool) *bool {
	return &b
}
