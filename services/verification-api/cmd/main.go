package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/snisid/platform/internal/platform/events"
	"github.com/snisid/platform/internal/platform/logger"
	"github.com/snisid/platform/internal/service/verification"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	brokers := []string{os.Getenv("KAFKA_BROKERS")}
	if len(brokers[0]) == 0 {
		brokers = []string{"localhost:9092"}
	}

	// 1. Initialize Connectors
	biometric := &verification.MockBiometricConnector{}
	passport := &verification.MockAgencyConnector{AgencyName: "passport-office"}
	police := &verification.MockAgencyConnector{AgencyName: "national-police"}

	// 2. Initialize Orchestrator
	orch := verification.NewOrchestrator(biometric, passport, police)

	// 3. Initialize Kafka Producer for results
	producer := events.NewProducer(brokers, "snisid.prod.identity.v1.verification")

	logger.Info(context.Background(), "SNISID Identity Verification Service starting...")

	// 4. Start consumer for incoming verification requests
	consumer := events.NewConsumer(brokers, "verification-group", "snisid.prod.identity.v1.requests")

	err := consumer.Start(ctx, func(ctx context.Context, payload []byte) error {
		// Mock request processing
		var req map[string]interface{}
		if err := json.Unmarshal(payload, &req); err != nil {
			return err
		}

		results, err := orch.VerifyIdentity(ctx, req)
		if err != nil {
			return err
		}

		verificationID := uuid.New().String()
		logger.Info(ctx, "Verification complete",
			zap.String("id", verificationID),
			zap.Any("results", results),
		)

		// Publish outcome
		return producer.Publish(ctx, verificationID, results)
	})

	if err != nil {
		logger.Fatal(context.Background(), "Verification service crashed", err)
	}

	logger.Info(context.Background(), "SNISID Identity Verification Service shutting down")
}
