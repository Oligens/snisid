package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snisid/civil-ht/internal/domain"
	"github.com/snisid/civil-ht/internal/repository"
)

type mockRepo struct {
	createBirthFunc         func(ctx context.Context, act *domain.CivilAct, birth *domain.BirthDeclaration) error
	createDeathFunc         func(ctx context.Context, act *domain.CivilAct, death *domain.DeathDeclaration) error
	createMarriageFunc      func(ctx context.Context, act *domain.CivilAct, marriage *domain.MarriageDeclaration) error
	findByActNumberFunc     func(ctx context.Context, actNumber string) (*domain.CivilAct, error)
	findByCitizenIDFunc     func(ctx context.Context, citizenID uuid.UUID) ([]domain.CivilAct, error)
	findBirthDetailsFunc    func(ctx context.Context, actID uuid.UUID) (*domain.BirthDeclaration, error)
	findDeathDetailsFunc    func(ctx context.Context, actID uuid.UUID) (*domain.DeathDeclaration, error)
	findMarriageDetailsFunc func(ctx context.Context, actID uuid.UUID) (*domain.MarriageDeclaration, error)
}

func (m *mockRepo) CreateBirth(ctx context.Context, act *domain.CivilAct, birth *domain.BirthDeclaration) error {
	return m.createBirthFunc(ctx, act, birth)
}
func (m *mockRepo) CreateDeath(ctx context.Context, act *domain.CivilAct, death *domain.DeathDeclaration) error {
	return m.createDeathFunc(ctx, act, death)
}
func (m *mockRepo) CreateMarriage(ctx context.Context, act *domain.CivilAct, marriage *domain.MarriageDeclaration) error {
	return m.createMarriageFunc(ctx, act, marriage)
}
func (m *mockRepo) FindByActNumber(ctx context.Context, actNumber string) (*domain.CivilAct, error) {
	return m.findByActNumberFunc(ctx, actNumber)
}
func (m *mockRepo) FindByCitizenID(ctx context.Context, citizenID uuid.UUID) ([]domain.CivilAct, error) {
	return m.findByCitizenIDFunc(ctx, citizenID)
}
func (m *mockRepo) FindBirthDetails(ctx context.Context, actID uuid.UUID) (*domain.BirthDeclaration, error) {
	return m.findBirthDetailsFunc(ctx, actID)
}
func (m *mockRepo) FindDeathDetails(ctx context.Context, actID uuid.UUID) (*domain.DeathDeclaration, error) {
	return m.findDeathDetailsFunc(ctx, actID)
}
func (m *mockRepo) FindMarriageDetails(ctx context.Context, actID uuid.UUID) (*domain.MarriageDeclaration, error) {
	return m.findMarriageDetailsFunc(ctx, actID)
}

func newTestService(repo repository.Repository) *CivilService {
	return NewCivilService(repo, nil)
}

func TestNewCivilService(t *testing.T) {
	svc := NewCivilService(nil, nil)
	require.NotNil(t, svc)
}

func TestDeclareBirth(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		birthReq    domain.BirthDeclaration
		act         domain.CivilAct
		repo        *mockRepo
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			birthReq: domain.BirthDeclaration{
				ChildFullName: "Marie Petit",
				ChildGender:   strPtr("F"),
			},
			act: domain.CivilAct{
				RegisteringOffice: "Mairie PAP",
				DeptCode:          "OU",
				Commune:           "Port-au-Prince",
				EventDate:         now,
				OfficerName:       strPtr("Officer Jean"),
			},
			repo: &mockRepo{
				createBirthFunc: func(ctx context.Context, act *domain.CivilAct, birth *domain.BirthDeclaration) error {
					return nil
				},
			},
		},
		{
			name: "repo error",
			birthReq: domain.BirthDeclaration{
				ChildFullName: "Fail Baby",
			},
			act: domain.CivilAct{
				RegisteringOffice: "Mairie PAP",
				DeptCode:          "OU",
				Commune:           "Port-au-Prince",
				EventDate:         now,
			},
			repo: &mockRepo{
				createBirthFunc: func(ctx context.Context, act *domain.CivilAct, birth *domain.BirthDeclaration) error {
					return errors.New("db error")
				},
			},
			wantErr:     true,
			errContains: "declare birth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			result, err := svc.DeclareBirth(context.Background(), tt.birthReq, tt.act)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, domain.ActBirth, result.ActType)
			assert.NotEmpty(t, result.ActID)
			assert.NotEmpty(t, result.ActNumber)
			assert.False(t, result.DeclaredDate.IsZero())
		})
	}
}

