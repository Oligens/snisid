package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/classification-mgmt-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	createRuleFn         func(ctx context.Context, r *domain.ClassificationRule) error
	getRulesByDataTypeFn func(ctx context.Context, dt string) ([]domain.ClassificationRule, error)
	createTagFn          func(ctx context.Context, t *domain.DataTag) error
	getTagByURIFn        func(ctx context.Context, uri string) (*domain.DataTag, error)
	createAuditLogFn     func(ctx context.Context, a *domain.ClassificationAudit) error
	getRecentAuditLogsFn func(ctx context.Context) ([]domain.ClassificationAudit, error)
	getDashboardStatsFn  func(ctx context.Context) (*domain.DashboardStats, error)
}

func (m *mockRepo) CreateRule(ctx context.Context, r *domain.ClassificationRule) error {
	return m.createRuleFn(ctx, r)
}
func (m *mockRepo) GetRulesByDataType(ctx context.Context, dt string) ([]domain.ClassificationRule, error) {
	return m.getRulesByDataTypeFn(ctx, dt)
}
func (m *mockRepo) CreateTag(ctx context.Context, t *domain.DataTag) error {
	return m.createTagFn(ctx, t)
}
func (m *mockRepo) GetTagByURI(ctx context.Context, uri string) (*domain.DataTag, error) {
	return m.getTagByURIFn(ctx, uri)
}
func (m *mockRepo) CreateAuditLog(ctx context.Context, a *domain.ClassificationAudit) error {
	return m.createAuditLogFn(ctx, a)
}
func (m *mockRepo) GetRecentAuditLogs(ctx context.Context) ([]domain.ClassificationAudit, error) {
	return m.getRecentAuditLogsFn(ctx)
}
func (m *mockRepo) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	return m.getDashboardStatsFn(ctx)
}

func TestCreateRule(t *testing.T) {
	repo := &mockRepo{
		createRuleFn: func(ctx context.Context, r *domain.ClassificationRule) error { return nil },
	}
	svc := NewClassificationService(repo, nil)
	req := domain.CreateRuleRequest{
		DataType:         "SSN",
		SensitivityLevel: "TOP_SECRET",
		CreatedBy:        uuid.New().String(),
	}
	rule, err := svc.CreateRule(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, rule.ID)
	assert.True(t, rule.Active)
}

func TestTagResource(t *testing.T) {
	repo := &mockRepo{
		createTagFn: func(ctx context.Context, t *domain.DataTag) error { return nil },
	}
	svc := NewClassificationService(repo, nil)
	req := domain.TagResourceRequest{
		ResourceURI:       "snisid://doc/123",
		ClassificationTop: "SECRET",
		OwnerAgency:       "DHS",
		TaggedBy:          uuid.New().String(),
	}
	tag, err := svc.TagResource(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, tag.ID)
	assert.Equal(t, req.ResourceURI, tag.ResourceURI)
}

func TestLogAudit(t *testing.T) {
	repo := &mockRepo{
		createAuditLogFn: func(ctx context.Context, a *domain.ClassificationAudit) error { return nil },
	}
	svc := NewClassificationService(repo, nil)
	req := domain.LogAuditRequest{
		ResourceURI:             "snisid://doc/123",
		Action:                  "CLASSIFY",
		AuthorizedBy:            uuid.New().String(),
		ClassificationAuthority: "EO 13526",
		IPAddress:               "10.0.0.1",
	}
	entry, err := svc.LogAudit(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, entry.ID)
	assert.Equal(t, domain.ActionClassify, entry.Action)
}
