package service

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/sigdc-svc/internal/domain"
)

type DisasterService struct {
	repo domain.DisasterRepository
	log  *zap.Logger
}

func NewDisasterService(repo domain.DisasterRepository, log *zap.Logger) *DisasterService {
	return &DisasterService{repo: repo, log: log}
}

func (s *DisasterService) DeclareDisaster(req *domain.DeclareDisasterRequest) (*domain.Disaster, error) {
	onset, err := time.Parse(time.RFC3339, req.OnsetDate)
	if err != nil {
		return nil, err
	}

	d := &domain.Disaster{
		DisasterType:     domain.DisasterType(req.DisasterType),
		AlertLevel:       domain.AlertLevel(req.AlertLevel),
		OnsetDate:        onset,
		AffectedDepts:    req.AffectedDepts,
		ResponseAgencies: []string{},
	}

	if req.DisasterName != "" {
		d.DisasterName = &req.DisasterName
	}
	if req.EpicenterLat != nil {
		d.EpicenterLat = req.EpicenterLat
	}
	if req.EpicenterLng != nil {
		d.EpicenterLng = req.EpicenterLng
	}
	if req.Magnitude != nil {
		d.Magnitude = req.Magnitude
	}

	return s.repo.CreateDisaster(d)
}

func (s *DisasterService) ListActiveDisasters() ([]domain.Disaster, error) {
	return s.repo.FindActiveDisasters()
}

func (s *DisasterService) IssueWarning(req *domain.IssueWarningRequest) error {
	w := &domain.EarlyWarning{
		DisasterType:  domain.DisasterType(req.DisasterType),
		AlertLevel:    domain.AlertLevel(req.AlertLevel),
		MessageText:   req.MessageText,
		AffectedDepts: req.AffectedDepts,
		IssuedAt:      time.Now(),
		ChannelsSent:  []string{},
	}

	if req.SourceAgency != "" {
		w.SourceAgency = &req.SourceAgency
	}

	return s.repo.SaveWarning(w)
}

func (s *DisasterService) RegisterVictim(req *domain.RegisterVictimRequest) (*domain.VictimRegistration, error) {
	disasterID, err := uuid.Parse(req.DisasterID)
	if err != nil {
		return nil, err
	}

	vr := &domain.VictimRegistration{
		DisasterID:       disasterID,
		Status:           req.Status,
		RegistrationDate: time.Now(),
		RegisteredBy:     uuid.New(),
	}

	if req.FullName != "" {
		vr.FullName = &req.FullName
	}
	if req.LocationFound != "" {
		vr.LocationFound = &req.LocationFound
	}
	if req.DeptCode != "" {
		vr.DeptCode = &req.DeptCode
	}

	return s.repo.CreateVictimRegistration(vr)
}

func (s *DisasterService) ListResources(disasterID uuid.UUID) ([]domain.Resource, error) {
	return s.repo.FindResources(disasterID)
}
