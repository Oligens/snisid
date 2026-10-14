package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	kafkago "github.com/segmentio/kafka-go"
	snkafka "github.com/snisid/platform/pkg/kafka"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		logger.Fatal("KAFKA_BROKERS environment variable is required")
	}

	cfg := snkafka.Config{
		Brokers: brokers,
		Topic:   os.Getenv("KAFKA_TOPIC"),
		GroupID: os.Getenv("KAFKA_GROUP_ID"),
	}

	consumer := snkafka.NewConsumer(cfg, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
	}()

	consumer.Start(ctx, func(ctx context.Context, msg kafkago.Message) error {
		fmt.Println("Received risk event:", string(msg.Value))
		return nil
	})
}
