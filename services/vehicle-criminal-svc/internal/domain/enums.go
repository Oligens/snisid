package domain

type CrimeCategory string

const (
	CrimeCategoryVehicleTheft     CrimeCategory = "VOL_VEHICULE"
	CrimeCategoryPlatTheft        CrimeCategory = "VOL_PLAQUE"
	CrimeCategoryDrugTraffic      CrimeCategory = "TRAFIC_STUPEFIANTS"
	CrimeCategoryKidnapping       CrimeCategory = "ENLEVEMENT"
	CrimeCategoryGangAffiliated   CrimeCategory = "GANG"
	CrimeCategoryArmsTraffic      CrimeCategory = "TRAFIC_ARMES"
	CrimeCategoryFakeStateVehicle CrimeCategory = "VEHICULE_ETATIQUE_CLONE"
	CrimeCategoryFakePolice       CrimeCategory = "FAUX_POLICIER"
	CrimeCategoryHumanTrafficking CrimeCategory = "TRAITE_PERSONNES"
	CrimeCategoryContraband       CrimeCategory = "CONTREBANDE"
	CrimeCategoryOtherFelony      CrimeCategory = "AUTRE_CRIME_GRAVE"
)

type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "INFO"
	AlertLevelCaution  AlertLevel = "CAUTION"
	AlertLevelWanted   AlertLevel = "WANTED"
	AlertLevelCritical AlertLevel = "CRITICAL"
)

type AlertStatus string

const (
	AlertStatusActive    AlertStatus = "ACTIVE"
	AlertStatusSuspended AlertStatus = "SUSPENDED"
	AlertStatusResolved  AlertStatus = "RESOLVED"
	AlertStatusExpired   AlertStatus = "EXPIRED"
	AlertStatusCancelled AlertStatus = "CANCELLED"
)

type PlateCategory string

const (
	PlateCategoryPrivate         PlateCategory = "PP"
	PlateCategoryHeavy           PlateCategory = "PL"
	PlateCategoryMoto            PlateCategory = "M"
	PlateCategoryPublicTransport PlateCategory = "TC"
	PlateCategoryState           PlateCategory = "SE"
	PlateCategoryDiplomatic      PlateCategory = "CD"
	PlateCategoryOrg             PlateCategory = "OA"
	PlateCategoryAgri            PlateCategory = "AG"
	PlateCategoryMedical         PlateCategory = "MD"
	PlateCategoryTaxi            PlateCategory = "TX"
)

type StolenPlateStatus string

const (
	StolenPlateStatusStolen      StolenPlateStatus = "STOLEN"
	StolenPlateStatusRecovered   StolenPlateStatus = "RECOVERED"
	StolenPlateStatusDestroyed   StolenPlateStatus = "DESTROYED"
	StolenPlateStatusUsedInCrime StolenPlateStatus = "USED_IN_CRIME"
)

type VehicleType string

const (
	VehicleTypeBerline     VehicleType = "BERLINE"
	VehicleTypeSUV         VehicleType = "SUV"
	VehicleTypePickup      VehicleType = "PICKUP"
	VehicleTypeCamion      VehicleType = "CAMION"
	VehicleTypeMoto        VehicleType = "MOTO"
	VehicleTypeTapTap      VehicleType = "TAP_TAP"
	VehicleTypeCamionnette VehicleType = "CAMIONNETTE"
	VehicleTypeBus         VehicleType = "BUS"
	VehicleTypeQuad        VehicleType = "QUAD"
	VehicleTypeBateau      VehicleType = "BATEAU"
	VehicleTypeAutre       VehicleType = "AUTRE"
)

type SyncDirection string

const (
	SyncDirectionOutbound SyncDirection = "OUTBOUND"
	SyncDirectionInbound  SyncDirection = "INBOUND"
)

type SyncStatus string

const (
	SyncStatusPending  SyncStatus = "PENDING"
	SyncStatusSuccess  SyncStatus = "SUCCESS"
	SyncStatusFailed   SyncStatus = "FAILED"
	SyncStatusRejected SyncStatus = "REJECTED"
)

type RouteType string

const (
	RouteTypeImport   RouteType = "IMPORT"
	RouteTypeExport   RouteType = "EXPORT"
	RouteTypeTransit  RouteType = "TRANSIT"
	RouteTypeDomestic RouteType = "DOMESTIC"
)

type KidnappingStatus string

const (
	KidnappingStatusInProgress     KidnappingStatus = "IN_PROGRESS"
	KidnappingStatusVictimRescued  KidnappingStatus = "VICTIM_RESCUED"
	KidnappingStatusVictimReleased KidnappingStatus = "VICTIM_RELEASED"
	KidnappingStatusVictimDeceased KidnappingStatus = "VICTIM_DECEASED"
	KidnappingStatusUnresolved     KidnappingStatus = "UNRESOLVED"
)

type SightingSource string

const (
	SightingSourceLAPI       SightingSource = "LAPI"
	SightingSourceManual     SightingSource = "MANUAL"
	SightingSourceTip        SightingSource = "TIP"
	SightingSourceCheckpoint SightingSource = "CHECKPOINT"
)

type ReportType string

const (
	ReportTypeIncident      ReportType = "INCIDENT"
	ReportTypePattern       ReportType = "PATTERN"
	ReportTypeAnalytical    ReportType = "ANALYTICAL"
	ReportTypeAlertBulletin ReportType = "ALERT_BULLETIN"
)
