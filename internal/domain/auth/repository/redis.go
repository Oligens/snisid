package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/snisid/platform/internal/domain/auth/entity"
)

type SessionRepository interface {
	StoreSession(ctx context.Context, session *entity.Session) error
	GetSession(ctx context.Context, sessionID string) (*entity.Session, error)
	RevokeSessionFamily(ctx context.Context, userID string) error

	IncrementFailedAttempts(ctx context.Context, ip, username string) (int, error)
	ResetFailedAttempts(ctx context.Context, ip, username string) error
}

type redisSessionRepo struct {
	client *redis.Client
}

func NewRedisSessionRepository(client *redis.Client) SessionRepository {
	return &redisSessionRepo{client: client}
}

func (r *redisSessionRepo) StoreSession(ctx context.Context, session *entity.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("auth:session:%s", session.SessionID)
	ttl := time.Until(session.ExpiresAt)
	if ttl < 0 {
		return nil // Already expired
	}

	// Keep a set of sessions per user for family revocation
	pipe := r.client.Pipeline()
	pipe.Set(ctx, key, data, ttl)
	pipe.SAdd(ctx, fmt.Sprintf("auth:user_sessions:%s", session.UserID), session.SessionID)
	// Optionally set expiry on the user_sessions set as well to prevent unbounded growth
	pipe.Expire(ctx, fmt.Sprintf("auth:user_sessions:%s", session.UserID), 30*24*time.Hour) // 30 days max

	_, err = pipe.Exec(ctx)
	return err
}

func (r *redisSessionRepo) GetSession(ctx context.Context, sessionID string) (*entity.Session, error) {
	key := fmt.Sprintf("auth:session:%s", sessionID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err // redis.Nil if not found
	}

	var session entity.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *redisSessionRepo) RevokeSessionFamily(ctx context.Context, userID string) error {
	setKey := fmt.Sprintf("auth:user_sessions:%s", userID)
	sessions, err := r.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	for _, sid := range sessions {
		pipe.Del(ctx, fmt.Sprintf("auth:session:%s", sid))
	}
	pipe.Del(ctx, setKey)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *redisSessionRepo) IncrementFailedAttempts(ctx context.Context, ip, username string) (int, error) {
	key := fmt.Sprintf("auth:bruteforce:%s:%s", ip, username)
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		// Set window for 15 minutes
		r.client.Expire(ctx, key, 15*time.Minute)
	}
	return int(count), nil
}

func (r *redisSessionRepo) ResetFailedAttempts(ctx context.Context, ip, username string) error {
	key := fmt.Sprintf("auth:bruteforce:%s:%s", ip, username)
	return r.client.Del(ctx, key).Err()
}