func TestDeclareDeath(t *testing.T) {
	now := time.Now()
	citizenID := uuid.New()

	tests := []struct {
		name        string
		deathReq    domain.DeathDeclaration
		act         domain.CivilAct
		repo        *mockRepo
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			deathReq: domain.DeathDeclaration{
				DeceasedCitizenID: citizenID,
				CauseOfDeath:      strPtr("Natural"),
			},
			act: domain.CivilAct{
				RegisteringOffice: "Mairie PAP",
				DeptCode:          "OU",
				Commune:           "Port-au-Prince",
				EventDate:         now,
			},
			repo: &mockRepo{
				createDeathFunc: func(ctx context.Context, act *domain.CivilAct, death *domain.DeathDeclaration) error {
					return nil
				},
			},
		},
		{
			name: "repo error",
			deathReq: domain.DeathDeclaration{
				DeceasedCitizenID: citizenID,
			},
			act: domain.CivilAct{
				RegisteringOffice: "Mairie PAP",
				DeptCode:          "OU",
				Commune:           "Port-au-Prince",
				EventDate:         now,
			},
			repo: &mockRepo{
				createDeathFunc: func(ctx context.Context, act *domain.CivilAct, death *domain.DeathDeclaration) error {
					return errors.New("db error")
				},
			},
			wantErr:     true,
			errContains: "declare death",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			result, err := svc.DeclareDeath(context.Background(), tt.deathReq, tt.act)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, domain.ActDeath, result.ActType)
			assert.NotEmpty(t, result.ActID)
		})
	}
}

func TestRegisterMarriage(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		marriageReq domain.MarriageDeclaration
		act         domain.CivilAct
		repo        *mockRepo
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			marriageReq: domain.MarriageDeclaration{
				SpouseACitizenID: uuid.New(),
				SpouseBCitizenID: uuid.New(),
				MarriageRegime:   strPtr("Community"),
			},
			act: domain.CivilAct{
				RegisteringOffice: "Mairie PAP",
				DeptCode:          "OU",
				Commune:           "Port-au-Prince",
				EventDate:         now,
			},
			repo: &mockRepo{
				createMarriageFunc: func(ctx context.Context, act *domain.CivilAct, marriage *domain.MarriageDeclaration) error {
					return nil
				},
			},
		},
		{
			name: "repo error",
			marriageReq: domain.MarriageDeclaration{
				SpouseACitizenID: uuid.New(),
				SpouseBCitizenID: uuid.New(),
			},
			act: domain.CivilAct{
				RegisteringOffice: "Mairie PAP",
				DeptCode:          "OU",
				Commune:           "Port-au-Prince",
				EventDate:         now,
			},
			repo: &mockRepo{
				createMarriageFunc: func(ctx context.Context, act *domain.CivilAct, marriage *domain.MarriageDeclaration) error {
					return errors.New("db error")
				},
			},
			wantErr:     true,
			errContains: "register marriage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			result, err := svc.RegisterMarriage(context.Background(), tt.marriageReq, tt.act)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, domain.ActMarriage, result.ActType)
			assert.NotEmpty(t, result.ActID)
		})
	}
}

