package domain

import (
	"time"

	"github.com/google/uuid"
)

type BLANCase struct {
	CaseID         uuid.UUID   `json:"case_id" db:"case_id"`
	NationalBlanID string      `json:"national_blan_id" db:"national_blan_id"`
	CaseTitle      string      `json:"case_title" db:"case_title"`
	Typology       Typology    `json:"typology" db:"typology"`
	Status         string      `json:"status" db:"status"`
	TotalAmountUSD *float64    `json:"total_amount_usd,omitempty" db:"total_amount_usd"`
	PredicateCrime *string     `json:"predicate_crime,omitempty" db:"predicate_crime"`
	SubjectIDs     []uuid.UUID `json:"subject_ids" db:"subject_ids"`
	GangID         *uuid.UUID  `json:"gang_id,omitempty" db:"gang_id"`
	StrIDs         []uuid.UUID `json:"str_ids" db:"str_ids"`
	OpenedAt       time.Time   `json:"opened_at" db:"opened_at"`
	AnalystID      *uuid.UUID  `json:"analyst_id,omitempty" db:"analyst_id"`
	ParquetRef     *string     `json:"parquet_ref,omitempty" db:"parquet_ref"`
	Notes          *string     `json:"notes,omitempty" db:"notes"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at" db:"updated_at"`
}

