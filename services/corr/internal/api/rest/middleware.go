package rest

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const (
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyUserUnit contextKey = "user_unit"
	ContextKeyUserRole contextKey = "user_role"
)

func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if uid := c.GetHeader("X-User-ID"); uid != "" {
			if parsed, err := uuid.Parse(uid); err == nil {
				ctx = context.WithValue(ctx, ContextKeyUserID, parsed)
			}
		}
		if unit := c.GetHeader("X-User-Unit"); unit != "" {
			ctx = context.WithValue(ctx, ContextKeyUserUnit, unit)
		}
		if role := c.GetHeader("X-User-Role"); role != "" {
			ctx = context.WithValue(ctx, ContextKeyUserRole, role)
		}
		c.Request = c.Request.WithContext(ctx)
		start := time.Now()
		c.Next()
		_ = start
	}
}
