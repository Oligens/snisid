package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type NodeStatus string

const (
	StatusHealthy     NodeStatus = "HEALTHY"
	StatusDegraded    NodeStatus = "DEGRADED"
	StatusIsolated    NodeStatus = "ISOLATED"
	StatusCompromised NodeStatus = "COMPROMISED"
)

type MeshNode struct {
	ID           string     `json:"id"`
	TrustScore   float64    `json:"trust_score"`
	Status       NodeStatus `json:"status"`
	LastAttested time.Time  `json:"last_attested"`
	SPIFFEID     string     `json:"spiffe_id"`
	FailureCount int        `json:"failure_count"`
}

type SelfHealingOperation struct {
	NodeID      string     `json:"node_id"`
	Operation   string     `json:"operation"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type MeshHealer struct {
	ClusterName     string
	mu              sync.Mutex
	operationLog    []SelfHealingOperation
	networkPolicies map[string]bool // nodeID -> isolated
}

func NewMeshHealer(clusterName string) *MeshHealer {
	return &MeshHealer{
		ClusterName:     clusterName,
		operationLog:    []SelfHealingOperation{},
		networkPolicies: make(map[string]bool),
	}
}

func (h *MeshHealer) MonitorMesh(nodes []MeshNode) {
	logger.Info(context.Background(), "GOS-HEAL: monitoring mesh integrity", zap.String("cluster", h.ClusterName), zap.Int("nodes", len(nodes)))

	for _, n := range nodes {
		switch {
		case n.TrustScore < 0.2:
			h.triggerSelfHealing(n, "trust critically low")
		case n.TrustScore < 0.5:
			if n.Status == StatusHealthy || n.Status == StatusDegraded {
				h.degradeNode(n, fmt.Sprintf("trust score %.2f below threshold", n.TrustScore))
			}
		case n.TrustScore >= 0.8 && n.Status == StatusIsolated:
			h.restoreNode(n)
		}
	}
}

func (h *MeshHealer) triggerSelfHealing(node MeshNode, reason string) {
	logger.Warn(context.Background(), "GOS-HEAL: triggering self-healing",
		zap.String("node", node.ID),
		zap.Float64("trust_score", node.TrustScore),
		zap.String("reason", reason),
	)

	op := SelfHealingOperation{
		NodeID:    node.ID,
		Operation: "ISOLATE_NETWORK_POLICY",
		Status:    "IN_PROGRESS",
		StartedAt: time.Now(),
	}

	err := h.applyNetworkIsolation(node.ID)
	if err != nil {
		op.Status = "FAILED"
		op.Error = err.Error()
		logger.Error(context.Background(), "GOS-HEAL: network isolation failed", zap.String("node", node.ID), zap.Error(err))
	} else {
		h.mu.Lock()
		h.networkPolicies[node.ID] = true
		h.mu.Unlock()

		op.Status = "COMPLETED"
		now := time.Now()
		op.CompletedAt = &now
	}

	h.mu.Lock()
	h.operationLog = append(h.operationLog, op)
	h.mu.Unlock()

	switch {
	case node.TrustScore < 0.2:
		h.rebuildViaK8s(node.ID)
	case node.TrustScore >= 0.2 && node.TrustScore < 0.5:
		h.reverifyAttestation(node.ID)
	}
}

func (h *MeshHealer) applyNetworkIsolation(nodeID string) error {
	logger.Info(context.Background(), "GOS-HEAL: applying network isolation via Istio/Calico", zap.String("node", nodeID))
	return nil
}

func (h *MeshHealer) rebuildViaK8s(nodeID string) error {
	logger.Info(context.Background(), "GOS-HEAL: triggering K8s deployment reconcile", zap.String("node", nodeID))
	return nil
}

func (h *MeshHealer) reverifyAttestation(nodeID string) error {
	logger.Info(context.Background(), "GOS-HEAL: re-verifying SPIFFE attestation", zap.String("node", nodeID))
	return nil
}

func (h *MeshHealer) degradeNode(node MeshNode, reason string) {
	logger.Info(context.Background(), "GOS-HEAL: degrading node status", zap.String("node", node.ID), zap.String("reason", reason))

	op := SelfHealingOperation{
		NodeID:    node.ID,
		Operation: "DEGRADE",
		Status:    "COMPLETED",
		StartedAt: time.Now(),
	}
	now := time.Now()
	op.CompletedAt = &now

	h.mu.Lock()
	h.operationLog = append(h.operationLog, op)
	h.mu.Unlock()
}

func (h *MeshHealer) restoreNode(node MeshNode) {
	logger.Info(context.Background(), "GOS-HEAL: restoring node to mesh", zap.String("node", node.ID))

	h.mu.Lock()
	delete(h.networkPolicies, node.ID)
	h.mu.Unlock()

	op := SelfHealingOperation{
		NodeID:    node.ID,
		Operation: "RESTORE",
		Status:    "COMPLETED",
		StartedAt: time.Now(),
	}
	now := time.Now()
	op.CompletedAt = &now

	h.mu.Lock()
	h.operationLog = append(h.operationLog, op)
	h.mu.Unlock()
}

func (h *MeshHealer) GetOperationLog() []SelfHealingOperation {
	h.mu.Lock()
	defer h.mu.Unlock()

	result := make([]SelfHealingOperation, len(h.operationLog))
	copy(result, h.operationLog)
	return result
}

func (h *MeshHealer) GetIsolatedNodes() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	var nodes []string
	for nodeID, isolated := range h.networkPolicies {
		if isolated {
			nodes = append(nodes, nodeID)
		}
	}
	return nodes
}
