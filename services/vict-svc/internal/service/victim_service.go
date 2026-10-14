package service

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/vict-svc/internal/domain"
)

type VictimService struct {
	repo domain.VictimRepository
	log  *zap.Logger
}

func NewVictimService(repo domain.VictimRepository, log *zap.Logger) *VictimService {
	return &VictimService{repo: repo, log: log}
}

func (s *VictimService) RegisterVictim(req *domain.RegisterVictimRequest) (*domain.Victim, error) {
	id, err := time.Parse(time.RFC3339, req.IncidentDate)
	if err != nil {
		return nil, err
	}

	victim := &domain.Victim{
		CrimeType:      domain.CrimeType(req.CrimeType),
		VictimStatus:   domain.VictimStatus(req.VictimStatus),
		IncidentDate:   id,
		PerpetratorIDs: []uuid.UUID{},
		CreatedBy:      uuid.New(),
	}

	if req.FullName != "" {
		victim.FullName = &req.FullName
	}
	if req.Gender != "" {
		victim.Gender = &req.Gender
	}
	if req.IncidentLocation != "" {
		victim.IncidentLocation = &req.IncidentLocation
	}
	if req.DeptCode != "" {
		victim.DeptCode = &req.DeptCode
	}
	if req.Commune != "" {
		victim.Commune = &req.Commune
	}
	if req.GangID != "" {
		gid, _ := uuid.Parse(req.GangID)
		victim.GangID = &gid
	}
	if req.DOB != "" {
		dob, _ := time.Parse("2006-01-02", req.DOB)
		victim.DOB = &dob
	}

	return s.repo.Create(victim)
}

func (s *VictimService) GetVictim(id uuid.UUID) (*domain.Victim, error) {
	return s.repo.FindByID(id)
}

func (s *VictimService) CreateMassIncident(req *domain.CreateMassIncidentRequest) (*domain.MassIncident, error) {
	id, err := time.Parse(time.RFC3339, req.IncidentDate)
	if err != nil {
		return nil, err
	}

	mi := &domain.MassIncident{
		IncidentName:    req.IncidentName,
		CrimeType:       domain.CrimeType(req.CrimeType),
		IncidentDate:    id,
		VictimCount:     req.VictimCount,
		DocumentedBy:    []string{},
		LinkedVictimIDs: []uuid.UUID{},
		CreatedBy:       uuid.New(),
	}

	if req.DeptCode != "" {
		mi.DeptCode = &req.DeptCode
	}
	if req.Commune != "" {
		mi.Commune = &req.Commune
	}
	if req.Description != "" {
		mi.Description = &req.Description
	}

	return s.repo.CreateMassIncident(mi)
}

func (s *VictimService) ListMassIncidents() ([]domain.MassIncident, error) {
	return s.repo.FindMassIncidents()
}

func (s *VictimService) ListByGang(gangID uuid.UUID) ([]domain.Victim, error) {
	return s.repo.FindByGang(gangID)
}

func (s *VictimService) GetStatsByType() ([]domain.CrimeStats, error) {
	return s.repo.GetStatsByType()
}

func (s *VictimService) GetReparationList() ([]domain.Victim, error) {
	return s.repo.GetReparationList()
}