type SuspiciousAsset struct {
	AssetID           uuid.UUID  `json:"asset_id" db:"asset_id"`
	CaseID            uuid.UUID  `json:"case_id" db:"case_id"`
	AssetType         AssetType  `json:"asset_type" db:"asset_type"`
	Description       string     `json:"description" db:"description"`
	Address           *string    `json:"address,omitempty" db:"address"`
	DeptCode          *string    `json:"dept_code,omitempty" db:"dept_code"`
	EstimatedValueUSD *float64   `json:"estimated_value_usd,omitempty" db:"estimated_value_usd"`
	AcquisitionDate   *time.Time `json:"acquisition_date,omitempty" db:"acquisition_date"`
	OwnerSnisidID     *uuid.UUID `json:"owner_snisid_id,omitempty" db:"owner_snisid_id"`
	OwnerName         *string    `json:"owner_name,omitempty" db:"owner_name"`
	RegisteredIn      *string    `json:"registered_in,omitempty" db:"registered_in"`
	IsFrozen          bool       `json:"is_frozen" db:"is_frozen"`
	FreezeOrderRef    *string    `json:"freeze_order_ref,omitempty" db:"freeze_order_ref"`
	IsSeized          bool       `json:"is_seized" db:"is_seized"`
	SeizureDate       *time.Time `json:"seizure_date,omitempty" db:"seizure_date"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

type TransactionChain struct {
	ChainID          uuid.UUID  `json:"chain_id" db:"chain_id"`
	CaseID           uuid.UUID  `json:"case_id" db:"case_id"`
	StepNumber       int        `json:"step_number" db:"step_number"`
	TransactionType  *string    `json:"transaction_type,omitempty" db:"transaction_type"`
	FromAccount      *string    `json:"from_account,omitempty" db:"from_account"`
	FromInstitution  *string    `json:"from_institution,omitempty" db:"from_institution"`
	ToAccount        *string    `json:"to_account,omitempty" db:"to_account"`
	ToInstitution    *string    `json:"to_institution,omitempty" db:"to_institution"`
	Amount           *float64   `json:"amount,omitempty" db:"amount"`
	Currency         *string    `json:"currency,omitempty" db:"currency"`
	AmountUSD        *float64   `json:"amount_usd,omitempty" db:"amount_usd"`
	TransactionDate  *time.Time `json:"transaction_date,omitempty" db:"transaction_date"`
	IsSuspiciousStep bool       `json:"is_suspicious_step" db:"is_suspicious_step"`
	Notes            *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

type RealEstateFlagged struct {
	PropertyID        uuid.UUID  `json:"property_id" db:"property_id"`
	CaseID            *uuid.UUID `json:"case_id,omitempty" db:"case_id"`
	Address           string     `json:"address" db:"address"`
	DeptCode          *string    `json:"dept_code,omitempty" db:"dept_code"`
	Commune           *string    `json:"commune,omitempty" db:"commune"`
	Lat               *float64   `json:"lat,omitempty" db:"lat"`
	Lng               *float64   `json:"lng,omitempty" db:"lng"`
	PropertyType      *string    `json:"property_type,omitempty" db:"property_type"`
	PurchasePriceUSD  *float64   `json:"purchase_price_usd,omitempty" db:"purchase_price_usd"`
	PurchaseDate      *time.Time `json:"purchase_date,omitempty" db:"purchase_date"`
	DeclaredOwner     *string    `json:"declared_owner,omitempty" db:"declared_owner"`
	BeneficialOwnerID *uuid.UUID `json:"beneficial_owner_id,omitempty" db:"beneficial_owner_id"`
	SuspiciousReasons []string   `json:"suspicious_reasons" db:"suspicious_reasons"`
	IsFrozen          bool       `json:"is_frozen" db:"is_frozen"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

type TypologyStats struct {
	Typology  Typology `json:"typology" db:"typology"`
	CaseCount int      `json:"case_count" db:"case_count"`
	TotalUSD  *float64 `json:"total_usd,omitempty" db:"total_usd"`
}

type CreateCaseRequest struct {
	CaseTitle      string      `json:"case_title" binding:"required"`
	Typology       Typology    `json:"typology" binding:"required"`
	TotalAmountUSD *float64    `json:"total_amount_usd,omitempty"`
	PredicateCrime *string     `json:"predicate_crime,omitempty"`
	SubjectIDs     []uuid.UUID `json:"subject_ids,omitempty"`
	GangID         *uuid.UUID  `json:"gang_id,omitempty"`
	StrIDs         []uuid.UUID `json:"str_ids,omitempty"`
	AnalystID      *uuid.UUID  `json:"analyst_id,omitempty"`
	Notes          *string     `json:"notes,omitempty"`
}

type AddAssetRequest struct {
	AssetType         AssetType  `json:"asset_type" binding:"required"`
	Description       string     `json:"description" binding:"required"`
	Address           *string    `json:"address,omitempty"`
	DeptCode          *string    `json:"dept_code,omitempty"`
	EstimatedValueUSD *float64   `json:"estimated_value_usd,omitempty"`
	AcquisitionDate   *time.Time `json:"acquisition_date,omitempty"`
	OwnerSnisidID     *uuid.UUID `json:"owner_snisid_id,omitempty"`
	OwnerName         *string    `json:"owner_name,omitempty"`
	RegisteredIn      *string    `json:"registered_in,omitempty"`
}

type AddChainStepRequest struct {
	StepNumber       int        `json:"step_number" binding:"required"`
	TransactionType  *string    `json:"transaction_type,omitempty"`
	FromAccount      *string    `json:"from_account,omitempty"`
	FromInstitution  *string    `json:"from_institution,omitempty"`
	ToAccount        *string    `json:"to_account,omitempty"`
	ToInstitution    *string    `json:"to_institution,omitempty"`
	Amount           *float64   `json:"amount,omitempty"`
	Currency         *string    `json:"currency,omitempty"`
	AmountUSD        *float64   `json:"amount_usd,omitempty"`
	TransactionDate  *time.Time `json:"transaction_date,omitempty"`
	IsSuspiciousStep *bool      `json:"is_suspicious_step,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
}

type Repository interface {
	CreateCase(c *BLANCase) (*BLANCase, error)
	FindByID(id uuid.UUID) (*BLANCase, error)
	AddAsset(a *SuspiciousAsset) (*SuspiciousAsset, error)
	AddChainStep(step *TransactionChain) (*TransactionChain, error)
	GetFlaggedRealEstate() ([]RealEstateFlagged, error)
	GetFrozenAssets() ([]SuspiciousAsset, error)
	GetStatsByTypology() ([]TypologyStats, error)
	CountCasesByPrefix(prefix string) (int, error)
}
