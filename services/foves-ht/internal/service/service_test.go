package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/foves-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockFovesRepo struct {
	createVehicleFn        func(ctx context.Context, v *domain.Vehicle) error
	findByPlateFn          func(ctx context.Context, plate string) (*domain.Vehicle, error)
	findByVINFn            func(ctx context.Context, vin string) (*domain.Vehicle, error)
	findByOwnerFn          func(ctx context.Context, citizenID uuid.UUID) ([]domain.Vehicle, error)
	createTransferFn       func(ctx context.Context, t *domain.OwnershipTransfer) error
	updateVehicleOwnerFn   func(ctx context.Context, vehicleID, newOwnerID uuid.UUID) error
	createLicenseFn        func(ctx context.Context, l *domain.DriverLicense) error
	findLicenseByCitizenFn func(ctx context.Context, citizenID uuid.UUID) (*domain.DriverLicense, error)
}

func (m *mockFovesRepo) CreateVehicle(ctx context.Context, v *domain.Vehicle) error {
	return m.createVehicleFn(ctx, v)
}
func (m *mockFovesRepo) FindByPlate(ctx context.Context, plate string) (*domain.Vehicle, error) {
	return m.findByPlateFn(ctx, plate)
}
func (m *mockFovesRepo) FindByVIN(ctx context.Context, vin string) (*domain.Vehicle, error) {
	return m.findByVINFn(ctx, vin)
}
func (m *mockFovesRepo) FindByOwner(ctx context.Context, citizenID uuid.UUID) ([]domain.Vehicle, error) {
	return m.findByOwnerFn(ctx, citizenID)
}
func (m *mockFovesRepo) CreateTransfer(ctx context.Context, t *domain.OwnershipTransfer) error {
	return m.createTransferFn(ctx, t)
}
func (m *mockFovesRepo) UpdateVehicleOwner(ctx context.Context, vehicleID, newOwnerID uuid.UUID) error {
	return m.updateVehicleOwnerFn(ctx, vehicleID, newOwnerID)
}
func (m *mockFovesRepo) CreateLicense(ctx context.Context, l *domain.DriverLicense) error {
	return m.createLicenseFn(ctx, l)
}
func (m *mockFovesRepo) FindLicenseByCitizen(ctx context.Context, citizenID uuid.UUID) (*domain.DriverLicense, error) {
	return m.findLicenseByCitizenFn(ctx, citizenID)
}

func TestRegisterVehicle(t *testing.T) {
	ownerID := uuid.New()
	tests := []struct {
		name    string
		vehicle *domain.Vehicle
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			vehicle: &domain.Vehicle{
				PlateNumber:    "ABC-1234",
				VIN:            "1HGCM82633A004352",
				Make:           "Toyota",
				Model:          "Corolla",
				Year:           2025,
				Category:       domain.VehiclePrivateCar,
				OwnerCitizenID: ownerID,
			},
		},
		{
			name: "repo error",
			vehicle: &domain.Vehicle{
				PlateNumber:    "ERR-0000",
				VIN:            "ERRVIN123456789",
				Make:           "Test",
				Model:          "Fail",
				Year:           2025,
				Category:       domain.VehiclePrivateCar,
				OwnerCitizenID: ownerID,
			},
			repoErr: errors.New("insert error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFovesRepo{
				createVehicleFn: func(ctx context.Context, v *domain.Vehicle) error {
					return tt.repoErr
				},
			}
			svc := NewFovesService(repo, nil)
			v, err := svc.RegisterVehicle(context.Background(), tt.vehicle)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, v.ID)
			assert.False(t, v.IsStolen)
			assert.True(t, v.IsActive)
			assert.Equal(t, tt.vehicle.PlateNumber, v.PlateNumber)
		})
	}
}

