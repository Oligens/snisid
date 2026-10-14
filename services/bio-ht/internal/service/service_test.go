package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snisid/bio-ht/internal/domain"
	"github.com/snisid/bio-ht/internal/milvus"
	"github.com/snisid/bio-ht/internal/repository"
)

type mockRepo struct {
	createTemplateFunc              func(ctx context.Context, t *domain.BioTemplate) error
	getTemplateFunc                 func(ctx context.Context, templateID uuid.UUID) (*domain.BioTemplate, error)
	getActiveTemplatesByCitizenFunc func(ctx context.Context, citizenID uuid.UUID) ([]domain.BioTemplate, error)
	deactivateTemplateFunc          func(ctx context.Context, templateID uuid.UUID) error
	logVerificationFunc             func(ctx context.Context, l *domain.VerificationLog) error
}

func (m *mockRepo) CreateTemplate(ctx context.Context, t *domain.BioTemplate) error {
	return m.createTemplateFunc(ctx, t)
}
func (m *mockRepo) GetTemplate(ctx context.Context, templateID uuid.UUID) (*domain.BioTemplate, error) {
	return m.getTemplateFunc(ctx, templateID)
}
func (m *mockRepo) GetActiveTemplatesByCitizen(ctx context.Context, citizenID uuid.UUID) ([]domain.BioTemplate, error) {
	return m.getActiveTemplatesByCitizenFunc(ctx, citizenID)
}
func (m *mockRepo) DeactivateTemplate(ctx context.Context, templateID uuid.UUID) error {
	return m.deactivateTemplateFunc(ctx, templateID)
}
func (m *mockRepo) LogVerification(ctx context.Context, l *domain.VerificationLog) error {
	return m.logVerificationFunc(ctx, l)
}

func newTestService(repo repository.Repository) *BioService {
	milvusClient, _ := milvus.NewClient("mock:0")
	return NewBioService(repo, milvusClient, nil)
}

func TestNewBioService(t *testing.T) {
	milvusClient, _ := milvus.NewClient("mock:0")
	svc := NewBioService(nil, milvusClient, nil)
	require.NotNil(t, svc)
}

func TestEnroll(t *testing.T) {
	citizenID := uuid.New()

	tests := []struct {
		name        string
		req         domain.EnrollRequest
		repo        *mockRepo
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			req: domain.EnrollRequest{
				CitizenID:       citizenID.String(),
				Modality:        "FINGERPRINT",
				ImageData:       []byte{0x01, 0x02},
				CaptureDevice:   "ScannerX",
				CaptureLocation: "Bureau PAP",
				CapturedBy:      uuid.New().String(),
			},
			repo: &mockRepo{
				createTemplateFunc: func(ctx context.Context, t *domain.BioTemplate) error { return nil },
			},
		},
		{
			name: "invalid citizen_id",
			req: domain.EnrollRequest{
				CitizenID:  "not-a-uuid",
				Modality:   "FINGERPRINT",
				CapturedBy: uuid.New().String(),
			},
			wantErr:     true,
			errContains: "invalid citizen_id",
		},
		{
			name: "invalid captured_by",
			req: domain.EnrollRequest{
				CitizenID:  citizenID.String(),
				Modality:   "FINGERPRINT",
				CapturedBy: "not-a-uuid",
			},
			wantErr:     true,
			errContains: "invalid captured_by",
		},
		{
			name: "repo error",
			req: domain.EnrollRequest{
				CitizenID:  citizenID.String(),
				Modality:   "FINGERPRINT",
				CapturedBy: uuid.New().String(),
			},
			repo: &mockRepo{
				createTemplateFunc: func(ctx context.Context, t *domain.BioTemplate) error {
					return errors.New("db error")
				},
			},
			wantErr:     true,
			errContains: "save template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			result, err := svc.Enroll(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, domain.Modality("FINGERPRINT"), result.Modality)
			assert.True(t, result.IsActive)
			assert.Equal(t, 0.85, result.QualityScore)
			assert.NotEmpty(t, result.TemplateID)
			assert.NotEmpty(t, result.MilvusVectorID)
		})
	}
}

func TestVerify(t *testing.T) {
	citizenID := uuid.New()

	tests := []struct {
		name        string
		req         domain.VerifyRequest
		repo        *mockRepo
		wantErr     bool
		errContains string
	}{
		{
			name: "success match found",
			req: domain.VerifyRequest{
				CitizenID:  citizenID.String(),
				Modality:   "FACE",
				SampleData: []byte{0x01},
			},
			repo: &mockRepo{
				logVerificationFunc: func(ctx context.Context, l *domain.VerificationLog) error { return nil },
			},
		},
		{
			name: "invalid citizen_id",
			req: domain.VerifyRequest{
				CitizenID: "bad-uuid",
				Modality:  "FACE",
			},
			wantErr:     true,
			errContains: "invalid citizen_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			result, err := svc.Verify(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.True(t, result.IsMatch)
			assert.Equal(t, 0.95, result.Score)
		})
	}
}

func TestIdentify(t *testing.T) {
	tests := []struct {
		name        string
		req         domain.IdentifyRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "success no candidates",
			req: domain.IdentifyRequest{
				Modality:   "FINGERPRINT",
				SampleData: []byte{0x01},
				Threshold:  0.85,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockRepo{})
			result, err := svc.Identify(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Empty(t, result.Candidates)
		})
	}
}

func TestGetQuality(t *testing.T) {
	templateID := uuid.New()

	tests := []struct {
		name       string
		templateID string
		repo       *mockRepo
		want       float64
		wantErr    bool
	}{
		{
			name:       "success",
			templateID: templateID.String(),
			repo: &mockRepo{
				getTemplateFunc: func(ctx context.Context, tid uuid.UUID) (*domain.BioTemplate, error) {
					return &domain.BioTemplate{TemplateID: tid, QualityScore: 0.92}, nil
				},
			},
			want: 0.92,
		},
		{
			name:       "invalid template ID",
			templateID: "bad-uuid",
			wantErr:    true,
		},
		{
			name:       "template not found",
			templateID: templateID.String(),
			repo: &mockRepo{
				getTemplateFunc: func(ctx context.Context, tid uuid.UUID) (*domain.BioTemplate, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.GetQuality(context.Background(), tt.templateID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
