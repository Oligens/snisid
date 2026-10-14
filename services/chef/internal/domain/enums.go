package domain

type ChefRoleType string

const (
	RoleSupremeLeader ChefRoleType = "SUPREME_LEADER"
	RoleZoneCommander ChefRoleType = "ZONE_COMMANDER"
	RoleLieutenant    ChefRoleType = "LIEUTENANT"
	RoleSoldier       ChefRoleType = "SOLDIER"
	RoleAssociate     ChefRoleType = "ASSOCIATE"
	RoleFinancier     ChefRoleType = "FINANCIER"
	RoleEnabler       ChefRoleType = "ENABLER"
	RoleInformant     ChefRoleType = "INFORMANT"
)

type ChefStatus string

const (
	StatusActive      ChefStatus = "ACTIVE"
	StatusArrested    ChefStatus = "ARRESTED"
	StatusDetained    ChefStatus = "DETAINED"
	StatusDeceased    ChefStatus = "DECEASED"
	StatusFledCountry ChefStatus = "FLED_COUNTRY"
	StatusUnknown     ChefStatus = "UNKNOWN"
)
