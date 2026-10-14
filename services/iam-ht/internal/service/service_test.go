package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snisid/iam-ht/internal/domain"
	"github.com/snisid/iam-ht/internal/repository"
)

type mockRepo struct {
	upsertAssuranceFunc func(ctx context.Context, a *domain.IdentityAssurance) error
	getAssuranceFunc    func(ctx context.Context, citizenID uuid.UUID) (*domain.IdentityAssurance, error)
	createClientFunc    func(ctx context.Context, c *domain.AgencyClient) error
	getClientFunc       func(ctx context.Context, oauthClientID string) (*domain.AgencyClient, error)
	logAccessFunc       func(ctx context.Context, l *domain.AccessLog) error
}

func (m *mockRepo) UpsertAssurance(ctx context.Context, a *domain.IdentityAssurance) error {
	return m.upsertAssuranceFunc(ctx, a)
}
func (m *mockRepo) GetAssurance(ctx context.Context, citizenID uuid.UUID) (*domain.IdentityAssurance, error) {
	return m.getAssuranceFunc(ctx, citizenID)
}
func (m *mockRepo) CreateClient(ctx context.Context, c *domain.AgencyClient) error {
	return m.createClientFunc(ctx, c)
}
func (m *mockRepo) GetClient(ctx context.Context, oauthClientID string) (*domain.AgencyClient, error) {
	return m.getClientFunc(ctx, oauthClientID)
}
func (m *mockRepo) LogAccess(ctx context.Context, l *domain.AccessLog) error {
	return m.logAccessFunc(ctx, l)
}

func newTestService(repo repository.Repository) *IAMService {
	return NewIAMService(repo, nil)
}

func TestNewIAMService(t *testing.T) {
	svc := NewIAMService(nil, nil)
	require.NotNil(t, svc)
}

func TestAuthorize(t *testing.T) {
	citizenID := uuid.New()
	assurance := &domain.IdentityAssurance{
		AssuranceID:    uuid.New(),
		CitizenID:      citizenID,
		AssuranceLevel: domain.IAL2BiometricVerified,
	}

	tests := []struct {
		name      string
		citizenID string
		repo      *mockRepo
		want      *domain.IdentityAssurance
		wantErr   bool
	}{
		{
			name:      "found",
			citizenID: citizenID.String(),
			repo: &mockRepo{
				getAssuranceFunc: func(ctx context.Context, cid uuid.UUID) (*domain.IdentityAssurance, error) {
					return assurance, nil
				},
			},
			want: assurance,
		},
		{
			name:      "not found",
			citizenID: uuid.New().String(),
			repo: &mockRepo{
				getAssuranceFunc: func(ctx context.Context, cid uuid.UUID) (*domain.IdentityAssurance, error) {
					return nil, errors.New("assurance not found")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.Authorize(context.Background(), tt.citizenID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStepUpAssurance(t *testing.T) {
	citizenID := uuid.New()
	existing := &domain.IdentityAssurance{
		AssuranceID:    uuid.New(),
		CitizenID:      citizenID,
		KeycloakUserID: "kc-" + citizenID.String(),
		AssuranceLevel: domain.IAL1SelfAsserted,
		MFAEnrolled:    false,
	}

	tests := []struct {
		name        string
		citizenID   string
		repo        *mockRepo
		wantLevel   domain.AssuranceLevel
		wantErr     bool
		errContains string
	}{
		{
			name:      "success",
			citizenID: citizenID.String(),
			repo: &mockRepo{
				getAssuranceFunc: func(ctx context.Context, cid uuid.UUID) (*domain.IdentityAssurance, error) {
					return existing, nil
				},
				upsertAssuranceFunc: func(ctx context.Context, a *domain.IdentityAssurance) error {
					return nil
				},
			},
			wantLevel: domain.IAL2BiometricVerified,
		},
		{
			name:      "identity not found",
			citizenID: uuid.New().String(),
			repo: &mockRepo{
				getAssuranceFunc: func(ctx context.Context, cid uuid.UUID) (*domain.IdentityAssurance, error) {
					return nil, errors.New("assurance not found")
				},
			},
			wantErr:     true,
			errContains: "identity not found",
		},
		{
			name:      "upsert fails",
			citizenID: citizenID.String(),
			repo: &mockRepo{
				getAssuranceFunc: func(ctx context.Context, cid uuid.UUID) (*domain.IdentityAssurance, error) {
					return existing, nil
				},
				upsertAssuranceFunc: func(ctx context.Context, a *domain.IdentityAssurance) error {
					return errors.New("db error")
				},
			},
			wantErr:     true,
			errContains: "step-up failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			result, err := svc.StepUpAssurance(context.Background(), tt.citizenID)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.wantLevel, result.AssuranceLevel)
			require.NotNil(t, result.BiometricVerifiedAt)
			assert.WithinDuration(t, time.Now(), *result.BiometricVerifiedAt, time.Second)
		})
	}
}
