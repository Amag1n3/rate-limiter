package main

import (
	"log"
	"net/http"
	"rate_limiter/limiter"
	"rate_limiter/middleware"
	"rate_limiter/store"
	"time"
)

func main() {
	// Fixed window: 5 requests per 10 seconds
	memStore := store.NewMemoryStore(10 * time.Second)
	fw := limiter.NewFixedWindow(memStore, 5)

	// Token bucket: 10 capacity, refills 1 token per second
	tb := limiter.NewTokenBucket(10, 1*time.Second)

	mux := http.NewServeMux()

	// Protected by fixed window
	mux.Handle("/api/fixed", middleware.RateLimit(fw)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK - fixed window"))
	})))

	// Protected by token bucket
	mux.Handle("/api/token", middleware.RateLimit(tb)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK - token bucket"))
	})))

	log.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
