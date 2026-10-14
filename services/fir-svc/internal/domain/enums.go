package domain

type OffenseClass string

const (
	OffenseClassContravention OffenseClass = "CONTRAVENTION"
	OffenseClassDelit         OffenseClass = "DELIT"
	OffenseClassCrime         OffenseClass = "CRIME"
	OffenseClassFelonyForeign OffenseClass = "FELONY_FOREIGN"
)

type CaseStatus string

const (
	CaseStatusOpen          CaseStatus = "OPEN"
	CaseStatusPendingTrial  CaseStatus = "PENDING_TRIAL"
	CaseStatusConvicted     CaseStatus = "CONVICTED"
	CaseStatusAcquitted     CaseStatus = "ACQUITTED"
	CaseStatusDismissed     CaseStatus = "DISMISSED"
	CaseStatusAppealPending CaseStatus = "APPEAL_PENDING"
	CaseStatusExpunged      CaseStatus = "EXPUNGED"
)

type SentenceType string

const (
	SentenceTypePrison           SentenceType = "PRISON"
	SentenceTypeSuspended        SentenceType = "SUSPENDED"
	SentenceTypeFine             SentenceType = "FINE"
	SentenceTypeCommunityService SentenceType = "COMMUNITY_SERVICE"
	SentenceTypeDeathPenalty     SentenceType = "DEATH_PENALTY"
	SentenceTypeAcquittal        SentenceType = "ACQUITTAL"
	SentenceTypeProbation        SentenceType = "PROBATION"
)

type CertificateResult string

const (
	CertificateResultClean     CertificateResult = "CLEAN"
	CertificateResultHasRecord CertificateResult = "HAS_RECORD"
)
