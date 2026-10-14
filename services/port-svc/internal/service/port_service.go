package service

import (
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/port-svc/internal/domain"
)

type PortService struct {
	repo domain.ContainerRepository
	log  *zap.Logger
}

func NewPortService(repo domain.ContainerRepository, log *zap.Logger) *PortService {
	return &PortService{repo: repo, log: log}
}

func (s *PortService) RecordArrival(req *domain.RecordArrivalRequest) (*domain.VesselArrival, error) {
	arrival := &domain.VesselArrival{
		PortCode:        req.PortCode,
		VesselIMO:       req.VesselIMO,
		VesselName:      req.VesselName,
		FlagCountry:     req.FlagCountry,
		ShippingCompany: req.ShippingCompany,
		ArrivalDate:     req.ArrivalDate,
		OriginPort:      req.OriginPort,
		OriginCountry:   req.OriginCountry,
		ContainerCount:  req.ContainerCount,
		ManifestRef:     req.ManifestRef,
		CBPTargetingRef: req.CBPTargetingRef,
	}

	return s.repo.CreateArrival(arrival)
}

func (s *PortService) GetArrival(id uuid.UUID) (*domain.VesselArrival, error) {
	return s.repo.FindArrivalByID(id)
}

func (s *PortService) GetHighRiskContainers() ([]domain.Container, error) {
	return s.repo.GetHighRiskContainers()
}

func (s *PortService) ScanContainer(id uuid.UUID, scanResult string) (*domain.Container, error) {
	return s.repo.ScanContainer(id, scanResult)
}

func (s *PortService) SeizeContainer(id uuid.UUID, description string, caseRef string) (*domain.Container, error) {
	return s.repo.SeizeContainer(id, description, caseRef)
}

func (s *PortService) GetSeizureStats() (*domain.SeizureStats, error) {
	return s.repo.GetSeizureStats()
}
