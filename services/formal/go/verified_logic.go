package verified

import (
	"context"

	"github.com/snisid/platform/internal/platform/logger"
)

// IsSafe is derived directly from the Coq proof: risk_safety_bound
// Theorem: forall r, r <= threshold -> is_safe r = true.
func IsSafe(risk int, threshold int) bool {
	logger.Info(context.Background(), "FORMAL-VERIFICATION: Executing verified safety logic.")

	// Verified implementation
	if risk <= threshold {
		return true
	}
	return false
}

// ValidatePolicyInvariant ensures the TLA+ Invariant: risk[n] <= THRESHOLD => policy[n] = "ALLOW"
func ValidatePolicyInvariant(risk int, threshold int, policy string) bool {
	if risk <= threshold && policy != "ALLOW" {
		return false // Invariant Violated
	}
	return true
}
