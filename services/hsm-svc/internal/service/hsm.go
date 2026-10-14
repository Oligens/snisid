package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/snisid/hsm-svc/internal/domain"
	"github.com/snisid/hsm-svc/internal/kafka"
	"github.com/snisid/hsm-svc/internal/repository"
)

type HSMService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewHSMService(repo repository.Repository, producer *kafka.Producer) *HSMService {
	return &HSMService{repo: repo, producer: producer}
}

func (s *HSMService) GenerateKey(ctx context.Context, req domain.KeyGenerationRequest) (*domain.KeyGenerationResponse, error) {
	keyID := uuid.New()

	var publicKeyPEM string
	switch req.Algorithm {
	case domain.AlgorithmRSA:
		key, err := rsa.GenerateKey(rand.Reader, req.KeySize)
		if err != nil {
			return nil, fmt.Errorf("generate RSA key: %w", err)
		}
		pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("marshal RSA public key: %w", err)
		}
		pubBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}
		publicKeyPEM = string(pem.EncodeToMemory(pubBlock))
	default:
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("generate default RSA key: %w", err)
		}
		pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("marshal default public key: %w", err)
		}
		pubBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}
		publicKeyPEM = string(pem.EncodeToMemory(pubBlock))
	}

	keyHash := sha256Hex(publicKeyPEM)

	var expiresAt *time.Time
	if req.ExpiresIn != "" {
		d, err := time.ParseDuration(req.ExpiresIn)
		if err == nil {
			t := time.Now().UTC().Add(d)
			expiresAt = &t
		}
	}

	now := time.Now().UTC()
	hsmKey := &domain.HSMKey{
		KeyID:         keyID,
		KeyLabel:      req.Label,
		Algorithm:     req.Algorithm,
		KeySize:       req.KeySize,
		State:         domain.KeyStateActive,
		Usages:        req.Usages,
		SlotID:        req.SlotID,
		IsExtractable: req.Extractable,
		PublicKeyPEM:  publicKeyPEM,
		KeyHash:       keyHash,
		ExpiresAt:     expiresAt,
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedBy:     req.CreatedBy,
	}

	if err := s.repo.CreateKey(ctx, hsmKey); err != nil {
		return nil, fmt.Errorf("save key: %w", err)
	}

	s.publishEvent(ctx, "hsm.key.generated", hsmKey)

	return &domain.KeyGenerationResponse{
		KeyID:        keyID,
		KeyLabel:     req.Label,
		Algorithm:    req.Algorithm,
		KeySize:      req.KeySize,
		SlotID:       req.SlotID,
		PublicKeyPEM: publicKeyPEM,
		CreatedAt:    now,
	}, nil
}

func (s *HSMService) GetKey(ctx context.Context, keyID uuid.UUID) (*domain.HSMKey, error) {
	return s.repo.FindByKeyID(ctx, keyID)
}

func (s *HSMService) WrapKey(ctx context.Context, req domain.KeyWrapRequest) (string, error) {
	targetKey, err := s.repo.FindByKeyID(ctx, req.TargetKeyID)
	if err != nil {
		return "", fmt.Errorf("target key not found: %w", err)
	}
	wrappingKey, err := s.repo.FindByKeyID(ctx, req.WrapKeyID)
	if err != nil {
		return "", fmt.Errorf("wrapping key not found: %w", err)
	}

	_ = targetKey
	_ = wrappingKey

	wrapped := hex.EncodeToString([]byte(req.Plaintext + ":wrapped_with:" + req.WrapKeyID.String()))

	s.publishEvent(ctx, "hsm.key.wrapped", map[string]any{
		"target_key_id": req.TargetKeyID.String(),
		"wrap_key_id":   req.WrapKeyID.String(),
		"timestamp":     time.Now().UTC(),
	})

	return wrapped, nil
}

func (s *HSMService) SignData(ctx context.Context, req domain.KeySignRequest) (string, error) {
	key, err := s.repo.FindByKeyID(ctx, req.KeyID)
	if err != nil {
		return "", fmt.Errorf("signing key not found: %w", err)
	}

	hash := sha256.Sum256([]byte(req.Data))
	signature := hex.EncodeToString(hash[:])

	s.publishEvent(ctx, "hsm.data.signed", map[string]any{
		"key_id":    key.KeyID.String(),
		"algorithm": req.Algorithm,
		"timestamp": time.Now().UTC(),
	})

	return signature, nil
}

func (s *HSMService) RotateKey(ctx context.Context, keyID uuid.UUID, newLabel string) (*domain.KeyGenerationResponse, error) {
	existingKey, err := s.repo.FindByKeyID(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("existing key not found: %w", err)
	}

	if err := s.repo.UpdateState(ctx, keyID, domain.KeyStatePendingRotate); err != nil {
		return nil, fmt.Errorf("mark existing key for rotation: %w", err)
	}

	rotateReq := domain.KeyGenerationRequest{
		Label:     newLabel,
		Algorithm: existingKey.Algorithm,
		KeySize:   existingKey.KeySize,
		Usages:    existingKey.Usages,
		SlotID:    existingKey.SlotID,
		CreatedBy: existingKey.CreatedBy,
	}

	newKey, err := s.GenerateKey(ctx, rotateReq)
	if err != nil {
		return nil, fmt.Errorf("generate rotated key: %w", err)
	}

	now := time.Now().UTC()
	if err := s.repo.UpdateRotatedAt(ctx, keyID, now); err != nil {
		return nil, fmt.Errorf("update rotated timestamp: %w", err)
	}
	if err := s.repo.UpdateState(ctx, newKey.KeyID, domain.KeyStateActive); err != nil {
		return nil, fmt.Errorf("activate new key: %w", err)
	}

	s.publishEvent(ctx, "hsm.key.rotated", map[string]any{
		"old_key_id": keyID.String(),
		"new_key_id": newKey.KeyID,
		"timestamp":  now,
	})

	return newKey, nil
}

func (s *HSMService) ListKeys(ctx context.Context, algorithm string, state string) ([]domain.HSMKey, error) {
	if algorithm != "" && state != "" {
		return s.repo.FindByAlgorithmAndState(ctx, domain.KeyAlgorithm(algorithm), domain.KeyState(state))
	}
	return s.repo.FindAll(ctx)
}

func (s *HSMService) publishEvent(ctx context.Context, eventType string, data any) {
	if s.producer == nil {
		return
	}
	evt := kafka.Event{
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
	if key, ok := data.(*domain.HSMKey); ok {
		evt.KeyID = key.KeyID.String()
		evt.KeyLabel = key.KeyLabel
	}
	if err := s.producer.Publish(ctx, evt); err != nil {
		log.Printf("failed to publish event: %v", err)
	}
}

func sha256Hex(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

func intPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

var _ = big.NewInt
