package fraud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type StateStore interface {
	IncrementVelocity(ctx context.Context, userID string) (int64, error)
	GetState(ctx context.Context, userID string) (*FraudState, error)
	SetState(ctx context.Context, userID string, state *FraudState) error
}

type FraudState struct {
	Velocity     int64     `json:"velocity"`
	LastAmount   float64   `json:"last_amount"`
	LastActivity time.Time `json:"last_activity"`
	WindowStart  time.Time `json:"window_start"`
}

type RedisStateStore struct {
	client *redis.Client
}

func NewRedisStateStore(client *redis.Client) *RedisStateStore {
	return &RedisStateStore{client: client}
}

func NewStateStore(addr string) *RedisStateStore {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return NewRedisStateStore(client)
}

func (s *RedisStateStore) IncrementVelocity(ctx context.Context, userID string) (int64, error) {
	key := fmt.Sprintf("snisid:fraud:%s:velocity", userID)
	val, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis incr velocity: %w", err)
	}
	s.client.Expire(ctx, key, time.Hour)
	return val, nil
}

func (s *RedisStateStore) GetState(ctx context.Context, userID string) (*FraudState, error) {
	key := fmt.Sprintf("snisid:fraud:%s:state", userID)
	val, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return &FraudState{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get state: %w", err)
	}
	_ = val
	return &FraudState{}, nil
}

func (s *RedisStateStore) SetState(ctx context.Context, userID string, state *FraudState) error {
	key := fmt.Sprintf("snisid:fraud:%s:state", userID)
	return s.client.Set(ctx, key, "state", time.Hour).Err()
}
