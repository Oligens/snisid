package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/snisid/platform/internal/domain/audit/handler"
	"github.com/snisid/platform/internal/domain/audit/repository"
	"github.com/snisid/platform/internal/domain/audit/usecase"
	"github.com/snisid/platform/internal/platform/events"
	"github.com/snisid/platform/internal/platform/logger"
	"github.com/snisid/platform/internal/platform/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/snisid/platform/internal/domain/audit/entity"
)

func main() {
	port := getEnv("PORT", "8085")
	dbURL := getEnv("DATABASE_URL", "host=localhost user=snisid password=snisid dbname=snisid port=5432 sslmode=disable")
	broker := getEnv("KAFKA_BROKER", "localhost:9092")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret")

	// Init PostgreSQL
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		logger.Fatal(context.Background(), "failed to connect database", err)
	}

	if getEnv("ENV", "dev") == "dev" {
		db.AutoMigrate(&entity.AuditEvent{})
	}

	// Init Kafka Consumer
	consumer := events.NewConsumer([]string{broker}, "audit-group", "audit.events")
	defer consumer.Close()

	// Dependencies
	postgresRepo := repository.NewPostgresAuditRepository(db)
	ingester := usecase.NewKafkaIngester(postgresRepo, consumer)
	forensics := usecase.NewForensicsService(postgresRepo)
	httpHandler := handler.NewHttpHandler(forensics)

	// Start Background Ingester
	go ingester.Start(context.Background())
	logger.Info(context.Background(), "audit kafka ingester started")

	// Router
	r := gin.Default()
	r.Use(middleware.RateLimit(50, 100))

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	api := r.Group("/v1/audit", middleware.Auth(jwtSecret))
	httpHandler.RegisterRoutes(api)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(context.Background(), "failed to run audit service", err)
		}
	}()

	logger.Info(context.Background(), "audit-service api started on port "+port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info(context.Background(), "shutting down audit-service...")
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
