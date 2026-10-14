package domain

type RecoveryContext string

const (
	RecoveryPoliceOperation    RecoveryContext = "POLICE_OPERATION"
	RecoveryCheckpoint         RecoveryContext = "CHECKPOINT"
	RecoveryPortSeizure        RecoveryContext = "PORT_SEIZURE"
	RecoveryAirportSeizure     RecoveryContext = "AIRPORT_SEIZURE"
	RecoveryCommunitySurrender RecoveryContext = "COMMUNITY_SURRENDER"
	RecoveryCrimeScene         RecoveryContext = "CRIME_SCENE"
	RecoveryRaid               RecoveryContext = "RAID"
	RecoveryBorderSeizure      RecoveryContext = "BORDER_SEIZURE"
	RecoveryOther              RecoveryContext = "OTHER"
)

type WeaponDisposition string

const (
	DispositionHeldAsEvidence      WeaponDisposition = "HELD_AS_EVIDENCE"
	DispositionDestroyed           WeaponDisposition = "DESTROYED"
	DispositionReturnedToOwner     WeaponDisposition = "RETURNED_TO_OWNER"
	DispositionTransferredToPolice WeaponDisposition = "TRANSFERRED_TO_POLICE"
	DispositionSentToInterpol      WeaponDisposition = "SENT_TO_INTERPOL"
	DispositionPending             WeaponDisposition = "PENDING"
)

type SyncDirection string

const (
	SyncOutbound SyncDirection = "OUTBOUND"
	SyncInbound  SyncDirection = "INBOUND"
)

type SyncStatus string

const (
	SyncPending SyncStatus = "PENDING"
	SyncSuccess SyncStatus = "SUCCESS"
	SyncFailed  SyncStatus = "FAILED"
)
