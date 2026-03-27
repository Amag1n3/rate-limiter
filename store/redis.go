package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var incrementWithExpiryScript = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return count
`)

type RedisStore struct {
	client redis.UniversalClient
	window time.Duration
}

func NewRedisStore(addr string, window time.Duration) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr: addr, // e.g. "localhost:6379"
	})
	return NewRedisStoreWithClient(client, window)
}

func NewRedisStoreWithClient(client redis.UniversalClient, window time.Duration) *RedisStore {
	return &RedisStore{client: client, window: window}
}

func (r *RedisStore) Increment(ctx context.Context, key string) (int64, error) {
	return incrementWithExpiryScript.Run(ctx, r.client, []string{key}, r.windowMillis()).Int64()
}

func (r *RedisStore) Get(ctx context.Context, key string) (int64, error) {
	count, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

func (r *RedisStore) Close() error {
	return r.client.Close()
}

func (r *RedisStore) windowMillis() int64 {
	if r.window <= 0 {
		return 1
	}

	windowMillis := r.window.Milliseconds()
	if windowMillis == 0 {
		return 1
	}

	return windowMillis
}
