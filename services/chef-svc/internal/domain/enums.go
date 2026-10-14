package domain

type RoleType string

const (
	RoleSupremeLeader RoleType = "SUPREME_LEADER"
	RoleZoneCommander RoleType = "ZONE_COMMANDER"
	RoleLieutenant    RoleType = "LIEUTENANT"
	RoleSoldier       RoleType = "SOLDIER"
	RoleAssociate     RoleType = "ASSOCIATE"
	RoleFinancier     RoleType = "FINANCIER"
	RoleEnabler       RoleType = "ENABLER"
	RoleInformant     RoleType = "INFORMANT"
)

type MemberStatus string

const (
	StatusActive      MemberStatus = "ACTIVE"
	StatusArrested    MemberStatus = "ARRESTED"
	StatusDetained    MemberStatus = "DETAINED"
	StatusDeceased    MemberStatus = "DECEASED"
	StatusFledCountry MemberStatus = "FLED_COUNTRY"
	StatusUnknown     MemberStatus = "UNKNOWN"
)
