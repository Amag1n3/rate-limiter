package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Amag1n3/rate-limiter/config"
	"github.com/Amag1n3/rate-limiter/limiter"
	"github.com/Amag1n3/rate-limiter/middleware"
	"github.com/Amag1n3/rate-limiter/store"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	for path, route := range config.Routes {
		route := route

		rateLimiter, err := buildLimiter(route)
		if err != nil {
			log.Fatalf("build limiter for %s: %v", path, err)
		}

		responseBody := "OK - " + strings.ReplaceAll(route.Algorithm, "_", " ")

		mux.Handle(path, middleware.RateLimit(rateLimiter, route.Window)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(responseBody))
		})))
	}

	addr := ":" + getenv("PORT", "8080")
	log.Printf("server running on %s", addr)
	log.Printf("fixed window store backend: %s", fixedWindowBackend())
	if fixedWindowBackend() == "redis" {
		log.Printf("redis address: %s", getenv("REDIS_ADDR", "localhost:6379"))
	}

	log.Fatal(http.ListenAndServe(addr, mux))
}

func buildLimiter(route config.RouteConfig) (limiter.Limiter, error) {
	switch route.Algorithm {
	case "fixed_window":
		return limiter.NewFixedWindow(newStore(route.Window), route.Limit), nil
	case "token_bucket":
		return limiter.NewTokenBucket(int(route.Limit), route.Window), nil
	default:
		return nil, fmt.Errorf("unknown algorithm %q", route.Algorithm)
	}
}

func newStore(window time.Duration) store.Store {
	if fixedWindowBackend() == "redis" {
		return store.NewRedisStore(getenv("REDIS_ADDR", "localhost:6379"), window)
	}
	return store.NewMemoryStore(window)
}

func fixedWindowBackend() string {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("RATE_LIMIT_STORE")), "redis") {
		return "redis"
	}
	return "memory"
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
