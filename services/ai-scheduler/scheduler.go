package scheduler

import (
	"context"
	"fmt"

	"github.com/snisid/platform/internal/platform/logger"
)

type Pod struct {
	Name   string
	CPUReq int
	MemReq int
}

type Node struct {
	ID       string
	Region   string
	Capacity int
	Free     int
	Health   float64 // 0.0 to 1.0
}

type AIScheduler struct{}

func (s *AIScheduler) OptimizePlacement(pods []Pod, nodes []Node) map[string]string {
	logger.Info(context.Background(), "AI-SCHEDULER: Optimizing pod placement for national resilience...")

	placement := make(map[string]string)
	for _, p := range pods {
		// Logic: Place on node with highest health and sufficient capacity
		bestNode := ""
		maxHealth := -1.0

		for _, n := range nodes {
			if n.Free >= p.CPUReq && n.Health > maxHealth {
				maxHealth = n.Health
				bestNode = n.ID
			}
		}

		if bestNode != "" {
			placement[p.Name] = bestNode
			fmt.Printf("AI-SCHEDULER: Assigned pod %s to node %s (Health: %.2f)\n", p.Name, bestNode, maxHealth)
		}
	}

	return placement
}
