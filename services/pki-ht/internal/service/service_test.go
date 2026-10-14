package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snisid/pki-ht/internal/domain"
	"github.com/snisid/pki-ht/internal/repository"
)

type mockRepo struct {
	createCertificateFunc func(ctx context.Context, cert *domain.IssuedCertificate) error
	findBySerialFunc      func(ctx context.Context, serial string) (*domain.IssuedCertificate, error)
	revokeCertificateFunc func(ctx context.Context, serial string, reason string) error
	getActiveCRLFunc      func(ctx context.Context, caID uuid.UUID) (*domain.CRL, error)
	updateCRLFunc         func(ctx context.Context, crl *domain.CRL) error
}

func (m *mockRepo) CreateCertificate(ctx context.Context, cert *domain.IssuedCertificate) error {
	return m.createCertificateFunc(ctx, cert)
}
func (m *mockRepo) FindBySerial(ctx context.Context, serial string) (*domain.IssuedCertificate, error) {
	return m.findBySerialFunc(ctx, serial)
}
func (m *mockRepo) RevokeCertificate(ctx context.Context, serial string, reason string) error {
	return m.revokeCertificateFunc(ctx, serial, reason)
}
func (m *mockRepo) GetActiveCRL(ctx context.Context, caID uuid.UUID) (*domain.CRL, error) {
	return m.getActiveCRLFunc(ctx, caID)
}
func (m *mockRepo) UpdateCRL(ctx context.Context, crl *domain.CRL) error {
	return m.updateCRLFunc(ctx, crl)
}

func newTestService(repo repository.Repository) *PKIService {
	return NewPKIService(repo, nil)
}

func TestNewPKIService(t *testing.T) {
	svc := NewPKIService(nil, nil)
	require.NotNil(t, svc)
}

func TestInitCA(t *testing.T) {
	svc := newTestService(&mockRepo{})
	err := svc.InitCA(context.Background())
	require.NoError(t, err)
	require.NotNil(t, svc.ca)
	assert.Equal(t, domain.CARoot, svc.ca.CAType)
	assert.Equal(t, "SNISID Root CA", svc.ca.CommonName)
	assert.True(t, svc.ca.IsActive)
}

func TestIssue(t *testing.T) {
	t.Run("success after InitCA", func(t *testing.T) {
		svc := newTestService(&mockRepo{
			createCertificateFunc: func(ctx context.Context, cert *domain.IssuedCertificate) error { return nil },
		})
		err := svc.InitCA(context.Background())
		require.NoError(t, err)

		req := domain.IssueRequest{
			SubjectType: "CITIZEN",
			SubjectRef:  uuid.New().String(),
			CommonName:  "Jean Dupont",
		}
		result, err := svc.Issue(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, domain.SubjectCitizen, result.SubjectType)
		assert.Equal(t, domain.CertValid, result.Status)
		assert.NotEmpty(t, result.SerialNumber)
		assert.NotNil(t, result.CommonName)
		assert.Equal(t, "Jean Dupont", *result.CommonName)
	})

	t.Run("success without subject ref", func(t *testing.T) {
		svc := newTestService(&mockRepo{
			createCertificateFunc: func(ctx context.Context, cert *domain.IssuedCertificate) error { return nil },
		})
		err := svc.InitCA(context.Background())
		require.NoError(t, err)

		req := domain.IssueRequest{
			SubjectType: "SERVICE",
			CommonName:  "fraud-engine",
		}
		result, err := svc.Issue(context.Background(), req)
		require.NoError(t, err)
		assert.Nil(t, result.SubjectRef)
	})

	t.Run("error without calling InitCA first", func(t *testing.T) {
		svc := newTestService(&mockRepo{})
		_, err := svc.Issue(context.Background(), domain.IssueRequest{
			SubjectType: "CITIZEN",
			CommonName:  "test",
		})
		require.Error(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		svc := newTestService(&mockRepo{
			createCertificateFunc: func(ctx context.Context, cert *domain.IssuedCertificate) error {
				return errors.New("db error")
			},
		})
		svc.InitCA(context.Background())

		_, err := svc.Issue(context.Background(), domain.IssueRequest{
			SubjectType: "CITIZEN",
			CommonName:  "test",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "save certificate")
	})
}

func TestRevoke(t *testing.T) {
	serial := "abc123"
	cert := &domain.IssuedCertificate{
		CertID:       uuid.New(),
		SerialNumber: serial,
		Status:       domain.CertValid,
	}

	t.Run("success", func(t *testing.T) {
		repo := &mockRepo{
			revokeCertificateFunc: func(ctx context.Context, s string, r string) error { return nil },
			findBySerialFunc: func(ctx context.Context, s string) (*domain.IssuedCertificate, error) {
				return cert, nil
			},
		}
		svc := newTestService(repo)
		err := svc.Revoke(context.Background(), serial, "key compromise")
		assert.NoError(t, err)
	})

	t.Run("repo revoke error", func(t *testing.T) {
		repo := &mockRepo{
			revokeCertificateFunc: func(ctx context.Context, s string, r string) error {
				return errors.New("not found")
			},
		}
		svc := newTestService(repo)
		err := svc.Revoke(context.Background(), serial, "test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "revoke")
	})

	t.Run("find after revoke error", func(t *testing.T) {
		repo := &mockRepo{
			revokeCertificateFunc: func(ctx context.Context, s string, r string) error { return nil },
			findBySerialFunc: func(ctx context.Context, s string) (*domain.IssuedCertificate, error) {
				return nil, errors.New("not found after revoke")
			},
		}
		svc := newTestService(repo)
		err := svc.Revoke(context.Background(), serial, "test")
		require.Error(t, err)
	})
}

func TestCheckOCSP(t *testing.T) {
	serial := "abc123"
	cert := &domain.IssuedCertificate{
		CertID:       uuid.New(),
		SerialNumber: serial,
		Status:       domain.CertValid,
	}

	tests := []struct {
		name    string
		serial  string
		repo    *mockRepo
		want    *domain.IssuedCertificate
		wantErr bool
	}{
		{
			name:   "valid certificate",
			serial: serial,
			repo: &mockRepo{
				findBySerialFunc: func(ctx context.Context, s string) (*domain.IssuedCertificate, error) { return cert, nil },
			},
			want: cert,
		},
		{
			name:   "not found",
			serial: "nonexistent",
			repo: &mockRepo{
				findBySerialFunc: func(ctx context.Context, s string) (*domain.IssuedCertificate, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.CheckOCSP(context.Background(), tt.serial)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetCRL(t *testing.T) {
	caID := uuid.New()
	now := time.Now()
	crl := &domain.CRL{
		CRLID:       uuid.New(),
		CAID:        caID,
		CRLNumber:   1,
		PublishedAt: now,
		NextUpdate:  now.AddDate(0, 0, 1),
	}

	tests := []struct {
		name    string
		caID    string
		repo    *mockRepo
		want    *domain.CRL
		wantErr bool
	}{
		{
			name: "success",
			caID: caID.String(),
			repo: &mockRepo{
				getActiveCRLFunc: func(ctx context.Context, cid uuid.UUID) (*domain.CRL, error) { return crl, nil },
			},
			want: crl,
		},
		{
			name:    "invalid CA ID",
			caID:    "bad-uuid",
			wantErr: true,
		},
		{
			name: "no CRL found",
			caID: caID.String(),
			repo: &mockRepo{
				getActiveCRLFunc: func(ctx context.Context, cid uuid.UUID) (*domain.CRL, error) {
					return nil, errors.New("no CRL found")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.GetCRL(context.Background(), tt.caID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
