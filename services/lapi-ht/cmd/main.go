package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/snisid/lapi-ht/internal/handler"
	"github.com/snisid/lapi-ht/internal/kafka"
	"github.com/snisid/lapi-ht/internal/repository"
	"github.com/snisid/lapi-ht/internal/service"
)

func main() {
	dsn := getEnv("DATABASE_URL", "postgres://localhost:5432/snisid_lapi?sslmode=disable")
	brokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	port := getEnv("PORT", "8096")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	producer := kafka.NewProducer(brokers, "snisid.lapi.alerts")
	defer producer.Close()

	repo := repository.NewPostgresRepository(db)
	svc := service.NewLapiService(repo, producer)
	h := handler.NewHandler(svc)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	h.RegisterRoutes(r)

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
