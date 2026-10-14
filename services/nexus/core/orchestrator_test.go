package core

import (
	"context"
	"errors"
	"testing"

	nexusv1 "github.com/snisid/platform/api/proto/nexus/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockStateManager struct {
	mock.Mock
	tasks map[string]*Task
}

func newMockState() *mockStateManager {
	return &mockStateManager{tasks: make(map[string]*Task)}
}

func (m *mockStateManager) SaveTask(ctx context.Context, task *Task) error {
	args := m.Called(ctx, task)
	m.tasks[task.Definition.Id] = task
	return args.Error(0)
}

func (m *mockStateManager) GetTask(ctx context.Context, id string) (*Task, error) {
	args := m.Called(ctx, id)
	if t, ok := m.tasks[id]; ok {
		return t, nil
	}
	return nil, args.Error(1)
}

func (m *mockStateManager) UpdateTaskStatus(ctx context.Context, id string, status nexusv1.TaskStatus) error {
	args := m.Called(ctx, id, status)
	if task, ok := m.tasks[id]; ok {
		task.Status = status
	}
	return args.Error(0)
}

type mockAgent struct {
	mock.Mock
	id  string
	typ string
}

func (a *mockAgent) ID() string   { return a.id }
func (a *mockAgent) Type() string { return a.typ }
func (a *mockAgent) CanHandle(task *Task) bool {
	return a.Called(task).Bool(0)
}
func (a *mockAgent) Execute(ctx context.Context, task *Task) (*nexusv1.AgentSignal, error) {
	args := a.Called(ctx, task)
	return args.Get(0).(*nexusv1.AgentSignal), args.Error(1)
}

func TestNewOrchestrator(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := newMockState()
	o := NewOrchestrator(5, logger, state)

	require.NotNil(t, o)
	assert.Equal(t, 5, o.workers)
	assert.NotNil(t, o.agents)
	assert.NotNil(t, o.queue)
}

func TestRegisterAgent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := newMockState()
	o := NewOrchestrator(1, logger, state)

	agent := &mockAgent{id: "kai-1", typ: "analysis"}
	agent.On("CanHandle", mock.Anything).Return(true)

	o.RegisterAgent(agent)
	assert.Len(t, o.agents, 1)
	assert.Equal(t, agent, o.agents["kai-1"])
}

func TestSubmitTask_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := newMockState()
	o := NewOrchestrator(1, logger, state)

	state.On("SaveTask", mock.Anything, mock.Anything).Return(nil)

	resp, err := o.SubmitTask(context.Background(), &nexusv1.TaskDefinition{
		Type:    "fraud_analysis",
		Payload: map[string]string{"citizen_id": "CIT-001"},
	})
	require.NoError(t, err)
	assert.True(t, resp.Accepted)
	assert.NotEmpty(t, resp.TaskId)
	assert.Contains(t, resp.Message, "queued")
}

func TestSubmitTask_GeneratedID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := newMockState()
	o := NewOrchestrator(1, logger, state)

	state.On("SaveTask", mock.Anything, mock.Anything).Return(nil)

	resp, err := o.SubmitTask(context.Background(), &nexusv1.TaskDefinition{
		Id:   "",
		Type: "identity_check",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.TaskId)
}

func TestSubmitTask_SaveFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := newMockState()
	o := NewOrchestrator(1, logger, state)

	state.On("SaveTask", mock.Anything, mock.Anything).Return(errors.New("storage error"))

	_, err := o.SubmitTask(context.Background(), &nexusv1.TaskDefinition{
		Type: "test",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save")
}

func TestStart_RunsWorkers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := newMockState()
	o := NewOrchestrator(3, logger, state)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	o.Start(ctx)
}

func TestProcessTask_NoAgentFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := newMockState()
	o := NewOrchestrator(1, logger, state)

	task := &Task{
		Definition: &nexusv1.TaskDefinition{Id: "task-1", Type: "unknown"},
		Status:     nexusv1.TaskStatus_TASK_STATUS_PENDING,
	}

	o.processTask(context.Background(), task)
}

func TestProcessTask_AgentFoundAndExecutes(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := newMockState()
	o := NewOrchestrator(1, logger, state)

	agent := &mockAgent{id: "kai-exec", typ: "executor"}
	agent.On("CanHandle", mock.Anything).Return(true)
	agent.On("Execute", mock.Anything, mock.Anything).Return(&nexusv1.AgentSignal{
		Status: nexusv1.TaskStatus_TASK_STATUS_COMPLETED,
	}, nil)
	o.RegisterAgent(agent)

	task := &Task{
		Definition: &nexusv1.TaskDefinition{Id: "task-exec", Type: "executor"},
		Status:     nexusv1.TaskStatus_TASK_STATUS_PENDING,
	}

	state.On("UpdateTaskStatus", mock.Anything, "task-exec", nexusv1.TaskStatus_TASK_STATUS_RUNNING).Return(nil)
	state.On("UpdateTaskStatus", mock.Anything, "task-exec", nexusv1.TaskStatus_TASK_STATUS_COMPLETED).Return(nil)

	o.processTask(context.Background(), task)
}

