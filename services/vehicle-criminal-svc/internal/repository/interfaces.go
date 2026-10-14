package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/vehicle-criminal-svc/internal/domain"
)

type CriminalAlertRepository interface {
	Create(ctx context.Context, alert *domain.CriminalAlert) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.CriminalAlert, error)
	FindActiveByPlate(ctx context.Context, plateNumber string) (*domain.CriminalAlert, error)
	FindAll(ctx context.Context, filter AlertFilter) ([]*domain.CriminalAlert, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.AlertStatus, updatedBy uuid.UUID) error
	UpdateLastSeen(ctx context.Context, id uuid.UUID, lat, lng float64, location, deptCode, commune string) error
	Search(ctx context.Context, query string, filters AlertFilter) ([]*domain.CriminalAlert, int, error)
}

type StolenPlateRepository interface {
	Create(ctx context.Context, plate *domain.StolenPlate) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.StolenPlate, error)
	FindByPlate(ctx context.Context, plateNumber string) (*domain.StolenPlate, error)
	MarkRecovered(ctx context.Context, id uuid.UUID, location string, deptCode string) error
	FindStolenByPlate(ctx context.Context, plateNumber string) (*domain.StolenPlate, error)
}

type SightingRepository interface {
	Create(ctx context.Context, sighting *domain.VehicleSighting) error
	FindByAlertID(ctx context.Context, alertID uuid.UUID) ([]*domain.VehicleSighting, error)
	FindByPlate(ctx context.Context, plateNumber string) ([]*domain.VehicleSighting, error)
}

type DrugIncidentRepository interface {
	Create(ctx context.Context, incident *domain.DrugIncident) error
	FindByAlertID(ctx context.Context, alertID uuid.UUID) (*domain.DrugIncident, error)
}

type KidnappingIncidentRepository interface {
	Create(ctx context.Context, incident *domain.KidnappingIncident) error
	FindByAlertID(ctx context.Context, alertID uuid.UUID) (*domain.KidnappingIncident, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.KidnappingStatus) error
}

type GangAssociationRepository interface {
	Create(ctx context.Context, assoc *domain.GangAssociation) error
	FindByAlertID(ctx context.Context, alertID uuid.UUID) ([]*domain.GangAssociation, error)
	FindByGang(ctx context.Context, gangIdentifier string) ([]*domain.GangAssociation, error)
}

type IntelReportRepository interface {
	Create(ctx context.Context, report *domain.IntelligenceReport) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.IntelligenceReport, error)
	FindByUnit(ctx context.Context, unit string) ([]*domain.IntelligenceReport, error)
}

type InterpolSyncRepository interface {
	Create(ctx context.Context, log *domain.InterpolSyncLog) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SyncStatus, response interface{}) error
	FindPending(ctx context.Context) ([]*domain.InterpolSyncLog, error)
}

type HotlistCache interface {
	SetPlateAlert(ctx context.Context, plate string, alert *domain.CriminalAlert, ttl time.Duration) error
	GetPlateAlert(ctx context.Context, plate string) (*domain.CriminalAlert, error)
	DeletePlateAlert(ctx context.Context, plate string) error
	BulkLoadHotlist(ctx context.Context, alerts []*domain.CriminalAlert) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic string, event interface{}) error
}

type InterpolClient interface {
	SubmitSMV(ctx context.Context, alert *domain.CriminalAlert) (string, error)
	SubmitSMVAsync(ctx context.Context, alert *domain.CriminalAlert)
}

type FovesClient interface {
	VerifyStatePlate(ctx context.Context, plate string) (*FovesStatePlateResult, error)
}

type FovesStatePlateResult struct {
	IsRegistered bool   `json:"is_registered"`
	VehicleID    string `json:"vehicle_id,omitempty"`
	Agency       string `json:"agency,omitempty"`
}

type AlertFilter struct {
	DeptCode      string
	Category      string
	Level         string
	Status        string
	ReportingUnit string
	Page          int
	Limit         int
}
