package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/snisid/mil-c2-ht/internal/domain"
	"github.com/snisid/mil-c2-ht/internal/kafka"
	"github.com/snisid/mil-c2-ht/internal/repository"
)

type MilC2ServiceInterface interface {
	CreateUnit(domain.MilitaryUnit) error
	GetDeployedUnits() ([]domain.MilitaryUnit, error)
	CreateOperation(domain.Operation) error
	GetActiveOperations() ([]domain.Operation, error)
	SubmitReport(uuid.UUID, domain.TacticalReport) error
	GetOperationTimeline(uuid.UUID) ([]domain.TacticalReport, error)
	GetCommonOperatingPicture() (*domain.CommonOperatingPicture, error)
}

type MilC2Service struct {
	repo     repository.MilC2Repo
	producer *kafka.Producer
}

func NewMilC2Service(repo repository.MilC2Repo, producer *kafka.Producer) *MilC2Service {
	return &MilC2Service{repo: repo, producer: producer}
}

func (s *MilC2Service) CreateUnit(u domain.MilitaryUnit) error {
	u.UnitID = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	if u.OperationalStatus == "" {
		u.OperationalStatus = domain.OpStatusStandby
	}
	if err := s.repo.CreateUnit(u); err != nil {
		return err
	}
	s.producer.Publish("unit.created", u.UnitID.String())
	return nil
}

func (s *MilC2Service) GetDeployedUnits() ([]domain.MilitaryUnit, error) {
	return s.repo.GetDeployedUnits()
}

func (s *MilC2Service) CreateOperation(o domain.Operation) error {
	o.OperationID = uuid.New()
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()
	o.Status = domain.OpStatusPlanning
	if err := s.repo.CreateOperation(o); err != nil {
		return err
	}
	s.producer.Publish("operation.created", o.OperationID.String())
	return nil
}

func (s *MilC2Service) GetActiveOperations() ([]domain.Operation, error) {
	return s.repo.GetActiveOperations()
}

func (s *MilC2Service) SubmitReport(operationID uuid.UUID, t domain.TacticalReport) error {
	t.ReportID = uuid.New()
	t.OperationID = operationID
	t.SubmittedAt = time.Now()
	if err := s.repo.CreateTacticalReport(t); err != nil {
		return err
	}
	s.producer.Publish("tactical.report.submitted", t.ReportID.String())
	return nil
}

func (s *MilC2Service) GetOperationTimeline(operationID uuid.UUID) ([]domain.TacticalReport, error) {
	return s.repo.GetReportsByOperation(operationID)
}

func (s *MilC2Service) GetCommonOperatingPicture() (*domain.CommonOperatingPicture, error) {
	units, err := s.repo.GetAllUnits()
	if err != nil {
		return nil, err
	}
	ops, err := s.repo.GetAllOperations()
	if err != nil {
		return nil, err
	}
	reports, err := s.repo.GetAllReports()
	if err != nil {
		return nil, err
	}
	return &domain.CommonOperatingPicture{
		Units:      units,
		Operations: ops,
		Reports:    reports,
	}, nil
}
