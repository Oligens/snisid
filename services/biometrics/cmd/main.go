package main

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/snisid/platform/internal/platform/logger"
	"github.com/snisid/platform/services/biometrics"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	port := getEnv("PORT", "8085")
	engine := biometrics.NewBiometricsEngine("v2.0")

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	r.POST("/v1/match-face", func(c *gin.Context) {
		var req struct {
			Image    string                       `json:"image"`
			Enrolled []biometrics.BiometricVector `json:"enrolled"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		image, err := base64.StdEncoding.DecodeString(req.Image)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 image"})
			return
		}
		result := engine.MatchFace(image, req.Enrolled)
		c.JSON(http.StatusOK, result)
	})

	r.POST("/v1/verify-liveness", func(c *gin.Context) {
		var req struct {
			Image string `json:"image"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		image, _ := base64.StdEncoding.DecodeString(req.Image)
		result := engine.VerifyLiveness(image)
		c.JSON(http.StatusOK, result)
	})

	r.POST("/v1/match-fingerprint", func(c *gin.Context) {
		var req struct {
			Template string   `json:"template"`
			Enrolled []string `json:"enrolled"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		tmpl, _ := base64.StdEncoding.DecodeString(req.Template)
		var enrolled [][]byte
		for _, e := range req.Enrolled {
			b, _ := base64.StdEncoding.DecodeString(e)
			enrolled = append(enrolled, b)
		}
		result := engine.MatchFingerprint(tmpl, enrolled)
		c.JSON(http.StatusOK, result)
	})

	r.POST("/v1/enroll-face", func(c *gin.Context) {
		var req struct {
			CitizenID string `json:"citizen_id"`
			Image     string `json:"image"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		image, _ := base64.StdEncoding.DecodeString(req.Image)
		enrollment, err := engine.EnrollFace(req.CitizenID, image)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, enrollment)
	})

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		logger.Info(ctx, "biometrics starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(ctx, "biometrics http server failed", err)
		}
	}()

	<-ctx.Done()
	logger.Info(ctx, "biometrics shutting down")
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
