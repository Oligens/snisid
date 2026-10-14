package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/vehicle-criminal-svc/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAlertRepo struct {
	mock.Mock
}

type MockHotlist struct {
	mock.Mock
}

type MockKafka struct {
	mock.Mock
}

type MockInterpol struct {
	mock.Mock
}

func (m *MockAlertRepo) Create(ctx context.Context, alert *domain.CriminalAlert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.CriminalAlert, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.CriminalAlert), args.Error(1)
}

func (m *MockAlertRepo) FindActiveByPlate(ctx context.Context, plate string) (*domain.CriminalAlert, error) {
	args := m.Called(ctx, plate)
	return args.Get(0).(*domain.CriminalAlert), args.Error(1)
}

func (m *MockAlertRepo) FindAll(ctx context.Context, filter interface{}) ([]*domain.CriminalAlert, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*domain.CriminalAlert), args.Int(1), args.Error(2)
}

func (m *MockAlertRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.AlertStatus, updatedBy uuid.UUID) error {
	args := m.Called(ctx, id, status, updatedBy)
	return args.Error(0)
}

func (m *MockAlertRepo) UpdateLastSeen(ctx context.Context, id uuid.UUID, lat, lng float64, location, deptCode, commune string) error {
	args := m.Called(ctx, id, lat, lng, location, deptCode, commune)
	return args.Error(0)
}

func (m *MockAlertRepo) Search(ctx context.Context, query string, filter interface{}) ([]*domain.CriminalAlert, int, error) {
	args := m.Called(ctx, query, filter)
	return args.Get(0).([]*domain.CriminalAlert), args.Int(1), args.Error(2)
}

func (m *MockHotlist) SetPlateAlert(ctx context.Context, plate string, alert *domain.CriminalAlert, ttl time.Duration) error {
	args := m.Called(ctx, plate, alert, ttl)
	return args.Error(0)
}

func (m *MockHotlist) GetPlateAlert(ctx context.Context, plate string) (*domain.CriminalAlert, error) {
	args := m.Called(ctx, plate)
	return args.Get(0).(*domain.CriminalAlert), args.Error(1)
}

func (m *MockHotlist) DeletePlateAlert(ctx context.Context, plate string) error {
	args := m.Called(ctx, plate)
	return args.Error(0)
}

func (m *MockHotlist) BulkLoadHotlist(ctx context.Context, alerts []*domain.CriminalAlert) error {
	args := m.Called(ctx, alerts)
	return args.Error(0)
}

func (m *MockKafka) Publish(ctx context.Context, topic string, event interface{}) error {
	args := m.Called(ctx, topic, event)
	return args.Error(0)
}

func (m *MockInterpol) SubmitSMV(ctx context.Context, alert *domain.CriminalAlert) (string, error) {
	args := m.Called(ctx, alert)
	return args.String(0), args.Error(1)
}

func (m *MockInterpol) SubmitSMVAsync(ctx context.Context, alert *domain.CriminalAlert) {
	m.Called(ctx, alert)
}

func TestValidatePlateNumber(t *testing.T) {
	tests := []struct {
		name    string
		plate   string
		wantErr bool
	}{
		{"valid plate PP-1234", "PP-1234", false},
		{"valid plate SE-00871", "SE-00871", false},
		{"valid plate ABC-123456", "ABC-123456", false},
		{"invalid plate empty", "", true},
		{"invalid plate too short", "AB", true},
		{"invalid plate numbers", "ABC-ABCDE", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidatePlateNumber(tt.plate)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsStatePlate(t *testing.T) {
	tests := []struct {
		plate string
		want  bool
	}{
		{"SE-00871", true},
		{"SE12345", true},
		{"SE 12345", true},
		{"PP-1234", false},
		{"ABC-123", false},
	}

	for _, tt := range tests {
		t.Run(tt.plate, func(t *testing.T) {
			got := domain.IsStatePlate(tt.plate)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCriminalAlert_IsHighRisk(t *testing.T) {
	tests := []struct {
		name   string
		alert  *domain.CriminalAlert
		expect bool
	}{
		{
			name: "armed and dangerous",
			alert: &domain.CriminalAlert{
				ArmedAndDangerous: true,
				AlertLevel:        domain.AlertLevelCaution,
			},
			expect: true,
		},
		{
			name: "critical level",
			alert: &domain.CriminalAlert{
				AlertLevel: domain.AlertLevelCritical,
			},
			expect: true,
		},
		{
			name: "kidnapping",
			alert: &domain.CriminalAlert{
				CrimeCategory: domain.CrimeCategoryKidnapping,
				AlertLevel:    domain.AlertLevelCaution,
			},
			expect: true,
		},
		{
			name: "low risk",
			alert: &domain.CriminalAlert{
				AlertLevel:    domain.AlertLevelInfo,
				CrimeCategory: domain.CrimeCategoryVehicleTheft,
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.alert.IsHighRisk())
		})
	}
}

func TestCriminalAlert_RequiresInterpolReport(t *testing.T) {
	tests := []struct {
		name   string
		alert  *domain.CriminalAlert
		expect bool
	}{
		{
			name: "vehicle theft not reported",
			alert: &domain.CriminalAlert{
				CrimeCategory:    domain.CrimeCategoryVehicleTheft,
				InterpolReported: false,
			},
			expect: true,
		},
		{
			name: "already reported",
			alert: &domain.CriminalAlert{
				CrimeCategory:    domain.CrimeCategoryVehicleTheft,
				InterpolReported: true,
			},
			expect: false,
		},
		{
			name: "critical level",
			alert: &domain.CriminalAlert{
				AlertLevel:       domain.AlertLevelCritical,
				InterpolReported: false,
			},
			expect: true,
		},
		{
			name: "low level not requiring report",
			alert: &domain.CriminalAlert{
				CrimeCategory:    domain.CrimeCategoryPlatTheft,
				AlertLevel:       domain.AlertLevelInfo,
				InterpolReported: false,
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.alert.RequiresInterpolReport())
		})
	}
}

func TestNewCriminalAlert(t *testing.T) {
	req := domain.CreateAlertRequest{
		PlateNumber:   "PP-1234",
		Make:          "Toyota",
		Model:         "Land Cruiser",
		ColorPrimary:  "Blanc",
		CrimeCategory: domain.CrimeCategoryVehicleTheft,
		ReportingUnit: "BLVV",
		IncidentDate:  time.Now(),
	}

	userID := uuid.New()
	alert := domain.NewCriminalAlert(req, userID)

	assert.NotEmpty(t, alert.AlertID)
	assert.Equal(t, "PP-1234", alert.PlateNumber)
	assert.Equal(t, "Toyota", alert.Make)
	assert.Equal(t, domain.CrimeCategoryVehicleTheft, alert.CrimeCategory)
	assert.Equal(t, domain.AlertLevelCaution, alert.AlertLevel)
	assert.Equal(t, domain.AlertStatusActive, alert.Status)
	assert.Equal(t, "BLVV", alert.ReportingUnit)
	assert.Equal(t, 1, alert.Version)
}
