package kafka

import (
	"time"
)

type IndexType string

const (
	IndexTypeBIO_CON IndexType = "BIO_CON"
	IndexTypeBIO_ARR IndexType = "BIO_ARR"
	IndexTypeBIO_FSC IndexType = "BIO_FSC"
	IndexTypeBIO_DIS IndexType = "BIO_DIS"
	IndexTypeBIO_RNI IndexType = "BIO_RNI"
)

type MatchType string

const (
	MatchTypeFull     MatchType = "FULL_MATCH"
	MatchTypePartial  MatchType = "PARTIAL"
	MatchTypeFamilial MatchType = "FAMILIAL"
)

type AlertLevel string

const (
	AlertLevelLow      AlertLevel = "LOW"
	AlertLevelMedium   AlertLevel = "MEDIUM"
	AlertLevelHigh     AlertLevel = "HIGH"
	AlertLevelCritical AlertLevel = "CRITICAL"
)

type EventEnvelope struct {
	EventID       string `json:"event_id"`
	EventType     string `json:"event_type"`
	CorrelationID string `json:"correlation_id,omitempty"`
	Timestamp     int64  `json:"timestamp"`
}

func newEnvelope(eventType string) EventEnvelope {
	return EventEnvelope{
		EventID:   newUUID(),
		EventType: eventType,
		Timestamp: time.Now().UnixMilli(),
	}
}

func newEnvelopeWithCorrelation(eventType, correlationID string) EventEnvelope {
	e := newEnvelope(eventType)
	e.CorrelationID = correlationID
	return e
}

type DNAProfileCreated struct {
	EventEnvelope
	SampleID       string    `json:"sample_id"`
	SpecimenNumber string    `json:"specimen_number"`
	IndexType      IndexType `json:"index_type"`
	LabID          string    `json:"lab_id"`
	LabLevel       string    `json:"lab_level"`
	QualityScore   float32   `json:"quality_score"`
	LociCount      int32     `json:"loci_count"`
	Amelogenin     *string   `json:"amelogenin,omitempty"`
	CaseNumber     *string   `json:"case_number,omitempty"`
	CollectedDate  string    `json:"collected_date"`
}

type DNAHitDetected struct {
	EventEnvelope
	HitID          string     `json:"hit_id"`
	QuerySampleID  string     `json:"query_sample_id"`
	MatchSampleID  string     `json:"match_sample_id"`
	MatchType      MatchType  `json:"match_type"`
	Confidence     float32    `json:"confidence"`
	MatchedLoci    int32      `json:"matched_loci"`
	TotalLoci      int32      `json:"total_loci"`
	HitLevel       string     `json:"hit_level"`
	AlertLevel     AlertLevel `json:"alert_level"`
	QueryIndexType string     `json:"query_index_type"`
	MatchIndexType string     `json:"match_index_type"`
	CaseNumber     *string    `json:"case_number,omitempty"`
}

type WantedPersonCreated struct {
	EventEnvelope
	RecordID       string   `json:"record_id"`
	RecordNumber   string   `json:"record_number"`
	NIU            *string  `json:"niu,omitempty"`
	WarrantType    string   `json:"warrant_type"`
	DangerLevel    string   `json:"danger_level"`
	ArmedDangerous bool     `json:"armed_dangerous"`
	Charges        []string `json:"charges"`
	EnteringAgency string   `json:"entering_agency"`
	InterpolNotice *string  `json:"interpol_notice,omitempty"`
	ExpiryDate     *string  `json:"expiry_date,omitempty"`
}

type LAPIPlateQuery struct {
	EventEnvelope
	PlateNumber string  `json:"plate_number"`
	CameraID    string  `json:"camera_id"`
	Location    string  `json:"location"`
	ImageRef    *string `json:"image_ref,omitempty"`
}

type LAPIPlateResponse struct {
	EventEnvelope
	PlateNumber  string  `json:"plate_number"`
	HitFound     bool    `json:"hit_found"`
	HitType      *string `json:"hit_type,omitempty"`
	RecordNumber *string `json:"record_number,omitempty"`
	AlertLevel   *string `json:"alert_level,omitempty"`
	MCOContact   *string `json:"mco_contact,omitempty"`
	ResponseMs   int32   `json:"response_ms"`
}

type MissingEvent struct {
	EventEnvelope
	RecordID        string `json:"record_id"`
	RecordNumber    string `json:"record_number"`
	Category        string `json:"category"`
	MissingDate     string `json:"missing_date"`
	MissingLocation string `json:"missing_location"`
	EnteringAgency  string `json:"entering_agency"`
}

