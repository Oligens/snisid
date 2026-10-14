package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/snisid/platform/internal/platform/logger"
)

func main() {
	broker := getEnv("KAFKA_BROKER", "localhost:9092")
	topic := getEnv("REPLAY_TOPIC", "identity.created")

	// Seeking to 24 hours ago
	startTime := time.Now().Add(-24 * time.Hour)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
	})
	defer reader.Close()

	err := reader.SetOffsetAt(context.Background(), startTime)
	if err != nil {
		logger.Fatal(context.Background(), "failed to set offset", err)
	}

	fmt.Printf("REPLAY: Starting event replay for topic %s from %s\n", topic, startTime.Format(time.RFC3339))

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			break
		}
		fmt.Printf("REPLAY_MSG: %s\n", string(m.Value))

		// Stop if we reached the current time
		if m.Time.After(time.Now().Add(-1 * time.Minute)) {
			fmt.Println("REPLAY: Caught up to real-time. Termination complete.")
			break
		}
	}
}

func getEnv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
