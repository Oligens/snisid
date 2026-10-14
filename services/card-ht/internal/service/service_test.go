package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snisid/card-ht/internal/domain"
	"github.com/snisid/card-ht/internal/repository"
)

type mockRepo struct {
	createFunc               func(ctx context.Context, doc *domain.CardDocument) error
	findByDocumentNumberFunc func(ctx context.Context, docNumber string) (*domain.CardDocument, error)
	findByCitizenIDFunc      func(ctx context.Context, citizenID uuid.UUID) ([]domain.CardDocument, error)
	updateStatusFunc         func(ctx context.Context, documentID uuid.UUID, status domain.CardStatus) error
}

func (m *mockRepo) Create(ctx context.Context, doc *domain.CardDocument) error {
	return m.createFunc(ctx, doc)
}
func (m *mockRepo) FindByDocumentNumber(ctx context.Context, docNumber string) (*domain.CardDocument, error) {
	return m.findByDocumentNumberFunc(ctx, docNumber)
}
func (m *mockRepo) FindByCitizenID(ctx context.Context, citizenID uuid.UUID) ([]domain.CardDocument, error) {
	return m.findByCitizenIDFunc(ctx, citizenID)
}
func (m *mockRepo) UpdateStatus(ctx context.Context, documentID uuid.UUID, status domain.CardStatus) error {
	return m.updateStatusFunc(ctx, documentID, status)
}

func newTestService(repo repository.Repository) *CardService {
	return NewCardService(repo, nil)
}

func TestNewCardService(t *testing.T) {
	svc := NewCardService(nil, nil)
	require.NotNil(t, svc)
}

func TestIssue(t *testing.T) {
	citizenID := uuid.New()

	tests := []struct {
		name        string
		req         domain.IssueRequest
		repo        *mockRepo
		wantErr     bool
		errContains string
	}{
		{
			name: "success national ID",
			req: domain.IssueRequest{
				DocType:       string(domain.DocNationalID),
				CitizenID:     citizenID.String(),
				IssuingOffice: "Bureau PAP",
				CreatedBy:     uuid.New().String(),
			},
			repo: &mockRepo{
				createFunc: func(ctx context.Context, doc *domain.CardDocument) error { return nil },
			},
		},
		{
			name: "success passport",
			req: domain.IssueRequest{
				DocType:   string(domain.DocPassport),
				CitizenID: citizenID.String(),
				CreatedBy: uuid.New().String(),
			},
			repo: &mockRepo{
				createFunc: func(ctx context.Context, doc *domain.CardDocument) error { return nil },
			},
		},
		{
			name: "invalid citizen_id",
			req: domain.IssueRequest{
				DocType:   string(domain.DocNationalID),
				CitizenID: "bad-uuid",
				CreatedBy: uuid.New().String(),
			},
			wantErr:     true,
			errContains: "invalid citizen_id",
		},
		{
			name: "invalid created_by",
			req: domain.IssueRequest{
				DocType:   string(domain.DocNationalID),
				CitizenID: citizenID.String(),
				CreatedBy: "bad-uuid",
			},
			wantErr:     true,
			errContains: "invalid created_by",
		},
		{
			name: "repo error",
			req: domain.IssueRequest{
				DocType:   string(domain.DocNationalID),
				CitizenID: citizenID.String(),
				CreatedBy: uuid.New().String(),
			},
			repo: &mockRepo{
				createFunc: func(ctx context.Context, doc *domain.CardDocument) error {
					return errors.New("db error")
				},
			},
			wantErr:     true,
			errContains: "create document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			result, err := svc.Issue(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, domain.CardIssued, result.Status)
			assert.NotEmpty(t, result.DocumentID)
			assert.NotEmpty(t, result.DocumentNumber)
			assert.False(t, result.IssueDate.IsZero())
			assert.False(t, result.ExpiryDate.IsZero())
			assert.Equal(t, "Imprimerie Nationale PAP", result.PersonalizationFacility)
		})
	}
}

func TestVerify(t *testing.T) {
	doc := &domain.CardDocument{
		DocumentID:     uuid.New(),
		DocumentNumber: "HTI-ID-2026-000001",
		DocType:        domain.DocNationalID,
	}

	tests := []struct {
		name      string
		docNumber string
		repo      *mockRepo
		want      *domain.CardDocument
		wantErr   bool
	}{
		{
			name:      "found",
			docNumber: "HTI-ID-2026-000001",
			repo: &mockRepo{
				findByDocumentNumberFunc: func(ctx context.Context, dn string) (*domain.CardDocument, error) { return doc, nil },
			},
			want: doc,
		},
		{
			name:      "not found",
			docNumber: "HTI-ID-2026-999999",
			repo: &mockRepo{
				findByDocumentNumberFunc: func(ctx context.Context, dn string) (*domain.CardDocument, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			got, err := svc.Verify(context.Background(), tt.docNumber)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReportLost(t *testing.T) {
	docID := uuid.New()

	tests := []struct {
		name        string
		documentID  string
		repo        *mockRepo
		wantErr     bool
		errContains string
	}{
		{
			name:       "success",
			documentID: docID.String(),
			repo: &mockRepo{
				updateStatusFunc: func(ctx context.Context, did uuid.UUID, status domain.CardStatus) error { return nil },
			},
		},
		{
			name:        "invalid document ID",
			documentID:  "bad-uuid",
			wantErr:     true,
			errContains: "invalid document_id",
		},
		{
			name:       "repo error",
			documentID: docID.String(),
			repo: &mockRepo{
				updateStatusFunc: func(ctx context.Context, did uuid.UUID, status domain.CardStatus) error {
					return errors.New("db error")
				},
			},
			wantErr:     true,
			errContains: "report lost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			err := svc.ReportLost(context.Background(), tt.documentID)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestRevoke(t *testing.T) {
	docID := uuid.New()

	tests := []struct {
		name        string
		documentID  string
		repo        *mockRepo
		wantErr     bool
		errContains string
	}{
		{
			name:       "success",
			documentID: docID.String(),
			repo: &mockRepo{
				updateStatusFunc: func(ctx context.Context, did uuid.UUID, status domain.CardStatus) error { return nil },
			},
		},
		{
			name:        "invalid document ID",
			documentID:  "bad-uuid",
			wantErr:     true,
			errContains: "invalid document_id",
		},
		{
			name:       "repo error",
			documentID: docID.String(),
			repo: &mockRepo{
				updateStatusFunc: func(ctx context.Context, did uuid.UUID, status domain.CardStatus) error {
					return errors.New("db error")
				},
			},
			wantErr:     true,
			errContains: "revoke",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.repo)
			err := svc.Revoke(context.Background(), tt.documentID)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}