type StolenVehicleEvent struct {
	EventEnvelope
	RecordID     string `json:"record_id"`
	RecordNumber string `json:"record_number"`
	VIN          string `json:"vin,omitempty"`
	PlateNumber  string `json:"plate_number"`
	Make         string `json:"vehicle_make"`
	Model        string `json:"vehicle_model"`
	Year         int32  `json:"vehicle_year"`
	TheftDate    string `json:"theft_date"`
	OwnerNIU     string `json:"owner_niu,omitempty"`
}

type VehicleRecoveredEvent struct {
	EventEnvelope
	RecordID          string `json:"record_id"`
	RecordNumber      string `json:"record_number"`
	RecoveredLocation string `json:"recovered_location"`
	RecoveringAgency  string `json:"recovering_agency"`
	Notes             string `json:"notes,omitempty"`
	NotifiedOwner     bool   `json:"notified_owner"`
}

type StolenDocumentEvent struct {
	EventEnvelope
	RecordID       string `json:"record_id"`
	RecordNumber   string `json:"record_number"`
	DocumentType   string `json:"document_type"`
	DocumentNumber string `json:"document_number"`
	ReportDate     string `json:"report_date"`
	OwnerNIU       string `json:"owner_niu,omitempty"`
	TheftType      string `json:"theft_type"`
}

type StolenVesselEvent struct {
	EventEnvelope
	RecordID           string  `json:"record_id"`
	RecordNumber       string  `json:"record_number"`
	VesselName         string  `json:"vessel_name,omitempty"`
	RegistrationNumber string  `json:"registration_number,omitempty"`
	HullIDNumber       string  `json:"hull_id_number,omitempty"`
	VesselType         string  `json:"vessel_type,omitempty"`
	VesselMake         string  `json:"vessel_make,omitempty"`
	VesselLengthM      float64 `json:"vessel_length_m,omitempty"`
	HullColor          string  `json:"hull_color,omitempty"`
	HomePort           string  `json:"home_port,omitempty"`
	EngineSerial       string  `json:"engine_serial,omitempty"`
	DistinctiveMarks   string  `json:"distinctive_marks,omitempty"`
	TheftLocation      string  `json:"theft_location"`
	TheftDate          string  `json:"theft_date"`
	OwnerNIU           string  `json:"owner_niu,omitempty"`
}

type StolenFirearmEvent struct {
	EventEnvelope
	RecordID     string  `json:"record_id"`
	RecordNumber string  `json:"record_number"`
	SerialNumber string  `json:"serial_number"`
	Make         string  `json:"make,omitempty"`
	Model        string  `json:"model,omitempty"`
	Caliber      string  `json:"caliber,omitempty"`
	FirearmType  string  `json:"firearm_type,omitempty"`
	BarrelLength float64 `json:"barrel_length,omitempty"`
	TheftDate    string  `json:"theft_date"`
}

type ArmCrimeSceneHitEvent struct {
	EventEnvelope
	HitID           string `json:"hit_id"`
	FirearmRecordID string `json:"firearm_record_id"`
	SerialNumber    string `json:"serial_number,omitempty"`
	CrimeSceneRef   string `json:"crime_scene_ref"`
	CaseNumber      string `json:"case_number"`
	AlertLevel      string `json:"alert_level"`
}

type ExpungeEvent struct {
	EventEnvelope
	SampleID      string `json:"sample_id"`
	CourtOrderRef string `json:"court_order_ref"`
	OrderedBy     string `json:"ordered_by"`
	Reason        string `json:"reason"`
	OfficerNIU    string `json:"officer_niu"`
}

type StolenArticleEvent struct {
	EventEnvelope
	RecordID       string  `json:"record_id"`
	RecordNumber   string  `json:"record_number"`
	Category       string  `json:"category"`
	Description    string  `json:"description"`
	SerialNumber   string  `json:"serial_number,omitempty"`
	EstimatedValue float64 `json:"estimated_value,omitempty"`
	CurrencyCode   string  `json:"currency_code,omitempty"`
	TheftDate      string  `json:"theft_date"`
	TheftLocation  string  `json:"theft_location"`
	OwnerNIU       string  `json:"owner_niu,omitempty"`
}

type StolenSecurityEvent struct {
	EventEnvelope
	RecordID       string  `json:"record_id"`
	RecordNumber   string  `json:"record_number"`
	SecurityType   string  `json:"security_type"`
	Issuer         string  `json:"issuer"`
	SecurityNumber string  `json:"security_number"`
	FaceValue      float64 `json:"face_value,omitempty"`
	CurrencyCode   string  `json:"currency_code,omitempty"`
	TheftDate      string  `json:"theft_date"`
	OwnerNIU       string  `json:"owner_niu,omitempty"`
}

