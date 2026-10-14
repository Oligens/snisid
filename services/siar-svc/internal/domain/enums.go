package domain

type WeaponType string

const (
	HANDGUN        WeaponType = "HANDGUN"
	RIFLE          WeaponType = "RIFLE"
	SHOTGUN        WeaponType = "SHOTGUN"
	SUBMACHINE_GUN WeaponType = "SUBMACHINE_GUN"
	ASSAULT_RIFLE  WeaponType = "ASSAULT_RIFLE"
	MACHINE_GUN    WeaponType = "MACHINE_GUN"
	SNIPER         WeaponType = "SNIPER"
	RPG            WeaponType = "RPG"
	GRENADE        WeaponType = "GRENADE"
	HOMEMADE       WeaponType = "HOMEMADE"
	OTHER          WeaponType = "OTHER"
)

type Status string

const (
	REGISTERED      Status = "REGISTERED"
	REPORTED_STOLEN Status = "REPORTED_STOLEN"
	SEIZED          Status = "SEIZED"
	DESTROYED       Status = "DESTROYED"
	REPORTED_LOST   Status = "REPORTED_LOST"
	TRANSFERRED     Status = "TRANSFERRED"
	DEACTIVATED     Status = "DEACTIVATED"
)

type RegistrationType string

const (
	CIVILIAN         RegistrationType = "CIVILIAN"
	POLICE           RegistrationType = "POLICE"
	MILITARY         RegistrationType = "MILITARY"
	SECURITY_COMPANY RegistrationType = "SECURITY_COMPANY"
	EMBASSY          RegistrationType = "EMBASSY"
	ILLEGAL_FOUND    RegistrationType = "ILLEGAL_FOUND"
	HISTORICAL       RegistrationType = "HISTORICAL"
)
