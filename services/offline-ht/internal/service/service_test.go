package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/offline-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockOfflineRepo struct {
	pushQueueFn             func(ctx context.Context, item *domain.SyncQueueItem) error
	syncTerminalFn          func(ctx context.Context, terminalID uuid.UUID) ([]domain.SyncQueueItem, error)
	getConflictItemsFn      func(ctx context.Context) ([]domain.SyncQueueItem, error)
	upsertTerminalFn        func(ctx context.Context, t *domain.OfflineTerminal) error
	getTerminalsStatusFn    func(ctx context.Context) ([]domain.OfflineTerminal, error)
	updateQueueItemStatusFn func(ctx context.Context, id uuid.UUID, status domain.SyncStatus, errMsg *string) error
}

func (m *mockOfflineRepo) PushQueue(ctx context.Context, item *domain.SyncQueueItem) error {
	return m.pushQueueFn(ctx, item)
}
func (m *mockOfflineRepo) SyncTerminal(ctx context.Context, terminalID uuid.UUID) ([]domain.SyncQueueItem, error) {
	return m.syncTerminalFn(ctx, terminalID)
}
func (m *mockOfflineRepo) GetConflictItems(ctx context.Context) ([]domain.SyncQueueItem, error) {
	return m.getConflictItemsFn(ctx)
}
func (m *mockOfflineRepo) UpsertTerminal(ctx context.Context, t *domain.OfflineTerminal) error {
	return m.upsertTerminalFn(ctx, t)
}
func (m *mockOfflineRepo) GetTerminalsStatus(ctx context.Context) ([]domain.OfflineTerminal, error) {
	return m.getTerminalsStatusFn(ctx)
}
func (m *mockOfflineRepo) UpdateQueueItemStatus(ctx context.Context, id uuid.UUID, status domain.SyncStatus, errMsg *string) error {
	return m.updateQueueItemStatusFn(ctx, id, status, errMsg)
}

func TestPushQueue(t *testing.T) {
	tid := uuid.New()
	tests := []struct {
		name    string
		req     domain.PushQueueRequest
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			req: domain.PushQueueRequest{
				TerminalID: tid.String(),
				EntityType: "biometric",
				EntityID:   "citizen-123",
				Action:     "sync",
				Payload:    `{"data":"test"}`,
			},
		},
		{
			name: "invalid terminal id",
			req: domain.PushQueueRequest{
				TerminalID: "not-a-uuid",
				EntityType: "test",
				EntityID:   "x",
				Action:     "sync",
				Payload:    "{}",
			},
			wantErr: true,
		},
		{
			name: "repo error",
			req: domain.PushQueueRequest{
				TerminalID: tid.String(),
				EntityType: "test",
				EntityID:   "x",
				Action:     "sync",
				Payload:    "{}",
			},
			repoErr: errors.New("insert error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOfflineRepo{
				pushQueueFn: func(ctx context.Context, item *domain.SyncQueueItem) error {
					return tt.repoErr
				},
			}
			svc := NewOfflineService(repo, nil)
			item, err := svc.PushQueue(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, domain.SyncPending, item.Status)
			assert.Equal(t, 0, item.RetryCount)
		})
	}
}

func TestSyncTerminal(t *testing.T) {
	tid := uuid.New()
	pendingItem := domain.SyncQueueItem{
		ID:         uuid.New(),
		TerminalID: tid,
		Status:     domain.SyncPending,
		CreatedAt:  time.Now(),
	}
	tests := []struct {
		name       string
		terminalID string
		repoRes    []domain.SyncQueueItem
		repoErr    error
		wantErr    bool
	}{
		{
			name:       "success with items",
			terminalID: tid.String(),
			repoRes:    []domain.SyncQueueItem{pendingItem},
		},
		{
			name:       "empty items",
			terminalID: tid.String(),
			repoRes:    []domain.SyncQueueItem{},
		},
		{
			name:       "invalid uuid",
			terminalID: "bad-uuid",
			wantErr:    true,
		},
		{
			name:       "repo error",
			terminalID: tid.String(),
			repoErr:    errors.New("query error"),
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOfflineRepo{
				syncTerminalFn: func(ctx context.Context, terminalID uuid.UUID) ([]domain.SyncQueueItem, error) {
					return tt.repoRes, tt.repoErr
				},
				updateQueueItemStatusFn: func(ctx context.Context, id uuid.UUID, status domain.SyncStatus, errMsg *string) error {
					return nil
				},
				upsertTerminalFn: func(ctx context.Context, t *domain.OfflineTerminal) error {
					return nil
				},
			}
			svc := NewOfflineService(repo, nil)
			items, err := svc.SyncTerminal(context.Background(), tt.terminalID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, items, len(tt.repoRes))
		})
	}
}

func TestGetConflicts(t *testing.T) {
	tests := []struct {
		name    string
		repoRes []domain.SyncQueueItem
		repoErr error
		wantErr bool
	}{
		{
			name:    "with conflicts",
			repoRes: []domain.SyncQueueItem{{ID: uuid.New(), Status: domain.SyncConflict}},
		},
		{
			name:    "no conflicts",
			repoRes: []domain.SyncQueueItem{},
		},
		{
			name:    "repo error",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOfflineRepo{
				getConflictItemsFn: func(ctx context.Context) ([]domain.SyncQueueItem, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewOfflineService(repo, nil)
			got, err := svc.GetConflicts(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, len(tt.repoRes))
		})
	}
}

func TestGetTerminalsStatus(t *testing.T) {
	tests := []struct {
		name    string
		repoRes []domain.OfflineTerminal
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			repoRes: []domain.OfflineTerminal{
				{ID: uuid.New(), Name: "Terminal-1", IsOnline: true},
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
			repo := &mockOfflineRepo{
				getTerminalsStatusFn: func(ctx context.Context) ([]domain.OfflineTerminal, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewOfflineService(repo, nil)
			got, err := svc.GetTerminalsStatus(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, len(tt.repoRes))
		})
	}
}