func TestGetAct(t *testing.T) {
	act := &domain.CivilAct{ActID: uuid.New(), ActNumber: "ACTE-HT-2026-OU-B-000001"}

	tests := []struct {
		name      string
		actNumber string
		repo      *mockRepo
		want      *domain.CivilAct
		wantErr   bool
	}{
		{
			name:      "found",
			actNumber: "ACTE-HT-2026-OU-B-000001",
			repo: &mockRepo{
				findByActNumberFunc: func(ctx context.Context, an string) (*domain.CivilAct, error) { return act, nil },
			},
			want: act,
		},
		{
			name:      "not found",
			actNumber: "ACTE-HT-2026-XX-Z-999999",
			repo: &mockRepo{
				findByActNumberFunc: func(ctx context.Context, an string) (*domain.CivilAct, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.GetAct(context.Background(), tt.actNumber)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetCitizenActs(t *testing.T) {
	citizenID := uuid.New()
	acts := []domain.CivilAct{{ActID: uuid.New(), CitizenID: &citizenID}}

	tests := []struct {
		name      string
		citizenID string
		repo      *mockRepo
		want      []domain.CivilAct
		wantErr   bool
	}{
		{
			name:      "success",
			citizenID: citizenID.String(),
			repo: &mockRepo{
				findByCitizenIDFunc: func(ctx context.Context, cid uuid.UUID) ([]domain.CivilAct, error) { return acts, nil },
			},
			want: acts,
		},
		{
			name:      "invalid UUID",
			citizenID: "not-a-uuid",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.GetCitizenActs(context.Background(), tt.citizenID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetBirthDetails(t *testing.T) {
	actID := uuid.New()
	details := &domain.BirthDeclaration{ActID: actID, ChildFullName: "Marie Petit"}

	tests := []struct {
		name    string
		actID   uuid.UUID
		repo    *mockRepo
		want    *domain.BirthDeclaration
		wantErr bool
	}{
		{
			name:  "found",
			actID: actID,
			repo: &mockRepo{
				findBirthDetailsFunc: func(ctx context.Context, aid uuid.UUID) (*domain.BirthDeclaration, error) { return details, nil },
			},
			want: details,
		},
		{
			name:  "not found",
			actID: uuid.New(),
			repo: &mockRepo{
				findBirthDetailsFunc: func(ctx context.Context, aid uuid.UUID) (*domain.BirthDeclaration, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.GetBirthDetails(context.Background(), tt.actID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetDeathDetails(t *testing.T) {
	actID := uuid.New()
	details := &domain.DeathDeclaration{ActID: actID, CauseOfDeath: strPtr("Natural")}

	tests := []struct {
		name    string
		actID   uuid.UUID
		repo    *mockRepo
		want    *domain.DeathDeclaration
		wantErr bool
	}{
		{
			name:  "found",
			actID: actID,
			repo: &mockRepo{
				findDeathDetailsFunc: func(ctx context.Context, aid uuid.UUID) (*domain.DeathDeclaration, error) { return details, nil },
			},
			want: details,
		},
		{
			name:  "not found",
			actID: uuid.New(),
			repo: &mockRepo{
				findDeathDetailsFunc: func(ctx context.Context, aid uuid.UUID) (*domain.DeathDeclaration, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.GetDeathDetails(context.Background(), tt.actID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetMarriageDetails(t *testing.T) {
	actID := uuid.New()
	details := &domain.MarriageDeclaration{ActID: actID, SpouseACitizenID: uuid.New(), SpouseBCitizenID: uuid.New()}

	tests := []struct {
		name    string
		actID   uuid.UUID
		repo    *mockRepo
		want    *domain.MarriageDeclaration
		wantErr bool
	}{
		{
			name:  "found",
			actID: actID,
			repo: &mockRepo{
				findMarriageDetailsFunc: func(ctx context.Context, aid uuid.UUID) (*domain.MarriageDeclaration, error) { return details, nil },
			},
			want: details,
		},
		{
			name:  "not found",
			actID: uuid.New(),
			repo: &mockRepo{
				findMarriageDetailsFunc: func(ctx context.Context, aid uuid.UUID) (*domain.MarriageDeclaration, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.GetMarriageDetails(context.Background(), tt.actID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func strPtr(s string) *string { return &s }
