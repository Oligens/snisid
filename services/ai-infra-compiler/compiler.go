package compiler

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type ClusterState struct {
	Nodes       []NodeStatus    `json:"nodes"`
	Pods        []PodStatus     `json:"pods"`
	Namespaces  []string        `json:"namespaces"`
	CollectedAt time.Time       `json:"collected_at"`
}

type NodeStatus struct {
	Name            string  `json:"name"`
	CPUCapacity    float64 `json:"cpu_capacity"`
	CPULimit       float64 `json:"cpu_limit"`
	MemoryCapacity float64 `json:"memory_capacity"`
	MemoryLimit    float64 `json:"memory_limit"`
	GPUCount       int     `json:"gpu_count"`
	AllocatableCPU float64 `json:"allocatable_cpu"`
	AllocatableMem float64 `json:"allocatable_mem"`
	HealthScore    float64 `json:"health_score"`
	Labels         map[string]string `json:"labels"`
	Taints         []Taint  `json:"taints"`
}

type Taint struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

type PodStatus struct {
	Name              string  `json:"name"`
	Namespace         string  `json:"namespace"`
	NodeName          string  `json:"node_name"`
	CPURequest        float64 `json:"cpu_request"`
	MemoryRequest     float64 `json:"memory_request"`
	CPULimit          float64 `json:"cpu_limit"`
	MemoryLimit       float64 `json:"memory_limit"`
	RestartCount      int     `json:"restart_count"`
	Phase             string  `json:"phase"`
	QoSClass          string  `json:"qos_class"`
	HasGPU            bool    `json:"has_gpu"`
}

type ScalingDecision struct {
	Namespace     string `json:"namespace"`
	Deployment    string `json:"deployment"`
	CurrentReplicas int  `json:"current_replicas"`
	TargetReplicas  int  `json:"target_replicas"`
	Reason        string `json:"reason"`
	Priority      int    `json:"priority"` // 1=critical, 2=high, 3=normal
	GeneratedAt   time.Time `json:"generated_at"`
}

type ScalingPlan struct {
	Decisions []ScalingDecision `json:"decisions"`
	Summary   ScalingSummary    `json:"summary"`
}

type ScalingSummary struct {
	TotalActions int `json:"total_actions"`
	CriticalActions int `json:"critical_actions"`
	EstimatedCPU    float64 `json:"estimated_cpu_savings"`
	EstimatedMem    float64 `json:"estimated_mem_savings"`
}

type InfrastructurePlan struct {
	Scaling   *ScalingPlan   `json:"scaling,omitempty"`
	Healing   *HealingPlan   `json:"healing,omitempty"`
	Placement *PlacementPlan `json:"placement,omitempty"`
}

type HealingPlan struct {
	Actions []HealingAction `json:"actions"`
}

type HealingAction struct {
	Target      string `json:"target"`       // node name or deployment name
	ActionType  string `json:"action_type"`  // DRAIN, CORDON, RESTART, EVICT
	Reason      string `json:"reason"`
	Priority    int    `json:"priority"`
}

type PlacementPlan struct {
	Suggestions []PlacementSuggestion `json:"suggestions"`
}

type PlacementSuggestion struct {
	PodName       string `json:"pod_name"`
	PreferredNode string `json:"preferred_node"`
	Reason        string `json:"reason"`
	Affinity      string `json:"affinity"`
}

type AIInfraCompiler struct {
	kafkaTopic string
}

func NewAIInfraCompiler() *AIInfraCompiler {
	return &AIInfraCompiler{
		kafkaTopic: "infra.scaling.decisions",
	}
}

