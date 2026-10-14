package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/infra-ht/internal/domain"
	"github.com/snisid/infra-ht/internal/kafka"
	"github.com/snisid/infra-ht/internal/repository"
)

type InfraService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewInfraService(repo repository.Repository, producer *kafka.Producer) *InfraService {
	return &InfraService{repo: repo, producer: producer}
}
func (s *InfraService) GetHealth(ctx context.Context) map[string]any {
	dcs, _ := s.repo.GetDatacenters(ctx)
	cls, _ := s.repo.GetClusters(ctx)
	return map[string]any{"datacenters": len(dcs), "clusters": len(cls), "status": "healthy"}
}
func (s *InfraService) GetClusters(ctx context.Context) ([]domain.K8sCluster, error) {
	return s.repo.GetClusters(ctx)
}
func (s *InfraService) RecordDRDrill(ctx context.Context, drill domain.DRDrill) (*domain.DRDrill, error) {
	drill.DrillID = uuid.New()
	drill.CreatedAt = time.Now().UTC()
	if drill.RTOTargetMin == 0 {
		drill.RTOTargetMin = 15
	}
	if drill.RPOTargetMin == 0 {
		drill.RPOTargetMin = 1
	}
	if err := s.repo.CreateDRDrill(ctx, &drill); err != nil {
		return nil, err
	}
	if s.producer != nil {
		s.producer.Publish(ctx, kafka.Event{EventType: "infra.dr.drill", Timestamp: time.Now().UTC(), Data: drill})
	}
	err := recover()
	if err != nil {
		log.Printf("publish: %v", err)
	}
	return &drill, nil
}
