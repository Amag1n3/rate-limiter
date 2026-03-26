package config

import "time"

type RouteConfig struct {
	Limit     int64
	Window    time.Duration
	Algorithm string // "fixed_window" or "token_bucket"
}

var Routes = map[string]RouteConfig{
	"/api/fixed": {
		Limit:     5,
		Window:    10 * time.Second,
		Algorithm: "fixed_window",
	},
	"/api/token": {
		Limit:     10,
		Window:    1 * time.Second,
		Algorithm: "token_bucket",
	},
}
