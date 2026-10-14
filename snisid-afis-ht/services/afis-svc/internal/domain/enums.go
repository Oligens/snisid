package domain

type FingerPosition string

const (
	FingerRightThumb  FingerPosition = "RIGHT_THUMB"
	FingerRightIndex  FingerPosition = "RIGHT_INDEX"
	FingerRightMiddle FingerPosition = "RIGHT_MIDDLE"
	FingerRightRing   FingerPosition = "RIGHT_RING"
	FingerRightLittle FingerPosition = "RIGHT_LITTLE"
	FingerLeftThumb   FingerPosition = "LEFT_THUMB"
	FingerLeftIndex   FingerPosition = "LEFT_INDEX"
	FingerLeftMiddle  FingerPosition = "LEFT_MIDDLE"
	FingerLeftRing    FingerPosition = "LEFT_RING"
	FingerLeftLittle  FingerPosition = "LEFT_LITTLE"
	FingerRightPalm   FingerPosition = "RIGHT_PALM"
	FingerLeftPalm    FingerPosition = "LEFT_PALM"
	FingerUnknown     FingerPosition = "UNKNOWN"
)

type CaptureMethod string

const (
	CaptureLiveScanner CaptureMethod = "LIVESCANNER"
	CaptureInkRoll     CaptureMethod = "INKROLL"
	CaptureLatentLift  CaptureMethod = "LATENT_LIFT"
	CapturePhoto       CaptureMethod = "PHOTO"
	CaptureUnknown     CaptureMethod = "UNKNOWN"
)

type SubjectType string

const (
	SubjectSuspect         SubjectType = "SUSPECT"
	SubjectCriminal        SubjectType = "CRIMINAL"
	SubjectVictim          SubjectType = "VICTIM"
	SubjectUnknownDeceased SubjectType = "UNKNOWN_DECEASED"
	SubjectMissingPerson   SubjectType = "MISSING_PERSON"
	SubjectEmployee        SubjectType = "EMPLOYEE"
)

type TransactionType string

const (
	TransactionTenToTen   TransactionType = "TEN2TEN"
	TransactionLatent2Ten TransactionType = "LATENT2TEN"
	TransactionPalm       TransactionType = "PALM"
)