func (c *AIInfraCompiler) Compile(state ClusterState) (*InfrastructurePlan, error) {
	logger.Info(context.Background(), "AI Compiler: analyzing cluster state",
		zap.Int("nodes", len(state.Nodes)),
		zap.Int("pods", len(state.Pods)),
	)

	plan := &InfrastructurePlan{}

	scalingPlan := c.analyzeScaling(state)
	if scalingPlan != nil && len(scalingPlan.Decisions) > 0 {
		plan.Scaling = scalingPlan
	}

	healingPlan := c.analyzeHealth(state)
	if healingPlan != nil && len(healingPlan.Actions) > 0 {
		plan.Healing = healingPlan
	}

	placementPlan := c.analyzePlacement(state)
	if placementPlan != nil && len(placementPlan.Suggestions) > 0 {
		plan.Placement = placementPlan
	}

	return plan, nil
}

func (c *AIInfraCompiler) analyzeScaling(state ClusterState) *ScalingPlan {
	type deploymentInfo struct {
		namespace      string
		name           string
		pods           []PodStatus
		avgCPU         float64
		avgMemory      float64
		cpuLimitTotal  float64
		memLimitTotal  float64
	}

	deployments := make(map[string]*deploymentInfo)

	for _, pod := range state.Pods {
		if pod.Phase != "Running" {
			continue
		}
		key := pod.Namespace + "/" + extractDeployment(pod.Name)
		if _, ok := deployments[key]; !ok {
			deployments[key] = &deploymentInfo{
				namespace: pod.Namespace,
				name:      extractDeployment(pod.Name),
			}
		}
		d := deployments[key]
		d.pods = append(d.pods, pod)
		d.avgCPU += pod.CPURequest
		d.avgMemory += pod.MemoryRequest
		d.cpuLimitTotal += pod.CPULimit
		d.memLimitTotal += pod.MemoryLimit
	}

	var decisions []ScalingDecision

	for key, d := range deployments {
		if len(d.pods) == 0 {
			continue
		}
		d.avgCPU /= float64(len(d.pods))
		d.avgMemory /= float64(len(d.pods))

		replicas := len(d.pods)

		threshold := d.avgCPU
		if d.avgMemory > threshold {
			threshold = d.avgMemory
		}

		var decision ScalingDecision
		decision.Namespace = d.namespace
		decision.Deployment = d.name
		decision.CurrentReplicas = replicas
		decision.GeneratedAt = time.Now()

		totalAllocatable := 0.0
		for _, n := range state.Nodes {
			totalAllocatable += n.AllocatableCPU
		}

		if threshold > 0.8 {
			target := int(float64(replicas) * (threshold / 0.6))
			if target > replicas*4 {
				target = replicas * 4
			}
			decision.TargetReplicas = target
			decision.Reason = fmt.Sprintf("high resource utilization (%.0f%%), scaling up to %d replicas", threshold*100, target)
			decision.Priority = 1
			decisions = append(decisions, decision)
		} else if threshold < 0.2 && replicas > 2 {
			target := int(float64(replicas) * 0.75)
			if target < 2 {
				target = 2
			}
			decision.TargetReplicas = target
			decision.Reason = fmt.Sprintf("low resource utilization (%.0f%%), scaling down to %d replicas", threshold*100, target)
			decision.Priority = 3
			decisions = append(decisions, decision)
		}

		logger.Info(context.Background(), "AI Compiler: deployment analysis",
			zap.String("key", key),
			zap.Int("replicas", replicas),
			zap.String("avgCPU", fmt.Sprintf("%.2f", d.avgCPU)),
			zap.String("avgMem", fmt.Sprintf("%.2f", d.avgMemory)),
		)
	}

	if len(decisions) == 0 {
		return nil
	}

	sort.Slice(decisions, func(i, j int) bool {
		return decisions[i].Priority < decisions[j].Priority
	})

	summary := ScalingSummary{
		TotalActions:    len(decisions),
		EstimatedCPU:    calculateEstimatedSavings(decisions),
	}

	for _, d := range decisions {
		if d.Priority == 1 {
			summary.CriticalActions++
		}
	}

	return &ScalingPlan{
		Decisions: decisions,
		Summary:   summary,
	}
}

