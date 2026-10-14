package agent

import (
	"context"
	"fmt"
	"time"

	nexusv1 "github.com/snisid/platform/api/proto/nexus/v1"
	"github.com/snisid/platform/nexus/core"
	"go.uber.org/zap"
)

type KaiAgent struct {
	id     string
	logger *zap.Logger
}

func NewKaiAgent(id string, logger *zap.Logger) *KaiAgent {
	return &KaiAgent{
		id:     id,
		logger: logger,
	}
}

func (a *KaiAgent) ID() string {
	return a.id
}

func (a *KaiAgent) Type() string {
	return "kai-executor"
}

func (a *KaiAgent) CanHandle(task *core.Task) bool {
	// For now, Kai handles everything. In production, we'd check task.Definition.Type
	return true
}

func (a *KaiAgent) Execute(ctx context.Context, task *core.Task) (*nexusv1.AgentSignal, error) {
	a.logger.Info("Kai executing task", zap.String("agent_id", a.id), zap.String("task_id", task.Definition.Id))

	// Simulate work
	select {
	case <-time.After(2 * time.Second):
		return &nexusv1.AgentSignal{
			AgentId:   a.id,
			TaskId:    task.Definition.Id,
			Status:    nexusv1.TaskStatus_TASK_STATUS_COMPLETED,
			Message:   "Execution finished successfully",
			Timestamp: nil, // Should be populated with current time
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