func TestGetByPlate(t *testing.T) {
	tests := []struct {
		name    string
		plate   string
		repoRes *domain.Vehicle
		repoErr error
		wantErr bool
	}{
		{
			name:  "found",
			plate: "ABC-123",
			repoRes: &domain.Vehicle{
				ID: uuid.New(), PlateNumber: "ABC-123", Make: "Honda",
			},
		},
		{
			name:    "not found",
			plate:   "ZZZ-999",
			repoErr: errors.New("vehicle not found"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFovesRepo{
				findByPlateFn: func(ctx context.Context, plate string) (*domain.Vehicle, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewFovesService(repo, nil)
			v, err := svc.GetByPlate(context.Background(), tt.plate)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.repoRes.PlateNumber, v.PlateNumber)
		})
	}
}

func TestGetByVIN(t *testing.T) {
	tests := []struct {
		name    string
		vin     string
		repoRes *domain.Vehicle
		repoErr error
		wantErr bool
	}{
		{
			name: "found",
			vin:  "1HGCM82633A004352",
			repoRes: &domain.Vehicle{
				ID: uuid.New(), VIN: "1HGCM82633A004352", Make: "Ford",
			},
		},
		{
			name:    "not found",
			vin:     "UNKNOWNVIN12345",
			repoErr: errors.New("vehicle not found by vin"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFovesRepo{
				findByVINFn: func(ctx context.Context, vin string) (*domain.Vehicle, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewFovesService(repo, nil)
			v, err := svc.GetByVIN(context.Background(), tt.vin)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.repoRes.VIN, v.VIN)
		})
	}
}

func TestGetByOwner(t *testing.T) {
	citizenID := uuid.New()
	tests := []struct {
		name      string
		citizenID uuid.UUID
		repoRes   []domain.Vehicle
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "success",
			citizenID: citizenID,
			repoRes: []domain.Vehicle{
				{ID: uuid.New(), PlateNumber: "CAR-001", OwnerCitizenID: citizenID},
			},
		},
		{
			name:      "empty",
			citizenID: citizenID,
			repoRes:   []domain.Vehicle{},
		},
		{
			name:      "repo error",
			citizenID: citizenID,
			repoErr:   errors.New("query error"),
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFovesRepo{
				findByOwnerFn: func(ctx context.Context, citizenID uuid.UUID) ([]domain.Vehicle, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewFovesService(repo, nil)
			vehicles, err := svc.GetByOwner(context.Background(), tt.citizenID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, vehicles, len(tt.repoRes))
		})
	}
}

func TestTransferOwnership(t *testing.T) {
	vehicleID := uuid.New()
	fromID := uuid.New()
	toID := uuid.New()
	contractRef := "CTR-001"
	tests := []struct {
		name        string
		vehicleID   uuid.UUID
		fromID      uuid.UUID
		toID        uuid.UUID
		contractRef *string
		createErr   error
		updateErr   error
		wantErr     bool
	}{
		{
			name:        "success",
			vehicleID:   vehicleID,
			fromID:      fromID,
			toID:        toID,
			contractRef: &contractRef,
		},
		{
			name:      "create transfer error",
			vehicleID: vehicleID,
			fromID:    fromID,
			toID:      toID,
			createErr: errors.New("insert error"),
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFovesRepo{
				createTransferFn: func(ctx context.Context, transfer *domain.OwnershipTransfer) error {
					return tt.createErr
				},
				updateVehicleOwnerFn: func(ctx context.Context, vehicleID, newOwnerID uuid.UUID) error {
					return tt.updateErr
				},
			}
			svc := NewFovesService(repo, nil)
			transfer, err := svc.TransferOwnership(context.Background(), tt.vehicleID, tt.fromID, tt.toID, tt.contractRef)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.vehicleID, transfer.VehicleID)
			assert.Equal(t, tt.fromID, transfer.FromCitizenID)
			assert.Equal(t, tt.toID, transfer.ToCitizenID)
		})
	}
}

func TestIssueLicense(t *testing.T) {
	citizenID := uuid.New()
	tests := []struct {
		name    string
		license *domain.DriverLicense
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			license: &domain.DriverLicense{
				CitizenID:     citizenID,
				LicenseNumber: "LIC-001",
				CategoryB:     true,
				ExpiryDate:    time.Now().AddDate(5, 0, 0),
			},
		},
		{
			name: "repo error",
			license: &domain.DriverLicense{
				CitizenID:     citizenID,
				LicenseNumber: "LIC-ERR",
				ExpiryDate:    time.Now().AddDate(5, 0, 0),
			},
			repoErr: errors.New("insert error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFovesRepo{
				createLicenseFn: func(ctx context.Context, l *domain.DriverLicense) error {
					return tt.repoErr
				},
			}
			svc := NewFovesService(repo, nil)
			l, err := svc.IssueLicense(context.Background(), tt.license)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, l.ID)
			assert.Equal(t, tt.license.LicenseNumber, l.LicenseNumber)
		})
	}
}

func TestGetLicense(t *testing.T) {
	citizenID := uuid.New()
	tests := []struct {
		name      string
		citizenID uuid.UUID
		repoRes   *domain.DriverLicense
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "found",
			citizenID: citizenID,
			repoRes: &domain.DriverLicense{
				ID: uuid.New(), CitizenID: citizenID, LicenseNumber: "LIC-001",
			},
		},
		{
			name:      "not found",
			citizenID: citizenID,
			repoErr:   errors.New("license not found"),
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFovesRepo{
				findLicenseByCitizenFn: func(ctx context.Context, id uuid.UUID) (*domain.DriverLicense, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewFovesService(repo, nil)
			l, err := svc.GetLicense(context.Background(), tt.citizenID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.repoRes.LicenseNumber, l.LicenseNumber)
		})
	}
}
