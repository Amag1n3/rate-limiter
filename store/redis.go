package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
	window time.Duration
}

func NewRedisStore(addr string, window time.Duration) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr: addr, // e.g. "localhost:6379"
	})
	return &RedisStore{client: client, window: window}
}

func (r *RedisStore) Increment(ctx context.Context, key string) (int64, error) {
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	// Set expiry only on first request (when count == 1)
	if count == 1 {
		r.client.Expire(ctx, key, r.window)
	}
	return count, nil
}

func (r *RedisStore) Get(ctx context.Context, key string) (int64, error) {
	count, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}
