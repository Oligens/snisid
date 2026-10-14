package proposer

import (
	"fmt"
	"time"
)

type InterventionStatus string

const (
	StatusProposed InterventionStatus = "PROPOSED"
	StatusApproved InterventionStatus = "APPROVED"
	StatusRejected InterventionStatus = "REJECTED"
	StatusExecuted InterventionStatus = "EXECUTED"
)

type Intervention struct {
	ID        string             `json:"id"`
	Type      string             `json:"type"`
	Target    string             `json:"target_id"`
	RiskScore float64            `json:"risk_score"`
	Action    string             `json:"action"`
	Reason    []string           `json:"reason"`
	Status    InterventionStatus `json:"status"`
	CreatedAt int64              `json:"created_at"`
}

type ProposalEngine struct{}

func (e *ProposalEngine) Propose(target string, riskScore float64, reasons []string) Intervention {
	fmt.Printf("🎯 NEXUS-INTERV: Proposing intervention for target %s (Score: %.2f)\n", target, riskScore)

	i := Intervention{
		ID:        fmt.Sprintf("INT-%d", time.Now().UnixNano()),
		Target:    target,
		RiskScore: riskScore,
		Status:    StatusProposed,
		CreatedAt: time.Now().Unix(),
		Reason:    reasons,
	}

	if riskScore > 0.9 {
		i.Type = "HIGH_RISK"
		i.Action = "FREEZE_ACCOUNT"
	} else if riskScore > 0.7 {
		i.Type = "MEDIUM_RISK"
		i.Action = "INVESTIGATE"
	} else {
		i.Type = "LOW_RISK"
		i.Action = "MONITOR"
	}

	return i
}
