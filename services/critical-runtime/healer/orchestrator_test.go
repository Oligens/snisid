package healer

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSnapshotStore struct {
	mu         sync.Mutex
	snapshots  map[string][]string
	restoreErr error
	listErr    error
}

func (m *mockSnapshotStore) RestoreLatest(snapshotID string) error {
	if m.restoreErr != nil {
		return m.restoreErr
	}
	return nil
}

func (m *mockSnapshotStore) ListSnapshots(component string) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.snapshots[component], nil
}

func TestNewHealer(t *testing.T) {
	store := &mockSnapshotStore{snapshots: map[string][]string{}}
	h := NewHealer("platform-1", store)
	assert.NotNil(t, h)
	assert.Equal(t, "platform-1", h.PlatformID)
	assert.Empty(t, h.activeHealings)
	assert.Empty(t, h.failureHistory)
}

func TestHeal_RateLimited(t *testing.T) {
	store := &mockSnapshotStore{snapshots: map[string][]string{}}
	h := NewHealer("p1", store)

	t0 := time.Now().Add(-6 * time.Minute)
	t1 := time.Now().Add(-4 * time.Minute)
	t2 := time.Now().Add(-3 * time.Minute)
	t3 := time.Now().Add(-2 * time.Minute)
	h.failureHistory["viol-1"] = []time.Time{t0, t1, t2, t3}

	violation := Violation{ID: "viol-1", Severity: SeverityHigh, Type: "SECURITY"}
	plan := h.Heal(violation)
	assert.Nil(t, plan)
}

func TestHeal_NotRateLimited(t *testing.T) {
	store := &mockSnapshotStore{snapshots: map[string][]string{}}
	h := NewHealer("p1", store)

	h.failureHistory["viol-1"] = []time.Time{time.Now().Add(-10 * time.Minute)}

	violation := Violation{ID: "viol-1", Severity: SeverityHigh, Type: "SECURITY"}
	plan := h.Heal(violation)
	require.NotNil(t, plan)
	assert.Equal(t, "viol-1", plan.Violation.ID)
}

func TestBuildCriticalPlan(t *testing.T) {
	store := &mockSnapshotStore{snapshots: map[string][]string{}}
	h := NewHealer("p1", store)

	v := Violation{
		ID:       "v1",
		Severity: SeverityCritical,
		Type:     "INTEGRITY",
		Affected: []string{"domain-1", "domain-2"},
	}
	plan := h.Heal(v)
	require.NotNil(t, plan)
	require.Len(t, plan.Steps, 6)
	assert.Equal(t, "ISOLATE", plan.Steps[0].Action)
	assert.Equal(t, "ISOLATE", plan.Steps[1].Action)
	assert.Equal(t, "ROLLBACK", plan.Steps[3].Action)
	assert.Equal(t, "RESTART", plan.Steps[4].Action)
	assert.Equal(t, "RECONFIGURE", plan.Steps[5].Action)
}

func TestBuildHighPlan(t *testing.T) {
	store := &mockSnapshotStore{snapshots: map[string][]string{}}
	h := NewHealer("p1", store)

	v := Violation{
		ID:       "v2",
		Severity: SeverityHigh,
		Affected: []string{"component-a"},
	}
	plan := h.Heal(v)
	require.NotNil(t, plan)
	require.Len(t, plan.Steps, 3)
	assert.Equal(t, "ISOLATE", plan.Steps[0].Action)
	assert.Equal(t, "ROLLBACK", plan.Steps[1].Action)
	assert.Equal(t, "RESTART", plan.Steps[2].Action)
}

func TestBuildMediumPlan(t *testing.T) {
	store := &mockSnapshotStore{snapshots: map[string][]string{}}
	h := NewHealer("p1", store)

	v := Violation{
		ID:       "v3",
		Severity: SeverityMedium,
		Affected: []string{"component-b"},
	}
	plan := h.Heal(v)
	require.NotNil(t, plan)
	require.Len(t, plan.Steps, 2)
	assert.Equal(t, "RESTART", plan.Steps[0].Action)
	assert.Equal(t, "RECONFIGURE", plan.Steps[1].Action)
}