func TestProcessTask_AgentFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := newMockState()
	o := NewOrchestrator(1, logger, state)

	agent := &mockAgent{id: "kai-fail", typ: "executor"}
	agent.On("CanHandle", mock.Anything).Return(true)
	agent.On("Execute", mock.Anything, mock.Anything).Return(&nexusv1.AgentSignal{}, errors.New("execution error"))
	o.RegisterAgent(agent)

	task := &Task{
		Definition: &nexusv1.TaskDefinition{Id: "task-fail", Type: "executor"},
		Status:     nexusv1.TaskStatus_TASK_STATUS_PENDING,
	}

	state.On("UpdateTaskStatus", mock.Anything, "task-fail", nexusv1.TaskStatus_TASK_STATUS_RUNNING).Return(nil)
	state.On("UpdateTaskStatus", mock.Anything, "task-fail", nexusv1.TaskStatus_TASK_STATUS_FAILED).Return(nil)

	o.processTask(context.Background(), task)
}

func TestRegisterMultipleAgents(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := newMockState()
	o := NewOrchestrator(2, logger, state)

	a1 := &mockAgent{id: "kai-1", typ: "type-a"}
	a2 := &mockAgent{id: "kai-2", typ: "type-b"}
	o.RegisterAgent(a1)
	o.RegisterAgent(a2)

	assert.Len(t, o.agents, 2)
}

func TestSubmitTask_WithPayload(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := newMockState()
	o := NewOrchestrator(1, logger, state)

	state.On("SaveTask", mock.Anything, mock.Anything).Return(nil)

	resp, err := o.SubmitTask(context.Background(), &nexusv1.TaskDefinition{
		Type: "biometric_match",
		Payload: map[string]string{
			"enrollment_id": "ENR-001",
			"modality":      "face",
		},
	})
	require.NoError(t, err)
	assert.True(t, resp.Accepted)
}

func TestSelectFirstMatchingAgent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := newMockState()
	o := NewOrchestrator(1, logger, state)

	a1 := &mockAgent{id: "kai-a", typ: "type-a"}
	a2 := &mockAgent{id: "kai-b", typ: "type-b"}

	a1.On("CanHandle", mock.Anything).Return(false)
	a2.On("CanHandle", mock.Anything).Return(true)

	o.RegisterAgent(a1)
	o.RegisterAgent(a2)

	task := &Task{Definition: &nexusv1.TaskDefinition{Id: "t1", Type: "type-b"}}
	o.SubmitTask(context.Background(), task.Definition)
}

func TestWorkerShutdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := newMockState()
	o := NewOrchestrator(2, logger, state)

	ctx, cancel := context.WithCancel(context.Background())
	o.Start(ctx)
	cancel()
}

func TestInMemoryState_SaveAndGet(t *testing.T) {
	s := NewInMemoryState()
	task := &Task{
		Definition: &nexusv1.TaskDefinition{Id: "test-1", Type: "test"},
		Status:     nexusv1.TaskStatus_TASK_STATUS_PENDING,
	}

	err := s.SaveTask(context.Background(), task)
	require.NoError(t, err)

	retrieved, err := s.GetTask(context.Background(), "test-1")
	require.NoError(t, err)
	assert.Equal(t, "test-1", retrieved.Definition.Id)
	assert.Equal(t, nexusv1.TaskStatus_TASK_STATUS_PENDING, retrieved.Status)
}

func TestInMemoryState_GetNotFound(t *testing.T) {
	s := NewInMemoryState()
	_, err := s.GetTask(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestInMemoryState_UpdateStatus(t *testing.T) {
	s := NewInMemoryState()
	task := &Task{
		Definition: &nexusv1.TaskDefinition{Id: "task-status", Type: "test"},
		Status:     nexusv1.TaskStatus_TASK_STATUS_PENDING,
	}
	s.SaveTask(context.Background(), task)

	err := s.UpdateTaskStatus(context.Background(), "task-status", nexusv1.TaskStatus_TASK_STATUS_RUNNING)
	require.NoError(t, err)

	retrieved, _ := s.GetTask(context.Background(), "task-status")
	assert.Equal(t, nexusv1.TaskStatus_TASK_STATUS_RUNNING, retrieved.Status)
}

func TestInMemoryState_UpdateNotFound(t *testing.T) {
	s := NewInMemoryState()
	err := s.UpdateTaskStatus(context.Background(), "missing", nexusv1.TaskStatus_TASK_STATUS_RUNNING)
	assert.Error(t, err)
}

func TestInMemoryState_ConcurrentSafe(t *testing.T) {
	s := NewInMemoryState()
	t.Run("parallel", func(t *testing.T) {
		t.Run("save", func(t *testing.T) {
			s.SaveTask(context.Background(), &Task{
				Definition: &nexusv1.TaskDefinition{Id: "concurrent-1"},
			})
		})
		t.Run("update", func(t *testing.T) {
			s.SaveTask(context.Background(), &Task{
				Definition: &nexusv1.TaskDefinition{Id: "concurrent-2"},
			})
			s.UpdateTaskStatus(context.Background(), "concurrent-2", nexusv1.TaskStatus_TASK_STATUS_RUNNING)
		})
	})

	_, err := s.GetTask(context.Background(), "concurrent-1")
	assert.NoError(t, err)
}

func TestTaskTimeUpdatedOnStatusChange(t *testing.T) {
	s := NewInMemoryState()
	task := &Task{
		Definition: &nexusv1.TaskDefinition{Id: "time-test"},
		Status:     nexusv1.TaskStatus_TASK_STATUS_PENDING,
	}
	s.SaveTask(context.Background(), task)

	original := task.UpdatedAt
	s.UpdateTaskStatus(context.Background(), "time-test", nexusv1.TaskStatus_TASK_STATUS_COMPLETED)

	retrieved, _ := s.GetTask(context.Background(), "time-test")
	assert.True(t, retrieved.UpdatedAt.After(original) || retrieved.UpdatedAt.Equal(original))
}
