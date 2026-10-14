package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/field-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockFieldRepo struct {
	createMissionFn     func(ctx context.Context, m *domain.Mission) error
	getActiveMissionsFn func(ctx context.Context) ([]domain.Mission, error)
	createMissionLogFn  func(ctx context.Context, l *domain.MissionLog) error
	getCoverageStatsFn  func(ctx context.Context) (*domain.CoverageStats, error)
}

func (m *mockFieldRepo) CreateMission(ctx context.Context, mission *domain.Mission) error {
	return m.createMissionFn(ctx, mission)
}
func (m *mockFieldRepo) GetActiveMissions(ctx context.Context) ([]domain.Mission, error) {
	return m.getActiveMissionsFn(ctx)
}
func (m *mockFieldRepo) CreateMissionLog(ctx context.Context, log *domain.MissionLog) error {
	return m.createMissionLogFn(ctx, log)
}
func (m *mockFieldRepo) GetCoverageStats(ctx context.Context) (*domain.CoverageStats, error) {
	return m.getCoverageStatsFn(ctx)
}

func TestCreateMission(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.CreateMissionRequest
		repoErr error
		wantErr bool
	}{
		{
			name: "success basic",
			req: domain.CreateMissionRequest{
				Title:    "Border Patrol Alpha",
				DeptCode: "DEPT-01",
			},
		},
		{
			name: "success with assigned unit",
			req: domain.CreateMissionRequest{
				Title:          "Intervention",
				Description:    "Test mission",
				AssignedUnitID: uuid.New().String(),
				DeptCode:       "DEPT-02",
			},
		},
		{
			name: "invalid assigned unit id",
			req: domain.CreateMissionRequest{
				Title:          "Bad Unit",
				AssignedUnitID: "not-a-uuid",
				DeptCode:       "DEPT-03",
			},
			wantErr: true,
		},
		{
			name: "repo error",
			req: domain.CreateMissionRequest{
				Title:    "Fail",
				DeptCode: "DEPT-99",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFieldRepo{
				createMissionFn: func(ctx context.Context, m *domain.Mission) error {
					return tt.repoErr
				},
			}
			svc := NewFieldService(repo, nil)
			m, err := svc.CreateMission(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, domain.MissionStatusPlanned, m.Status)
			assert.Equal(t, tt.req.Title, m.Title)
			assert.Equal(t, tt.req.DeptCode, m.DeptCode)
		})
	}
}

func TestGetActiveMissions(t *testing.T) {
	tests := []struct {
		name    string
		repoRes []domain.Mission
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			repoRes: []domain.Mission{
				{ID: uuid.New(), Title: "Mission A", Status: domain.MissionStatusInProgress},
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
			repo := &mockFieldRepo{
				getActiveMissionsFn: func(ctx context.Context) ([]domain.Mission, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewFieldService(repo, nil)
			got, err := svc.GetActiveMissions(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, len(tt.repoRes))
		})
	}
}

func TestCreateMissionLog(t *testing.T) {
	tests := []struct {
		name      string
		missionID string
		req       domain.CreateMissionLogRequest
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "success",
			missionID: uuid.New().String(),
			req: domain.CreateMissionLogRequest{
				LoggedBy:  uuid.New().String(),
				Action:    "PATROL_START",
				Latitude:  18.5,
				Longitude: -72.3,
			},
		},
		{
			name:      "invalid mission id",
			missionID: "bad-uuid",
			req: domain.CreateMissionLogRequest{
				LoggedBy: uuid.New().String(),
				Action:   "CHECK",
			},
			wantErr: true,
		},
		{
			name:      "invalid logged_by",
			missionID: uuid.New().String(),
			req: domain.CreateMissionLogRequest{
				LoggedBy: "bad-uuid",
				Action:   "CHECK",
			},
			wantErr: true,
		},
		{
			name:      "repo error",
			missionID: uuid.New().String(),
			req: domain.CreateMissionLogRequest{
				LoggedBy: uuid.New().String(),
				Action:   "LOG",
			},
			repoErr: errors.New("insert error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFieldRepo{
				createMissionLogFn: func(ctx context.Context, l *domain.MissionLog) error {
					return tt.repoErr
				},
			}
			svc := NewFieldService(repo, nil)
			got, err := svc.CreateMissionLog(context.Background(), tt.missionID, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, got.ID)
			assert.Equal(t, tt.req.Action, got.Action)
		})
	}
}

func TestGetCoverageStats(t *testing.T) {
	tests := []struct {
		name    string
		repoRes *domain.CoverageStats
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			repoRes: &domain.CoverageStats{
				TotalMissions:    5,
				ActiveUnits:      3,
				CoverageRadiusKm: 50.0,
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
			repo := &mockFieldRepo{
				getCoverageStatsFn: func(ctx context.Context) (*domain.CoverageStats, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewFieldService(repo, nil)
			got, err := svc.GetCoverageStats(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.repoRes.TotalMissions, got.TotalMissions)
		})
	}
}
