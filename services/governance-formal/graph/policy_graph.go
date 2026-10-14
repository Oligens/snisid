package graph

import (
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/snisid/platform/governance-formal/compiler"
)

type CompatibilityLevel int

const (
	Compatible     CompatibilityLevel = 3
	PartialOverlap CompatibilityLevel = 2
	RequiresReview CompatibilityLevel = 1
	Incompatible   CompatibilityLevel = 0
)

type PolicyNode struct {
	Country     string                  `json:"country"`
	Policy      compiler.CompiledPolicy `json:"policy"`
	Constraints []PolicyConstraint      `json:"constraints"`
	TrustScore  float64                 `json:"trust_score"`
}

type PolicyConstraint struct {
	Domain     string `json:"domain"` // DATA_RETENTION, BIOMETRIC_SHARING, FPR_CROSSCHECK
	Rule       string `json:"rule"`
	Strictness int    `json:"strictness"` // 1-10
}

type ProofEdge struct {
	From          string             `json:"from"`
	To            string             `json:"to"`
	Valid         bool               `json:"valid"`
	Compatibility CompatibilityLevel `json:"compatibility_level"`
	VerifiedAt    int64              `json:"verified_at"`
	Confidence    float64            `json:"confidence"`
	Violations    []string           `json:"violations,omitempty"`
}

type GlobalPolicyLattice struct {
	Nodes map[string]*PolicyNode `json:"nodes"`
	Edges []ProofEdge            `json:"edges"`
	mu    sync.RWMutex
}

func NewGlobalPolicyLattice() *GlobalPolicyLattice {
	return &GlobalPolicyLattice{
		Nodes: make(map[string]*PolicyNode),
		Edges: []ProofEdge{},
	}
}

func (l *GlobalPolicyLattice) ProveCompatibility(from, to string) (CompatibilityLevel, []string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	fromNode, ok := l.Nodes[from]
	if !ok {
		return Incompatible, nil, fmt.Errorf("source country %s not registered", from)
	}
	toNode, ok := l.Nodes[to]
	if !ok {
		return Incompatible, nil, fmt.Errorf("target country %s not registered", to)
	}

	violations := []string{}
	overlapScore := 0.0
	totalChecks := 0.0

	for _, fromCons := range fromNode.Constraints {
		for _, toCons := range toNode.Constraints {
			if fromCons.Domain != toCons.Domain {
				continue
			}
			totalChecks++
			if fromCons.Strictness <= toCons.Strictness+2 {
				overlapScore++
			} else {
				violations = append(violations, fmt.Sprintf(
					"%s: %s strictness %d > %s strictness %d (gap: %d)",
					fromCons.Domain, from, fromCons.Strictness, to, toCons.Strictness,
					fromCons.Strictness-toCons.Strictness,
				))
			}
		}
	}

	commonRules := 0
	for _, r1 := range fromNode.Policy.Rules {
		for _, r2 := range toNode.Policy.Rules {
			if r1 == r2 {
				commonRules++
			}
		}
	}

	totalUnique := len(fromNode.Policy.Rules) + len(toNode.Policy.Rules) - commonRules
	if totalUnique > 0 {
		policyOverlap := float64(commonRules) / float64(totalUnique)
		overlapScore += policyOverlap * 10
		totalChecks += 10
	}

	confidence := 1.0
	if totalChecks > 0 {
		confidence = overlapScore / totalChecks
	}

	var level CompatibilityLevel
	switch {
	case confidence >= 0.9 && len(violations) == 0:
		level = Compatible
	case confidence >= 0.6 && len(violations) <= 2:
		level = PartialOverlap
	case confidence >= 0.3:
		level = RequiresReview
	default:
		level = Incompatible
	}

	edge := ProofEdge{
		From:          from,
		To:            to,
		Valid:         level >= PartialOverlap,
		Compatibility: level,
		VerifiedAt:    now(),
		Confidence:    math.Round(confidence*100) / 100,
		Violations:    violations,
	}

	l.upsertEdge(edge)

	return level, violations, nil
}

func (l *GlobalPolicyLattice) AddCountry(country string, p compiler.CompiledPolicy) {
	l.mu.Lock()
	defer l.mu.Unlock()

	node := &PolicyNode{
		Country:    country,
		Policy:     p,
		TrustScore: 0.8,
		Constraints: []PolicyConstraint{
			{Domain: "DATA_RETENTION", Rule: "retain_biometrics_10y", Strictness: 7},
			{Domain: "BIOMETRIC_SHARING", Rule: "consent_required", Strictness: 8},
			{Domain: "FPR_CROSSCHECK", Rule: "real_time", Strictness: 9},
			{Domain: "DATA_SOVEREIGNTY", Rule: "local_storage_required", Strictness: 6},
			{Domain: "AUDIT_LOG", Rule: "immutable", Strictness: 8},
		},
	}

	if p.TrustScore > 0 {
		node.TrustScore = p.TrustScore
	}

	l.Nodes[country] = node
}

func (l *GlobalPolicyLattice) upsertEdge(edge ProofEdge) {
	for i, e := range l.Edges {
		if e.From == edge.From && e.To == edge.To {
			l.Edges[i] = edge
			return
		}
	}
	l.Edges = append(l.Edges, edge)
}

func (l *GlobalPolicyLattice) GetEdge(from, to string) *ProofEdge {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, e := range l.Edges {
		if e.From == from && e.To == to {
			return &e
		}
	}
	return nil
}

func (l *GlobalPolicyLattice) GetCompatiblePeers(country string, minLevel CompatibilityLevel) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var peers []string
	for _, e := range l.Edges {
		if e.From == country && e.Compatibility >= minLevel {
			peers = append(peers, e.To)
		}
	}
	return peers
}

func (l *GlobalPolicyLattice) RankPeersByTrust(country string) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	type scored struct {
		country string
		score   float64
	}
	var scores []scored

	for _, e := range l.Edges {
		if e.From == country {
			node, ok := l.Nodes[e.To]
			if !ok {
				continue
			}
			trustScore := node.TrustScore * e.Confidence
			scores = append(scores, scored{country: e.To, score: trustScore})
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	result := make([]string, len(scores))
	for i, s := range scores {
		result[i] = s.country
	}
	return result
}

func (l *GlobalPolicyLattice) VerifyGlobalConsistency() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var globalIssues []string
	for _, edge := range l.Edges {
		if !edge.Valid {
			continue
		}
		for _, reverse := range l.Edges {
			if reverse.From == edge.To && reverse.To == edge.From && !reverse.Valid {
				globalIssues = append(globalIssues,
					fmt.Sprintf("asymmetric compatibility: %s->%s valid but %s->%s invalid",
						edge.From, edge.To, edge.To, edge.From))
			}
		}
	}
	return globalIssues
}

func now() int64 {
	return 1748880000
}
