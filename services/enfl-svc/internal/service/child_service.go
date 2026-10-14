package service

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/enfl-svc/internal/domain"
)

type ChildService struct {
	repo domain.ChildRepository
	log  *zap.Logger
}

func NewChildService(repo domain.ChildRepository, log *zap.Logger) *ChildService {
	return &ChildService{repo: repo, log: log}
}

func (s *ChildService) RegisterChild(req *domain.RegisterChildRequest) (*domain.Child, error) {
	dob, err := time.Parse("2006-01-02", req.DOB)
	if err != nil {
		return nil, err
	}

	child := &domain.Child{
		RiskCategory:        domain.RiskCategory(req.RiskCategory),
		Status:              domain.Missing,
		FullName:            req.FullName,
		DOB:                 dob,
		Gender:              &req.Gender,
		Nationality:         &req.Nationality,
		DistinguishingMarks: &req.DistinguishingMarks,
		HeightCm:            req.HeightCm,
		SkinTone:            &req.SkinTone,
		GuardianName:        &req.GuardianName,
		GuardianPhone:       &req.GuardianPhone,
		DeptCode:            &req.DeptCode,
		Commune:             &req.Commune,
		PhotoRefs:           []string{},
		AssistanceType:      []string{},
		CreatedBy:           uuid.New(),
	}

	if req.DisappearanceDate != "" {
		dd, _ := time.Parse(time.RFC3339, req.DisappearanceDate)
		child.DisappearanceDate = &dd
	}
	if req.GangID != "" {
		gid, _ := uuid.Parse(req.GangID)
		child.GangID = &gid
	}

	return s.repo.Create(child)
}

func (s *ChildService) GetChild(id uuid.UUID) (*domain.Child, error) {
	return s.repo.FindByID(id)
}

func (s *ChildService) ListMissing() ([]domain.Child, error) {
	return s.repo.FindMissing()
}

func (s *ChildService) ListRestaveks() ([]domain.Restavek, error) {
	return s.repo.FindRestaveks()
}

func (s *ChildService) LocateChild(id uuid.UUID, req *domain.LocateChildRequest) error {
	status := domain.LocatedSafe
	if req.Status != "" {
		status = domain.ChildStatus(req.Status)
	}
	return s.repo.UpdateStatus(id, status, req.Location)
}

func (s *ChildService) ListGangRecruited() ([]domain.Child, error) {
	return s.repo.FindGangRecruited()
}
