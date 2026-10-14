package healer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type ViolationSeverity int

const (
	SeverityLow      ViolationSeverity = 1
	SeverityMedium   ViolationSeverity = 2
	SeverityHigh     ViolationSeverity = 3
	SeverityCritical ViolationSeverity = 4
)

type Violation struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"` // SECURITY, PERFORMANCE, AVAILABILITY, INTEGRITY
	Description string            `json:"description"`
	Severity    ViolationSeverity `json:"severity"`
	Affected    []string          `json:"affected"` // affected components
	DetectedAt  time.Time         `json:"detected_at"`
	Source      string            `json:"source"` // monitor, formal verification, audit
}

type HealingStep struct {
	Action      string     `json:"action"` // ISOLATE, ROLLBACK, RESTART, RECONFIGURE, SCALE
	Target      string     `json:"target"`
	Status      string     `json:"status"` // PENDING, IN_PROGRESS, COMPLETED, FAILED
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type HealingPlan struct {
	Violation Violation     `json:"violation"`
	Steps     []HealingStep `json:"steps"`
	CreatedAt time.Time     `json:"created_at"`
	Completed bool          `json:"completed"`
	Success   bool          `json:"success"`
}

type SnapshotStore interface {
	RestoreLatest(snapshotID string) error
	ListSnapshots(component string) ([]string, error)
}

type Healer struct {
	PlatformID     string
	snapshotStore  SnapshotStore
	mu             sync.Mutex
	activeHealings map[string]*HealingPlan
	failureHistory map[string][]time.Time
}

func NewHealer(platformID string, store SnapshotStore) *Healer {
	return &Healer{
		PlatformID:     platformID,
		snapshotStore:  store,
		activeHealings: make(map[string]*HealingPlan),
		failureHistory: make(map[string][]time.Time),
	}
}

func (h *Healer) Heal(violation Violation) *HealingPlan {
	h.mu.Lock()
	defer h.mu.Unlock()

	logger.Warn(context.Background(), "SELF-HEALING: Initiating recovery",
		zap.String("violation", violation.ID),
		zap.String("type", violation.Type),
		zap.Int("severity", int(violation.Severity)),
	)

	if h.isRateLimited(violation.ID) {
		logger.Warn(context.Background(), "SELF-HEALING: healing rate limited for", zap.String("violation", violation.ID))
		return nil
	}

	plan := &HealingPlan{
		Violation: violation,
		CreatedAt: time.Now(),
	}

	switch violation.Severity {
	case SeverityCritical:
		plan.Steps = h.buildCriticalPlan(violation)
	case SeverityHigh:
		plan.Steps = h.buildHighPlan(violation)
	case SeverityMedium:
		plan.Steps = h.buildMediumPlan(violation)
	default:
		plan.Steps = h.buildLowPlan(violation)
	}

	h.activeHealings[violation.ID] = plan
	h.recordFailure(violation.ID)
	go h.executePlan(plan)

	return plan
}

func (h *Healer) buildCriticalPlan(v Violation) []HealingStep {
	steps := []HealingStep{
		{Action: "ISOLATE", Target: "affected_domains", Status: "PENDING"},
	}
	for _, affected := range v.Affected {
		steps = append(steps, HealingStep{Action: "ISOLATE", Target: affected, Status: "PENDING"})
	}
	steps = append(steps, HealingStep{Action: "ROLLBACK", Target: "last_verified_snapshot", Status: "PENDING"})
	steps = append(steps, HealingStep{Action: "RESTART", Target: "isolated_components", Status: "PENDING"})
	steps = append(steps, HealingStep{Action: "RECONFIGURE", Target: "security_policies", Status: "PENDING"})
	return steps
}

func (h *Healer) buildHighPlan(v Violation) []HealingStep {
	steps := []HealingStep{
		{Action: "ISOLATE", Target: v.Affected[0], Status: "PENDING"},
		{Action: "ROLLBACK", Target: "last_verified_snapshot", Status: "PENDING"},
		{Action: "RESTART", Target: v.Affected[0], Status: "PENDING"},
	}
	return steps
}

func (h *Healer) buildMediumPlan(v Violation) []HealingStep {
	steps := []HealingStep{
		{Action: "RESTART", Target: v.Affected[0], Status: "PENDING"},
		{Action: "RECONFIGURE", Target: v.Affected[0], Status: "PENDING"},
	}
	return steps
}

func (h *Healer) buildLowPlan(v Violation) []HealingStep {
	return []HealingStep{
		{Action: "RECONFIGURE", Target: v.Affected[0], Status: "PENDING"},
	}
}

func (h *Healer) executePlan(plan *HealingPlan) {
	for i := range plan.Steps {
		step := &plan.Steps[i]
		now := time.Now()
		step.StartedAt = &now
		step.Status = "IN_PROGRESS"

		logger.Info(context.Background(), "SELF-HEALING: executing step",
			zap.String("action", step.Action),
			zap.String("target", step.Target),
			zap.String("violation", plan.Violation.ID),
		)

		err := h.executeStep(step)
		completed := time.Now()
		step.CompletedAt = &completed

		if err != nil {
			step.Status = "FAILED"
			step.Error = err.Error()
			plan.Completed = true
			plan.Success = false
			logger.Error(context.Background(), "SELF-HEALING: step failed", err,
				zap.String("action", step.Action),
			)

			if plan.Violation.Severity >= SeverityHigh {
				h.escalate(plan.Violation, fmt.Sprintf("healing step %s failed: %s", step.Action, err.Error()))
			}
			return
		}

		step.Status = "COMPLETED"
	}

	plan.Completed = true
	plan.Success = true
	logger.Info(context.Background(), "SELF-HEALING: healing plan completed successfully",
		zap.String("violation", plan.Violation.ID),
	)
}

func (h *Healer) executeStep(step *HealingStep) error {
	switch step.Action {
	case "ISOLATE":
		return h.isolate(step.Target)
	case "ROLLBACK":
		return h.RollbackToLastVerified()
	case "RESTART":
		return h.restart(step.Target)
	case "RECONFIGURE":
		return h.reconfigure(step.Target)
	case "SCALE":
		return h.scale(step.Target)
	}
	return fmt.Errorf("unknown action: %s", step.Action)
}

func (h *Healer) isolate(target string) error {
	logger.Info(context.Background(), "SELF-HEALING: isolating", zap.String("target", target))
	return nil
}

func (h *Healer) RollbackToLastVerified() error {
	logger.Info(context.Background(), "SELF-HEALING: rolling back to last verified snapshot")
	if h.snapshotStore != nil {
		snapshots, err := h.snapshotStore.ListSnapshots("critical")
		if err != nil {
			return fmt.Errorf("failed to list snapshots: %w", err)
		}
		if len(snapshots) > 0 {
			if err := h.snapshotStore.RestoreLatest(snapshots[len(snapshots)-1]); err != nil {
				return fmt.Errorf("failed to restore snapshot: %w", err)
			}
		}
	}
	return nil
}

func (h *Healer) restart(target string) error {
	logger.Info(context.Background(), "SELF-HEALING: restarting", zap.String("target", target))
	return nil
}

func (h *Healer) reconfigure(target string) error {
	logger.Info(context.Background(), "SELF-HEALING: reconfiguring", zap.String("target", target))
	return nil
}

func (h *Healer) scale(target string) error {
	logger.Info(context.Background(), "SELF-HEALING: scaling", zap.String("target", target))
	return nil
}

func (h *Healer) Resume() {
	logger.Info(context.Background(), "SELF-HEALING: system stability restored, resuming operations")
	h.mu.Lock()
	defer h.mu.Unlock()
	h.activeHealings = make(map[string]*HealingPlan)
}

func (h *Healer) GetActiveHealings() map[string]*HealingPlan {
	h.mu.Lock()
	defer h.mu.Unlock()
	result := make(map[string]*HealingPlan)
	for k, v := range h.activeHealings {
		result[k] = v
	}
	return result
}

func (h *Healer) isRateLimited(violationID string) bool {
	history, ok := h.failureHistory[violationID]
	if !ok || len(history) < 3 {
		return false
	}
	window := time.Now().Add(-5 * time.Minute)
	count := 0
	for _, t := range history {
		if t.After(window) {
			count++
		}
	}
	return count >= 3
}

func (h *Healer) recordFailure(violationID string) {
	if h.failureHistory[violationID] == nil {
		h.failureHistory[violationID] = []time.Time{}
	}
	h.failureHistory[violationID] = append(h.failureHistory[violationID], time.Now())
}

func (h *Healer) escalate(violation Violation, reason string) {
	logger.Error(context.Background(), "SELF-HEALING: escalating to human operator", fmt.Errorf(reason),
		zap.String("violation", violation.ID),
		zap.String("type", violation.Type),
	)
}
