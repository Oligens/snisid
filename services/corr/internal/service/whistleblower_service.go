package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/corr/internal/domain"
	"github.com/snisid/platform/services/corr/internal/repository"
)

type WhistleblowerService struct {
	wbRepo   repository.WhistleblowerRepository
	caseRepo repository.CaseRepository
}

func NewWhistleblowerService(wbRepo repository.WhistleblowerRepository, caseRepo repository.CaseRepository) *WhistleblowerService {
	return &WhistleblowerService{wbRepo: wbRepo, caseRepo: caseRepo}
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *WhistleblowerService) SubmitReport(ctx context.Context, req domain.CreateWhistleblowerRequest, ipHash string) (*domain.WhistleblowerReport, error) {
	now := time.Now()
	r := &domain.WhistleblowerReport{
		ReportID:            uuid.New(),
		ReportToken:         generateToken(),
		AllegationType:      req.AllegationType,
		SeverityEstimate:    req.SeverityEstimate,
		OfficerUnitHint:     req.OfficerUnitHint,
		OfficerRankHint:     req.OfficerRankHint,
		Description:         req.Description,
		EvidenceDescription: req.EvidenceDescription,
		SubmissionDate:      now,
		IPHash:              ipHash,
		Processed:           false,
		CreatedAt:           now,
	}
	if err := s.wbRepo.Create(ctx, r); err != nil {
		return nil, fmt.Errorf("erreur soumission signalement: %w", err)
	}
	return r, nil
}

func (s *WhistleblowerService) GetReportByToken(ctx context.Context, token string) (*domain.WhistleblowerReport, error) {
	return s.wbRepo.GetByToken(ctx, token)
}
