package models

// ── Allele & IndexType (CODIS Standard) ────────────────────────────────────

type Allele struct {
	Locus  string   `json:"locus"`
	Value1 float64  `json:"value1"`
	Value2 *float64 `json:"value2,omitempty"`
}

type IndexType string

const (
	IndexConvicted     IndexType = "BIO-CON"
	IndexArrestee      IndexType = "BIO-ARR"
	IndexForensic      IndexType = "BIO-FSC"
	IndexMissingPerson IndexType = "BIO-DIS"
	IndexUnidentified  IndexType = "BIO-RNI"
)

// 20 CODIS Core Loci + Amelogenin (NIST 2017)
var CODISCoreLoci = []string{
	"CSF1PO", "D3S1358", "D5S818", "D7S820", "D8S1179",
	"D13S317", "D16S539", "D18S51", "D21S11", "FGA",
	"TH01", "TPOX", "vWA", "D1S1656", "D2S441",
	"D2S1338", "D10S1248", "D12S391", "D19S433", "D22S1045",
	"Amelogenin",
}

// ── BioIdentityLink (dissociation ADN / identité civile) ────────────────────

type BioIdentityLink struct {
	SampleID   string `json:"sample_id"`
	NIU        string `json:"niu"`
	LinkedBy   string `json:"linked_by"`
	LinkedAt   string `json:"linked_at"`
	CourtOrder string `json:"court_order,omitempty"`
}

type DNAProfile struct {
	SampleID       string  `json:"sample_id"`
	SpecimenNumber string  `json:"specimen_number"`
	IndexType      string  `json:"index_type"`
	LociHash       string  `json:"loci_hash"`
	Amelogenin     string  `json:"amelogenin,omitempty"`
	QualityScore   float64 `json:"quality_score"`
	LociCount      int     `json:"loci_count"`
	LabID          string  `json:"lab_id"`
	CaseNumber     string  `json:"case_number,omitempty"`
	CollectedDate  string  `json:"collected_date"`
	IsExpunged     bool    `json:"is_expunged"`
}

type WantedPerson struct {
	RecordID        string   `json:"record_id"`
	RecordNumber    string   `json:"record_number"`
	NIU             string   `json:"niu,omitempty"`
	LastName        string   `json:"last_name,omitempty"`
	FirstName       string   `json:"first_name,omitempty"`
	Aliases         []string `json:"aliases,omitempty"`
	DateOfBirth     string   `json:"date_of_birth,omitempty"`
	Gender          string   `json:"gender,omitempty"`
	Nationality     string   `json:"nationality,omitempty"`
	WarrantType     string   `json:"warrant_type"`
	WarrantNumber   string   `json:"warrant_number,omitempty"`
	IssuingCourt    string   `json:"issuing_court,omitempty"`
	IssuingDate     string   `json:"issuing_date"`
	Charges         []string `json:"charges"`
	DangerLevel     string   `json:"danger_level"`
	ArmedDangerous  bool     `json:"armed_dangerous"`
	EnteringAgency  string   `json:"entering_agency"`
	MCOContact      string   `json:"mco_contact,omitempty"`
	Status          string   `json:"status"`
	ExpiryDate      string   `json:"expiry_date,omitempty"`
	EnteringOfficer string   `json:"entering_officer,omitempty"`
	InterpolNotice  string   `json:"interpol_notice,omitempty"`
}

type WantedQuery struct {
	LastName    string
	FirstName   string
	NIU         string
	PlateNumber string
	Status      string
	Limit       int
	Offset      int
}

type PlateHitResult struct {
	HitFound     bool   `json:"hit_found"`
	HitType      string `json:"hit_type,omitempty"`
	RecordNumber string `json:"record_number,omitempty"`
	AlertLevel   string `json:"alert_level,omitempty"`
	MCOContact   string `json:"mco_contact,omitempty"`
	ResponseMs   int    `json:"response_ms"`
}

