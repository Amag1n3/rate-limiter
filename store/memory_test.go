package store_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Amag1n3/rate-limiter/store"
)

func TestMemoryStore_IncrementSequential(t *testing.T) {
	ms := store.NewMemoryStore(10 * time.Second)
	ctx := context.Background()

	for i := int64(1); i <= 5; i++ {
		count, err := ms.Increment(ctx, "key")
		if err != nil {
			t.Fatalf("increment %d: unexpected error: %v", i, err)
		}
		if count != i {
			t.Fatalf("expected count %d, got %d", i, count)
		}
	}
}

func TestMemoryStore_GetReturnsCurrentCount(t *testing.T) {
	ms := store.NewMemoryStore(10 * time.Second)
	ctx := context.Background()

	ms.Increment(ctx, "key") //nolint
	ms.Increment(ctx, "key") //nolint

	count, err := ms.Get(ctx, "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestMemoryStore_GetMissingKeyReturnsZero(t *testing.T) {
	ms := store.NewMemoryStore(10 * time.Second)
	ctx := context.Background()

	count, err := ms.Get(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 for missing key, got %d", count)
	}
}

func TestMemoryStore_WindowExpiry(t *testing.T) {
	window := 50 * time.Millisecond
	ms := store.NewMemoryStore(window)
	ctx := context.Background()

	ms.Increment(ctx, "key") //nolint
	ms.Increment(ctx, "key") //nolint

	time.Sleep(window + 10*time.Millisecond)

	// After expiry, should start a fresh window from 1
	count, err := ms.Increment(ctx, "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count to reset to 1 after window expiry, got %d", count)
	}
}

func TestMemoryStore_ConcurrentIncrements(t *testing.T) {
	const goroutines = 100
	ms := store.NewMemoryStore(10 * time.Second)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ms.Increment(ctx, "shared") //nolint
		}()
	}
	wg.Wait()

	count, err := ms.Get(ctx, "shared")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != goroutines {
		t.Fatalf("expected %d increments, got %d (possible race condition)", goroutines, count)
	}
}
