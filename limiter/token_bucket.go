package limiter

import (
	"context"
	"sync"
	"time"
)

type TokenBucket struct {
	mu         sync.Mutex
	buckets    map[string]*bucket
	capacity   int
	refillRate time.Duration
}

type bucket struct {
	tokens     int
	lastRefill time.Time
}

func NewTokenBucket(capacity int, refillRate time.Duration) *TokenBucket {
	return &TokenBucket{
		buckets:    make(map[string]*bucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, error) {
	if tb.capacity <= 0 {
		return false, nil
	}

	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	b, exists := tb.buckets[key]

	if !exists {
		// First request from this client — give them a full bucket minus 1
		tb.buckets[key] = &bucket{
			tokens:     tb.capacity - 1,
			lastRefill: now,
		}
		return true, nil
	}

	// Refill tokens based on how much time has passed
	elapsed := now.Sub(b.lastRefill)
	tokensToAdd := int(elapsed / tb.refillRate)

	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > tb.capacity {
			b.tokens = tb.capacity
		}
		b.lastRefill = b.lastRefill.Add(time.Duration(tokensToAdd) * tb.refillRate)
	}

	// Check if request can be served
	if b.tokens <= 0 {
		return false, nil
	}

	b.tokens--
	return true, nil
}
