package alerter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type PushAlerter struct {
	client *redis.Client
}

func NewPushAlerter(client *redis.Client) *PushAlerter {
	return &PushAlerter{client: client}
}

type PushNotification struct {
	AlertID   string `json:"alert_id"`
	Plate     string `json:"plate_number"`
	Level     string `json:"alert_level"`
	Category  string `json:"crime_category"`
	Timestamp string `json:"timestamp"`
}

func (a *PushAlerter) SendPush(ctx context.Context, unit string, alert interface{}) error {
	data, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("sivc:push:%s", unit)
	return a.client.LPush(ctx, key, data).Err()
}

func (a *PushAlerter) GetPendingPushes(ctx context.Context, unit string) ([]string, error) {
	key := fmt.Sprintf("sivc:push:%s", unit)
	return a.client.LRange(ctx, key, 0, -1).Result()
}

func (a *PushAlerter) Start(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
