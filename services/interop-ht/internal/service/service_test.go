package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snisid/interop-ht/internal/domain"
	"github.com/snisid/interop-ht/internal/repository"
)

type mockRepo struct {
	createAgreementFunc func(ctx context.Context, a *domain.DataExchangeAgreement) error
	getAgreementFunc    func(ctx context.Context, id uuid.UUID) (*domain.DataExchangeAgreement, error)
	logExchangeFunc     func(ctx context.Context, l *domain.ExchangeLog) error
	getExchangeLogsFunc func(ctx context.Context, agreementID uuid.UUID) ([]domain.ExchangeLog, error)
}

func (m *mockRepo) CreateAgreement(ctx context.Context, a *domain.DataExchangeAgreement) error {
	return m.createAgreementFunc(ctx, a)
}
func (m *mockRepo) GetAgreement(ctx context.Context, id uuid.UUID) (*domain.DataExchangeAgreement, error) {
	return m.getAgreementFunc(ctx, id)
}
func (m *mockRepo) LogExchange(ctx context.Context, l *domain.ExchangeLog) error {
	return m.logExchangeFunc(ctx, l)
}
func (m *mockRepo) GetExchangeLogs(ctx context.Context, agreementID uuid.UUID) ([]domain.ExchangeLog, error) {
	return m.getExchangeLogsFunc(ctx, agreementID)
}

func newTestService(repo repository.Repository) *InteropService {
	return NewInteropService(repo, nil)
}

func TestNewInteropService(t *testing.T) {
	svc := NewInteropService(nil, nil)
	require.NotNil(t, svc)
}

func TestCreateAgreement(t *testing.T) {
	providerID := uuid.New()
	consumerID := uuid.New()

	tests := []struct {
		name    string
		req     domain.DataExchangeAgreement
		repo    *mockRepo
		wantErr bool
	}{
		{
			name: "success",
			req: domain.DataExchangeAgreement{
				ProviderAgencyID: providerID,
				ConsumerAgencyID: consumerID,
				ServiceName:      "identity-verify",
				AllowedFields:    []string{"full_name", "dob"},
				LegalBasis:       strPtr("LOI-2025-001"),
				RateLimitPerMin:  500,
			},
			repo: &mockRepo{
				createAgreementFunc: func(ctx context.Context, a *domain.DataExchangeAgreement) error { return nil },
			},
		},
		{
			name: "success with default rate limit",
			req: domain.DataExchangeAgreement{
				ProviderAgencyID: providerID,
				ConsumerAgencyID: consumerID,
				ServiceName:      "bio-verify",
			},
			repo: &mockRepo{
				createAgreementFunc: func(ctx context.Context, a *domain.DataExchangeAgreement) error { return nil },
			},
		},
		{
			name: "repo error",
			req: domain.DataExchangeAgreement{
				ProviderAgencyID: providerID,
				ConsumerAgencyID: consumerID,
				ServiceName:      "fail",
			},
			repo: &mockRepo{
				createAgreementFunc: func(ctx context.Context, a *domain.DataExchangeAgreement) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			result, err := svc.CreateAgreement(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.NotEmpty(t, result.AgreementID)
			assert.True(t, result.IsActive)
			assert.False(t, result.CreatedAt.IsZero())
			if tt.req.RateLimitPerMin == 0 {
				assert.Equal(t, 1000, result.RateLimitPerMin)
			} else {
				assert.Equal(t, tt.req.RateLimitPerMin, result.RateLimitPerMin)
			}
		})
	}
}

func TestExchange(t *testing.T) {
	svc := newTestService(&mockRepo{})
	err := svc.Exchange(context.Background(), uuid.New().String())
	assert.NoError(t, err)
}

func TestGetLogs(t *testing.T) {
	agreementID := uuid.New()
	logs := []domain.ExchangeLog{
		{LogID: uuid.New(), AgreementID: agreementID, StatusCode: 200, DurationMs: 45},
	}

	tests := []struct {
		name        string
		agreementID string
		repo        *mockRepo
		want        []domain.ExchangeLog
		wantErr     bool
	}{
		{
			name:        "success",
			agreementID: agreementID.String(),
			repo: &mockRepo{
				getExchangeLogsFunc: func(ctx context.Context, aid uuid.UUID) ([]domain.ExchangeLog, error) {
					return logs, nil
				},
			},
			want: logs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.GetLogs(context.Background(), tt.agreementID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func strPtr(s string) *string { return &s }
