package service

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type UserNode struct {
	UID       string            `json:"uid"`
	Role      string            `json:"role"`
	RiskLevel float64           `json:"risk_level"`
	Labels    map[string]string `json:"labels"`
}

type ActionEdge struct {
	Type        string  `json:"type"`
	Target      string  `json:"target"`
	Timestamp   int64   `json:"timestamp"`
	RiskScore   float64 `json:"risk_score"`
	AnomalyFlag bool    `json:"anomaly_flag"`
}

type GraphQuery struct {
	SourceUID string  `json:"source_uid"`
	TargetUID string  `json:"target_uid"`
	MaxDepth  int     `json:"max_depth"`
	MinScore  float64 `json:"min_score"`
}

type PathResult struct {
	Path      []string `json:"path"`
	TotalRisk float64  `json:"total_risk"`
	Length    int      `json:"length"`
	Anomalous bool     `json:"anomalous"`
}

type ClusterResult struct {
	Nodes       []string `json:"nodes"`
	Centrality  float64  `json:"centrality"`
	AvgRisk     float64  `json:"avg_risk"`
	ThreatScore float64  `json:"threat_score"`
}

type ThreatAnalyzer struct {
	mu         sync.RWMutex
	users      map[string]*UserNode
	edges      map[string][]ActionEdge
	thresholds ThreatThresholds
}

type ThreatThresholds struct {
	InsiderRiskThreshold   float64 `json:"insider_risk_threshold"`
	SuspiciousPathScore    float64 `json:"suspicious_path_score"`
	ClusterThreatThreshold float64 `json:"cluster_threat_threshold"`
	AnomalyActionCount     int     `json:"anomaly_action_count"`
}

func NewThreatAnalyzer() *ThreatAnalyzer {
	return &ThreatAnalyzer{
		users: make(map[string]*UserNode),
		edges: make(map[string][]ActionEdge),
		thresholds: ThreatThresholds{
			InsiderRiskThreshold:   0.7,
			SuspiciousPathScore:    1.5,
			ClusterThreatThreshold: 0.6,
			AnomalyActionCount:     5,
		},
	}
}

func (a *ThreatAnalyzer) MapRelationship(user UserNode, action ActionEdge) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.users[user.UID]; !ok {
		a.users[user.UID] = &UserNode{
			UID:       user.UID,
			Role:      user.Role,
			RiskLevel: user.RiskLevel,
			Labels:    user.Labels,
		}
	}

	key := user.UID + ":" + action.Type
	a.edges[key] = append(a.edges[key], action)

	actionCount := len(a.edges[key])
	if actionCount > a.thresholds.AnomalyActionCount {
		logger.Warn(context.Background(), "GRAPH: high action frequency detected",
			zap.String("user", user.UID),
			zap.String("action", action.Type),
			zap.Int("count", actionCount),
		)
	}
}

func (a *ThreatAnalyzer) DetectInsiderThreat(uid string) float64 {
	a.mu.RLock()
	user, ok := a.users[uid]
	if !ok {
		a.mu.RUnlock()
		return 0.0
	}

	var userEdges []ActionEdge
	for key, edges := range a.edges {
		if len(key) > len(uid) && key[:len(uid)] == uid {
			userEdges = append(userEdges, edges...)
		}
	}
	a.mu.RUnlock()

	if len(userEdges) == 0 {
		return 0.0
	}

	threatScore := user.RiskLevel * 0.3

	anomalyCount := 0
	totalRisk := 0.0
	timeWindows := make(map[string]int)

	for _, e := range userEdges {
		totalRisk += e.RiskScore
		if e.AnomalyFlag {
			anomalyCount++
		}

		t := time.Unix(e.Timestamp, 0).Format("2006-01-02 15:00")
		timeWindows[t]++
	}

	avgRisk := totalRisk / float64(len(userEdges))
	threatScore += avgRisk * 0.4

	if anomalyCount > 0 {
		threatScore += float64(anomalyCount) * 0.05
	}

	for _, count := range timeWindows {
		if count > 5 {
			threatScore += float64(count) * 0.02
		}
	}

	threatScore = math.Min(1.0, threatScore)

	if threatScore > a.thresholds.InsiderRiskThreshold {
		logger.Warn(context.Background(), "GRAPH: insider threat detected",
			zap.String("user", uid),
			zap.Float64("score", threatScore),
			zap.Int("anomalies", anomalyCount),
			zap.Int("actions", len(userEdges)),
		)
	}

	return math.Round(threatScore*100) / 100
}

