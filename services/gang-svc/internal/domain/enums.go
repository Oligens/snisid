package domain

type StructureType string

const (
	StructureHierarchy StructureType = "HIERARCHY"
	StructureNetwork   StructureType = "NETWORK"
	StructureCell      StructureType = "CELL"
	StructureCoalition StructureType = "COALITION"
	StructureFranchise StructureType = "FRANCHISE"
)

type ActivityLevel string

const (
	ActivityDormant  ActivityLevel = "DORMANT"
	ActivityLow      ActivityLevel = "LOW"
	ActivityModerate ActivityLevel = "MODERATE"
	ActivityHigh     ActivityLevel = "HIGH"
	ActivityExtreme  ActivityLevel = "EXTREME"
)

type PrimaryActivity string

const (
	ActivityKidnapping       PrimaryActivity = "KIDNAPPING"
	ActivityDrugTrafficking  PrimaryActivity = "DRUG_TRAFFICKING"
	ActivityArmsTrafficking  PrimaryActivity = "ARMS_TRAFFICKING"
	ActivityExtortion        PrimaryActivity = "EXTORTION"
	ActivityTerritoryControl PrimaryActivity = "TERRITORY_CONTROL"
	ActivityContractKilling  PrimaryActivity = "CONTRACT_KILLING"
	ActivityHumanTrafficking PrimaryActivity = "HUMAN_TRAFFICKING"
	ActivityMoneyLaundering  PrimaryActivity = "MONEY_LAUNDERING"
	ActivityMixed            PrimaryActivity = "MIXED"
)