func (c *AIInfraCompiler) analyzeHealth(state ClusterState) *HealingPlan {
	var actions []HealingAction

	for _, node := range state.Nodes {
		if node.HealthScore < 0.5 {
			cordonAction := HealingAction{
				Target:     node.Name,
				ActionType: "CORDON",
				Reason:     fmt.Sprintf("node health score %.2f is below threshold", node.HealthScore),
				Priority:   1,
			}
			actions = append(actions, cordonAction)

			drainAction := HealingAction{
				Target:     node.Name,
				ActionType: "DRAIN",
				Reason:     fmt.Sprintf("draining unhealthy node (score: %.2f)", node.HealthScore),
				Priority:   2,
			}
			actions = append(actions, drainAction)
		}

		cpuUtil := 1.0
		if node.AllocatableCPU > 0 {
			cpuUtil = (node.CPULimit - node.AllocatableCPU) / node.CPULimit
		}
		if cpuUtil > 0.95 {
			actions = append(actions, HealingAction{
				Target:     node.Name,
				ActionType: "RESTART",
				Reason:     fmt.Sprintf("node CPU at %.0f%% capacity", cpuUtil*100),
				Priority:   2,
			})
		}
	}

	crashLoopMap := make(map[string]int)
	for _, pod := range state.Pods {
		if pod.RestartCount > 5 {
			key := pod.Namespace + "/" + extractDeployment(pod.Name)
			crashLoopMap[key]++
		}
	}

	for key, count := range crashLoopMap {
		actions = append(actions, HealingAction{
			Target:     key,
			ActionType: "RESTART",
			Reason:     fmt.Sprintf("deployment %s has %d pods in crash loop", key, count),
			Priority:   1,
		})
	}

	if len(actions) == 0 {
		return nil
	}

	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Priority < actions[j].Priority
	})

	return &HealingPlan{Actions: actions}
}

func (c *AIInfraCompiler) analyzePlacement(state ClusterState) *PlacementPlan {
	var suggestions []PlacementSuggestion

	if len(state.Nodes) == 0 {
		return nil
	}

	sort.Slice(state.Nodes, func(i, j int) bool {
		return state.Nodes[i].HealthScore > state.Nodes[j].HealthScore
	})
	bestNode := state.Nodes[0]

	for _, pod := range state.Pods {
		if pod.Phase != "Pending" {
			continue
		}
		if pod.HasGPU && bestNode.GPUCount == 0 {
			for _, n := range state.Nodes {
				if n.GPUCount > 0 {
					suggestions = append(suggestions, PlacementSuggestion{
						PodName:       pod.Name,
						PreferredNode: n.Name,
						Reason:        "GPU-required pod needs GPU-enabled node",
						Affinity:      "nodeAffinity",
					})
					break
				}
			}
		} else {
			suggestions = append(suggestions, PlacementSuggestion{
				PodName:       pod.Name,
				PreferredNode: bestNode.Name,
				Reason:        fmt.Sprintf("highest health score (%.2f)", bestNode.HealthScore),
				Affinity:      "preferred",
			})
		}
	}

	if len(suggestions) == 0 {
		return nil
	}
	return &PlacementPlan{Suggestions: suggestions}
}

func (c *AIInfraCompiler) EncodeDecisions(plan *InfrastructurePlan) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"topic":     c.kafkaTopic,
		"plan":      plan,
		"timestamp": time.Now().UTC(),
		"version":   "1.0",
	})
}

func extractDeployment(podName string) string {
	parts := strings.Split(podName, "-")
	if len(parts) <= 2 {
		return podName
	}
	return strings.Join(parts[:len(parts)-1], "-")
}

func calculateEstimatedSavings(decisions []ScalingDecision) float64 {
	var savings float64
	for _, d := range decisions {
		if d.TargetReplicas < d.CurrentReplicas {
			savings += float64(d.CurrentReplicas - d.TargetReplicas) * 0.5
		}
	}
	return savings
}
