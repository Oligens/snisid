package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/afis/internal/domain"
)

type LatentService struct {
	mu      sync.RWMutex
	latents map[uuid.UUID]*domain.LatentPrint
	search  *SearchService
	quality *QualityService
	seq     int
}

func NewLatentService(search *SearchService, quality *QualityService) *LatentService {
	return &LatentService{
		latents: make(map[uuid.UUID]*domain.LatentPrint),
		search:  search,
		quality: quality,
	}
}

func (s *LatentService) Submit(ctx context.Context, latent domain.LatentPrint) (*domain.LatentPrint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	latent.LatentID = uuid.New()
	latent.CreatedAt = time.Now()
	latent.IsIdentified = false

	s.latents[latent.LatentID] = &latent
	return &latent, nil
}

func (s *LatentService) ConfirmMatch(ctx context.Context, latentID uuid.UUID, subjectID uuid.UUID, score float64, examinerID uuid.UUID) (*domain.LatentPrint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	latent, ok := s.latents[latentID]
	if !ok {
		return nil, fmt.Errorf("empreinte latente non trouvée: %s", latentID)
	}

	latent.IsIdentified = true
	latent.MatchedSubjectID = &subjectID
	scoreCopy := score
	latent.MatchScore = &scoreCopy
	latent.ExaminedBy = &examinerID

	return latent, nil
}

func (s *LatentService) GetByID(ctx context.Context, id uuid.UUID) (*domain.LatentPrint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	latent, ok := s.latents[id]
	if !ok {
		return nil, fmt.Errorf("empreinte latente non trouvée: %s", id)
	}
	return latent, nil
}

func (s *LatentService) List(ctx context.Context) ([]domain.LatentPrint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]domain.LatentPrint, 0, len(s.latents))
	for _, l := range s.latents {
		result = append(result, *l)
	}
	return result, nil
}

func (s *LatentService) Search(ctx context.Context, latentID uuid.UUID) ([]domain.SearchResult, error) {
	latent, err := s.GetByID(ctx, latentID)
	if err != nil {
		return nil, err
	}

	return s.search.SearchLatent(ctx, *latent)
}
