package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/interop-ht/internal/domain"
	"github.com/snisid/interop-ht/internal/kafka"
	"github.com/snisid/interop-ht/internal/repository"
)

type InteropService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewInteropService(repo repository.Repository, producer *kafka.Producer) *InteropService {
	return &InteropService{repo: repo, producer: producer}
}

func (s *InteropService) CreateAgreement(ctx context.Context, req domain.DataExchangeAgreement) (*domain.DataExchangeAgreement, error) {
	req.AgreementID = uuid.New()
	req.CreatedAt = time.Now().UTC()
	req.IsActive = true
	if req.RateLimitPerMin == 0 {
		req.RateLimitPerMin = 1000
	}
	if err := s.repo.CreateAgreement(ctx, &req); err != nil {
		return nil, err
	}
	s.publish(ctx, "interop.agreement.created", &req)
	return &req, nil
}
func (s *InteropService) Exchange(ctx context.Context, agreementID string) error { return nil }
func (s *InteropService) GetLogs(ctx context.Context, agreementID string) ([]domain.ExchangeLog, error) {
	id, _ := uuid.Parse(agreementID)
	return s.repo.GetExchangeLogs(ctx, id)
}
func (s *InteropService) publish(ctx context.Context, t string, a *domain.DataExchangeAgreement) {
	if s.producer == nil {
		return
	}
	s.producer.Publish(ctx, kafka.Event{EventType: t, AgreementID: a.AgreementID.String(), Timestamp: time.Now().UTC(), Data: a})
	err := recover()
	if err != nil {
		log.Printf("publish failed: %v", err)
	}
}
