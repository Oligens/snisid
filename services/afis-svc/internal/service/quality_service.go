package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/snisid/platform/services/afis-svc/internal/domain"
)

type QualityService struct {
	mu            sync.RWMutex
	nfiq2MinScore int16
	pendingChecks int
}

func NewQualityService(minScore int16) *QualityService {
	return &QualityService{
		nfiq2MinScore: minScore,
	}
}

func (s *QualityService) ValidateScore(score int16) error {
	if score < 0 || score > 100 {
		return fmt.Errorf("NFIQ2 score must be between 0 and 100, got %d", score)
	}
	if score < s.nfiq2MinScore {
		return domain.ErrQualityTooLow
	}
	return nil
}

func (s *QualityService) IsHighQuality(score int16) bool {
	return score >= 80
}

func (s *QualityService) ScheduleQualityCheck(ctx context.Context, fp *domain.Fingerprint) error {
	if err := s.ValidateScore(fp.NFIQ2Score); err != nil {
		return err
	}
	s.mu.Lock()
	s.pendingChecks++
	s.mu.Unlock()
	return nil
}

func (s *QualityService) PendingChecksCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingChecks
}