type StolenVehicle struct {
	RecordID       string `json:"record_id"`
	RecordNumber   string `json:"record_number"`
	VIN            string `json:"vin,omitempty"`
	PlateNumber    string `json:"plate_number"`
	VehicleMake    string `json:"vehicle_make"`
	VehicleModel   string `json:"vehicle_model"`
	VehicleYear    int    `json:"vehicle_year"`
	VehicleColor   string `json:"vehicle_color,omitempty"`
	TheftDate      string `json:"theft_date"`
	TheftLocation  string `json:"theft_location"`
	OwnerNIU       string `json:"owner_niu,omitempty"`
	OwnerName      string `json:"owner_name,omitempty"`
	Status         string `json:"status"`
	EnteringAgency string `json:"entering_agency"`
}

type StolenFirearm struct {
	RecordID       string  `json:"record_id"`
	RecordNumber   string  `json:"record_number"`
	SerialNumber   string  `json:"serial_number"`
	Make           string  `json:"make,omitempty"`
	Model          string  `json:"model,omitempty"`
	Caliber        string  `json:"caliber,omitempty"`
	FirearmType    string  `json:"firearm_type,omitempty"`
	BarrelLength   float64 `json:"barrel_length,omitempty"`
	TheftDate      string  `json:"theft_date"`
	TheftLocation  string  `json:"theft_location,omitempty"`
	OwnerNIU       string  `json:"owner_niu,omitempty"`
	Status         string  `json:"status"`
	EnteringAgency string  `json:"entering_agency"`
}

type StolenDocument struct {
	RecordID       string `json:"record_id"`
	RecordNumber   string `json:"record_number"`
	DocumentType   string `json:"document_type"`
	DocumentNumber string `json:"document_number,omitempty"`
	IssuingAgency  string `json:"issuing_agency,omitempty"`
	IssueDate      string `json:"issue_date,omitempty"`
	ReportDate     string `json:"report_date"`
	OwnerNIU       string `json:"owner_niu,omitempty"`
	TheftType      string `json:"theft_type"`
	Status         string `json:"status"`
}

type StolenVessel struct {
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
	OwnerName          string  `json:"owner_name,omitempty"`
	Status             string  `json:"status"`
}

type StolenArticle struct {
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
	Status         string  `json:"status"`
	EnteringAgency string  `json:"entering_agency"`
}

type ForeignFugitive struct {
	RecordID             string   `json:"record_id"`
	RecordNumber         string   `json:"record_number"`
	InterpolNoticeNumber string   `json:"interpol_notice_number"`
	NoticeType           string   `json:"notice_type"`
	LastName             string   `json:"last_name"`
	FirstName            string   `json:"first_name,omitempty"`
	Aliases              []string `json:"aliases,omitempty"`
	DateOfBirth          string   `json:"date_of_birth,omitempty"`
	Gender               string   `json:"gender,omitempty"`
	Nationality          string   `json:"nationality,omitempty"`
	Charges              []string `json:"charges"`
	IssuingCountry       string   `json:"issuing_country"`
	EnteringAgency       string   `json:"entering_agency"`
	Status               string   `json:"status"`
}

type UnidentifiedPerson struct {
	RecordID            string `json:"record_id"`
	RecordNumber        string `json:"record_number"`
	DiscoveryDate       string `json:"discovery_date"`
	DiscoveryLocation   string `json:"discovery_location"`
	DiscoveryDepartment string `json:"discovery_department,omitempty"`
	EstimatedAgeMin     int    `json:"estimated_age_min,omitempty"`
	EstimatedAgeMax     int    `json:"estimated_age_max,omitempty"`
	Gender              string `json:"gender,omitempty"`
	EstimatedHeightCM   int    `json:"estimated_height_cm,omitempty"`
	EstimatedWeightKG   int    `json:"estimated_weight_kg,omitempty"`
	HairColor           string `json:"hair_color,omitempty"`
	EyeColor            string `json:"eye_color,omitempty"`
	DistinctiveFeatures string `json:"distinctive_features,omitempty"`
	ClothingDescription string `json:"clothing_description,omitempty"`
	DNASampleRef        string `json:"dna_sample_ref,omitempty"`
	FingerprintRef      string `json:"fingerprint_ref,omitempty"`
	DentalRecordsRef    string `json:"dental_records_ref,omitempty"`
	EnteringAgency      string `json:"entering_agency"`
	Status              string `json:"status"`
}