func TestBuildLowPlan(t *testing.T) {
	store := &mockSnapshotStore{snapshots: map[string][]string{}}
	h := NewHealer("p1", store)

	v := Violation{
		ID:       "v4",
		Severity: SeverityLow,
		Affected: []string{"component-c"},
	}
	plan := h.Heal(v)
	require.NotNil(t, plan)
	require.Len(t, plan.Steps, 1)
	assert.Equal(t, "RECONFIGURE", plan.Steps[0].Action)
}

func TestExecuteStep_RollbackWithSnapshots(t *testing.T) {
	store := &mockSnapshotStore{
		snapshots: map[string][]string{"critical": {"snap-1", "snap-2"}},
	}
	h := NewHealer("p1", store)

	err := h.RollbackToLastVerified()
	require.NoError(t, err)
}

func TestExecuteStep_RollbackNoSnapshots(t *testing.T) {
	store := &mockSnapshotStore{
		snapshots: map[string][]string{"critical": {}},
	}
	h := NewHealer("p1", store)

	err := h.RollbackToLastVerified()
	require.NoError(t, err)
}

func TestExecuteStep_RollbackListError(t *testing.T) {
	store := &mockSnapshotStore{
		snapshots: map[string][]string{},
		listErr:   errors.New("db error"),
	}
	h := NewHealer("p1", store)

	err := h.RollbackToLastVerified()
	assert.Error(t, err)
}

func TestExecuteStep_RollbackRestoreError(t *testing.T) {
	store := &mockSnapshotStore{
		snapshots:  map[string][]string{"critical": {"snap-1"}},
		restoreErr: errors.New("restore failed"),
	}
	h := NewHealer("p1", store)

	err := h.RollbackToLastVerified()
	assert.Error(t, err)
}

func TestExecuteStep_UnknownAction(t *testing.T) {
	store := &mockSnapshotStore{snapshots: map[string][]string{}}
	h := NewHealer("p1", store)

	err := h.executeStep(&HealingStep{Action: "UNKNOWN"})
	assert.Error(t, err)
}

func TestGetActiveHealings(t *testing.T) {
	store := &mockSnapshotStore{snapshots: map[string][]string{}}
	h := NewHealer("p1", store)

	v := Violation{ID: "v5", Severity: SeverityLow, Affected: []string{"x"}}
	h.Heal(v)

	active := h.GetActiveHealings()
	require.Len(t, active, 1)
	assert.Contains(t, active, "v5")
}

func TestResume(t *testing.T) {
	store := &mockSnapshotStore{snapshots: map[string][]string{}}
	h := NewHealer("p1", store)

	v := Violation{ID: "v6", Severity: SeverityLow, Affected: []string{"x"}}
	h.Heal(v)
	h.Resume()

	active := h.GetActiveHealings()
	assert.Empty(t, active)
}

func TestIsRateLimited(t *testing.T) {
	store := &mockSnapshotStore{snapshots: map[string][]string{}}
	h := NewHealer("p1", store)

	assert.False(t, h.isRateLimited("new-viol"))

	h.failureHistory["viol-1"] = []time.Time{time.Now(), time.Now(), time.Now()}
	assert.True(t, h.isRateLimited("viol-1"))

	h.failureHistory["viol-1"] = []time.Time{time.Now().Add(-10 * time.Minute)}
	assert.False(t, h.isRateLimited("viol-1"))
}

func TestConcurrentHealing(t *testing.T) {
	store := &mockSnapshotStore{snapshots: map[string][]string{}}
	h := NewHealer("p1", store)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			v := Violation{
				ID:       "concurrent-heal",
				Severity: SeverityLow,
				Affected: []string{"c"},
			}
			h.Heal(v)
		}(i)
	}
	wg.Wait()

	active := h.GetActiveHealings()
	assert.Len(t, active, 1)
}
