package domain

type Typology string

const (
	SMURFING                Typology = "SMURFING"
	TRADE_BASED_ML          Typology = "TRADE_BASED_ML"
	REAL_ESTATE             Typology = "REAL_ESTATE"
	SHELL_COMPANY           Typology = "SHELL_COMPANY"
	CASH_INTENSIVE_BUSINESS Typology = "CASH_INTENSIVE_BUSINESS"
	CRYPTO_MIXING           Typology = "CRYPTO_MIXING"
	DIASPORA_TRANSFER       Typology = "DIASPORA_TRANSFER"
	RANSOM_LAUNDERING       Typology = "RANSOM_LAUNDERING"
	CORRUPTION_PROCEEDS     Typology = "CORRUPTION_PROCEEDS"
)

type AssetType string

const (
	ASSET_REAL_ESTATE   AssetType = "REAL_ESTATE"
	ASSET_VEHICLE       AssetType = "VEHICLE"
	ASSET_BUSINESS      AssetType = "BUSINESS"
	ASSET_BANK_ACCOUNT  AssetType = "BANK_ACCOUNT"
	ASSET_CRYPTO_WALLET AssetType = "CRYPTO_WALLET"
	ASSET_CASH          AssetType = "CASH"
	ASSET_JEWELRY       AssetType = "JEWELRY"
	ASSET_LIVESTOCK     AssetType = "LIVESTOCK"
	ASSET_OTHER         AssetType = "OTHER"
)
