package domain

import (
	"time"

	"github.com/google/uuid"
)

type FlaggedWallet struct {
	ID                  uuid.UUID     `json:"wallet_id" db:"wallet_id"`
	NationalCryptID     string        `json:"national_crypt_id" db:"national_crypt_id"`
	WalletAddress       string        `json:"wallet_address" db:"wallet_address"`
	AssetType           AssetType     `json:"asset_type" db:"asset_type"`
	BlockchainNetwork   *string       `json:"blockchain_network,omitempty" db:"blockchain_network"`
	SuspicionType       SuspicionType `json:"suspicion_type" db:"suspicion_type"`
	SNISIDPersonID      *uuid.UUID    `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	GangID              *uuid.UUID    `json:"gang_id,omitempty" db:"gang_id"`
	EstimatedBalanceUSD *float64      `json:"estimated_balance_usd,omitempty" db:"estimated_balance_usd"`
	TotalReceivedUSD    *float64      `json:"total_received_usd,omitempty" db:"total_received_usd"`
	TotalSentUSD        *float64      `json:"total_sent_usd,omitempty" db:"total_sent_usd"`
	FirstTxDate         *time.Time    `json:"first_tx_date,omitempty" db:"first_tx_date"`
	LastTxDate          *time.Time    `json:"last_tx_date,omitempty" db:"last_tx_date"`
	IsSanctioned        bool          `json:"is_sanctioned" db:"is_sanctioned"`
	OfacSDNRef          *string       `json:"ofac_sdn_ref,omitempty" db:"ofac_sdn_ref"`
	ChainalysisRef      *string       `json:"chainalysis_ref,omitempty" db:"chainalysis_ref"`
	EllipticRef         *string       `json:"elliptic_ref,omitempty" db:"elliptic_ref"`
	SourceIntel         *string       `json:"source_intel,omitempty" db:"source_intel"`
	LinkedCases         []uuid.UUID   `json:"linked_cases" db:"linked_cases"`
	IsFrozen            bool          `json:"is_frozen" db:"is_frozen"`
	FreezeJurisdiction  *string       `json:"freeze_jurisdiction,omitempty" db:"freeze_jurisdiction"`
	CreatedBy           uuid.UUID     `json:"created_by" db:"created_by"`
	CreatedAt           time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at" db:"updated_at"`
}

