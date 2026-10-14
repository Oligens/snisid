package domain

type STRStatus string

const (
	STRStatusReceived      STRStatus = "RECEIVED"
	STRStatusUnderAnalysis STRStatus = "UNDER_ANALYSIS"
	STRStatusDisseminated  STRStatus = "DISSEMINATED"
	STRStatusArchived      STRStatus = "ARCHIVED"
	STRStatusNoAction      STRStatus = "NO_ACTION"
)

type ReportType string

const (
	ReportTypeSTR               ReportType = "STR"
	ReportTypeCTR               ReportType = "CTR"
	ReportTypeInternationalWire ReportType = "INTERNATIONAL_WIRE"
	ReportTypeRealEstate        ReportType = "REAL_ESTATE"
	ReportTypeMoncashPattern    ReportType = "MONCASH_PATTERN"
	ReportTypeCryptoPattern     ReportType = "CRYPTO_PATTERN"
)
