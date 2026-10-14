package domain

type TraiffickingType string

const (
	LABOR_EXPLOITATION               TraiffickingType = "LABOR_EXPLOITATION"
	SEXUAL_EXPLOITATION              TraiffickingType = "SEXUAL_EXPLOITATION"
	FORCED_MARRIAGE                  TraiffickingType = "FORCED_MARRIAGE"
	CHILD_DOMESTIC_SERVITUDE         TraiffickingType = "CHILD_DOMESTIC_SERVITUDE"
	GANG_RECRUITMENT_FORCED          TraiffickingType = "GANG_RECRUITMENT_FORCED"
	IRREGULAR_MIGRATION_FACILITATION TraiffickingType = "IRREGULAR_MIGRATION_FACILITATION"
	ORGAN_TRAFFICKING                TraiffickingType = "ORGAN_TRAFFICKING"
	OTHER                            TraiffickingType = "OTHER"
)

type VictimStatus string

const (
	IDENTIFIED_VICTIM VictimStatus = "IDENTIFIED_VICTIM"
	POTENTIAL_VICTIM  VictimStatus = "POTENTIAL_VICTIM"
	WITNESS           VictimStatus = "WITNESS"
	RESCUED           VictimStatus = "RESCUED"
	REPATRIATED       VictimStatus = "REPATRIATED"
	DECEASED          VictimStatus = "DECEASED"
	MISSING           VictimStatus = "MISSING"
)
