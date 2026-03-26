package store

import "context"

type Store interface {
	// Increment adds 1 to the counter for a given key within a window.
	// Returns the new count after incrementing.
	Increment(ctx context.Context, key string) (int64, error)

	// Get returns the current count for a key.
	Get(ctx context.Context, key string) (int64, error)
}
