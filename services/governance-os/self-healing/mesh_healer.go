package selfhealing

import (
	"context"
	"sync"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type NodeRole string

const (
	RoleValidator NodeRole = "VALIDATOR"
	RoleGateway   NodeRole = "GATEWAY"
	RoleExecutor  NodeRole = "EXECUTOR"
)

type NodeState struct {
	NodeID       string    `json:"node_id"`
	TrustScore   float64   `json:"trust_score"`
	Status       string    `json:"status"`
	Role         NodeRole  `json:"role"`
	LastSeen     time.Time `json:"last_seen"`
	FailureCount int       `json:"failure_count"`
	Latency      float64   `json:"latency_ms"`
}

type IsolationAction struct {
	NodeID     string    `json:"node_id"`
	Action     string    `json:"action"` // ISOLATE, CORDON, REMOVE
	Reason     string    `json:"reason"`
	ExecutedAt time.Time `json:"executed_at"`
}

type RebuildPlan struct {
	NodeID    string   `json:"node_id"`
	Steps     []string `json:"steps"`
	Completed bool     `json:"completed"`
}

type MeshHealer struct {
	ID             string
	mu             sync.Mutex
	isolationLog   []IsolationAction
	rebuildPlans   map[string]*RebuildPlan
	quorumRequired int
}

func NewMeshHealer(id string, quorum int) *MeshHealer {
	return &MeshHealer{
		ID:             id,
		isolationLog:   []IsolationAction{},
		rebuildPlans:   make(map[string]*RebuildPlan),
		quorumRequired: quorum,
	}
}

func (h *MeshHealer) MonitorTrust(nodes []NodeState) {
	var unhealthyNodes []NodeState

	for _, node := range nodes {
		switch {
		case node.TrustScore < 0.3:
			logger.Warn(context.Background(), "GOS-HEAL: node trust critically degraded",
				zap.String("node", node.NodeID),
				zap.Float64("trust_score", node.TrustScore),
			)
			h.Isolate(node.NodeID, "trust score below 0.3")
			h.TriggerRebuild(node.NodeID)
			unhealthyNodes = append(unhealthyNodes, node)

		case node.TrustScore < 0.5:
			logger.Warn(context.Background(), "GOS-HEAL: node trust degraded",
				zap.String("node", node.NodeID),
				zap.Float64("trust_score", node.TrustScore),
			)
			h.Isolate(node.NodeID, "trust score below 0.5")

		case node.FailureCount > 10:
			logger.Warn(context.Background(), "GOS-HEAL: node excessive failures",
				zap.String("node", node.NodeID),
				zap.Int("failures", node.FailureCount),
			)
			h.Isolate(node.NodeID, "excessive failure count")
		}
	}

	if len(unhealthyNodes) > 0 && len(nodes)-len(unhealthyNodes) < h.quorumRequired {
		logger.Error(context.Background(), "GOS-HEAL: quorum lost - too many nodes isolated",
			zap.Int("healthy", len(nodes)-len(unhealthyNodes)),
			zap.Int("required", h.quorumRequired),
		)
	}
}

func (h *MeshHealer) Isolate(nodeID string, reason string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	action := IsolationAction{
		NodeID:     nodeID,
		Action:     "ISOLATE",
		Reason:     reason,
		ExecutedAt: time.Now(),
	}
	h.isolationLog = append(h.isolationLog, action)

	logger.Warn(context.Background(), "GOS-HEAL: node isolated from zero-trust lattice",
		zap.String("node", nodeID),
		zap.String("reason", reason),
	)
}

func (h *MeshHealer) TriggerRebuild(nodeID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	plan := &RebuildPlan{
		NodeID: nodeID,
		Steps: []string{
			"drain_node",
			"verify_attestation",
			"reissue_spiffe_cert",
			"rejoin_mesh",
			"verify_trust_score",
		},
		Completed: false,
	}
	h.rebuildPlans[nodeID] = plan

	logger.Info(context.Background(), "GOS-HEAL: rebuild plan created", zap.String("node", nodeID), zap.Int("steps", len(plan.Steps)))
}

func (h *MeshHealer) GetIsolationLog() []IsolationAction {
	h.mu.Lock()
	defer h.mu.Unlock()

	result := make([]IsolationAction, len(h.isolationLog))
	copy(result, h.isolationLog)
	return result
}

func (h *MeshHealer) GetRebuildPlan(nodeID string) *RebuildPlan {
	h.mu.Lock()
	defer h.mu.Unlock()

	plan, ok := h.rebuildPlans[nodeID]
	if !ok {
		return nil
	}
	return plan
}

func (h *MeshHealer) CompleteRebuild(nodeID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if plan, ok := h.rebuildPlans[nodeID]; ok {
		plan.Completed = true
		logger.Info(context.Background(), "GOS-HEAL: node rebuild completed", zap.String("node", nodeID))
	}
}
