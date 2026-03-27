# rate-limiter

A production-grade HTTP rate limiting library in Go — pluggable algorithms, Redis-backed distributed storage, per-client enforcement, and zero external framework dependencies.

[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Amag1n3/rate-limiter)](https://goreportcard.com/report/github.com/Amag1n3/rate-limiter)

---

## Features

- **Two algorithms** — Fixed Window and Token Bucket, each suited to different traffic patterns
- **Pluggable storage** — swap between in-memory and Redis without changing application code
- **Atomic Redis operations** — Lua script ensures INCR + EXPIRY are race-free across distributed instances
- **Per-client, per-route enforcement** — middleware keys on `path:clientIP` with `X-Forwarded-For` / `X-Real-IP` support
- **Correct HTTP semantics** — 429 responses include a `Retry-After` header reflecting the actual window duration
- **Race-tested** — concurrent correctness verified with Go's `-race` detector
- **Docker + Compose ready** — single command to spin up app + Redis

---

## Algorithms

### Fixed Window
Counts requests in fixed time buckets. Simple and memory-efficient. Best for enforcing hard per-minute or per-hour limits.

```
Window:  |--- 10s ---|--- 10s ---|--- 10s ---|
Limit:       5 req       5 req       5 req
```

### Token Bucket
Tokens refill at a steady rate up to a capacity cap. Allows controlled bursting while smoothing traffic over time. Best for APIs where occasional spikes are acceptable.

```
Capacity: 10 tokens
Refill:    1 token / second
Burst:     up to 10 requests instantly, then throttled
```

---

## Project Layout

```
.
├── config/         # Route configuration for the demo server
├── limiter/        # Algorithm implementations + benchmarks
│   ├── fixed_window.go
│   ├── fixed_window_test.go
│   ├── token_bucket.go
│   └── token_bucket_test.go
├── middleware/     # net/http middleware
│   ├── middleware.go
│   └── middleware_test.go
├── store/          # Store interface, memory + Redis implementations
│   ├── store.go
│   ├── memory.go
│   ├── memory_test.go
│   └── redis.go
├── docker-compose.yml
├── Dockerfile
└── main.go         # Demo server
```

---

## Quick Start

```bash
go run .
```

Server starts on `:8080`. Available endpoints:

| Endpoint | Algorithm | Limit |
|---|---|---|
| `GET /api/fixed` | Fixed Window | 5 req / 10s |
| `GET /api/token` | Token Bucket | 10 cap, 1 token/s refill |
| `GET /healthz` | — | — |

**Try it:**

```bash
# Hit the fixed window limit
for i in $(seq 1 7); do curl -i http://localhost:8080/api/fixed; done

# Hit the token bucket limit
for i in $(seq 1 12); do curl -i http://localhost:8080/api/token; done
```

On the 6th request (fixed) and 11th (token), you'll get:

```
HTTP/1.1 429 Too Many Requests
Retry-After: 10
```

---

## Redis Backend

The fixed window limiter supports Redis for distributed rate limiting across multiple instances.

```bash
RATE_LIMIT_STORE=redis REDIS_ADDR=localhost:6379 go run .
```

The Redis implementation uses a Lua script to atomically increment the counter and set expiry on first write — no TOCTOU race condition:

```lua
local count = redis.call("INCR", KEYS[1])
if count == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return count
```

**Environment variables:**

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `RATE_LIMIT_STORE` | `memory` | `memory` or `redis` |
| `REDIS_ADDR` | `localhost:6379` | Redis address |

---

## Docker

**In-memory store:**
```bash
docker build -t rate-limiter .
docker run --rm -p 8080:8080 rate-limiter
```

**App + Redis with Compose (recommended):**
```bash
docker compose up
```

---

## Usage as a Library

```go
import (
    "net/http"
    "time"

    "github.com/Amag1n3/rate-limiter/limiter"
    "github.com/Amag1n3/rate-limiter/middleware"
    "github.com/Amag1n3/rate-limiter/store"
)

func main() {
    // Fixed window: 100 requests per minute per client
    fw := limiter.NewFixedWindow(store.NewMemoryStore(time.Minute), 100)

    // Token bucket: burst of 20, refill 1 token/sec
    tb := limiter.NewTokenBucket(20, time.Second)

    mux := http.NewServeMux()

    // Middleware automatically keys by path + client IP
    // and sets Retry-After on 429 responses
    mux.Handle("/api/read",  middleware.RateLimit(fw, time.Minute)(readHandler))
    mux.Handle("/api/write", middleware.RateLimit(tb, time.Second)(writeHandler))

    http.ListenAndServe(":8080", mux)
}
```

---

## Tests

```bash
# Run all tests with race detector
go test -race ./...

# With coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Benchmarks
go test -bench=. -benchmem ./...
```

---

## Design Decisions

**Why a `Store` interface?** Decouples the algorithm from storage. The same `FixedWindow` struct works identically against in-memory state and Redis — no algorithm code changes when you scale horizontally.

**Why Lua for Redis?** `INCR` and `PEXPIRE` are two separate commands. Without atomicity, a server crash between them leaves a key that never expires. The Lua script executes as a single atomic unit on the Redis server.

**Why variadic `window` in middleware?** Keeps the API backward-compatible — existing callers passing just a limiter still work, while new callers can pass the window to get an accurate `Retry-After` header.