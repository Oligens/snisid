package domain

type DeportationCountry string

const (
	CountryUSA   DeportationCountry = "USA"
	CountryCAN   DeportationCountry = "CAN"
	CountryDOM   DeportationCountry = "DOM"
	CountryBHS   DeportationCountry = "BHS"
	CountryCUB   DeportationCountry = "CUB"
	CountryJAM   DeportationCountry = "JAM"
	CountryTTO   DeportationCountry = "TTO"
	CountryMEX   DeportationCountry = "MEX"
	CountryBRA   DeportationCountry = "BRA"
	CountryFRA   DeportationCountry = "FRA"
	CountryOther DeportationCountry = "OTHER"
)

type CriminalRisk string

const (
	RiskNone     CriminalRisk = "NONE"
	RiskLow      CriminalRisk = "LOW"
	RiskMedium   CriminalRisk = "MEDIUM"
	RiskHigh     CriminalRisk = "HIGH"
	RiskVeryHigh CriminalRisk = "VERY_HIGH"
)

type MonitoringStatus string

const (
	MonitoringActive    MonitoringStatus = "ACTIVE"
	MonitoringSuspended MonitoringStatus = "SUSPENDED"
	MonitoringCompleted MonitoringStatus = "COMPLETED"
	MonitoringFled      MonitoringStatus = "FLED"
	MonitoringDeceased  MonitoringStatus = "DECEASED"
)
