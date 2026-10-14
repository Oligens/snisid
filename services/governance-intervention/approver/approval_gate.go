package approver

import (
	"fmt"
	"github.com/snisid/platform/governance-intervention/proposer"
)

type ApprovalGate struct{}

func (g *ApprovalGate) Authorize(i *proposer.Intervention, userRole string, justification string) error {
	if justification == "" {
		return fmt.Errorf("JUSTIFICATION_REQUIRED")
	}

	// Policy: Only investigators can approve FREEZE_ACCOUNT
	if i.Action == "FREEZE_ACCOUNT" && userRole != "investigator" && userRole != "admin" {
		return fmt.Errorf("INSUFFICIENT_PRIVILEGES_FOR_ACTION")
	}

	i.Status = proposer.StatusApproved
	fmt.Printf("✅ NEXUS-INTERV: Intervention %s approved by role %s. Justification: %s\n",
		i.ID, userRole, justification)

	return nil
}

func (g *ApprovalGate) Reject(i *proposer.Intervention, reason string) {
	i.Status = proposer.StatusRejected
	fmt.Printf("❌ NEXUS-INTERV: Intervention %s rejected. Reason: %s\n", i.ID, reason)
}
