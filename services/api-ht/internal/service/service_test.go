package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/api-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAPIRepo struct {
	createAccountFn      func(ctx context.Context, acc *domain.DeveloperAccount) error
	findAccountByEmailFn func(ctx context.Context, email string) (*domain.DeveloperAccount, error)
	createAPIKeyFn       func(ctx context.Context, key *domain.APIKey) error
	findKeyByValueFn     func(ctx context.Context, keyValue string) (*domain.APIKey, error)
	findKeysByAccountFn  func(ctx context.Context, accountID uuid.UUID) ([]domain.APIKey, error)
	revokeKeyFn          func(ctx context.Context, id uuid.UUID) error
	findKeyByIDFn        func(ctx context.Context, id uuid.UUID) (*domain.APIKey, error)
	listCatalogFn        func(ctx context.Context) ([]domain.APIEndpoint, error)
	listUsageByKeyFn     func(ctx context.Context, keyID uuid.UUID) ([]domain.UsageLog, error)
	logUsageFn           func(ctx context.Context, log *domain.UsageLog) error
}

func (m *mockAPIRepo) CreateAccount(ctx context.Context, acc *domain.DeveloperAccount) error {
	return m.createAccountFn(ctx, acc)
}
func (m *mockAPIRepo) FindAccountByEmail(ctx context.Context, email string) (*domain.DeveloperAccount, error) {
	return m.findAccountByEmailFn(ctx, email)
}
func (m *mockAPIRepo) CreateAPIKey(ctx context.Context, key *domain.APIKey) error {
	return m.createAPIKeyFn(ctx, key)
}
func (m *mockAPIRepo) FindKeyByValue(ctx context.Context, keyValue string) (*domain.APIKey, error) {
	return m.findKeyByValueFn(ctx, keyValue)
}
func (m *mockAPIRepo) FindKeysByAccount(ctx context.Context, accountID uuid.UUID) ([]domain.APIKey, error) {
	return m.findKeysByAccountFn(ctx, accountID)
}
func (m *mockAPIRepo) RevokeKey(ctx context.Context, id uuid.UUID) error {
	return m.revokeKeyFn(ctx, id)
}
func (m *mockAPIRepo) FindKeyByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	return m.findKeyByIDFn(ctx, id)
}
func (m *mockAPIRepo) ListCatalog(ctx context.Context) ([]domain.APIEndpoint, error) {
	return m.listCatalogFn(ctx)
}
func (m *mockAPIRepo) ListUsageByKey(ctx context.Context, keyID uuid.UUID) ([]domain.UsageLog, error) {
	return m.listUsageByKeyFn(ctx, keyID)
}
func (m *mockAPIRepo) LogUsage(ctx context.Context, log *domain.UsageLog) error {
	return m.logUsageFn(ctx, log)
}

func TestRegisterDeveloper(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		contactName  string
		orgName      *string
		contactPhone *string
		findResult   *domain.DeveloperAccount
		findErr      error
		createErr    error
		wantErr      bool
	}{
		{
			name:        "success",
			email:       "dev@test.com",
			contactName: "John Doe",
		},
		{
			name:        "account already exists",
			email:       "existing@test.com",
			contactName: "Jane",
			findResult:  &domain.DeveloperAccount{Email: "existing@test.com"},
			wantErr:     true,
		},
		{
			name:        "create error",
			email:       "fail@test.com",
			contactName: "Bob",
			createErr:   errors.New("insert error"),
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAPIRepo{
				findAccountByEmailFn: func(ctx context.Context, email string) (*domain.DeveloperAccount, error) {
					return tt.findResult, tt.findErr
				},
				createAccountFn: func(ctx context.Context, acc *domain.DeveloperAccount) error {
					return tt.createErr
				},
			}
			svc := NewAPIService(repo, nil)
			acc, err := svc.RegisterDeveloper(context.Background(), tt.email, tt.contactName, tt.orgName, tt.contactPhone)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.email, acc.Email)
			assert.False(t, acc.IsApproved)
		})
	}
}

func TestRequestKey(t *testing.T) {
	aid := uuid.New()
	desc := "test key"
	tests := []struct {
		name        string
		accountID   uuid.UUID
		description *string
		createErr   error
		wantErr     bool
	}{
		{
			name:        "success",
			accountID:   aid,
			description: &desc,
		},
		{
			name:      "repo error",
			accountID: aid,
			createErr: errors.New("insert error"),
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAPIRepo{
				createAPIKeyFn: func(ctx context.Context, key *domain.APIKey) error {
					return tt.createErr
				},
			}
			svc := NewAPIService(repo, nil)
			key, err := svc.RequestKey(context.Background(), tt.accountID, tt.description)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.True(t, key.IsActive)
			assert.Equal(t, tt.accountID, key.AccountID)
			assert.Contains(t, key.KeyValue, "snisid_")
		})
	}
}

func TestGetCatalog(t *testing.T) {
	tests := []struct {
		name    string
		repoRes []domain.APIEndpoint
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			repoRes: []domain.APIEndpoint{
				{Path: "/v1/identities", Method: "GET"},
			},
		},
		{
			name:    "repo error",
			repoErr: errors.New("query error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAPIRepo{
				listCatalogFn: func(ctx context.Context) ([]domain.APIEndpoint, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewAPIService(repo, nil)
			got, err := svc.GetCatalog(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, len(tt.repoRes))
		})
	}
}

func TestGetUsage(t *testing.T) {
	kid := uuid.New()
	tests := []struct {
		name    string
		keyID   uuid.UUID
		repoRes []domain.UsageLog
		repoErr error
		wantErr bool
	}{
		{
			name:  "success",
			keyID: kid,
			repoRes: []domain.UsageLog{
				{Endpoint: "/v1/check", Status: 200, LatencyMs: 42},
			},
		},
		{
			name:    "repo error",
			keyID:   kid,
			repoErr: errors.New("query error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAPIRepo{
				listUsageByKeyFn: func(ctx context.Context, keyID uuid.UUID) ([]domain.UsageLog, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewAPIService(repo, nil)
			got, err := svc.GetUsage(context.Background(), tt.keyID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, len(tt.repoRes))
		})
	}
}

func TestRevokeKey(t *testing.T) {
	kid := uuid.New()
	tests := []struct {
		name    string
		keyID   uuid.UUID
		key     *domain.APIKey
		findErr error
		wantErr bool
	}{
		{
			name:  "success",
			keyID: kid,
			key:   &domain.APIKey{ID: kid, IsActive: true},
		},
		{
			name:    "key not found",
			keyID:   kid,
			findErr: errors.New("not found"),
			wantErr: true,
		},
		{
			name:    "already revoked",
			keyID:   kid,
			key:     &domain.APIKey{ID: kid, IsActive: false},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAPIRepo{
				findKeyByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
					return tt.key, tt.findErr
				},
				revokeKeyFn: func(ctx context.Context, id uuid.UUID) error {
					return nil
				},
			}
			svc := NewAPIService(repo, nil)
			err := svc.RevokeKey(context.Background(), tt.keyID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
