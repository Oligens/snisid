package domain

type RestrictionType string

const (
	ENTRY_BAN       RestrictionType = "ENTRY_BAN"
	EXIT_BAN        RestrictionType = "EXIT_BAN"
	BOTH_BAN        RestrictionType = "BOTH_BAN"
	CONDITIONAL_BAN RestrictionType = "CONDITIONAL_BAN"
)

type Source string

const (
	JUDICIAL_ORDER         Source = "JUDICIAL_ORDER"
	WANTED_WARRANT         Source = "WANTED_WARRANT"
	UN_SANCTIONS           Source = "UN_SANCTIONS"
	OFAC_SANCTIONS         Source = "OFAC_SANCTIONS"
	MINISTERIAL_ORDER      Source = "MINISTERIAL_ORDER"
	EXPULSION              Source = "EXPULSION"
	OPR_TRAVEL_RESTRICTION Source = "OPR_TRAVEL_RESTRICTION"
	INTERPOL_NOTICE        Source = "INTERPOL_NOTICE"
)
