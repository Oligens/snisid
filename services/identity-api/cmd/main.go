package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/snisid/platform/services/identity-api/internal/handlers"
	"github.com/snisid/platform/services/identity-api/internal/kafka"
	"github.com/snisid/platform/services/identity-api/internal/models"
)

type jwtClaims struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
}

func main() {
	env := getEnv("ENV", "dev")
	jwtSecret, err := requiredSecret("JWT_SECRET", env)
	if err != nil { log.Fatal(err) }
	broker := getEnv("KAFKA_BROKER", getEnv("KAFKA_BROKERS", "localhost:9092"))
	dbURL, err := requiredDatabaseURL(env)
	if err != nil { log.Fatal(err) }
	port := getEnv("PORT", "8081")

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil { log.Fatalf("failed to connect database: %v", err) }

	if env == "dev" || env == "development" {
		if err := db.AutoMigrate(&models.Identity{}, &models.IdentityHistory{}, &models.BiometricReference{}, &models.DocumentAssociation{}); err != nil {
			log.Fatalf("failed to migrate: %v", err)
		}
	}

	topic := getEnv("KAFKA_TOPIC", "snisid.prod.identity.v1.events")
	producer := kafka.NewProducer([]string{broker}, topic)
	defer producer.Close()

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	h := handlers.New(db, producer)
	api := r.Group("/api/v1")
	api.Use(authMiddleware(jwtSecret, env))
	h.RegisterRoutes(api)

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		log.Printf("identity-service started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run identity service: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("shutting down identity-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil { log.Printf("server forced to shutdown: %v", err) }
}

func authMiddleware(secret, env string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			if env == "dev" || env == "development" {
				if actorID := c.GetHeader("X-Actor-ID"); actorID != "" {
					c.Set("actor_id", actorID); c.Next(); return
				}
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"}); return
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization scheme"}); return
		}
		claims, err := verifyHS256JWT(strings.TrimSpace(strings.TrimPrefix(auth, prefix)), secret)
		if err != nil || claims.Sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"}); return
		}
		c.Set("actor_id", claims.Sub)
		c.Next()
	}
}

func verifyHS256JWT(token, secret string) (*jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 { return nil, errors.New("malformed jwt") }
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signingInput))
	expected := mac.Sum(nil)
	got, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || subtle.ConstantTimeCompare(expected, got) != 1 { return nil, errors.New("invalid signature") }
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil { return nil, errors.New("invalid payload") }
	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil { return nil, err }
	if claims.Exp > 0 && time.Now().Unix() >= claims.Exp { return nil, errors.New("expired token") }
	return &claims, nil
}

func requiredSecret(key, env string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" || value == "dev-secret" || value == "your-jwt-secret" {
		if env == "dev" || env == "development" { return "dev-secret", nil }
		return "", fmt.Errorf("%s must be set to a strong secret outside development", key)
	}
	if len(value) < 32 { return "", fmt.Errorf("%s must contain at least 32 characters", key) }
	return value, nil
}

func requiredDatabaseURL(env string) (string, error) {
	if value := strings.TrimSpace(os.Getenv("DATABASE_URL")); value != "" { return value, nil }
	if env == "dev" || env == "development" {
		return "host=localhost user=snisid password=snisid dbname=snisid port=5432 sslmode=disable", nil
	}
	return "", errors.New("DATABASE_URL is required outside development")
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}
