package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snisid/idcore-svc/internal/domain"
	"github.com/snisid/idcore-svc/internal/milvus"
	"github.com/snisid/idcore-svc/internal/nin"
	"github.com/snisid/idcore-svc/internal/repository"
)

type mockRepo struct {
	createFunc                 func(ctx context.Context, citizen *domain.Citizen) error
	findByNINFunc              func(ctx context.Context, nin string) (*domain.Citizen, error)
	findByIDFunc               func(ctx context.Context, id uuid.UUID) (*domain.Citizen, error)
	updateFunc                 func(ctx context.Context, citizen *domain.Citizen) error
	updateStatusFunc           func(ctx context.Context, nin string, status domain.IDStatus, reason string, authorizedBy uuid.UUID) error
	findDemographicMatchesFunc func(ctx context.Context, fullName string, dob time.Time) ([]domain.DemographicMatch, error)
	createDedupCandidateFunc   func(ctx context.Context, candidate domain.DedupCandidate) error
	getHistoryFunc             func(ctx context.Context, citizenID uuid.UUID) ([]domain.ChangeHistory, error)
	getPopulationStatsFunc     func(ctx context.Context) (*domain.PopulationStats, error)
}

func (m *mockRepo) Create(ctx context.Context, citizen *domain.Citizen) error {
	return m.createFunc(ctx, citizen)
}
func (m *mockRepo) FindByNIN(ctx context.Context, nin string) (*domain.Citizen, error) {
	return m.findByNINFunc(ctx, nin)
}
func (m *mockRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Citizen, error) {
	return m.findByIDFunc(ctx, id)
}
func (m *mockRepo) Update(ctx context.Context, citizen *domain.Citizen) error {
	return m.updateFunc(ctx, citizen)
}
func (m *mockRepo) UpdateStatus(ctx context.Context, nin string, status domain.IDStatus, reason string, authorizedBy uuid.UUID) error {
	return m.updateStatusFunc(ctx, nin, status, reason, authorizedBy)
}
func (m *mockRepo) FindDemographicMatches(ctx context.Context, fullName string, dob time.Time) ([]domain.DemographicMatch, error) {
	return m.findDemographicMatchesFunc(ctx, fullName, dob)
}
func (m *mockRepo) CreateDedupCandidate(ctx context.Context, candidate domain.DedupCandidate) error {
	return m.createDedupCandidateFunc(ctx, candidate)
}
func (m *mockRepo) GetHistory(ctx context.Context, citizenID uuid.UUID) ([]domain.ChangeHistory, error) {
	return m.getHistoryFunc(ctx, citizenID)
}
func (m *mockRepo) GetPopulationStats(ctx context.Context) (*domain.PopulationStats, error) {
	return m.getPopulationStatsFunc(ctx)
}

func newTestService(repo repository.Repository) *IdentityService {
	milvusClient, _ := milvus.NewClient("mock:0")
	ninGen := nin.NewGenerator(nil)
	return NewIdentityService(repo, milvusClient, nil, ninGen, "0.95", "0.85")
}

func TestNewIdentityService(t *testing.T) {
	milvusClient, _ := milvus.NewClient("mock:0")
	ninGen := nin.NewGenerator(nil)
	svc := NewIdentityService(nil, milvusClient, nil, ninGen, "0.95", "0.85")
	require.NotNil(t, svc)
	assert.Equal(t, 0.95, svc.bioThreshold)
	assert.Equal(t, 0.85, svc.demoThreshold)
}

func TestNewIdentityService_Defaults(t *testing.T) {
	milvusClient, _ := milvus.NewClient("mock:0")
	ninGen := nin.NewGenerator(nil)
	svc := NewIdentityService(nil, milvusClient, nil, ninGen, "", "")
	require.NotNil(t, svc)
	assert.Equal(t, 0.95, svc.bioThreshold)
	assert.Equal(t, 0.85, svc.demoThreshold)
}

