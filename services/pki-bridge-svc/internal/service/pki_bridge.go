package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/snisid/pki-bridge-svc/internal/domain"
	"github.com/snisid/pki-bridge-svc/internal/kafka"
	"github.com/snisid/pki-bridge-svc/internal/repository"
)

type PKIBridgeService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewPKIBridgeService(repo repository.Repository, producer *kafka.Producer) *PKIBridgeService {
	return &PKIBridgeService{repo: repo, producer: producer}
}

func (s *PKIBridgeService) RegisterForeignCA(ctx context.Context, ca domain.ForeignCA) (*domain.ForeignCA, error) {
	ca.CAID = uuid.New()
	ca.RegisteredAt = time.Now().UTC()
	ca.Status = "ACTIVE"

	if err := s.repo.CreateForeignCA(ctx, &ca); err != nil {
		return nil, fmt.Errorf("register foreign ca: %w", err)
	}

	s.publishEvent(ctx, "pki-bridge.foreign-ca.registered", ca.CAID.String(), &ca)
	return &ca, nil
}

func (s *PKIBridgeService) IssueCrossCert(ctx context.Context, subject string, issuerCAID uuid.UUID, serialNumber string, notBefore, notAfter time.Time, certPEM string) (*domain.CrossCertificate, error) {
	cert := &domain.CrossCertificate{
		CrossCertID:    uuid.New(),
		Subject:        subject,
		IssuerCAID:     issuerCAID,
		SerialNumber:   serialNumber,
		NotBefore:      notBefore,
		NotAfter:       notAfter,
		CertificatePEM: certPEM,
		CreatedAt:      time.Now().UTC(),
	}

	if err := s.repo.CreateCrossCert(ctx, cert); err != nil {
		return nil, fmt.Errorf("issue cross cert: %w", err)
	}

	s.publishEvent(ctx, "pki-bridge.cross-cert.issued", cert.CrossCertID.String(), cert)
	return cert, nil
}

func (s *PKIBridgeService) GetCrossCert(ctx context.Context, subject string) (*domain.CrossCertificate, error) {
	return s.repo.FindCrossCertBySubject(ctx, subject)
}

func (s *PKIBridgeService) ListTrustAnchors(ctx context.Context) ([]domain.TrustAnchor, error) {
	return s.repo.ListTrustAnchors(ctx)
}

func (s *PKIBridgeService) ValidatePath(ctx context.Context, leafSubject string, intermediates []string, rootSubject string) (*domain.PathValidation, error) {
	path := domain.CertificatePath{
		PathID:        uuid.New(),
		LeafSubject:   leafSubject,
		Intermediates: intermediates,
		RootSubject:   rootSubject,
		Valid:         false,
		ValidatedAt:   time.Now().UTC(),
	}

	validation := &domain.PathValidation{
		ValidationID: uuid.New(),
		PathID:       path.PathID,
		Result:       true,
		Errors:       []string{},
		ValidatedAt:  time.Now().UTC(),
	}

	for i := 0; i < len(intermediates); i++ {
		if intermediates[i] == "" {
			validation.Result = false
			validation.Errors = append(validation.Errors, fmt.Sprintf("empty intermediate at index %d", i))
		}
	}

	if leafSubject == "" {
		validation.Result = false
		validation.Errors = append(validation.Errors, "leaf subject is empty")
	}

	path.Valid = validation.Result

	if err := s.repo.SavePathValidation(ctx, validation); err != nil {
		return nil, fmt.Errorf("validate path: %w", err)
	}

	s.publishEvent(ctx, "pki-bridge.path.validated", validation.ValidationID.String(), validation)
	return validation, nil
}

func (s *PKIBridgeService) ListAgreements(ctx context.Context) ([]domain.BridgeAgreement, error) {
	return s.repo.ListBridgeAgreements(ctx)
}

func (s *PKIBridgeService) CreateAgreement(ctx context.Context, name, partnerCA string, policyID uuid.UUID, expiresAt *time.Time) (*domain.BridgeAgreement, error) {
	agreement := &domain.BridgeAgreement{
		AgreementID: uuid.New(),
		Name:        name,
		PartnerCA:   partnerCA,
		PolicyID:    policyID,
		SignedAt:    time.Now().UTC(),
		ExpiresAt:   expiresAt,
		Status:      "ACTIVE",
	}

	if err := s.repo.CreateBridgeAgreement(ctx, agreement); err != nil {
		return nil, fmt.Errorf("create agreement: %w", err)
	}

	s.publishEvent(ctx, "pki-bridge.agreement.created", agreement.AgreementID.String(), agreement)
	return agreement, nil
}

func (s *PKIBridgeService) publishEvent(ctx context.Context, eventType, id string, data any) {
	if s.producer == nil {
		return
	}
	evt := kafka.Event{
		EventType:     eventType,
		CertificateID: id,
		Timestamp:     time.Now().UTC(),
		Data:          data,
	}
	if err := s.producer.Publish(ctx, evt); err != nil {
		log.Printf("failed to publish event: %v", err)
	}
}