func (a *ThreatAnalyzer) FindShortestPath(query GraphQuery) *PathResult {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if _, ok := a.users[query.SourceUID]; !ok {
		return nil
	}
	if _, ok := a.users[query.TargetUID]; !ok {
		return nil
	}

	if query.MaxDepth == 0 {
		query.MaxDepth = 5
	}
	if query.MinScore == 0 {
		query.MinScore = 0.1
	}

	type nodeDist struct {
		node string
		dist int
		risk float64
		path []string
	}

	visited := make(map[string]bool)
	queue := []nodeDist{{node: query.SourceUID, dist: 0, path: []string{query.SourceUID}}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.node == query.TargetUID {
			return &PathResult{
				Path:      current.path,
				TotalRisk: math.Round(current.risk*100) / 100,
				Length:    current.dist,
				Anomalous: current.risk > a.thresholds.SuspiciousPathScore,
			}
		}

		if current.dist >= query.MaxDepth {
			continue
		}

		for key, edges := range a.edges {
			var edgeUID string
			if len(key) > len(current.node) && key[:len(current.node)] == current.node && key[len(current.node):len(current.node)+1] == ":" {
				edgeUID = current.node
			} else {
				continue
			}

			_ = edgeUID
			if visited[key] {
				continue
			}

			for _, e := range edges {
				if e.RiskScore < query.MinScore {
					continue
				}

				visited[key] = true
				newPath := make([]string, len(current.path))
				copy(newPath, current.path)

				queue = append(queue, nodeDist{
					node: current.node,
					dist: current.dist + 1,
					risk: current.risk + e.RiskScore,
					path: newPath,
				})
			}
		}
	}

	return nil
}

func (a *ThreatAnalyzer) DetectClusters(minSize int) []ClusterResult {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if minSize == 0 {
		minSize = 3
	}

	adjacency := make(map[string]map[string]bool)
	for _, user := range a.users {
		adjacency[user.UID] = make(map[string]bool)
	}

	clusters := a.findConnectedComponents(adjacency)

	var results []ClusterResult
	for _, cluster := range clusters {
		if len(cluster) < minSize {
			continue
		}

		totalRisk := 0.0
		nodeCount := len(cluster)

		for _, uid := range cluster {
			if user, ok := a.users[uid]; ok {
				totalRisk += user.RiskLevel
			}
		}

		avgRisk := totalRisk / float64(nodeCount)
		centrality := a.computeCentrality(cluster, adjacency)
		threatScore := avgRisk*0.6 + centrality*0.4

		if threatScore > a.thresholds.ClusterThreatThreshold {
			logger.Warn(context.Background(), "GRAPH: high-threat cluster detected",
				zap.Int("size", nodeCount),
				zap.Float64("avg_risk", avgRisk),
				zap.Float64("centrality", centrality),
				zap.Float64("threat_score", threatScore),
			)
		}

		results = append(results, ClusterResult{
			Nodes:       cluster,
			Centrality:  math.Round(centrality*100) / 100,
			AvgRisk:     math.Round(avgRisk*100) / 100,
			ThreatScore: math.Round(threatScore*100) / 100,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ThreatScore > results[j].ThreatScore
	})

	return results
}

func (a *ThreatAnalyzer) findConnectedComponents(adjacency map[string]map[string]bool) [][]string {
	visited := make(map[string]bool)
	var components [][]string

	for uid := range a.users {
		if visited[uid] {
			continue
		}

		component := []string{}
		stack := []string{uid}

		for len(stack) > 0 {
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if visited[current] {
				continue
			}
			visited[current] = true
			component = append(component, current)

			for neighbor := range adjacency[current] {
				if !visited[neighbor] {
					stack = append(stack, neighbor)
				}
			}
		}

		components = append(components, component)
	}

	return components
}

func (a *ThreatAnalyzer) computeCentrality(cluster []string, adj map[string]map[string]bool) float64 {
	if len(cluster) <= 1 {
		return 0
	}

	totalEdges := 0
	for _, uid := range cluster {
		for neighbor := range adj[uid] {
			if contains(cluster, neighbor) {
				totalEdges++
			}
		}
	}

	maxEdges := len(cluster) * (len(cluster) - 1)
	if maxEdges == 0 {
		return 0
	}

	return float64(totalEdges) / float64(maxEdges)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (a *ThreatAnalyzer) DetectAnomalousAccess(uid string) []ActionEdge {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var anomalous []ActionEdge
	for key, edges := range a.edges {
		if len(key) > len(uid) && key[:len(uid)] == uid {
			for _, e := range edges {
				if e.AnomalyFlag || e.RiskScore > 0.8 {
					anomalous = append(anomalous, e)
				}
			}
		}
	}

	sort.Slice(anomalous, func(i, j int) bool {
		return anomalous[i].RiskScore > anomalous[j].RiskScore
	})

	return anomalous
}
