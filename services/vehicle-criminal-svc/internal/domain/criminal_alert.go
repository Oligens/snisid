package domain

import (
	"time"

	"github.com/google/uuid"
)

type CriminalAlert struct {
	AlertID             uuid.UUID     `json:"alert_id" db:"alert_id"`
	PlateNumber         string        `json:"plate_number" db:"plate_number"`
	PlateCategory       PlateCategory `json:"plate_category" db:"plate_category"`
	VIN                 *string       `json:"vin,omitempty" db:"vin"`
	ChassisNumber       *string       `json:"chassis_number,omitempty" db:"chassis_number"`
	VehicleType         VehicleType   `json:"vehicle_type" db:"vehicle_type"`
	Make                string        `json:"make" db:"make"`
	Model               string        `json:"model" db:"model"`
	Year                *int16        `json:"year,omitempty" db:"year"`
	ColorPrimary        string        `json:"color_primary" db:"color_primary"`
	ColorSecondary      *string       `json:"color_secondary,omitempty" db:"color_secondary"`
	DistinguishingMarks *string       `json:"distinguishing_marks,omitempty" db:"distinguishing_marks"`

	FovesVehicleID *uuid.UUID `json:"foves_vehicle_id,omitempty" db:"foves_vehicle_id"`

	CrimeCategory    CrimeCategory `json:"crime_category" db:"crime_category"`
	CrimeSubcategory *string       `json:"crime_subcategory,omitempty" db:"crime_subcategory"`
	AlertLevel       AlertLevel    `json:"alert_level" db:"alert_level"`
	Status           AlertStatus   `json:"status" db:"status"`

	ArmedAndDangerous  bool    `json:"armed_and_dangerous" db:"armed_and_dangerous"`
	DoNotStopAlone     bool    `json:"do_not_stop_alone" db:"do_not_stop_alone"`
	OfficerSafetyNotes *string `json:"officer_safety_notes,omitempty" db:"officer_safety_notes"`

	ReportingUnit      string     `json:"reporting_unit" db:"reporting_unit"`
	ReportingOfficerID *uuid.UUID `json:"reporting_officer_id,omitempty" db:"reporting_officer_id"`
	IncidentReference  *string    `json:"incident_reference,omitempty" db:"incident_reference"`
	IncidentDate       time.Time  `json:"incident_date" db:"incident_date"`
	ExpiryDate         *time.Time `json:"expiry_date,omitempty" db:"expiry_date"`

	AssociatedPersonIDs []uuid.UUID `json:"associated_person_ids" db:"associated_person_ids"`
	AssociatedCaseIDs   []uuid.UUID `json:"associated_case_ids" db:"associated_case_ids"`

	InterpolSMVID      *string    `json:"interpol_smv_id,omitempty" db:"interpol_smv_id"`
	InterpolReported   bool       `json:"interpol_reported" db:"interpol_reported"`
	InterpolReportedAt *time.Time `json:"interpol_reported_at,omitempty" db:"interpol_reported_at"`

	LastSeenLat      *float64   `json:"last_seen_lat,omitempty" db:"last_seen_lat"`
	LastSeenLng      *float64   `json:"last_seen_lng,omitempty" db:"last_seen_lng"`
	LastSeenLocation *string    `json:"last_seen_location,omitempty" db:"last_seen_location"`
	LastSeenDeptCode *string    `json:"last_seen_dept_code,omitempty" db:"last_seen_dept_code"`
	LastSeenCommune  *string    `json:"last_seen_commune,omitempty" db:"last_seen_commune"`
	LastSeenAt       *time.Time `json:"last_seen_at,omitempty" db:"last_seen_at"`

	PhotoRefs []string   `json:"photo_refs" db:"photo_refs"`
	CreatedBy uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	Version   int        `json:"version" db:"version"`
}

func (a *CriminalAlert) IsHighRisk() bool {
	return a.ArmedAndDangerous ||
		a.DoNotStopAlone ||
		a.AlertLevel == AlertLevelCritical ||
		a.CrimeCategory == CrimeCategoryKidnapping ||
		a.CrimeCategory == CrimeCategoryFakePolice
}

func (a *CriminalAlert) RequiresInterpolReport() bool {
	return !a.InterpolReported && (a.CrimeCategory == CrimeCategoryVehicleTheft ||
		a.CrimeCategory == CrimeCategoryDrugTraffic ||
		a.CrimeCategory == CrimeCategoryArmsTraffic ||
		a.AlertLevel == AlertLevelCritical)
}

