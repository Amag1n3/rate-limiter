package limiter

import (
	"context"
	"math"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Amag1n3/rate-limiter/store"
)

func BenchmarkFixedWindowAllow(b *testing.B) {
	ctx := context.Background()
	rateLimiter := NewFixedWindow(store.NewMemoryStore(time.Minute), math.MaxInt64)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := rateLimiter.Allow(ctx, "fixed-window-client"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFixedWindowAllowParallel(b *testing.B) {
	ctx := context.Background()
	rateLimiter := NewFixedWindow(store.NewMemoryStore(time.Minute), math.MaxInt64)
	var goroutineID uint64

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		key := "fixed-window-client-" + strconv.FormatUint(atomic.AddUint64(&goroutineID, 1), 10)
		for pb.Next() {
			if _, err := rateLimiter.Allow(ctx, key); err != nil {
				panic(err)
			}
		}
	})
}

func BenchmarkTokenBucketAllow(b *testing.B) {
	ctx := context.Background()
	rateLimiter := NewTokenBucket(b.N+1, time.Hour)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := rateLimiter.Allow(ctx, "token-bucket-client"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTokenBucketAllowParallel(b *testing.B) {
	ctx := context.Background()
	rateLimiter := NewTokenBucket((2*b.N)+1, time.Hour)
	var goroutineID uint64

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		key := "token-bucket-client-" + strconv.FormatUint(atomic.AddUint64(&goroutineID, 1), 10)
		for pb.Next() {
			if _, err := rateLimiter.Allow(ctx, key); err != nil {
				panic(err)
			}
		}
	})
}
