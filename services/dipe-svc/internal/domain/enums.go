package domain

type CaseType string

const (
	CaseTypeKidnappingSuspected    CaseType = "KIDNAPPING_SUSPECTED"
	CaseTypeVoluntaryDisappearance CaseType = "VOLUNTARY_DISAPPEARANCE"
	CaseTypeDisasterRelated        CaseType = "DISASTER_RELATED"
	CaseTypeGangViolence           CaseType = "GANG_VIOLENCE"
	CaseTypeMigrationRelated       CaseType = "MIGRATION_RELATED"
	CaseTypeChildAbduction         CaseType = "CHILD_ABDUCTION"
	CaseTypeTraffickingSuspected   CaseType = "TRAFFICKING_SUSPECTED"
	CaseTypeUnknown                CaseType = "UNKNOWN"
)

type CaseStatus string

const (
	CaseStatusOpen             CaseStatus = "OPEN"
	CaseStatusLocatedAlive     CaseStatus = "LOCATED_ALIVE"
	CaseStatusBodyIdentified   CaseStatus = "BODY_IDENTIFIED"
	CaseStatusBodyUnidentified CaseStatus = "BODY_UNIDENTIFIED"
	CaseStatusCancelled        CaseStatus = "CANCELLED"
	CaseStatusColdCase         CaseStatus = "COLD_CASE"
)