type CreateAlertRequest struct {
	PlateNumber         string        `json:"plate_number" validate:"required"`
	PlateCategory       PlateCategory `json:"plate_category"`
	VIN                 *string       `json:"vin,omitempty"`
	ChassisNumber       *string       `json:"chassis_number,omitempty"`
	VehicleType         VehicleType   `json:"vehicle_type"`
	Make                string        `json:"make" validate:"required"`
	Model               string        `json:"model" validate:"required"`
	Year                *int16        `json:"year,omitempty"`
	ColorPrimary        string        `json:"color_primary" validate:"required"`
	ColorSecondary      *string       `json:"color_secondary,omitempty"`
	DistinguishingMarks *string       `json:"distinguishing_marks,omitempty"`
	CrimeCategory       CrimeCategory `json:"crime_category" validate:"required"`
	CrimeSubcategory    *string       `json:"crime_subcategory,omitempty"`
	AlertLevel          AlertLevel    `json:"alert_level"`
	ArmedAndDangerous   bool          `json:"armed_and_dangerous"`
	DoNotStopAlone      bool          `json:"do_not_stop_alone"`
	OfficerSafetyNotes  *string       `json:"officer_safety_notes,omitempty"`
	ReportingUnit       string        `json:"reporting_unit" validate:"required"`
	IncidentReference   *string       `json:"incident_reference,omitempty"`
	IncidentDate        time.Time     `json:"incident_date" validate:"required"`
	ExpiryDate          *time.Time    `json:"expiry_date,omitempty"`
	AssociatedPersonIDs []uuid.UUID   `json:"associated_person_ids,omitempty"`
	AssociatedCaseIDs   []uuid.UUID   `json:"associated_case_ids,omitempty"`
	LastSeenLat         *float64      `json:"last_seen_lat,omitempty"`
	LastSeenLng         *float64      `json:"last_seen_lng,omitempty"`
	LastSeenLocation    *string       `json:"last_seen_location,omitempty"`
	LastSeenDeptCode    *string       `json:"last_seen_dept_code,omitempty"`
	LastSeenCommune     *string       `json:"last_seen_commune,omitempty"`
	PhotoRefs           []string      `json:"photo_refs,omitempty"`
}

func NewCriminalAlert(req CreateAlertRequest, createdBy uuid.UUID) *CriminalAlert {
	if req.AlertLevel == "" {
		req.AlertLevel = AlertLevelCaution
	}
	return &CriminalAlert{
		AlertID:             uuid.New(),
		PlateNumber:         req.PlateNumber,
		PlateCategory:       req.PlateCategory,
		VIN:                 req.VIN,
		ChassisNumber:       req.ChassisNumber,
		VehicleType:         req.VehicleType,
		Make:                req.Make,
		Model:               req.Model,
		Year:                req.Year,
		ColorPrimary:        req.ColorPrimary,
		ColorSecondary:      req.ColorSecondary,
		DistinguishingMarks: req.DistinguishingMarks,
		CrimeCategory:       req.CrimeCategory,
		CrimeSubcategory:    req.CrimeSubcategory,
		AlertLevel:          req.AlertLevel,
		Status:              AlertStatusActive,
		ArmedAndDangerous:   req.ArmedAndDangerous,
		DoNotStopAlone:      req.DoNotStopAlone,
		OfficerSafetyNotes:  req.OfficerSafetyNotes,
		ReportingUnit:       req.ReportingUnit,
		IncidentReference:   req.IncidentReference,
		IncidentDate:        req.IncidentDate,
		ExpiryDate:          req.ExpiryDate,
		AssociatedPersonIDs: req.AssociatedPersonIDs,
		AssociatedCaseIDs:   req.AssociatedCaseIDs,
		LastSeenLat:         req.LastSeenLat,
		LastSeenLng:         req.LastSeenLng,
		LastSeenLocation:    req.LastSeenLocation,
		LastSeenDeptCode:    req.LastSeenDeptCode,
		LastSeenCommune:     req.LastSeenCommune,
		PhotoRefs:           req.PhotoRefs,
		CreatedBy:           createdBy,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		Version:             1,
	}
}

type PlateCheckResult struct {
	PlateNumber      string         `json:"plate_number"`
	CheckedAt        time.Time      `json:"checked_at"`
	HasCriminalAlert bool           `json:"has_criminal_alert"`
	HasStolenPlate   bool           `json:"has_stolen_plate"`
	Alert            *CriminalAlert `json:"alert,omitempty"`
	StolenPlate      *StolenPlate   `json:"stolen_plate,omitempty"`
	AlertLevel       AlertLevel     `json:"alert_level,omitempty"`
	Source           string         `json:"source,omitempty"`
}
