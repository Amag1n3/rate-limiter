package limiter

import (
	"context"
	"rate_limiter/store"
)

type FixedWindow struct {
	store store.Store
	limit int64
}

func NewFixedWindow(store store.Store, limit int64) *FixedWindow {
	return &FixedWindow{
		store: store,
		limit: limit,
	}
}

func (fw *FixedWindow) Allow(ctx context.Context, key string) (bool, error) {
	count, err := fw.store.Increment(ctx, key)
	if err != nil {
		return false, err
	}
	return count <= fw.limit, nil
}
