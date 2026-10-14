package domain

type OffenseClass string

const (
	OffenseContravention OffenseClass = "CONTRAVENTION"
	OffenseDelit         OffenseClass = "DELIT"
	OffenseCrime         OffenseClass = "CRIME"
	OffenseFelonyForeign OffenseClass = "FELONY_FOREIGN"
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
	SentencePrison           SentenceType = "PRISON"
	SentenceSuspended        SentenceType = "SUSPENDED"
	SentenceFine             SentenceType = "FINE"
	SentenceCommunityService SentenceType = "COMMUNITY_SERVICE"
	SentenceDeathPenalty     SentenceType = "DEATH_PENALTY"
	SentenceAcquittal        SentenceType = "ACQUITTAL"
	SentenceProbation        SentenceType = "PROBATION"
)

type MovementType string

const (
	MovementRecordCreated     MovementType = "RECORD_CREATED"
	MovementArrestAdded       MovementType = "ARREST_ADDED"
	MovementConvictionAdded   MovementType = "CONVICTION_ADDED"
	MovementAliasAdded        MovementType = "ALIAS_ADDED"
	MovementAliasRemoved      MovementType = "ALIAS_REMOVED"
	MovementRecordExpunged    MovementType = "RECORD_EXPUNGED"
	MovementRecordReactivated MovementType = "RECORD_REACTIVATED"
	MovementCertificateIssued MovementType = "CERTIFICATE_ISSUED"
)

type CertificateResult string

const (
	CertificateClean     CertificateResult = "CLEAN"
	CertificateHasRecord CertificateResult = "HAS_RECORD"
)
