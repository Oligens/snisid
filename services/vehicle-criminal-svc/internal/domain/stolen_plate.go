package domain

import (
	"time"

	"github.com/google/uuid"
)

type StolenPlate struct {
	PlateID            uuid.UUID         `json:"plate_id" db:"plate_id"`
	PlateNumber        string            `json:"plate_number" db:"plate_number"`
	PlateCategory      PlateCategory     `json:"plate_category" db:"plate_category"`
	OriginalVehicleID  *uuid.UUID        `json:"original_vehicle_id,omitempty" db:"original_vehicle_id"`
	OriginalMake       *string           `json:"original_make,omitempty" db:"original_make"`
	OriginalModel      *string           `json:"original_model,omitempty" db:"original_model"`
	OriginalVIN        *string           `json:"original_vin,omitempty" db:"original_vin"`
	TheftDate          time.Time         `json:"theft_date" db:"theft_date"`
	TheftLocation      *string           `json:"theft_location,omitempty" db:"theft_location"`
	TheftDeptCode      string            `json:"theft_dept_code" db:"theft_dept_code"`
	TheftCommune       *string           `json:"theft_commune,omitempty" db:"theft_commune"`
	TheftContext       *string           `json:"theft_context,omitempty" db:"theft_context"`
	ReportingUnit      string            `json:"reporting_unit" db:"reporting_unit"`
	ReportingOfficerID *uuid.UUID        `json:"reporting_officer_id,omitempty" db:"reporting_officer_id"`
	BlvvCaseNumber     *string           `json:"blvv_case_number,omitempty" db:"blvv_case_number"`
	Status             StolenPlateStatus `json:"status" db:"status"`
	RecoveredDate      *time.Time        `json:"recovered_date,omitempty" db:"recovered_date"`
	RecoveryLocation   *string           `json:"recovery_location,omitempty" db:"recovery_location"`
	RecoveryDeptCode   *string           `json:"recovery_dept_code,omitempty" db:"recovery_dept_code"`
	UsedInCrime        bool              `json:"used_in_crime" db:"used_in_crime"`
	CrimeCategories    []CrimeCategory   `json:"crime_categories" db:"crime_categories"`
	CrimeAlertIDs      []uuid.UUID       `json:"crime_alert_ids" db:"crime_alert_ids"`
	IsStatePlateClone  bool              `json:"is_state_plate_clone" db:"is_state_plate_clone"`
	ImpersonatedAgency *string           `json:"impersonated_agency,omitempty" db:"impersonated_agency"`
	InterpolSADID      *string           `json:"interpol_sad_id,omitempty" db:"interpol_sad_id"`
	InterpolReported   bool              `json:"interpol_reported" db:"interpol_reported"`
	Notes              *string           `json:"notes,omitempty" db:"notes"`
	CreatedBy          uuid.UUID         `json:"created_by" db:"created_by"`
	CreatedAt          time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at" db:"updated_at"`
}

type DeclareStolenPlateRequest struct {
	PlateNumber        string        `json:"plate_number" validate:"required"`
	PlateCategory      PlateCategory `json:"plate_category" validate:"required"`
	OriginalMake       *string       `json:"original_make,omitempty"`
	OriginalModel      *string       `json:"original_model,omitempty"`
	OriginalVIN        *string       `json:"original_vin,omitempty"`
	TheftDate          time.Time     `json:"theft_date" validate:"required"`
	TheftLocation      *string       `json:"theft_location,omitempty"`
	TheftDeptCode      string        `json:"theft_dept_code" validate:"required"`
	TheftCommune       *string       `json:"theft_commune,omitempty"`
	TheftContext       *string       `json:"theft_context,omitempty"`
	ReportingUnit      string        `json:"reporting_unit" validate:"required"`
	BlvvCaseNumber     *string       `json:"blvv_case_number,omitempty"`
	IsStatePlateClone  bool          `json:"is_state_plate_clone"`
	ImpersonatedAgency *string       `json:"impersonated_agency,omitempty"`
	Notes              *string       `json:"notes,omitempty"`
}

func NewStolenPlate(req DeclareStolenPlateRequest, createdBy uuid.UUID) *StolenPlate {
	return &StolenPlate{
		PlateID:            uuid.New(),
		PlateNumber:        req.PlateNumber,
		PlateCategory:      req.PlateCategory,
		OriginalMake:       req.OriginalMake,
		OriginalModel:      req.OriginalModel,
		OriginalVIN:        req.OriginalVIN,
		TheftDate:          req.TheftDate,
		TheftLocation:      req.TheftLocation,
		TheftDeptCode:      req.TheftDeptCode,
		TheftCommune:       req.TheftCommune,
		TheftContext:       req.TheftContext,
		ReportingUnit:      req.ReportingUnit,
		BlvvCaseNumber:     req.BlvvCaseNumber,
		Status:             StolenPlateStatusStolen,
		IsStatePlateClone:  req.IsStatePlateClone,
		ImpersonatedAgency: req.ImpersonatedAgency,
		Notes:              req.Notes,
		CreatedBy:          createdBy,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}
