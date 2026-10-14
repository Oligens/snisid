package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snisid/pki-bridge-svc/internal/domain"
	"github.com/snisid/pki-bridge-svc/internal/repository"
)

type mockPKIRepo struct {
	createForeignCAFunc        func(ctx context.Context, ca *domain.ForeignCA) error
	createCrossCertFunc        func(ctx context.Context, cert *domain.CrossCertificate) error
	findCrossCertBySubjectFunc func(ctx context.Context, subject string) (*domain.CrossCertificate, error)
	listTrustAnchorsFunc       func(ctx context.Context) ([]domain.TrustAnchor, error)
	savePathValidationFunc     func(ctx context.Context, pv *domain.PathValidation) error
	createBridgeAgreementFunc  func(ctx context.Context, a *domain.BridgeAgreement) error
	listBridgeAgreementsFunc   func(ctx context.Context) ([]domain.BridgeAgreement, error)
}

func (m *mockPKIRepo) CreateForeignCA(ctx context.Context, ca *domain.ForeignCA) error {
	return m.createForeignCAFunc(ctx, ca)
}
func (m *mockPKIRepo) FindForeignCAByID(ctx context.Context, id uuid.UUID) (*domain.ForeignCA, error) {
	return nil, nil
}
func (m *mockPKIRepo) ListForeignCAs(ctx context.Context) ([]domain.ForeignCA, error) {
	return nil, nil
}
func (m *mockPKIRepo) CreateCrossCert(ctx context.Context, cert *domain.CrossCertificate) error {
	return m.createCrossCertFunc(ctx, cert)
}
func (m *mockPKIRepo) FindCrossCertBySubject(ctx context.Context, subject string) (*domain.CrossCertificate, error) {
	return m.findCrossCertBySubjectFunc(ctx, subject)
}
func (m *mockPKIRepo) ListTrustAnchors(ctx context.Context) ([]domain.TrustAnchor, error) {
	return m.listTrustAnchorsFunc(ctx)
}
func (m *mockPKIRepo) SavePathValidation(ctx context.Context, pv *domain.PathValidation) error {
	return m.savePathValidationFunc(ctx, pv)
}
func (m *mockPKIRepo) CreateBridgeAgreement(ctx context.Context, a *domain.BridgeAgreement) error {
	return m.createBridgeAgreementFunc(ctx, a)
}
func (m *mockPKIRepo) ListBridgeAgreements(ctx context.Context) ([]domain.BridgeAgreement, error) {
	return m.listBridgeAgreementsFunc(ctx)
}

func newTestPKIService(repo repository.Repository) *PKIBridgeService {
	return NewPKIBridgeService(repo, nil)
}

func TestNewPKIBridgeService(t *testing.T) {
	svc := NewPKIBridgeService(nil, nil)
	require.NotNil(t, svc)
}

func TestRegisterForeignCA(t *testing.T) {
	tests := []struct {
		name    string
		ca      domain.ForeignCA
		repo    *mockPKIRepo
		wantErr bool
	}{
		{
			name: "success",
			ca:   domain.ForeignCA{Name: "ForeignCA", Country: "FR", PublicKeyPEM: "LS0t..."},
			repo: &mockPKIRepo{
				createForeignCAFunc: func(ctx context.Context, ca *domain.ForeignCA) error { return nil },
			},
		},
		{
			name: "repo error",
			ca:   domain.ForeignCA{Name: "FailCA", Country: "XX"},
			repo: &mockPKIRepo{
				createForeignCAFunc: func(ctx context.Context, ca *domain.ForeignCA) error { return errors.New("db error") },
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestPKIService(tt.repo)
			result, err := svc.RegisterForeignCA(context.Background(), tt.ca)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "ACTIVE", result.Status)
			assert.NotEmpty(t, result.CAID)
		})
	}
}

func TestIssueCrossCert(t *testing.T) {
	caID := uuid.New()
	repo := &mockPKIRepo{
		createCrossCertFunc: func(ctx context.Context, cert *domain.CrossCertificate) error { return nil },
	}
	svc := newTestPKIService(repo)
	cert, err := svc.IssueCrossCert(context.Background(), "CN=Test", caID, "01", time.Now(), time.Now().AddDate(1, 0, 0), "LS0t...")
	require.NoError(t, err)
	assert.Equal(t, "CN=Test", cert.Subject)
	assert.NotEmpty(t, cert.CrossCertID)
}

func TestGetCrossCert(t *testing.T) {
	repo := &mockPKIRepo{
		findCrossCertBySubjectFunc: func(ctx context.Context, subject string) (*domain.CrossCertificate, error) {
			return &domain.CrossCertificate{Subject: subject}, nil
		},
	}
	svc := newTestPKIService(repo)
	cert, err := svc.GetCrossCert(context.Background(), "CN=Test")
	require.NoError(t, err)
	assert.Equal(t, "CN=Test", cert.Subject)
}

func TestListTrustAnchors(t *testing.T) {
	repo := &mockPKIRepo{
		listTrustAnchorsFunc: func(ctx context.Context) ([]domain.TrustAnchor, error) {
			return []domain.TrustAnchor{{Subject: "CN=Root"}}, nil
		},
	}
	svc := newTestPKIService(repo)
	anchors, err := svc.ListTrustAnchors(context.Background())
	require.NoError(t, err)
	assert.Len(t, anchors, 1)
}

func TestValidatePath(t *testing.T) {
	repo := &mockPKIRepo{
		savePathValidationFunc: func(ctx context.Context, pv *domain.PathValidation) error { return nil },
	}
	svc := newTestPKIService(repo)
	result, err := svc.ValidatePath(context.Background(), "CN=Leaf", []string{"CN=Intermediate"}, "CN=Root")
	require.NoError(t, err)
	assert.True(t, result.Result)
}

func TestCreateAgreement(t *testing.T) {
	policyID := uuid.New()
	repo := &mockPKIRepo{
		createBridgeAgreementFunc: func(ctx context.Context, a *domain.BridgeAgreement) error { return nil },
	}
	svc := newTestPKIService(repo)
	agreement, err := svc.CreateAgreement(context.Background(), "TestAgreement", "ForeignCA", policyID, nil)
	require.NoError(t, err)
	assert.Equal(t, "TestAgreement", agreement.Name)
}

func TestListAgreements(t *testing.T) {
	repo := &mockPKIRepo{
		listBridgeAgreementsFunc: func(ctx context.Context) ([]domain.BridgeAgreement, error) {
			return []domain.BridgeAgreement{{Name: "Agr1"}}, nil
		},
	}
	svc := newTestPKIService(repo)
	agreements, err := svc.ListAgreements(context.Background())
	require.NoError(t, err)
	assert.Len(t, agreements, 1)
}
