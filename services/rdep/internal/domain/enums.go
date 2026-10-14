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

type ExtraditionStatus string

const (
	ExtraditionRequested ExtraditionStatus = "REQUESTED"
	ExtraditionApproved  ExtraditionStatus = "APPROVED"
	ExtraditionInTransit ExtraditionStatus = "IN_TRANSIT"
	ExtraditionExecuted  ExtraditionStatus = "EXECUTED"
	ExtraditionDenied    ExtraditionStatus = "DENIED"
	ExtraditionCancelled ExtraditionStatus = "CANCELLED"
)

type FlightType string

const (
	FlightCharterICE FlightType = "CHARTER_ICE"
	FlightCommercial FlightType = "COMMERCIAL"
	FlightMilitary   FlightType = "MILITARY"
	FlightGovernment FlightType = "GOVERNMENT"
	FlightOther      FlightType = "OTHER"
)

type MonitoringEventType string

const (
	EventCheckIn       MonitoringEventType = "CHECK_IN"
	EventViolation     MonitoringEventType = "VIOLATION"
	EventAddressChange MonitoringEventType = "ADDRESS_CHANGE"
	EventTravelAuth    MonitoringEventType = "TRAVEL_AUTHORIZATION"
	EventAlert         MonitoringEventType = "ALERT"
)
