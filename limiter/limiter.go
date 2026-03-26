package limiter

import "context"

// Limiter is the interface every algorithm must implement.
type Limiter interface {
	// Allow returns true if the request should be permitted.
	Allow(ctx context.Context, key string) (bool, error)
}

// Config holds the rules for a single route.
type Config struct {
	RequestsPerWindow int
	Algorithm         string // "fixed_window" or "token_bucket"
}