func TestEnrollCitizen(t *testing.T) {
	t.Parallel()

	now := time.Now()
	oldDOB := now.AddDate(-20, 0, 0)
	youngDOB := now.AddDate(-1, 0, 0)

	tests := []struct {
		name        string
		req         domain.EnrollmentRequest
		repo        *mockRepo
		wantErr     bool
		errContains string
	}{
		{
			name: "success adult enrollment",
			req: domain.EnrollmentRequest{
				Age:             20,
				EnrollmentType:  domain.EnrollmentAdultFirst,
				FullNameLegal:   "Jean Dupont",
				FirstName:       "Jean",
				LastName:        "Dupont",
				DOB:             oldDOB,
				Nationality:     "HTI",
				DeptCode:        "OU",
				CreatedBy:       uuid.New().String(),
				BiometricSample: []byte{0x01, 0x02},
			},
			repo: &mockRepo{
				createFunc: func(ctx context.Context, c *domain.Citizen) error { return nil },
				findDemographicMatchesFunc: func(ctx context.Context, fullName string, dob time.Time) ([]domain.DemographicMatch, error) {
					return nil, nil
				},
				createDedupCandidateFunc: func(ctx context.Context, c domain.DedupCandidate) error { return nil },
				updateFunc:               func(ctx context.Context, c *domain.Citizen) error { return nil },
			},
		},
		{
			name: "success child enrollment skips biometric",
			req: domain.EnrollmentRequest{
				Age:            1,
				EnrollmentType: domain.EnrollmentBirth,
				FullNameLegal:  "Marie Petit",
				FirstName:      "Marie",
				LastName:       "Petit",
				DOB:            youngDOB,
				Nationality:    "HTI",
				DeptCode:       "ND",
				CreatedBy:      uuid.New().String(),
			},
			repo: &mockRepo{
				createFunc: func(ctx context.Context, c *domain.Citizen) error { return nil },
				findDemographicMatchesFunc: func(ctx context.Context, fullName string, dob time.Time) ([]domain.DemographicMatch, error) {
					return nil, nil
				},
				createDedupCandidateFunc: func(ctx context.Context, c domain.DedupCandidate) error { return nil },
			},
		},
		{
			name: "invalid dept code for NIN",
			req: domain.EnrollmentRequest{
				Age:            20,
				EnrollmentType: domain.EnrollmentAdultFirst,
				FullNameLegal:  "Bad Dept",
				FirstName:      "Bad",
				LastName:       "Dept",
				DOB:            oldDOB,
				Nationality:    "HTI",
				DeptCode:       "XX",
				CreatedBy:      uuid.New().String(),
			},
			repo: &mockRepo{
				createFunc: func(ctx context.Context, c *domain.Citizen) error { return nil },
				findDemographicMatchesFunc: func(ctx context.Context, fullName string, dob time.Time) ([]domain.DemographicMatch, error) {
					return nil, nil
				},
				createDedupCandidateFunc: func(ctx context.Context, c domain.DedupCandidate) error { return nil },
			},
			wantErr:     true,
			errContains: "NIN generation",
		},
		{
			name: "repo create fails",
			req: domain.EnrollmentRequest{
				Age:            20,
				EnrollmentType: domain.EnrollmentAdultFirst,
				FullNameLegal:  "Fail Create",
				FirstName:      "Fail",
				LastName:       "Create",
				DOB:            oldDOB,
				Nationality:    "HTI",
				DeptCode:       "OU",
				CreatedBy:      uuid.New().String(),
			},
			repo: &mockRepo{
				createFunc: func(ctx context.Context, c *domain.Citizen) error { return errors.New("db error") },
				findDemographicMatchesFunc: func(ctx context.Context, fullName string, dob time.Time) ([]domain.DemographicMatch, error) {
					return nil, nil
				},
				createDedupCandidateFunc: func(ctx context.Context, c domain.DedupCandidate) error { return nil },
			},
			wantErr:     true,
			errContains: "create citizen",
		},
		{
			name: "demographic match logged as dedup candidate",
			req: domain.EnrollmentRequest{
				Age:            20,
				EnrollmentType: domain.EnrollmentAdultFirst,
				FullNameLegal:  "Match Test",
				FirstName:      "Match",
				LastName:       "Test",
				DOB:            oldDOB,
				Nationality:    "HTI",
				DeptCode:       "OU",
				CreatedBy:      uuid.New().String(),
			},
			repo: &mockRepo{
				createFunc: func(ctx context.Context, c *domain.Citizen) error { return nil },
				findDemographicMatchesFunc: func(ctx context.Context, fullName string, dob time.Time) ([]domain.DemographicMatch, error) {
					return []domain.DemographicMatch{{CitizenID: uuid.New(), Score: 0.95}}, nil
				},
				createDedupCandidateFunc: func(ctx context.Context, c domain.DedupCandidate) error { return nil },
				updateFunc:               func(ctx context.Context, c *domain.Citizen) error { return nil },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			result, err := svc.EnrollCitizen(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.NotEmpty(t, result.NIN)
			require.NotNil(t, result.Citizen)
			assert.Equal(t, domain.StatusActive, result.Citizen.Status)
			assert.Equal(t, tt.req.FullNameLegal, result.Citizen.FullNameLegal)
		})
	}
}

func TestVerifyIdentity(t *testing.T) {
	citizen := &domain.Citizen{CitizenID: uuid.New(), NIN: "HTI-OU-20-000001", FullNameLegal: "Jean Dupont"}
	tests := []struct {
		name    string
		nin     string
		repo    *mockRepo
		want    *domain.Citizen
		wantErr bool
	}{
		{
			name: "found",
			nin:  "HTI-OU-20-000001",
			repo: &mockRepo{
				findByNINFunc: func(ctx context.Context, nin string) (*domain.Citizen, error) { return citizen, nil },
			},
			want: citizen,
		},
		{
			name: "not found",
			nin:  "HTI-OU-20-999999",
			repo: &mockRepo{
				findByNINFunc: func(ctx context.Context, nin string) (*domain.Citizen, error) { return nil, errors.New("not found") },
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.VerifyIdentity(context.Background(), tt.nin)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetCitizen(t *testing.T) {
	citizen := &domain.Citizen{CitizenID: uuid.New(), NIN: "HTI-OU-20-000001"}

	tests := []struct {
		name    string
		id      string
		repo    *mockRepo
		want    *domain.Citizen
		wantErr bool
	}{
		{
			name: "by UUID",
			id:   uuid.New().String(),
			repo: &mockRepo{
				findByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Citizen, error) { return citizen, nil },
			},
			want: citizen,
		},
		{
			name: "by NIN when UUID parse fails",
			id:   "HTI-OU-20-000001",
			repo: &mockRepo{
				findByNINFunc: func(ctx context.Context, nin string) (*domain.Citizen, error) { return citizen, nil },
			},
			want: citizen,
		},
		{
			name: "not found",
			id:   uuid.New().String(),
			repo: &mockRepo{
				findByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Citizen, error) { return nil, errors.New("not found") },
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.GetCitizen(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUpdateStatus(t *testing.T) {
	tests := []struct {
		name    string
		nin     string
		status  domain.IDStatus
		repo    *mockRepo
		wantErr bool
	}{
		{
			name:   "success",
			nin:    "HTI-OU-20-000001",
			status: domain.StatusSuspended,
			repo: &mockRepo{
				updateStatusFunc: func(ctx context.Context, nin string, status domain.IDStatus, reason string, authorizedBy uuid.UUID) error {
					return nil
				},
			},
		},
		{
			name:   "repo error",
			nin:    "HTI-OU-20-000001",
			status: domain.StatusCancelled,
			repo: &mockRepo{
				updateStatusFunc: func(ctx context.Context, nin string, status domain.IDStatus, reason string, authorizedBy uuid.UUID) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			err := svc.UpdateStatus(context.Background(), tt.nin, tt.status, "test reason", uuid.New().String())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGetHistory(t *testing.T) {
	citizenID := uuid.New()
	history := []domain.ChangeHistory{
		{HistoryID: uuid.New(), CitizenID: citizenID, FieldChanged: "status", ChangeReason: "test"},
	}

	tests := []struct {
		name    string
		id      string
		repo    *mockRepo
		want    []domain.ChangeHistory
		wantErr bool
	}{
		{
			name: "by UUID",
			id:   citizenID.String(),
			repo: &mockRepo{
				getHistoryFunc: func(ctx context.Context, cid uuid.UUID) ([]domain.ChangeHistory, error) { return history, nil },
			},
			want: history,
		},
		{
			name: "by NIN",
			id:   "HTI-OU-20-000001",
			repo: &mockRepo{
				findByNINFunc: func(ctx context.Context, nin string) (*domain.Citizen, error) {
					return &domain.Citizen{CitizenID: citizenID}, nil
				},
				getHistoryFunc: func(ctx context.Context, cid uuid.UUID) ([]domain.ChangeHistory, error) { return history, nil },
			},
			want: history,
		},
		{
			name: "citizen not found via NIN",
			id:   "HTI-OU-20-999999",
			repo: &mockRepo{
				findByNINFunc: func(ctx context.Context, nin string) (*domain.Citizen, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.GetHistory(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetPopulationStats(t *testing.T) {
	stats := &domain.PopulationStats{Total: 1000, Active: 800}

	tests := []struct {
		name    string
		repo    *mockRepo
		want    *domain.PopulationStats
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockRepo{
				getPopulationStatsFunc: func(ctx context.Context) (*domain.PopulationStats, error) { return stats, nil },
			},
			want: stats,
		},
		{
			name: "repo error",
			repo: &mockRepo{
				getPopulationStatsFunc: func(ctx context.Context) (*domain.PopulationStats, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.GetPopulationStats(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSearchCitizens_NotImplemented(t *testing.T) {
	svc := newTestService(&mockRepo{})
	result, err := svc.SearchCitizens(context.Background(), "test")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestResolveDedup_NotImplemented(t *testing.T) {
	svc := newTestService(&mockRepo{})
	err := svc.ResolveDedup(context.Background(), "id", "merge", "reviewer")
	assert.Error(t, err)
}
