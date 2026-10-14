package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/snisid/mdl-svc/internal/domain"
	"github.com/snisid/mdl-svc/internal/kafka"
	"github.com/snisid/mdl-svc/internal/repository"
)

type MDLService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewMDLService(repo repository.Repository, producer *kafka.Producer) *MDLService {
	return &MDLService{repo: repo, producer: producer}
}

func (s *MDLService) IssueMDL(ctx context.Context, identityID uuid.UUID, deviceID string) (*domain.MDLIssuance, error) {
	issuance := &domain.MDLIssuance{
		IssuanceID: uuid.New(),
		IdentityID: identityID,
		DeviceID:   deviceID,
		IssuedAt:   time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().AddDate(4, 0, 0),
		IsRevoked:  false,
	}

	engagement := &domain.DeviceEngagement{
		EngagementID:   uuid.New(),
		IssuanceID:     issuance.IssuanceID,
		QRPayload:      fmt.Sprintf("mdl://%s/%s", issuance.IssuanceID.String(), deviceID),
		EngagementCode: uuid.New().String()[:8],
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(5 * time.Minute),
	}

	qr := &domain.QRBarcode{
		BarcodeID:    uuid.New(),
		EngagementID: engagement.EngagementID,
		EncodedData:  engagement.QRPayload,
		Format:       "QR",
		GeneratedAt:  time.Now().UTC(),
	}

	if err := s.repo.CreateIssuance(ctx, issuance); err != nil {
		return nil, fmt.Errorf("issue mdl: %w", err)
	}
	if err := s.repo.InsertDeviceEngagement(ctx, engagement); err != nil {
		return nil, fmt.Errorf("insert engagement: %w", err)
	}
	if err := s.repo.InsertQRBarcode(ctx, qr); err != nil {
		return nil, fmt.Errorf("insert qr: %w", err)
	}

	s.publishEvent(ctx, "mdl.issued", issuance)
	return issuance, nil
}

func (s *MDLService) GetMDLByIdentity(ctx context.Context, identityID uuid.UUID) (*domain.MDLIssuance, error) {
	return s.repo.FindIssuanceByIdentity(ctx, identityID)
}

func (s *MDLService) GetMDLByID(ctx context.Context, issuanceID uuid.UUID) (*domain.MDLIssuance, error) {
	return s.repo.FindIssuanceByID(ctx, issuanceID)
}

func (s *MDLService) VerifyMDLPresentation(ctx context.Context, issuanceID uuid.UUID, readerID string, elements map[string]string) (*domain.MDLVerification, error) {
	issuance, err := s.repo.FindIssuanceByID(ctx, issuanceID)
	if err != nil {
		return nil, fmt.Errorf("issuance not found: %w", err)
	}
	if issuance.IsRevoked {
		return nil, fmt.Errorf("mdl is revoked")
	}
	if time.Now().UTC().After(issuance.ExpiresAt) {
		return nil, fmt.Errorf("mdl is expired")
	}

	trustEntry, err := s.repo.FindTrustEntryByReaderID(ctx, readerID)
	if err != nil {
		return nil, fmt.Errorf("reader not trusted: %w", err)
	}
	if !trustEntry.IsTrusted {
		return nil, fmt.Errorf("reader is not trusted")
	}

	presentation := &domain.MDLPresentation{
		PresentationID:     uuid.New(),
		IssuanceID:         issuanceID,
		ReaderID:           readerID,
		PresentedAt:        time.Now().UTC(),
		IsVerified:         true,
		VerificationResult: "authentic",
	}
	if err := s.repo.InsertPresentation(ctx, presentation); err != nil {
		return nil, fmt.Errorf("insert presentation: %w", err)
	}

	verification := &domain.MDLVerification{
		VerificationID: uuid.New(),
		PresentationID: presentation.PresentationID,
		VerifiedBy:     readerID,
		VerifiedAt:     time.Now().UTC(),
		IsAuthentic:    true,
		Reason:         "all checks passed",
	}

	if err := s.repo.UpdatePresentationVerification(ctx, presentation.PresentationID, true, "authentic"); err != nil {
		return nil, fmt.Errorf("update verification: %w", err)
	}

	s.publishEvent(ctx, "mdl.verified", verification)
	return verification, nil
}

func (s *MDLService) RegisterTrustedReader(ctx context.Context, readerID, readerName, publicKey string) (*domain.MDLTrustRegistry, error) {
	entry := &domain.MDLTrustRegistry{
		EntryID:      uuid.New(),
		ReaderID:     readerID,
		ReaderName:   readerName,
		PublicKey:    publicKey,
		IsTrusted:    true,
		RegisteredAt: time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().AddDate(1, 0, 0),
	}
	if err := s.repo.InsertTrustRegistry(ctx, entry); err != nil {
		return nil, fmt.Errorf("register reader: %w", err)
	}
	s.publishEvent(ctx, "mdl.reader.registered", entry)
	return entry, nil
}

func (s *MDLService) GetTrustRegistry(ctx context.Context) ([]domain.MDLTrustRegistry, error) {
	return s.repo.FindTrustRegistry(ctx)
}

func (s *MDLService) ReissueMDL(ctx context.Context, identityID uuid.UUID, deviceID string) (*domain.MDLIssuance, error) {
	existing, err := s.repo.FindIssuanceByIdentity(ctx, identityID)
	if err == nil {
		if err := s.repo.UpdateIssuanceRevoked(ctx, existing.IssuanceID); err != nil {
			return nil, fmt.Errorf("revoke old mdl: %w", err)
		}
		s.publishEvent(ctx, "mdl.revoked", existing)
	}

	return s.IssueMDL(ctx, identityID, deviceID)
}

func (s *MDLService) GenerateQRCode(ctx context.Context, issuanceID uuid.UUID) (*domain.QRBarcode, error) {
	engagement, err := s.repo.FindEngagementByIssuance(ctx, issuanceID)
	if err != nil {
		return nil, fmt.Errorf("engagement not found: %w", err)
	}
	qr, err := s.repo.FindQRByEngagement(ctx, engagement.EngagementID)
	if err != nil {
		return nil, fmt.Errorf("qr not found: %w", err)
	}
	return qr, nil
}

func (s *MDLService) publishEvent(ctx context.Context, eventType string, data any) {
	if s.producer == nil {
		return
	}
	var issuanceID string
	if iss, ok := data.(*domain.MDLIssuance); ok {
		issuanceID = iss.IssuanceID.String()
	}
	evt := kafka.Event{
		EventType:  eventType,
		IssuanceID: issuanceID,
		Timestamp:  time.Now().UTC(),
		Data:       data,
	}
	if err := s.producer.Publish(ctx, evt); err != nil {
		log.Printf("failed to publish event: %v", err)
	}
}
