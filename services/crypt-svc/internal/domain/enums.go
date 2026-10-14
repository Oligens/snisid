package domain

type AssetType string

const (
	BITCOIN     AssetType = "BITCOIN"
	ETHEREUM    AssetType = "ETHEREUM"
	USDT        AssetType = "USDT"
	USDC        AssetType = "USDC"
	MONERO      AssetType = "MONERO"
	ZCASH       AssetType = "ZCASH"
	LITECOIN    AssetType = "LITECOIN"
	OTHER_ERC20 AssetType = "OTHER_ERC20"
	UNKNOWN     AssetType = "UNKNOWN"
)

type SuspicionType string

const (
	RANSOM_RECEIPT           SuspicionType = "RANSOM_RECEIPT"
	SANCTIONS_EVASION        SuspicionType = "SANCTIONS_EVASION"
	DARKWEB_PAYMENT          SuspicionType = "DARKWEB_PAYMENT"
	MIXER_SERVICE            SuspicionType = "MIXER_SERVICE"
	PEER_TO_PEER_UNREGULATED SuspicionType = "PEER_TO_PEER_UNREGULATED"
	EXCHANGE_HIGH_RISK       SuspicionType = "EXCHANGE_HIGH_RISK"
	GANG_PAYMENT             SuspicionType = "GANG_PAYMENT"
	UNKNOWN_SUSPICION        SuspicionType = "UNKNOWN"
)
