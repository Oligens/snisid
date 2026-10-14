package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type PathogenType string

const (
	PathogenVirus    PathogenType = "VIRUS"
	PathogenBacteria PathogenType = "BACTERIA"
	PathogenParasite PathogenType = "PARASITE"
	PathogenFungus   PathogenType = "FUNGUS"
	PathogenUnknown  PathogenType = "UNKNOWN"
)

type AlertLevel string

const (
	AlertGreen  AlertLevel = "GREEN"
	AlertYellow AlertLevel = "YELLOW"
	AlertOrange AlertLevel = "ORANGE"
	AlertRed    AlertLevel = "RED"
)

type TransmissionMode string

const (
	TransmissionAirborne   TransmissionMode = "AIRBORNE"
	TransmissionFoodborne  TransmissionMode = "FOODBORNE"
	TransmissionWaterborne TransmissionMode = "WATERBORNE"
	TransmissionVector     TransmissionMode = "VECTOR"
	TransmissionContact    TransmissionMode = "CONTACT"
	TransmissionUnknown    TransmissionMode = "UNKNOWN"
)

type FacilityType string

const (
	FacilityHospital      FacilityType = "HOSPITAL"
	FacilityClinic        FacilityType = "CLINIC"
	FacilityLab           FacilityType = "LAB"
	FacilityPharmacy      FacilityType = "PHARMACY"
	FacilityEmergencyPost FacilityType = "EMERGENCY_POST"
)

type StockStatus string

const (
	StockAdequate   StockStatus = "ADEQUATE"
	StockLow        StockStatus = "LOW"
	StockCritical   StockStatus = "CRITICAL"
	StockOutOfStock StockStatus = "OUT_OF_STOCK"
)

type DiseaseAlert struct {
	ID                  uuid.UUID        `json:"id"`
	DiseaseName         string           `json:"disease_name"`
	PathogenType        PathogenType     `json:"pathogen_type"`
	Icd10Code           string           `json:"icd10_code"`
	AlertLevel          AlertLevel       `json:"alert_level"`
	FirstCaseDetected   time.Time        `json:"first_case_detected_at"`
	SymptomsHallmark    *string          `json:"symptoms_hallmark,omitempty"`
	TransmissionMode    TransmissionMode `json:"transmission_mode"`
	IncubationDays      int              `json:"incubation_days"`
	FatalityRate        float64          `json:"fatality_rate"`
	CasesConfirmed      int              `json:"cases_confirmed"`
	CasesSuspected      int              `json:"cases_suspected"`
	CasesDeaths         int              `json:"cases_deaths"`
	AffectedRegions     pq.StringArray   `json:"affected_regions"`
	SourceLab           *string          `json:"source_lab,omitempty"`
	WhoAlertRef         *string          `json:"who_alert_ref,omitempty"`
	ContainmentMeasures *string          `json:"containment_measures,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
}

type VaccinationCampaign struct {
	ID                uuid.UUID      `json:"id"`
	CampaignName      string         `json:"campaign_name"`
	TargetDisease     string         `json:"target_disease"`
	VaccineType       string         `json:"vaccine_type"`
	TargetPopulation  int            `json:"target_population"`
	DosesAdministered int            `json:"doses_administered"`
	CoveragePct       float64        `json:"coverage_pct"`
	RegionsActive     pq.StringArray `json:"regions_active"`
	StartDate         time.Time      `json:"start_date"`
	EndDate           *time.Time     `json:"end_date,omitempty"`
	CoordinatorAgency string         `json:"coordinator_agency"`
	CreatedAt         time.Time      `json:"created_at"`
}

type HealthFacility struct {
	ID             uuid.UUID    `json:"id"`
	FacilityName   string       `json:"facility_name"`
	FacilityType   FacilityType `json:"facility_type"`
	Region         string       `json:"region"`
	Commune        string       `json:"commune"`
	DeptCode       string       `json:"dept_code"`
	CapacityBeds   int          `json:"capacity_beds"`
	BedsAvailable  int          `json:"beds_available"`
	StockStatus    StockStatus  `json:"stock_status"`
	HasVentilators bool         `json:"has_ventilators"`
	HasAmbulance   bool         `json:"has_ambulance"`
	LastReportAt   *time.Time   `json:"last_report_at,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

type CreateDiseaseAlertRequest struct {
	DiseaseName         string   `json:"disease_name"`
	PathogenType        string   `json:"pathogen_type"`
	Icd10Code           string   `json:"icd10_code"`
	AlertLevel          string   `json:"alert_level"`
	FirstCaseDetected   string   `json:"first_case_detected,omitempty"`
	SymptomsHallmark    string   `json:"symptoms_hallmark,omitempty"`
	TransmissionMode    string   `json:"transmission_mode"`
	IncubationDays      int      `json:"incubation_days"`
	FatalityRate        float64  `json:"fatality_rate"`
	CasesConfirmed      int      `json:"cases_confirmed"`
	CasesSuspected      int      `json:"cases_suspected"`
	CasesDeaths         int      `json:"cases_deaths"`
	AffectedRegions     []string `json:"affected_regions"`
	SourceLab           string   `json:"source_lab,omitempty"`
	WhoAlertRef         string   `json:"who_alert_ref,omitempty"`
	ContainmentMeasures string   `json:"containment_measures,omitempty"`
}

type CreateVaccinationCampaignRequest struct {
	CampaignName      string   `json:"campaign_name"`
	TargetDisease     string   `json:"target_disease"`
	VaccineType       string   `json:"vaccine_type"`
	TargetPopulation  int      `json:"target_population"`
	DosesAdministered int      `json:"doses_administered"`
	CoveragePct       float64  `json:"coverage_pct"`
	RegionsActive     []string `json:"regions_active"`
	StartDate         string   `json:"start_date"`
	EndDate           string   `json:"end_date,omitempty"`
	CoordinatorAgency string   `json:"coordinator_agency"`
}

type UpdateFacilityStockRequest struct {
	StockStatus   string `json:"stock_status"`
	BedsAvailable int    `json:"beds_available"`
}

type DashboardNational struct {
	TotalAlerts         int     `json:"total_alerts"`
	ActiveAlerts        int     `json:"active_alerts"`
	TotalCampaigns      int     `json:"total_campaigns"`
	TotalFacilities     int     `json:"total_facilities"`
	CriticalFacilities  int     `json:"critical_facilities"`
	AvgCoveragePct      float64 `json:"avg_coverage_pct"`
	TotalCasesConfirmed int     `json:"total_cases_confirmed"`
	TotalDeaths         int     `json:"total_deaths"`
}