type TerrorismWatch struct {
	RecordID           string   `json:"record_id"`
	RecordNumber       string   `json:"record_number"`
	NIU                string   `json:"niu,omitempty"`
	LastName           string   `json:"last_name"`
	FirstName          string   `json:"first_name,omitempty"`
	Aliases            []string `json:"aliases,omitempty"`
	DateOfBirth        string   `json:"date_of_birth,omitempty"`
	Nationality        string   `json:"nationality,omitempty"`
	RiskLevel          string   `json:"risk_level"`
	ThreatType         string   `json:"threat_type"`
	Groups             []string `json:"groups,omitempty"`
	LastKnownLocation  string   `json:"last_known_location,omitempty"`
	EnteringAgency     string   `json:"entering_agency"`
	ApprovedByDirector string   `json:"approved_by_director"`
	ApprovedByPG       string   `json:"approved_by_pg"`
	Status             string   `json:"status"`
}

type ProtectionOrder struct {
	RecordID         string   `json:"record_id"`
	RecordNumber     string   `json:"record_number"`
	OrderType        string   `json:"order_type"`
	IssuingCourt     string   `json:"issuing_court"`
	IssuingJudge     string   `json:"issuing_judge"`
	BeneficiaryNIU   string   `json:"beneficiary_niu,omitempty"`
	BeneficiaryName  string   `json:"beneficiary_name"`
	RestrainedPerson string   `json:"restrained_person"`
	Restrictions     []string `json:"restrictions"`
	IssueDate        string   `json:"issue_date"`
	ExpirationDate   string   `json:"expiration_date,omitempty"`
	EmergencyContact string   `json:"emergency_contact,omitempty"`
	Status           string   `json:"status"`
}

type SupervisedRelease struct {
	RecordID           string   `json:"record_id"`
	RecordNumber       string   `json:"record_number"`
	NIU                string   `json:"niu"`
	LastName           string   `json:"last_name"`
	FirstName          string   `json:"first_name,omitempty"`
	SupervisionType    string   `json:"supervision_type"`
	StartDate          string   `json:"start_date"`
	EndDate            string   `json:"end_date,omitempty"`
	Conditions         []string `json:"conditions"`
	SupervisingOfficer string   `json:"supervising_officer"`
	SupervisingAgency  string   `json:"supervising_agency"`
	Status             string   `json:"status"`
}

type StolenSecurity struct {
	RecordID       string  `json:"record_id"`
	RecordNumber   string  `json:"record_number"`
	SecurityType   string  `json:"security_type"`
	Issuer         string  `json:"issuer"`
	SecurityNumber string  `json:"security_number"`
	FaceValue      float64 `json:"face_value,omitempty"`
	CurrencyCode   string  `json:"currency_code,omitempty"`
	IssueDate      string  `json:"issue_date,omitempty"`
	TheftDate      string  `json:"theft_date"`
	TheftLocation  string  `json:"theft_location"`
	OwnerNIU       string  `json:"owner_niu,omitempty"`
	Status         string  `json:"status"`
	EnteringAgency string  `json:"entering_agency"`
}

type LabEquipment struct {
	ID              string `json:"id"`
	LabCode         string `json:"lab_code"`
	EquipmentName   string `json:"equipment_name"`
	Model           string `json:"model,omitempty"`
	SerialNumber    string `json:"serial_number"`
	Role            string `json:"role"`
	CalibrationDate string `json:"calibration_date,omitempty"`
	CalibrationDue  string `json:"calibration_due,omitempty"`
	Status          string `json:"status"`
}

type StaffTraining struct {
	ID            string `json:"id"`
	StaffNIU      string `json:"staff_niu"`
	TrainingName  string `json:"training_name"`
	TrainingCode  string `json:"training_code"`
	DurationHours int    `json:"duration_hours,omitempty"`
	CompletedDate string `json:"completed_date"`
	ValidUntil    string `json:"valid_until,omitempty"`
	IssuedBy      string `json:"issued_by"`
	Frequency     string `json:"frequency,omitempty"`
}

