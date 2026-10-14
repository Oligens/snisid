package service

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/afis/internal/domain"
)

const (
	MinMatchScore     = 0.85
	CandidateListSize = 15
)

type MilvusRepo interface {
	SearchNearest(ctx context.Context, vectors [][]float32, topK int) ([]domain.SearchResult, error)
}

type Matcher interface {
	CompareMinutiae(ctx context.Context, captures []domain.FingerprintCapture, subjectID uuid.UUID) (float64, error)
}

type SearchService struct {
	mu           sync.RWMutex
	subjects     map[uuid.UUID]*domain.Subject
	fingerprints map[uuid.UUID][]domain.Fingerprint
	matcher      Matcher
}

func NewSearchService() *SearchService {
	return &SearchService{
		subjects:     make(map[uuid.UUID]*domain.Subject),
		fingerprints: make(map[uuid.UUID][]domain.Fingerprint),
		matcher:      &mockMatcher{},
	}
}

type mockMatcher struct{}

func (m *mockMatcher) CompareMinutiae(ctx context.Context, captures []domain.FingerprintCapture, subjectID uuid.UUID) (float64, error) {
	score := 0.85 + rand.Float64()*0.14
	return math.Round(score*100) / 100, nil
}

type dummyVectorizer struct{}

func (v *dummyVectorizer) Vectorize(ctx context.Context, captures []domain.FingerprintCapture) ([][]float32, error) {
	vectors := make([][]float32, len(captures))
	for i := range captures {
		vec := make([]float32, 512)
		for j := range vec {
			vec[j] = rand.Float32()
		}
		vectors[i] = vec
	}
	return vectors, nil
}

func (s *SearchService) SearchTenprint(ctx context.Context, req domain.EnrollmentRequest) ([]domain.SearchResult, error) {
	vec := &dummyVectorizer{}
	vectors, err := vec.Vectorize(ctx, req.Fingerprints)
	if err != nil {
		return nil, fmt.Errorf("vectorisation échouée: %w", err)
	}

	_ = vectors

	s.mu.RLock()
	candidates := s.mockMilvusSearch(ctx, req.Fingerprints)
	s.mu.RUnlock()

	var results []domain.SearchResult
	for _, c := range candidates {
		score, err := s.matcher.CompareMinutiae(ctx, req.Fingerprints, c.SubjectID)
		if err != nil {
			continue
		}
		if score >= MinMatchScore {
			results = append(results, domain.SearchResult{
				CandidateID:    c.CandidateID,
				SubjectID:      c.SubjectID,
				Score:          score,
				NationalAFISID: c.NationalAFISID,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	for i := range results {
		results[i].Rank = i + 1
	}

	if len(results) > CandidateListSize {
		results = results[:CandidateListSize]
	}

	return results, nil
}

func (s *SearchService) SearchLatent(ctx context.Context, latent domain.LatentPrint) ([]domain.SearchResult, error) {
	vec := &dummyVectorizer{}
	captures := []domain.FingerprintCapture{{
		Position:    latent.FingerPosition,
		Method:      domain.CaptureUnknown,
		ImageBase64: latent.ImageRef,
		NFIQ2Score:  0,
	}}
	vectors, err := vec.Vectorize(ctx, captures)
	if err != nil {
		return nil, fmt.Errorf("vectorisation latente échouée: %w", err)
	}
	_ = vectors

	s.mu.RLock()
	candidates := s.mockMilvusSearch(ctx, captures)
	s.mu.RUnlock()

	var results []domain.SearchResult
	for _, c := range candidates {
		score, err := s.matcher.CompareMinutiae(ctx, captures, c.SubjectID)
		if err != nil {
			continue
		}
		if score >= MinMatchScore {
			results = append(results, domain.SearchResult{
				CandidateID:    c.CandidateID,
				SubjectID:      c.SubjectID,
				Score:          score,
				NationalAFISID: c.NationalAFISID,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	for i := range results {
		results[i].Rank = i + 1
	}

	if len(results) > CandidateListSize {
		results = results[:CandidateListSize]
	}

	return results, nil
}

func (s *SearchService) mockMilvusSearch(ctx context.Context, captures []domain.FingerprintCapture) []domain.SearchResult {
	s.mu.RLock()
	ids := make([]uuid.UUID, 0, len(s.subjects))
	for id := range s.subjects {
		ids = append(ids, id)
	}
	s.mu.RUnlock()

	rand.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })

	n := CandidateListSize
	if len(ids) < n {
		n = len(ids)
	}

	if n == 0 {
		return nil
	}

	results := make([]domain.SearchResult, n)
	now := time.Now().UnixNano()
	for i := 0; i < n; i++ {
		r := rand.New(rand.NewSource(now + int64(i)))
		score := 0.75 + r.Float64()*0.24
		s.mu.RLock()
		subj := s.subjects[ids[i]]
		var nationalID string
		if subj != nil && subj.NationalAFISID != nil {
			nationalID = *subj.NationalAFISID
		}
		s.mu.RUnlock()
		results[i] = domain.SearchResult{
			CandidateID:    uuid.New(),
			SubjectID:      ids[i],
			Score:          math.Round(score*100) / 100,
			Rank:           0,
			NationalAFISID: nationalID,
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	for i := range results {
		results[i].Rank = i + 1
	}
	return results
}

func (s *SearchService) IndexSubject(subject *domain.Subject, fps []domain.Fingerprint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subjects[subject.SubjectID] = subject
	s.fingerprints[subject.SubjectID] = fps
}

func (s *SearchService) GetSubject(ctx context.Context, id uuid.UUID) (*domain.Subject, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	subj, ok := s.subjects[id]
	if !ok {
		return nil, fmt.Errorf("sujet non trouvé: %s", id)
	}
	return subj, nil
}
