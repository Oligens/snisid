package consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/snisid/vehicle-alert-svc/internal/alerter"
)

type KafkaConsumer struct {
	reader *kafka.Reader
	radio  *alerter.RadioAlerter
	sms    *alerter.SMSAlerter
	push   *alerter.PushAlerter
	logger *zap.Logger
}

func NewKafkaConsumer(
	reader *kafka.Reader,
	radio *alerter.RadioAlerter,
	sms *alerter.SMSAlerter,
	push *alerter.PushAlerter,
	logger *zap.Logger,
) *KafkaConsumer {
	return &KafkaConsumer{
		reader: reader,
		radio:  radio,
		sms:    sms,
		push:   push,
		logger: logger,
	}
}

type AlertCreatedEvent struct {
	AlertID       string    `json:"alert_id"`
	PlateNumber   string    `json:"plate_number"`
	CrimeCategory string    `json:"crime_category"`
	AlertLevel    string    `json:"alert_level"`
	ReportingUnit string    `json:"reporting_unit"`
	Timestamp     time.Time `json:"timestamp"`
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.logger.Error("Failed to read message", zap.Error(err))
				continue
			}

			var event AlertCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				c.logger.Error("Failed to unmarshal event", zap.Error(err))
				continue
			}

			c.processAlert(ctx, event)
		}
	}
}

func (c *KafkaConsumer) processAlert(ctx context.Context, event AlertCreatedEvent) {
	c.logger.Info("Processing alert",
		zap.String("alert_id", event.AlertID),
		zap.String("plate", event.PlateNumber),
		zap.String("level", event.AlertLevel),
	)

	units := c.getUnitsForLevel(event.AlertLevel)

	for _, unit := range units {
		if err := c.radio.BroadcastToUnit(ctx, unit, event); err != nil {
			c.logger.Error("Failed to broadcast to radio",
				zap.String("unit", unit),
				zap.Error(err),
			)
		}

		if err := c.push.SendPush(ctx, unit, event); err != nil {
			c.logger.Error("Failed to send push",
				zap.String("unit", unit),
				zap.Error(err),
			)
		}
	}
}

func (c *KafkaConsumer) getUnitsForLevel(level string) []string {
	switch level {
	case "CRITICAL":
		return []string{"BRI", "GIPNH", "DCPJ", "CAE", "BLVV", "BLTS", "BAC"}
	case "WANTED":
		return []string{"BLVV", "BLTS", "BAC", "DCPJ", "BRI"}
	case "CAUTION":
		return []string{"BLVV", "BAC"}
	default:
		return []string{"BAC"}
	}
}
