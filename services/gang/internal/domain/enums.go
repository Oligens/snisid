package domain

type GangStructureType string

const (
	GangStructureHierarchy GangStructureType = "HIERARCHY"
	GangStructureNetwork   GangStructureType = "NETWORK"
	GangStructureCell      GangStructureType = "CELL"
	GangStructureCoalition GangStructureType = "COALITION"
	GangStructureFranchise GangStructureType = "FRANCHISE"
)

type GangActivityLevel string

const (
	GangActivityDormant  GangActivityLevel = "DORMANT"
	GangActivityLow      GangActivityLevel = "LOW"
	GangActivityModerate GangActivityLevel = "MODERATE"
	GangActivityHigh     GangActivityLevel = "HIGH"
	GangActivityExtreme  GangActivityLevel = "EXTREME"
)

type GangPrimaryActivity string

const (
	ActivityKidnapping      GangPrimaryActivity = "KIDNAPPING"
	ActivityDrugTrafficking GangPrimaryActivity = "DRUG_TRAFFICKING"
	ActivityArmsTrafficking GangPrimaryActivity = "ARMS_TRAFFICKING"
	ActivityExtortion       GangPrimaryActivity = "EXTORTION"
	ActivityTerritoryCtrl   GangPrimaryActivity = "TERRITORY_CONTROL"
	ActivityContractKilling GangPrimaryActivity = "CONTRACT_KILLING"
	ActivityHumanTraffic    GangPrimaryActivity = "HUMAN_TRAFFICKING"
	ActivityMoneyLaundering GangPrimaryActivity = "MONEY_LAUNDERING"
	ActivityMixed           GangPrimaryActivity = "MIXED"
)
