package domain

type AllegationType string

const (
	AllegationDataLeak           AllegationType = "DATA_LEAK_TO_GANG"
	AllegationRecordTampering    AllegationType = "RECORD_TAMPERING"
	AllegationUnauthorizedAccess AllegationType = "UNAUTHORIZED_ACCESS"
	AllegationBribery            AllegationType = "BRIBERY"
	AllegationExtortion          AllegationType = "EXTORTION_OF_CIVILIANS"
	AllegationPrisonEscape       AllegationType = "FACILITATED_PRISON_ESCAPE"
	AllegationStolenCredentials  AllegationType = "STOLEN_CREDENTIALS"
	AllegationFinancial          AllegationType = "FINANCIAL_CORRUPTION"
	AllegationGangAffiliation    AllegationType = "GANG_AFFILIATION"
	AllegationOther              AllegationType = "OTHER"
)

type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

type CaseStatus string

const (
	StatusReported           CaseStatus = "REPORTED"
	StatusUnderInvestigation CaseStatus = "UNDER_INVESTIGATION"
	StatusSubstantiated      CaseStatus = "SUBSTANTIATED"
	StatusUnsubstantiated    CaseStatus = "UNSUBSTANTIATED"
	StatusReferredToParquet  CaseStatus = "REFERRED_TO_PARQUET"
	StatusClosed             CaseStatus = "CLOSED"
)
