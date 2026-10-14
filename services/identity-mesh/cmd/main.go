package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/snisid/platform/internal/platform/events"
	"github.com/snisid/platform/internal/platform/logger"
	"github.com/snisid/platform/internal/platform/middleware"
	identitymesh "github.com/snisid/platform/services/identity-mesh"
)

func main() {
	broker := getEnv("KAFKA_BROKER", "localhost:9092")
	neo4jURI := getEnv("NEO4J_URI", "neo4j://localhost:7687")
	neo4jUser := getEnv("NEO4J_USER", "neo4j")
	neo4jPass := getEnv("NEO4J_PASSWORD", "dev_password")
	port := getEnv("PORT", "8084")

	driver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth(neo4jUser, neo4jPass, ""))
	if err != nil {
		log.Fatalf("failed to connect to Neo4j: %v", err)
	}
	defer driver.Close(context.Background())

	if err := driver.VerifyConnectivity(context.Background()); err != nil {
		log.Printf("WARNING: Neo4j not reachable: %v", err)
	}

	mesh := identitymesh.NewIdentityMesh("SNISID-NSIM-01", driver)
	producer := events.NewProducer([]string{broker}, "identity-mesh.events")
	defer producer.Close()

	r := gin.Default()
	r.Use(middleware.RateLimit(30, 60))

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	api := r.Group("/v1/mesh", middleware.Auth(getEnv("JWT_SECRET", "dev-secret")))
	{
		api.POST("/fuse", func(c *gin.Context) {
			var req struct {
				ONI  identitymesh.Record `json:"oni" binding:"required"`
				DGI  identitymesh.Record `json:"dgi" binding:"required"`
				ANH  identitymesh.Record `json:"anh" binding:"required"`
				DCPJ identitymesh.Record `json:"dcpj" binding:"required"`
			}
			if err := c.BindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			fused := mesh.FuseRecords(req.ONI, req.DGI, req.ANH, req.DCPJ)
			c.JSON(http.StatusOK, fused)
		})

		api.POST("/resolve", func(c *gin.Context) {
			var req struct {
				Fused    identitymesh.FusedIdentity `json:"fused" binding:"required"`
				Agency   string                     `json:"agency" binding:"required"`
				Field    string                     `json:"field" binding:"required"`
				Resolved interface{}                `json:"resolved" binding:"required"`
			}
			if err := c.BindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			result := mesh.ResolveConflict(req.Fused, req.Agency, req.Field, req.Resolved)
			c.JSON(http.StatusOK, result)
		})

		api.GET("/identity/:nnu", func(c *gin.Context) {
			nnu := c.Param("nnu")
			identity, err := mesh.GetIdentityGraph(c.Request.Context(), nnu)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, identity)
		})

		api.POST("/inconsistencies", func(c *gin.Context) {
			var fused identitymesh.FusedIdentity
			if err := c.BindJSON(&fused); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			inconsistencies := mesh.DetectInconsistency(fused)
			c.JSON(http.StatusOK, gin.H{"inconsistencies": inconsistencies})
		})
	}

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(context.Background(), "identity-mesh server failed", err)
		}
	}()
	logger.Info(context.Background(), "Identity Mesh started on port "+port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
