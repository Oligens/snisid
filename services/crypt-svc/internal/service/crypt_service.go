package service

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/crypt-svc/internal/domain"
)

type CryptService struct {
	repo domain.Repository
	log  *zap.Logger
}

func NewCryptService(repo domain.Repository, log *zap.Logger) *CryptService {
	return &CryptService{repo: repo, log: log}
}

func (s *CryptService) AnalyzeWalletRisk(address string) (*domain.WalletRiskReport, error) {
	wallet, err := s.repo.FindByAddress(address)
	if err != nil {
		return &domain.WalletRiskReport{
			WalletAddress: address,
			RiskScore:     0,
			AnalyzedAt:    time.Now(),
		}, nil
	}

	txs, err := s.repo.GetTransactionsByWallet(wallet.ID)
	if err != nil {
		s.log.Error("failed to fetch transactions", zap.Error(err))
	}

	riskScore := 0
	if wallet.IsSanctioned {
		riskScore += 40
	}
	if wallet.IsFrozen {
		riskScore += 20
	}
	switch wallet.SuspicionType {
	case domain.RANSOM_RECEIPT:
		riskScore += 25
	case domain.SANCTIONS_EVASION:
		riskScore += 30
	case domain.DARKWEB_PAYMENT:
		riskScore += 20
	case domain.MIXER_SERVICE:
		riskScore += 15
	case domain.GANG_PAYMENT:
		riskScore += 35
	case domain.EXCHANGE_HIGH_RISK:
		riskScore += 10
	case domain.PEER_TO_PEER_UNREGULATED:
		riskScore += 5
	}

	mixerExposure := false
	for _, tx := range txs {
		if tx.IsMixerInvolved {
			mixerExposure = true
			riskScore += 10
			break
		}
	}

	if riskScore > 100 {
		riskScore = 100
	}

	isKnownCriminal := wallet.SNISIDPersonID != nil || wallet.GangID != nil

	return &domain.WalletRiskReport{
		WalletAddress:   wallet.WalletAddress,
		NationalCryptID: &wallet.NationalCryptID,
		IsSanctioned:    wallet.IsSanctioned,
		IsFrozen:        wallet.IsFrozen,
		SuspicionType:   wallet.SuspicionType,
		SuspicionCount:  1,
		RiskScore:       riskScore,
		IsKnownCriminal: isKnownCriminal,
		MixerExposure:   mixerExposure,
		LinkedCases:     wallet.LinkedCases,
		Transactions:    len(txs),
		AnalyzedAt:      time.Now(),
	}, nil
}

func (s *CryptService) FlagWallet(req *domain.CreateWalletRequest) (*domain.FlaggedWallet, error) {
	wallet := &domain.FlaggedWallet{
		WalletAddress:       req.WalletAddress,
		AssetType:           req.AssetType,
		BlockchainNetwork:   req.BlockchainNetwork,
		SuspicionType:       req.SuspicionType,
		SNISIDPersonID:      req.SNISIDPersonID,
		GangID:              req.GangID,
		EstimatedBalanceUSD: req.EstimatedBalanceUSD,
		TotalReceivedUSD:    req.TotalReceivedUSD,
		TotalSentUSD:        req.TotalSentUSD,
		FirstTxDate:         req.FirstTxDate,
		LastTxDate:          req.LastTxDate,
		IsSanctioned:        req.IsSanctioned,
		OfacSDNRef:          req.OfacSDNRef,
		ChainalysisRef:      req.ChainalysisRef,
		EllipticRef:         req.EllipticRef,
		SourceIntel:         req.SourceIntel,
		LinkedCases:         []uuid.UUID{},
		IsFrozen:            req.IsFrozen,
		FreezeJurisdiction:  req.FreezeJurisdiction,
		CreatedBy:           req.CreatedBy,
	}

	return s.repo.CreateWallet(wallet)
}

func (s *CryptService) AddTransaction(walletID uuid.UUID, req *domain.AddTransactionRequest) (*domain.CryptoTransaction, error) {
	tx := &domain.CryptoTransaction{
		WalletID:        walletID,
		TxHash:          req.TxHash,
		AssetType:       req.AssetType,
		Direction:       req.Direction,
		FromAddress:     req.FromAddress,
		ToAddress:       req.ToAddress,
		AmountCrypto:    req.AmountCrypto,
		AmountUSDAtTx:   req.AmountUSDAtTx,
		TxTimestamp:     *req.TxTimestamp,
		BlockNumber:     req.BlockNumber,
		IsMixerInvolved: req.IsMixerInvolved,
		MixerService:    req.MixerService,
		RiskScore:       req.RiskScore,
		SuspicionFlags:  req.SuspicionFlags,
		ExtorsCaseID:    req.ExtorsCaseID,
		UCRefStrID:      req.UCRefStrID,
	}

	if tx.SuspicionFlags == nil {
		tx.SuspicionFlags = []string{}
	}

	return s.repo.AddTransaction(tx)
}

func (s *CryptService) GetSanctionedWallets() ([]domain.FlaggedWallet, error) {
	return s.repo.GetSanctionedWallets()
}

func (s *CryptService) GetWalletsByGang(gangID uuid.UUID) ([]domain.FlaggedWallet, error) {
	return s.repo.GetWalletsByGang(gangID)
}

func (s *CryptService) GetStatsByAsset() ([]domain.AssetStats, error) {
	return s.repo.GetStatsByAsset()
}
