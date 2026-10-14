package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/snisid/platform/internal/platform/logger"
	apiplatform "github.com/snisid/platform/services/api-platform/tenancy"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	port := getEnv("PORT", "8083")
	manager := &apiplatform.TenantManager{
		Tenants: make(map[string]apiplatform.Tenant),
	}

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	r.POST("/v1/validate", func(c *gin.Context) {
		var req struct {
			APIKey   string                 `json:"api_key"`
			Required apiplatform.Permission `json:"required"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ok, reason := manager.ValidateAccess(req.APIKey, req.Required)
		c.JSON(http.StatusOK, gin.H{"allowed": ok, "reason": reason})
	})

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		logger.Info(ctx, "api-platform starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(ctx, "api-platform http server failed", err)
		}
	}()

	<-ctx.Done()
	logger.Info(ctx, "api-platform shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
