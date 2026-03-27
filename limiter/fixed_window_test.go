package limiter_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Amag1n3/rate-limiter/limiter"
	"github.com/Amag1n3/rate-limiter/store"
)

func TestFixedWindow_AllowsUpToLimit(t *testing.T) {
	const limit = 5
	fw := limiter.NewFixedWindow(store.NewMemoryStore(10*time.Second), limit)
	ctx := context.Background()

	for i := 1; i <= limit; i++ {
		allowed, err := fw.Allow(ctx, "key")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
		if !allowed {
			t.Fatalf("request %d: expected allowed, got denied", i)
		}
	}
}

func TestFixedWindow_DeniesOverLimit(t *testing.T) {
	const limit = 3
	fw := limiter.NewFixedWindow(store.NewMemoryStore(10*time.Second), limit)
	ctx := context.Background()

	for i := 0; i < limit; i++ {
		fw.Allow(ctx, "key") //nolint
	}

	allowed, err := fw.Allow(ctx, "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("expected request to be denied after limit, but was allowed")
	}
}

func TestFixedWindow_IndependentKeys(t *testing.T) {
	const limit = 2
	fw := limiter.NewFixedWindow(store.NewMemoryStore(10*time.Second), limit)
	ctx := context.Background()

	// Exhaust key-A
	for i := 0; i < limit; i++ {
		fw.Allow(ctx, "key-A") //nolint
	}

	// key-B should still be allowed
	allowed, err := fw.Allow(ctx, "key-B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("key-B should be independent from key-A")
	}
}

func TestFixedWindow_Concurrent(t *testing.T) {
	const limit = 50
	const goroutines = 200

	fw := limiter.NewFixedWindow(store.NewMemoryStore(10*time.Second), limit)
	ctx := context.Background()

	var allowed atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := fw.Allow(ctx, "shared-key")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if ok {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()

	if got := allowed.Load(); got != limit {
		t.Fatalf("expected exactly %d allowed requests, got %d", limit, got)
	}
}
