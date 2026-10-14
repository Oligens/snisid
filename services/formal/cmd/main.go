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
	formal_verified "github.com/snisid/platform/services/formal/go"
	"github.com/snisid/platform/services/formal/healing"
	"github.com/snisid/platform/services/formal/runtime"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	port := getEnv("PORT", "8090")
	mon := &runtime.FormalMonitor{Threshold: 100}
	healingEngine := &healing.HealingEngine{}

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	r.POST("/v1/validate", func(c *gin.Context) {
		var req struct {
			Risk      int `json:"risk"`
			Threshold int `json:"threshold"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		safe := verified.IsSafe(req.Risk, req.Threshold)
		c.JSON(http.StatusOK, gin.H{"safe": safe})
	})

	r.POST("/v1/validate-policy", func(c *gin.Context) {
		var req struct {
			Risk      int    `json:"risk"`
			Threshold int    `json:"threshold"`
			Policy    string `json:"policy"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		valid := verified.ValidatePolicyInvariant(req.Risk, req.Threshold, req.Policy)
		c.JSON(http.StatusOK, gin.H{"valid": valid})
	})

	r.POST("/v1/monitor", func(c *gin.Context) {
		var event runtime.SystemEvent
		if err := c.BindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		valid := mon.ValidateEvent(event)
		if !valid {
			mon.TriggerEmergencyResponse(event)
		}
		c.JSON(http.StatusOK, gin.H{"valid": valid})
	})

	r.POST("/v1/heal", func(c *gin.Context) {
		var state healing.SystemState
		if err := c.BindJSON(&state); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		healingEngine.DetectAndHeal(&state, true)
		c.JSON(http.StatusOK, gin.H{"status": "checked"})
	})

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		logger.Info(ctx, "formal starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(ctx, "formal http server failed", err)
		}
	}()

	<-ctx.Done()
	logger.Info(ctx, "formal shutting down")
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
