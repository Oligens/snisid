package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/afis-svc/internal/domain"
	"github.com/snisid/platform/services/afis-svc/internal/repository/milvus"
)

const (
	MinMatchScore     = 0.85
	CandidateListSize = 15
)

type Matcher interface {
	CompareMinutiae(ctx context.Context, captures []domain.FingerprintCapture, subjectID uuid.UUID) (float64, error)
}

type SearchService struct {
	milvusRepo      *milvus.VectorRepo
	fingerprintRepo FingerprintRepo
	subjectRepo     SubjectRepo
	matcher         Matcher
	vectorizer      Vectorizer
}

func NewSearchService(mr *milvus.VectorRepo, fpr FingerprintRepo, sr SubjectRepo, m Matcher, vz Vectorizer) *SearchService {
	return &SearchService{
		milvusRepo:      mr,
		fingerprintRepo: fpr,
		subjectRepo:     sr,
		matcher:         m,
		vectorizer:      vz,
	}
}

func (s *SearchService) SearchTenprint(ctx context.Context, req domain.EnrollmentRequest) ([]domain.SearchResult, error) {
	vectors, err := s.vectorizer.Vectorize(ctx, req.Fingerprints)
	if err != nil {
		return nil, fmt.Errorf("vectorisation échouée: %w", err)
	}

	candidates, err := s.milvusRepo.SearchNearest(ctx, vectors, CandidateListSize)
	if err != nil {
		return nil, fmt.Errorf("recherche Milvus échouée: %w", err)
	}

	var results []domain.SearchResult
	for _, c := range candidates {
		subjectID, _ := uuid.Parse(c.SubjectID)
		score, _ := s.matcher.CompareMinutiae(ctx, req.Fingerprints, subjectID)
		if score >= MinMatchScore {
			results = append(results, domain.SearchResult{
				CandidateID:    uuid.MustParse(c.PrintID),
				SubjectID:      subjectID,
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
	return results, nil
}
