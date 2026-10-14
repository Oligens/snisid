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
	"github.com/snisid/platform/services/governance-formal/compiler"
	"github.com/snisid/platform/services/governance-formal/graph"
	"github.com/snisid/platform/services/governance-formal/kernel"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	port := getEnv("PORT", "8092")
	lattice := graph.NewGlobalPolicyLattice()
	k := &kernel.GovernanceKernel{}

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	r.POST("/v1/compile", func(c *gin.Context) {
		var ast compiler.PolicyAST
		if err := c.BindJSON(&ast); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		policy := compiler.Compile(ast)
		c.JSON(http.StatusOK, gin.H{"compiled": true, "name": ast.Name})
		_ = policy
	})

	r.POST("/v1/lattice/country", func(c *gin.Context) {
		var req struct {
			Country string                  `json:"country"`
			Policy  compiler.CompiledPolicy `json:"-"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		lattice.AddCountry(req.Country, req.Policy)
		c.JSON(http.StatusOK, gin.H{"status": "added"})
	})

	r.POST("/v1/lattice/prove", func(c *gin.Context) {
		var req struct {
			From string `json:"from"`
			To   string `json:"to"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		level, violations, err := lattice.ProveCompatibility(req.From, req.To)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"compatibility": level, "violations": violations})
	})

	r.GET("/v1/lattice/peers/:country", func(c *gin.Context) {
		country := c.Param("country")
		peers := lattice.GetCompatiblePeers(country, graph.Compatible)
		c.JSON(http.StatusOK, gin.H{"peers": peers})
	})

	r.POST("/v1/kernel/execute", func(c *gin.Context) {
		var req struct {
			State  compiler.State  `json:"state"`
			Action compiler.Action `json:"action"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		newState, err := k.Execute(req.State, req.Action)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "state": req.State})
			return
		}
		c.JSON(http.StatusOK, newState)
	})

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		logger.Info(ctx, "governance-formal starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(ctx, "governance-formal http server failed", err)
		}
	}()

	<-ctx.Done()
	logger.Info(ctx, "governance-formal shutting down")
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