type CryptoTransaction struct {
	ID              uuid.UUID  `json:"tx_id" db:"tx_id"`
	WalletID        uuid.UUID  `json:"wallet_id" db:"wallet_id"`
	TxHash          string     `json:"tx_hash" db:"tx_hash"`
	AssetType       AssetType  `json:"asset_type" db:"asset_type"`
	Direction       string     `json:"direction" db:"direction"`
	FromAddress     *string    `json:"from_address,omitempty" db:"from_address"`
	ToAddress       *string    `json:"to_address,omitempty" db:"to_address"`
	AmountCrypto    *float64   `json:"amount_crypto,omitempty" db:"amount_crypto"`
	AmountUSDAtTx   *float64   `json:"amount_usd_at_tx,omitempty" db:"amount_usd_at_tx"`
	TxTimestamp     time.Time  `json:"tx_timestamp" db:"tx_timestamp"`
	BlockNumber     *int64     `json:"block_number,omitempty" db:"block_number"`
	IsMixerInvolved bool       `json:"is_mixer_involved" db:"is_mixer_involved"`
	MixerService    *string    `json:"mixer_service,omitempty" db:"mixer_service"`
	RiskScore       *int       `json:"risk_score,omitempty" db:"risk_score"`
	SuspicionFlags  []string   `json:"suspicion_flags" db:"suspicion_flags"`
	ExtorsCaseID    *uuid.UUID `json:"extors_case_id,omitempty" db:"extors_case_id"`
	UCRefStrID      *uuid.UUID `json:"ucref_str_id,omitempty" db:"ucref_str_id"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

type ExchangeAccount struct {
	ID               uuid.UUID  `json:"exchange_id" db:"exchange_id"`
	SNISIDPersonID   *uuid.UUID `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	ExchangeName     string     `json:"exchange_name" db:"exchange_name"`
	ExchangeCountry  *string    `json:"exchange_country,omitempty" db:"exchange_country"`
	AccountRef       *string    `json:"account_ref,omitempty" db:"account_ref"`
	KycLevel         *string    `json:"kyc_level,omitempty" db:"kyc_level"`
	TotalVolumeUSD   *float64   `json:"total_volume_usd,omitempty" db:"total_volume_usd"`
	IsFlagged        bool       `json:"is_flagged" db:"is_flagged"`
	FlaggingReason   *string    `json:"flagging_reason,omitempty" db:"flagging_reason"`
	LegalHoldRequest bool       `json:"legal_hold_request" db:"legal_hold_request"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

type WalletRiskReport struct {
	WalletAddress   string        `json:"wallet_address"`
	NationalCryptID *string       `json:"national_crypt_id,omitempty"`
	IsSanctioned    bool          `json:"is_sanctioned"`
	IsFrozen        bool          `json:"is_frozen"`
	SuspicionType   SuspicionType `json:"suspicion_type"`
	SuspicionCount  int           `json:"suspicion_count"`
	RiskScore       int           `json:"risk_score"`
	IsKnownCriminal bool          `json:"is_known_criminal"`
	MixerExposure   bool          `json:"mixer_exposure"`
	LinkedCases     []uuid.UUID   `json:"linked_cases"`
	Transactions    int           `json:"transaction_count"`
	AnalyzedAt      time.Time     `json:"analyzed_at"`
}

type CreateWalletRequest struct {
	WalletAddress       string        `json:"wallet_address" binding:"required"`
	AssetType           AssetType     `json:"asset_type" binding:"required"`
	BlockchainNetwork   *string       `json:"blockchain_network,omitempty"`
	SuspicionType       SuspicionType `json:"suspicion_type" binding:"required"`
	SNISIDPersonID      *uuid.UUID    `json:"snisid_person_id,omitempty"`
	GangID              *uuid.UUID    `json:"gang_id,omitempty"`
	EstimatedBalanceUSD *float64      `json:"estimated_balance_usd,omitempty"`
	TotalReceivedUSD    *float64      `json:"total_received_usd,omitempty"`
	TotalSentUSD        *float64      `json:"total_sent_usd,omitempty"`
	FirstTxDate         *time.Time    `json:"first_tx_date,omitempty"`
	LastTxDate          *time.Time    `json:"last_tx_date,omitempty"`
	IsSanctioned        bool          `json:"is_sanctioned"`
	OfacSDNRef          *string       `json:"ofac_sdn_ref,omitempty"`
	ChainalysisRef      *string       `json:"chainalysis_ref,omitempty"`
	EllipticRef         *string       `json:"elliptic_ref,omitempty"`
	SourceIntel         *string       `json:"source_intel,omitempty"`
	IsFrozen            bool          `json:"is_frozen"`
	FreezeJurisdiction  *string       `json:"freeze_jurisdiction,omitempty"`
	CreatedBy           uuid.UUID     `json:"created_by" binding:"required"`
}

type AddTransactionRequest struct {
	TxHash          string     `json:"tx_hash" binding:"required"`
	AssetType       AssetType  `json:"asset_type" binding:"required"`
	Direction       string     `json:"direction" binding:"required"`
	FromAddress     *string    `json:"from_address,omitempty"`
	ToAddress       *string    `json:"to_address,omitempty"`
	AmountCrypto    *float64   `json:"amount_crypto,omitempty"`
	AmountUSDAtTx   *float64   `json:"amount_usd_at_tx,omitempty"`
	TxTimestamp     *time.Time `json:"tx_timestamp" binding:"required"`
	BlockNumber     *int64     `json:"block_number,omitempty"`
	IsMixerInvolved bool       `json:"is_mixer_involved"`
	MixerService    *string    `json:"mixer_service,omitempty"`
	RiskScore       *int       `json:"risk_score,omitempty"`
	SuspicionFlags  []string   `json:"suspicion_flags,omitempty"`
	ExtorsCaseID    *uuid.UUID `json:"extors_case_id,omitempty"`
	UCRefStrID      *uuid.UUID `json:"ucref_str_id,omitempty"`
}

type AssetStats struct {
	AssetType       AssetType `json:"asset_type"`
	TotalWallets    int       `json:"total_wallets"`
	SanctionedCount int       `json:"sanctioned_wallets"`
	FrozenCount     int       `json:"frozen_wallets"`
	TotalBalanceUSD float64   `json:"total_balance_usd"`
	AvgRiskScore    float64   `json:"avg_risk_score"`
}

type Repository interface {
	CreateWallet(wallet *FlaggedWallet) (*FlaggedWallet, error)
	FindByAddress(address string) (*FlaggedWallet, error)
	FindByID(id uuid.UUID) (*FlaggedWallet, error)
	GetSanctionedWallets() ([]FlaggedWallet, error)
	GetWalletsByGang(gangID uuid.UUID) ([]FlaggedWallet, error)
	AddTransaction(tx *CryptoTransaction) (*CryptoTransaction, error)
	GetTransactionsByWallet(walletID uuid.UUID) ([]CryptoTransaction, error)
	GetStatsByAsset() ([]AssetStats, error)
}
