package service

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/blkl-svc/internal/domain"
)

type BLKLService struct {
	repo domain.Repository
	log  *zap.Logger
}

func NewBLKLService(repo domain.Repository, log *zap.Logger) *BLKLService {
	return &BLKLService{repo: repo, log: log}
}

func (s *BLKLService) CheckPerson(personID uuid.UUID) (*domain.BlacklistCheckResult, error) {
	return s.repo.CheckPerson(personID)
}

func (s *BLKLService) AddEntry(req *domain.AddEntryRequest) (*domain.BlklBlacklist, error) {
	now := time.Now()
	entry := &domain.BlklBlacklist{
		SNISIDPersonID:  req.SNISIDPersonID,
		RestrictionType: req.RestrictionType,
		Source:          req.Source,
		SourceRecordID:  req.SourceRecordID,
		Reason:          req.Reason,
		CourtOrderRef:   req.CourtOrderRef,
		OrderedBy:       req.OrderedBy,
		EffectiveDate:   now,
		ExpiryDate:      req.ExpiryDate,
		IsPermanent:     req.IsPermanent,
		AlertLevel:      req.AlertLevel,
		ArmedDangerous:  req.ArmedDangerous,
		CreatedBy:       req.CreatedBy,
	}
	if req.EffectiveDate != nil {
		entry.EffectiveDate = *req.EffectiveDate
	}

	return s.repo.AddEntry(entry)
}

func (s *BLKLService) LiftEntry(id uuid.UUID, liftedBy string) error {
	return s.repo.LiftEntry(id, liftedBy)
}

func (s *BLKLService) GetActiveEntries() ([]domain.BlklBlacklist, error) {
	return s.repo.GetActiveEntries()
}

func (s *BLKLService) GetExpiringSoon(days int) ([]domain.BlklBlacklist, error) {
	return s.repo.GetExpiringSoon(days)
}
