package limiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/Amag1n3/rate-limiter/limiter"
)

func TestTokenBucket_AllowsUpToCapacity(t *testing.T) {
	const capacity = 5
	tb := limiter.NewTokenBucket(capacity, time.Second)
	ctx := context.Background()

	for i := 1; i <= capacity; i++ {
		allowed, err := tb.Allow(ctx, "key")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
		if !allowed {
			t.Fatalf("request %d: expected allowed within capacity, got denied", i)
		}
	}
}

func TestTokenBucket_DeniesWhenEmpty(t *testing.T) {
	const capacity = 3
	tb := limiter.NewTokenBucket(capacity, time.Hour) // very slow refill
	ctx := context.Background()

	for i := 0; i < capacity; i++ {
		tb.Allow(ctx, "key") //nolint
	}

	allowed, err := tb.Allow(ctx, "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("expected denial after bucket empty, but was allowed")
	}
}

func TestTokenBucket_RefillsOverTime(t *testing.T) {
	const capacity = 5
	refillRate := 50 * time.Millisecond

	tb := limiter.NewTokenBucket(capacity, refillRate)
	ctx := context.Background()

	// Drain the bucket
	for i := 0; i < capacity; i++ {
		tb.Allow(ctx, "key") //nolint
	}

	// Confirm it's empty
	denied, _ := tb.Allow(ctx, "key")
	if denied {
		t.Fatal("bucket should be empty at this point")
	}

	// Wait for 3 tokens to refill
	time.Sleep(3 * refillRate)

	var allowed int
	for i := 0; i < 3; i++ {
		ok, err := tb.Allow(ctx, "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			allowed++
		}
	}
	if allowed == 0 {
		t.Fatal("expected tokens to have refilled after waiting, but all denied")
	}
}

func TestTokenBucket_ZeroCapacityDeniesAll(t *testing.T) {
	tb := limiter.NewTokenBucket(0, time.Second)
	ctx := context.Background()

	allowed, err := tb.Allow(ctx, "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("zero-capacity bucket should deny all requests")
	}
}

func TestTokenBucket_IndependentKeys(t *testing.T) {
	const capacity = 2
	tb := limiter.NewTokenBucket(capacity, time.Hour)
	ctx := context.Background()

	for i := 0; i < capacity; i++ {
		tb.Allow(ctx, "key-A") //nolint
	}

	allowed, err := tb.Allow(ctx, "key-B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("key-B should be independent from key-A")
	}
}
