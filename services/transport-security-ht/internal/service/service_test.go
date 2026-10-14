package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/transport-security-ht/internal/domain"
	"github.com/snisid/transport-security-ht/internal/kafka"
	"github.com/snisid/transport-security-ht/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTransportRepo struct {
	repository.TransportRepository
	createScreeningFn     func(ctx context.Context, s *domain.PassengerScreening) error
	getRecentScreeningsFn func(ctx context.Context, limit int) ([]domain.PassengerScreening, error)
	addNoFlyFn            func(ctx context.Context, p *domain.NoFlyPassenger) error
	checkNoFlyFn          func(ctx context.Context, identityRef string) (*domain.NoFlyPassenger, error)
	getZonesByAirportFn   func(ctx context.Context, airportCode string) ([]domain.AirportSecurityZone, error)
	reportZoneBreachFn    func(ctx context.Context, zoneID uuid.UUID) error
}

func (m *mockTransportRepo) CreateScreening(ctx context.Context, s *domain.PassengerScreening) error {
	return m.createScreeningFn(ctx, s)
}
func (m *mockTransportRepo) GetRecentScreenings(ctx context.Context, limit int) ([]domain.PassengerScreening, error) {
	return m.getRecentScreeningsFn(ctx, limit)
}
func (m *mockTransportRepo) AddNoFly(ctx context.Context, p *domain.NoFlyPassenger) error {
	return m.addNoFlyFn(ctx, p)
}
func (m *mockTransportRepo) CheckNoFly(ctx context.Context, identityRef string) (*domain.NoFlyPassenger, error) {
	return m.checkNoFlyFn(ctx, identityRef)
}
func (m *mockTransportRepo) GetZonesByAirport(ctx context.Context, airportCode string) ([]domain.AirportSecurityZone, error) {
	return m.getZonesByAirportFn(ctx, airportCode)
}
func (m *mockTransportRepo) ReportZoneBreach(ctx context.Context, zoneID uuid.UUID) error {
	return m.reportZoneBreachFn(ctx, zoneID)
}

type mockKafka struct {
	kafka.Producer
	publishFn func(ctx context.Context, key string, msg interface{}) error
}

func (m *mockKafka) Publish(ctx context.Context, key string, msg interface{}) error {
	if m.publishFn != nil {
		return m.publishFn(ctx, key, msg)
	}
	return nil
}

func TestLogScreening(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{name: "success", repoErr: nil, wantErr: false},
		{name: "repo error", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTransportRepo{
				createScreeningFn: func(_ context.Context, _ *domain.PassengerScreening) error {
					return tt.repoErr
				},
			}
			svc := NewTransportService(repo, &mockKafka{})
			s := &domain.PassengerScreening{TravelerIdentityRef: "T123"}
			err := svc.LogScreening(context.Background(), s)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, s.ScreeningID)
		})
	}
}

func TestGetRecentScreenings(t *testing.T) {
	tests := []struct {
		name    string
		items   []domain.PassengerScreening
		repoErr error
		wantErr bool
	}{
		{name: "success", items: []domain.PassengerScreening{{TravelerIdentityRef: "T1"}}, wantErr: false},
		{name: "repo error", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTransportRepo{
				getRecentScreeningsFn: func(_ context.Context, _ int) ([]domain.PassengerScreening, error) {
					return tt.items, tt.repoErr
				},
			}
			svc := NewTransportService(repo, &mockKafka{})
			result, err := svc.GetRecentScreenings(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.items, result)
		})
	}
}

func TestAddNoFlyEntry(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{name: "success", wantErr: false},
		{name: "repo error", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTransportRepo{
				addNoFlyFn: func(_ context.Context, _ *domain.NoFlyPassenger) error {
					return tt.repoErr
				},
			}
			svc := NewTransportService(repo, &mockKafka{})
			p := &domain.NoFlyPassenger{IdentityRef: "T123"}
			err := svc.AddNoFlyEntry(context.Background(), p)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestCheckNoFly(t *testing.T) {
	entry := &domain.NoFlyPassenger{IdentityRef: "T123"}
	tests := []struct {
		name    string
		result  *domain.NoFlyPassenger
		repoErr error
		wantErr bool
	}{
		{name: "found", result: entry, wantErr: false},
		{name: "not found", result: nil, wantErr: false},
		{name: "repo error", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTransportRepo{
				checkNoFlyFn: func(_ context.Context, _ string) (*domain.NoFlyPassenger, error) {
					return tt.result, tt.repoErr
				},
			}
			svc := NewTransportService(repo, &mockKafka{})
			result, err := svc.CheckNoFly(context.Background(), "T123")
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.result, result)
		})
	}
}

func TestGetZonesByAirport(t *testing.T) {
	zones := []domain.AirportSecurityZone{{AirportCode: "JFK"}}
	tests := []struct {
		name    string
		result  []domain.AirportSecurityZone
		repoErr error
		wantErr bool
	}{
		{name: "success", result: zones, wantErr: false},
		{name: "repo error", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTransportRepo{
				getZonesByAirportFn: func(_ context.Context, _ string) ([]domain.AirportSecurityZone, error) {
					return tt.result, tt.repoErr
				},
			}
			svc := NewTransportService(repo, &mockKafka{})
			result, err := svc.GetZonesByAirport(context.Background(), "JFK")
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.result, result)
		})
	}
}

func TestReportBreach(t *testing.T) {
	zoneID := uuid.New()
	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{name: "success", wantErr: false},
		{name: "repo error", repoErr: errors.New("not found"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTransportRepo{
				reportZoneBreachFn: func(_ context.Context, _ uuid.UUID) error {
					return tt.repoErr
				},
			}
			svc := NewTransportService(repo, &mockKafka{})
			err := svc.ReportBreach(context.Background(), zoneID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestNewTransportService(t *testing.T) {
	svc := NewTransportService(&mockTransportRepo{}, &mockKafka{})
	require.NotNil(t, svc)
}