type NdisCrossDeptHit struct {
	HitID         string  `json:"hit_id"`
	QuerySampleID string  `json:"query_sample_id"`
	MatchSampleID string  `json:"match_sample_id"`
	MatchType     string  `json:"match_type"`
	Confidence    float64 `json:"confidence"`
	QuerySDIS     string  `json:"query_sdis"`
	MatchSDIS     string  `json:"match_sdis"`
	NotifiedAt    string  `json:"notified_at,omitempty"`
	AlertLevel    string  `json:"alert_level"`
}

type NdisStats struct {
	TotalBIOCon           int     `json:"total_bio_con"`
	TotalBIOArr           int     `json:"total_bio_arr"`
	TotalBIOFsc           int     `json:"total_bio_fsc"`
	TotalBIODis           int     `json:"total_bio_dis"`
	TotalBIORni           int     `json:"total_bio_rni"`
	CrossDeptHitsThisWeek int     `json:"cross_dept_hits_this_week"`
	HitRatePercent        float64 `json:"hit_rate_percent"`
}

type NdisReport struct {
	ID          string `json:"id"`
	ReportType  string `json:"report_type"`
	GeneratedAt string `json:"generated_at"`
	Status      string `json:"status"`
	FilePath    string `json:"file_path,omitempty"`
}

type InterpolSubmission struct {
	ID          string   `json:"id"`
	SampleIDs   []string `json:"sample_ids"`
	Reason      string   `json:"reason"`
	CaseNumber  string   `json:"case_number,omitempty"`
	Status      string   `json:"status"`
	SubmittedAt string   `json:"submitted_at,omitempty"`
}

type SdisNode struct {
	Code          string `json:"code"`
	Department    string `json:"department"`
	DCLocation    string `json:"dc_location"`
	DCType        string `json:"dc_type"`
	Status        string `json:"status"`
	LastHeartbeat string `json:"last_heartbeat,omitempty"`
}

type SdisMatch struct {
	ID            string  `json:"id"`
	SdisCode      string  `json:"sdis_code"`
	QuerySampleID string  `json:"query_sample_id"`
	MatchSampleID string  `json:"match_sample_id"`
	MatchType     string  `json:"match_type"`
	Confidence    float64 `json:"confidence"`
	Alerted       bool    `json:"alerted"`
}

type SdisSyncError struct {
	ID         string `json:"id"`
	SdisCode   string `json:"sdis_code"`
	ErrorType  string `json:"error_type"`
	Details    string `json:"details,omitempty"`
	RetryCount int    `json:"retry_count"`
	Resolved   bool   `json:"resolved"`
}

type SdisQualityReview struct {
	ID           string  `json:"id"`
	SampleID     string  `json:"sample_id"`
	SdisCode     string  `json:"sdis_code"`
	QualityScore float64 `json:"quality_score"`
	Reason       string  `json:"reason"`
	Reviewed     bool    `json:"reviewed"`
}

type ViolenceRecord struct {
	RecordID        string `json:"record_id"`
	RecordNumber    string `json:"record_number"`
	NIU             string `json:"niu,omitempty"`
	LastName        string `json:"last_name,omitempty"`
	FirstName       string `json:"first_name,omitempty"`
	IncidentType    string `json:"incident_type"`
	IncidentDate    string `json:"incident_date"`
	Location        string `json:"location"`
	VictimNIU       string `json:"victim_niu,omitempty"`
	VictimName      string `json:"victim_name,omitempty"`
	ArrestingAgency string `json:"arresting_agency"`
	CourtCaseRef    string `json:"court_case_ref,omitempty"`
	RiskLevel       string `json:"risk_level"`
	Status          string `json:"status"`
}

type IdentityTheft struct {
	RecordID         string `json:"record_id"`
	RecordNumber     string `json:"record_number"`
	VictimNIU        string `json:"victim_niu"`
	VictimName       string `json:"victim_name,omitempty"`
	FraudType        string `json:"fraud_type"`
	DocumentTypeUsed string `json:"document_type_used,omitempty"`
	PerpetratorKnown bool   `json:"perpetrator_known"`
	PerpetratorName  string `json:"perpetrator_name,omitempty"`
	ReportDate       string `json:"report_date"`
	ReportingAgency  string `json:"reporting_agency"`
	Status           string `json:"status"`
}
