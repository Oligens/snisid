package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/blan-svc/internal/domain"
)

type BLANService struct {
	repo domain.Repository
	log  *zap.Logger
}

func NewBLANService(repo domain.Repository, log *zap.Logger) *BLANService {
	return &BLANService{repo: repo, log: log}
}

func (s *BLANService) OpenCase(req *domain.CreateCaseRequest) (*domain.BLANCase, error) {
	prefix := "BLAN-HT-AAAA-"
	count, err := s.repo.CountCasesByPrefix("BLAN-HT-AAAA-")
	if err != nil {
		return nil, fmt.Errorf("count existing cases: %w", err)
	}
	nationalID := fmt.Sprintf("%s%06d", prefix, count+1)

	now := time.Now()
	caseData := &domain.BLANCase{
		NationalBlanID: nationalID,
		CaseTitle:      req.CaseTitle,
		Typology:       req.Typology,
		TotalAmountUSD: req.TotalAmountUSD,
		PredicateCrime: req.PredicateCrime,
		SubjectIDs:     req.SubjectIDs,
		GangID:         req.GangID,
		StrIDs:         req.StrIDs,
		AnalystID:      req.AnalystID,
		Notes:          req.Notes,
		OpenedAt:       now,
	}

	if caseData.SubjectIDs == nil {
		caseData.SubjectIDs = []uuid.UUID{}
	}
	if caseData.StrIDs == nil {
		caseData.StrIDs = []uuid.UUID{}
	}

	result, err := s.repo.CreateCase(caseData)
	if err != nil {
		s.log.Error("open case failed", zap.Error(err))
		return nil, fmt.Errorf("open case: %w", err)
	}

	return result, nil
}

func (s *BLANService) GetCaseDetail(id uuid.UUID) (*domain.BLANCase, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("get case detail: %w", err)
	}
	return c, nil
}

func (s *BLANService) AddSuspiciousAsset(caseID uuid.UUID, req *domain.AddAssetRequest) (*domain.SuspiciousAsset, error) {
	_, err := s.repo.FindByID(caseID)
	if err != nil {
		return nil, fmt.Errorf("case not found: %w", err)
	}

	asset := &domain.SuspiciousAsset{
		CaseID:            caseID,
		AssetType:         req.AssetType,
		Description:       req.Description,
		Address:           req.Address,
		DeptCode:          req.DeptCode,
		EstimatedValueUSD: req.EstimatedValueUSD,
		AcquisitionDate:   req.AcquisitionDate,
		OwnerSnisidID:     req.OwnerSnisidID,
		OwnerName:         req.OwnerName,
		RegisteredIn:      req.RegisteredIn,
	}

	result, err := s.repo.AddAsset(asset)
	if err != nil {
		s.log.Error("add suspicious asset failed", zap.Error(err))
		return nil, fmt.Errorf("add asset: %w", err)
	}

	return result, nil
}

func (s *BLANService) DocumentTransactionChain(caseID uuid.UUID, req *domain.AddChainStepRequest) (*domain.TransactionChain, error) {
	_, err := s.repo.FindByID(caseID)
	if err != nil {
		return nil, fmt.Errorf("case not found: %w", err)
	}

	step := &domain.TransactionChain{
		CaseID:          caseID,
		StepNumber:      req.StepNumber,
		TransactionType: req.TransactionType,
		FromAccount:     req.FromAccount,
		FromInstitution: req.FromInstitution,
		ToAccount:       req.ToAccount,
		ToInstitution:   req.ToInstitution,
		Amount:          req.Amount,
		Currency:        req.Currency,
		AmountUSD:       req.AmountUSD,
		TransactionDate: req.TransactionDate,
		Notes:           req.Notes,
	}

	if req.IsSuspiciousStep != nil {
		step.IsSuspiciousStep = *req.IsSuspiciousStep
	}

	result, err := s.repo.AddChainStep(step)
	if err != nil {
		s.log.Error("document transaction chain failed", zap.Error(err))
		return nil, fmt.Errorf("add chain step: %w", err)
	}

	return result, nil
}

func (s *BLANService) GetFlaggedRealEstate() ([]domain.RealEstateFlagged, error) {
	return s.repo.GetFlaggedRealEstate()
}

func (s *BLANService) GetFrozenAssets() ([]domain.SuspiciousAsset, error) {
	return s.repo.GetFrozenAssets()
}

func (s *BLANService) GetStatsByTypology() ([]domain.TypologyStats, error) {
	return s.repo.GetStatsByTypology()
}

func generateNationalID(seq int) string {
	return fmt.Sprintf("BLAN-HT-AAAA-%06d", seq)
}

func validateTypology(t domain.Typology) bool {
	valid := map[domain.Typology]bool{
		domain.SMURFING:                true,
		domain.TRADE_BASED_ML:          true,
		domain.REAL_ESTATE:             true,
		domain.SHELL_COMPANY:           true,
		domain.CASH_INTENSIVE_BUSINESS: true,
		domain.CRYPTO_MIXING:           true,
		domain.DIASPORA_TRANSFER:       true,
		domain.RANSOM_LAUNDERING:       true,
		domain.CORRUPTION_PROCEEDS:     true,
	}
	return valid[t]
}

func normalizeCurrency(c string) string {
	return strings.ToUpper(c)
}
