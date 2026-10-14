package domain

type DetentionBasis string

const (
	DetentionBasisPreventive     DetentionBasis = "PREVENTIVE"
	DetentionBasisSentenced      DetentionBasis = "SENTENCED"
	DetentionBasisAdministrative DetentionBasis = "ADMINISTRATIVE"
	DetentionBasisContempt       DetentionBasis = "CONTEMPT"
)

type LegalStatus string

const (
	LegalStatusAwaitingTrial LegalStatus = "AWAITING_TRIAL"
	LegalStatusOnTrial       LegalStatus = "ON_TRIAL"
	LegalStatusSentenced     LegalStatus = "SENTENCED"
	LegalStatusAppealPending LegalStatus = "APPEAL_PENDING"
	LegalStatusCondemned     LegalStatus = "CONDEMNED"
)

type ReleaseType string

const (
	ReleaseTypeSentenceServed     ReleaseType = "SENTENCE_SERVED"
	ReleaseTypeConditionalRelease ReleaseType = "CONDITIONAL_RELEASE"
	ReleaseTypeBail               ReleaseType = "BAIL"
	ReleaseTypeJudicialOrder      ReleaseType = "JUDICIAL_ORDER"
	ReleaseTypeDeath              ReleaseType = "DEATH"
	ReleaseTypeEscape             ReleaseType = "ESCAPE"
	ReleaseTypeTransferOut        ReleaseType = "TRANSFER_OUT"
)
