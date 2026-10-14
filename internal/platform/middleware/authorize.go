package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

// Authorize is a middleware hook for other microservices to enforce access.
// In a real scenario, this would use a fast gRPC client. For simplicity, we use HTTP here.
func Authorize(authzServiceURL string, action string, resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract subject from Context (previously set by Auth middleware)
		userID, ok := c.Request.Context().Value(userContextKey{}).(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		rolesIntf, _ := c.Request.Context().Value(roleContextKey{}).([]interface{})
		var roles []string
		for _, r := range rolesIntf {
			roles = append(roles, r.(string))
		}

		// Build Authorization Request
		reqBody := map[string]interface{}{
			"subject": map[string]interface{}{
				"userId": userID,
				"roles":  roles,
			},
			"action":   action,
			"resource": resource,
		}

		payload, _ := json.Marshal(reqBody)

		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Post(authzServiceURL+"/v1/authz/enforce", "application/json", bytes.NewBuffer(payload))
		if err != nil || resp.StatusCode != http.StatusOK {
			logger.Error(context.Background(), "failed to call authz service", err, zap.String("url", authzServiceURL))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "authorization service unavailable"})
			return
		}
		defer resp.Body.Close()

		var decision struct {
			Allowed bool   `json:"allowed"`
			Reason  string `json:"reason"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&decision); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid authz response"})
			return
		}

		if !decision.Allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden", "reason": decision.Reason})
			return
		}

		c.Next()
	}
}
