package store

import (
	"context"
	"sync"
	"time"
)

type entry struct {
	count     int64
	expiresAt time.Time
}

type MemoryStore struct {
	mu      sync.Mutex
	entries map[string]entry
	window  time.Duration
}

func NewMemoryStore(window time.Duration) *MemoryStore {
	return &MemoryStore{
		entries: make(map[string]entry),
		window:  window,
	}
}

func (m *MemoryStore) Increment(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	e, exists := m.entries[key]

	if !exists || now.After(e.expiresAt) {
		// Fresh window
		m.entries[key] = entry{count: 1, expiresAt: now.Add(m.window)}
		return 1, nil
	}

	e.count++
	m.entries[key] = e
	return e.count, nil
}

func (m *MemoryStore) Get(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	e, exists := m.entries[key]
	if !exists || time.Now().After(e.expiresAt) {
		return 0, nil
	}
	return e.count, nil
}
