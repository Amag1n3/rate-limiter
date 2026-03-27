# rate-limiter

A small Go rate limiter project with two algorithms, pluggable storage for fixed windows, HTTP middleware, Docker packaging, and benchmarks.

## Features

- Fixed window limiter backed by an in-memory store or Redis
- Token bucket limiter for burst-friendly throttling
- HTTP middleware for per-route, per-client enforcement
- Runnable demo server with `/api/fixed`, `/api/token`, and `/healthz`
- Benchmarks for the limiter implementations

## Project Layout

```text
.
├── config/       # Route configuration for the demo server
├── limiter/      # Rate limiting algorithms + benchmarks
├── middleware/   # net/http middleware
├── store/        # Store interface, memory store, Redis store
└── main.go       # Demo application entrypoint
```

## Run Locally

```bash
go run .
```

The server listens on `:8080` by default.

### Endpoints

- `GET /healthz`
- `GET /api/fixed`
- `GET /api/token`

### Try It

Fixed window route, limited to `5` requests per `10s`:

```bash
for i in $(seq 1 7); do curl -i http://localhost:8080/api/fixed; done
```

Token bucket route, capacity `10`, refilling `1 token/sec`:

```bash
for i in $(seq 1 12); do curl -i http://localhost:8080/api/token; done
```

## Redis Store

The fixed window route can use Redis instead of the in-memory store:

```bash
RATE_LIMIT_STORE=redis REDIS_ADDR=localhost:6379 go run .
```

Environment variables:

- `PORT`: HTTP port for the demo server. Default: `8080`
- `RATE_LIMIT_STORE`: `memory` or `redis`. Default: `memory`
- `REDIS_ADDR`: Redis address used when `RATE_LIMIT_STORE=redis`. Default: `localhost:6379`

The Redis implementation performs the increment and TTL setup atomically, which avoids race conditions around first-write expiration.

## Docker

Build the image:

```bash
docker build -t rate-limiter .
```

Run with the in-memory store:

```bash
docker run --rm -p 8080:8080 rate-limiter
```

Run with Redis:

```bash
docker run --rm -p 8080:8080 \
  -e RATE_LIMIT_STORE=redis \
  -e REDIS_ADDR=host.docker.internal:6379 \
  rate-limiter
```

## Benchmarks

Run all benchmarks with allocation stats:

```bash
go test -bench=. -benchmem ./...
```

## Package Example

```go
package main

import (
	"net/http"
	"time"

	"github.com/Amag1n3/rate-limiter/limiter"
	"github.com/Amag1n3/rate-limiter/middleware"
	"github.com/Amag1n3/rate-limiter/store"
)

func main() {
	fixedWindow := limiter.NewFixedWindow(store.NewMemoryStore(10*time.Second), 5)

	mux := http.NewServeMux()
	mux.Handle("/api", middleware.RateLimit(fixedWindow)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})))

	http.ListenAndServe(":8080", mux)
}
```
