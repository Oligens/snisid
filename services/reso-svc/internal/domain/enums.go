package domain

type ActivityLevel string

const (
	ACTIVE    ActivityLevel = "ACTIVE"
	INACTIVE  ActivityLevel = "INACTIVE"
	DISBANDED ActivityLevel = "DISBANDED"
	EMERGING  ActivityLevel = "EMERGING"
)

type RelationType string

const (
	ALLIANCE    RelationType = "ALLIANCE"
	RIVALRY     RelationType = "RIVALRY"
	SUBORDINATE RelationType = "SUBORDINATE"
	SUPPLIER    RelationType = "SUPPLIER"
	NEUTRAL     RelationType = "NEUTRAL"
)

type AssociationType string

const (
	FAMILY          AssociationType = "FAMILY"
	BUSINESS        AssociationType = "BUSINESS"
	CRIMINAL        AssociationType = "CRIMINAL"
	COMMUNICATION   AssociationType = "COMMUNICATION"
	KNOWN_ASSOCIATE AssociationType = "KNOWN_ASSOCIATE"
)

type GangRole string

const (
	LEADER     GangRole = "LEADER"
	LIEUTENANT GangRole = "LIEUTENANT"
	MEMBER     GangRole = "MEMBER"
	ASSOCIATE  GangRole = "ASSOCIATE"
	CANDIDATE  GangRole = "CANDIDATE"
)
