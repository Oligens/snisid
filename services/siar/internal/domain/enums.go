package domain

type WeaponType string

const (
	WeaponHandgun      WeaponType = "HANDGUN"
	WeaponRifle        WeaponType = "RIFLE"
	WeaponShotgun      WeaponType = "SHOTGUN"
	WeaponSubmachine   WeaponType = "SUBMACHINE_GUN"
	WeaponAssaultRifle WeaponType = "ASSAULT_RIFLE"
	WeaponMachineGun   WeaponType = "MACHINE_GUN"
	WeaponSniper       WeaponType = "SNIPER"
	WeaponRPG          WeaponType = "RPG"
	WeaponGrenade      WeaponType = "GRENADE"
	WeaponHomemade     WeaponType = "HOMEMADE"
	WeaponOther        WeaponType = "OTHER"
)

type FirearmStatus string

const (
	StatusRegistered     FirearmStatus = "REGISTERED"
	StatusReportedStolen FirearmStatus = "REPORTED_STOLEN"
	StatusSeized         FirearmStatus = "SEIZED"
	StatusDestroyed      FirearmStatus = "DESTROYED"
	StatusReportedLost   FirearmStatus = "REPORTED_LOST"
	StatusTransferred    FirearmStatus = "TRANSFERRED"
	StatusDeactivated    FirearmStatus = "DEACTIVATED"
)

type RegistrationType string

const (
	RegCivilian     RegistrationType = "CIVILIAN"
	RegPolice       RegistrationType = "POLICE"
	RegMilitary     RegistrationType = "MILITARY"
	RegSecurityCo   RegistrationType = "SECURITY_COMPANY"
	RegEmbassy      RegistrationType = "EMBASSY"
	RegIllegalFound RegistrationType = "ILLEGAL_FOUND"
	RegHistorical   RegistrationType = "HISTORICAL"
)

type LicenseType string

const (
	LicenseCarry     LicenseType = "CARRY"
	LicensePossess   LicenseType = "POSSESS"
	LicenseDealer    LicenseType = "DEALER"
	LicenseCollector LicenseType = "COLLECTOR"
	LicenseImport    LicenseType = "IMPORT"
	LicenseExport    LicenseType = "EXPORT"
)

type TransferType string

const (
	TransferSale         TransferType = "SALE"
	TransferGift         TransferType = "GIFT"
	TransferInheritance  TransferType = "INHERITANCE"
	TransferConfiscation TransferType = "CONFISCATION"
	TransferReturn       TransferType = "RETURN"
	TransferTheftReport  TransferType = "THEFT_REPORTED"
)

type DealerStatus string

const (
	DealerActive    DealerStatus = "ACTIVE"
	DealerSuspended DealerStatus = "SUSPENDED"
	DealerRevoked   DealerStatus = "REVOKED"
	DealerExpired   DealerStatus = "EXPIRED"
)

type DisposalMethod string

const (
	DisposalDestroyed     DisposalMethod = "DESTROYED"
	DisposalKeptEvidence  DisposalMethod = "KEPT_AS_EVIDENCE"
	DisposalReturnedOwner DisposalMethod = "RETURNED_TO_OWNER"
	DisposalSold          DisposalMethod = "SOLD"
	DisposalTransferred   DisposalMethod = "TRANSFERRED"
)
