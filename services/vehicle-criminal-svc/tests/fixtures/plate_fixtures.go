package fixtures

import (
	"time"

	"github.com/google/uuid"
	"github.com/snisid/vehicle-criminal-svc/internal/domain"
)

func CreateTestStolenStatePlate() *domain.StolenPlate {
	agency := "PNH"
	return &domain.StolenPlate{
		PlateID:            uuid.New(),
		PlateNumber:        "SE-00871",
		PlateCategory:      domain.PlateCategoryState,
		TheftDate:          time.Now().Add(-12 * time.Hour),
		TheftDeptCode:      "OU",
		TheftLocation:      strPtr("Delmas 31, Port-au-Prince"),
		TheftContext:       strPtr("Plaque clonée utilisée pour faux policiers kidnapping"),
		Status:             domain.StolenPlateStatusStolen,
		IsStatePlateClone:  true,
		ImpersonatedAgency: &agency,
		ReportingUnit:      "DCPJ",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func CreateTestDrugVehicleAlert() *domain.CriminalAlert {
	return &domain.CriminalAlert{
		AlertID:       uuid.New(),
		PlateNumber:   "PL-4444",
		PlateCategory: domain.PlateCategoryHeavy,
		VehicleType:   domain.VehicleTypeCamion,
		Make:          "Mitsubishi",
		Model:         "Canter",
		Year:          int16Ptr(2019),
		ColorPrimary:  "Bleu",
		CrimeCategory: domain.CrimeCategoryDrugTraffic,
		AlertLevel:    domain.AlertLevelWanted,
		Status:        domain.AlertStatusActive,
		ReportingUnit: "BLTS",
		IncidentDate:  time.Now().Add(-6 * time.Hour),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Version:       1,
	}
}

func CreateTestGangVehicleAlert() *domain.CriminalAlert {
	return &domain.CriminalAlert{
		AlertID:           uuid.New(),
		PlateNumber:       "M-7777",
		PlateCategory:     domain.PlateCategoryMoto,
		VehicleType:       domain.VehicleTypeMoto,
		Make:              "Yamaha",
		Model:             "DT125",
		ColorPrimary:      "Noir",
		CrimeCategory:     domain.CrimeCategoryGangAffiliated,
		AlertLevel:        domain.AlertLevelCritical,
		Status:            domain.AlertStatusActive,
		ArmedAndDangerous: true,
		DoNotStopAlone:    true,
		ReportingUnit:     "BAC",
		LastSeenDeptCode:  strPtr("OU"),
		LastSeenCommune:   strPtr("Cité Soleil"),
		LastSeenLocation:  strPtr("Cité Soleil, Avenue 15"),
		IncidentDate:      time.Now().Add(-3 * time.Hour),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Version:           1,
	}
}