type ONIDocumentRevokedEvent struct {
	EventEnvelope
	DocumentType     string `json:"document_type"`
	DocumentNumber   string `json:"document_number"`
	RevocationReason string `json:"revocation_reason"`
	RevokedBy        string `json:"revoked_by"`
	RevokedAt        string `json:"revoked_at"`
}

type ForeignFugitiveCreated struct {
	EventEnvelope
	RecordID             string `json:"record_id"`
	RecordNumber         string `json:"record_number"`
	InterpolNoticeNumber string `json:"interpol_notice_number"`
	NoticeType           string `json:"notice_type"`
	IssuingCountry       string `json:"issuing_country"`
	EnteringAgency       string `json:"entering_agency"`
}

type UnidentifiedPersonCreated struct {
	EventEnvelope
	RecordID          string `json:"record_id"`
	RecordNumber      string `json:"record_number"`
	DiscoveryDate     string `json:"discovery_date"`
	DiscoveryLocation string `json:"discovery_location"`
	EnteringAgency    string `json:"entering_agency"`
	DNASampleRef      string `json:"dna_sample_ref,omitempty"`
}

type TerrorismWatchCreated struct {
	EventEnvelope
	RecordID       string `json:"record_id"`
	RecordNumber   string `json:"record_number"`
	ThreatType     string `json:"threat_type"`
	RiskLevel      string `json:"risk_level"`
	EnteringAgency string `json:"entering_agency"`
}

type ProtectionOrderCreated struct {
	EventEnvelope
	RecordID         string `json:"record_id"`
	RecordNumber     string `json:"record_number"`
	OrderType        string `json:"order_type"`
	BeneficiaryName  string `json:"beneficiary_name"`
	RestrainedPerson string `json:"restrained_person"`
	IssuingCourt     string `json:"issuing_court"`
}

type SupervisedReleaseCreated struct {
	EventEnvelope
	RecordID           string `json:"record_id"`
	RecordNumber       string `json:"record_number"`
	NIU                string `json:"niu"`
	SupervisionType    string `json:"supervision_type"`
	SupervisingOfficer string `json:"supervising_officer"`
	SupervisingAgency  string `json:"supervising_agency"`
}

type DuplicateSpecimenDetected struct {
	EventEnvelope
	SpecimenNumber   string `json:"specimen_number"`
	ExistingSampleID string `json:"existing_sample_id"`
	NewSubmissionID  string `json:"new_submission_id"`
}

type EquipmentRegistered struct {
	EventEnvelope
	EquipmentID   string `json:"equipment_id"`
	LabCode       string `json:"lab_code"`
	EquipmentName string `json:"equipment_name"`
}

type TrainingRecorded struct {
	EventEnvelope
	TrainingID   string `json:"training_id"`
	StaffNIU     string `json:"staff_niu"`
	TrainingName string `json:"training_name"`
}

type LDISUploadCompleted struct {
	EventEnvelope
	LabCode       string `json:"lab_code"`
	UploadedCount int    `json:"uploaded_count"`
	OperatorNIU   string `json:"operator_niu"`
}

type CrossDeptHitDetected struct {
	EventEnvelope
	HitID         string  `json:"hit_id"`
	QuerySampleID string  `json:"query_sample_id"`
	MatchSampleID string  `json:"match_sample_id"`
	MatchType     string  `json:"match_type"`
	Confidence    float64 `json:"confidence"`
	QuerySDIS     string  `json:"query_sdis"`
	MatchSDIS     string  `json:"match_sdis"`
}

type InterpolSubmissionRequested struct {
	EventEnvelope
	SubmissionID string   `json:"submission_id"`
	SampleIDs    []string `json:"sample_ids"`
	Reason       string   `json:"reason"`
}

type NDISReportGenerated struct {
	EventEnvelope
	ReportID    string `json:"report_id"`
	ReportType  string `json:"report_type"`
	GeneratedBy string `json:"generated_by"`
}

type ViolenceRecordCreated struct {
	EventEnvelope
	RecordID        string `json:"record_id"`
	RecordNumber    string `json:"record_number"`
	IncidentType    string `json:"incident_type"`
	NIU             string `json:"niu,omitempty"`
	ArrestingAgency string `json:"arresting_agency"`
	RiskLevel       string `json:"risk_level"`
}

type IdentityTheftRecorded struct {
	EventEnvelope
	RecordID        string `json:"record_id"`
	RecordNumber    string `json:"record_number"`
	VictimNIU       string `json:"victim_niu"`
	FraudType       string `json:"fraud_type"`
	ReportingAgency string `json:"reporting_agency"`
}

type BioIdentityLinked struct {
	EventEnvelope
	SampleID   string `json:"sample_id"`
	NIU        string `json:"niu"`
	LinkedBy   string `json:"linked_by"`
	CourtOrder string `json:"court_order,omitempty"`
}
