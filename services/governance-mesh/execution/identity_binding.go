package execution

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type IdentityContext struct {
	ServiceIdentity string `json:"service_identity"`
	UserIdentity    string `json:"user_identity"`
	Role            string `json:"role"`
	MTLSVerified    bool   `json:"mtls_verified"`
}

func IdentityBinding() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mock: Extracting identity from SPIFFE SVID or mTLS headers (X-Forwarded-Client-Cert)
		ctx := IdentityContext{
			ServiceIdentity: c.GetHeader("X-Service-ID"),
			UserIdentity:    c.GetHeader("X-User-ID"),
			Role:            c.GetHeader("X-User-Role"),
			MTLSVerified:    true, // Enforced by Istio Envoy sidecar
		}

		if ctx.ServiceIdentity == "" {
			fmt.Println("🚨 NEXUS-MESH: Anonymous service request detected. REJECTING.")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Set("identity", ctx)
		fmt.Printf("🔐 NEXUS-MESH: Request verified. Service=%s, User=%s, Role=%s\n",
			ctx.ServiceIdentity, ctx.UserIdentity, ctx.Role)
		c.Next()
	}
}
