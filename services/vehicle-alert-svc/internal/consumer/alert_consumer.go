package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/vehicle-alert-svc/internal/alerter"
)

type AlertConsumer struct {
	reader  *kafka.Reader
	alerter *alerter.Alerter
	logger  *zap.Logger
}

func New(brokers []string, groupID string, alerter *alerter.Alerter, logger *zap.Logger) *AlertConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    "sivc.alerts",
		MinBytes: 10,
		MaxBytes: 10e6,
		MaxWait:  1 * time.Second,
	})
	return &AlertConsumer{reader: reader, alerter: alerter, logger: logger}
}

func (c *AlertConsumer) Run(ctx context.Context) error {
	c.logger.Info("Démarrage consommateur d'alertes SIVC")
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return fmt.Errorf("fetch message: %w", err)
		}

		switch string(msg.Key) {
		case "alert.created", "alert.updated":
			if err := c.alerter.DispatchAlert(ctx, msg.Value); err != nil {
				c.logger.Error("Erreur dispatch alerte", zap.Error(err))
			}
		case "sighting.alert":
			if err := c.alerter.DispatchSightingAlert(ctx, msg.Value); err != nil {
				c.logger.Error("Erreur dispatch visuel", zap.Error(err))
			}
		default:
			c.logger.Warn("Type d'événement inconnu", zap.ByteString("key", msg.Key))
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("Erreur commit message", zap.Error(err))
		}
	}
}

func (c *AlertConsumer) Close() error {
	return c.reader.Close()
}
