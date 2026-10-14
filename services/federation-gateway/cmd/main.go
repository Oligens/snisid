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
	federationgateway "github.com/snisid/platform/services/federation-gateway"
	federation "github.com/snisid/platform/services/federation-gateway/internal/service"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	port := getEnv("PORT", "8089")
	sharedSecret := getEnv("FEDERATION_SECRET", "dev-secret")
	peers := map[string]string{
		"ht": "http://federation-peer-ht:8090",
		"do": "http://federation-peer-do:8090",
	}

	gateway := federation.NewFederationGateway("federation-gw-ht", sharedSecret, peers)
	aggregator, err := federationgateway.NewAggregator("fed-aggregator-1")
	if err != nil {
		logger.Fatal(ctx, "failed to create aggregator", err)
	}

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	r.POST("/v1/federation/events", func(c *gin.Context) {
		var event federation.FederatedEvent
		if err := c.BindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ack, err := gateway.HandleIncoming(event)
		if err != nil {
			c.JSON(http.StatusBadRequest, ack)
			return
		}
		c.JSON(http.StatusOK, ack)
	})

	r.POST("/v1/federation/exchange", func(c *gin.Context) {
		var event federation.FederatedEvent
		if err := c.BindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ack, err := gateway.Exchange(event)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ack)
	})

	r.POST("/v1/federation/aggregate", func(c *gin.Context) {
		var updates []federationgateway.ModelUpdate
		if err := c.BindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		model, err := aggregator.AggregateWeights(updates)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, model)
	})

	r.GET("/v1/federation/reputations", func(c *gin.Context) {
		c.JSON(http.StatusOK, aggregator.GetNodeReputations())
	})

	r.GET("/v1/federation/outbox/process", func(c *gin.Context) {
		gateway.ProcessOutbox()
		c.JSON(http.StatusOK, gin.H{"status": "processed"})
	})

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		logger.Info(ctx, "federation-gateway starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(ctx, "federation-gateway http server failed", err)
		}
	}()

	<-ctx.Done()
	logger.Info(ctx, "federation-gateway shutting down")
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
