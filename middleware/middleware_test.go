package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Amag1n3/rate-limiter/limiter"
	"github.com/Amag1n3/rate-limiter/middleware"
	"github.com/Amag1n3/rate-limiter/store"
)

// okHandler is a simple handler that writes 200 OK.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func newFixedWindow(limit int64) limiter.Limiter {
	return limiter.NewFixedWindow(store.NewMemoryStore(10*time.Second), limit)
}

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	handler := middleware.RateLimit(newFixedWindow(5))(okHandler)

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.RemoteAddr = "1.2.3.4:9999"
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}
}

func TestRateLimit_Returns429WhenExceeded(t *testing.T) {
	const limit = 3
	handler := middleware.RateLimit(newFixedWindow(limit))(okHandler)

	for i := 0; i < limit; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.RemoteAddr = "1.2.3.4:9999"
		handler.ServeHTTP(rec, req)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "1.2.3.4:9999"
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestRateLimit_RetryAfterHeader(t *testing.T) {
	const limit = 1
	handler := middleware.RateLimit(newFixedWindow(limit), 10*time.Second)(okHandler)

	for i := 0; i <= limit; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.2.3.4:9999"
		handler.ServeHTTP(rec, req)

		if rec.Code == http.StatusTooManyRequests {
			retryAfter := rec.Header().Get("Retry-After")
			if retryAfter == "" {
				t.Fatal("expected Retry-After header on 429 response")
			}
			if retryAfter != "10" {
				t.Fatalf("expected Retry-After: 10, got: %s", retryAfter)
			}
			return
		}
	}
	t.Fatal("never got a 429 response")
}

func TestRateLimit_PerIPIsolation(t *testing.T) {
	const limit = 2
	handler := middleware.RateLimit(newFixedWindow(limit))(okHandler)

	// Exhaust limit for IP A
	for i := 0; i < limit; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		handler.ServeHTTP(rec, req)
	}

	// IP B should still be allowed
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("IP B should not be rate-limited by IP A's usage, got %d", rec.Code)
	}
}

func TestClientIP_XForwardedFor(t *testing.T) {
	var capturedKey string

	// Spy limiter to capture the key used
	spy := &spyLimiter{}
	handler := middleware.RateLimit(spy)(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	capturedKey = spy.lastKey
	if capturedKey == "" {
		t.Fatal("expected a key to be captured")
	}

	// Key format is "path:ip" — client IP should be the first XFF entry
	expected := "203.0.113.5"
	if len(capturedKey) < len(expected) || capturedKey[len(capturedKey)-len(expected):] != expected {
		t.Fatalf("expected key to end with %q, got %q", expected, capturedKey)
	}
}

func TestClientIP_XRealIP(t *testing.T) {
	spy := &spyLimiter{}
	handler := middleware.RateLimit(spy)(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	req.Header.Set("X-Real-IP", "198.51.100.7")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expected := "198.51.100.7"
	if key := spy.lastKey; len(key) < len(expected) || key[len(key)-len(expected):] != expected {
		t.Fatalf("expected key to end with %q, got %q", expected, spy.lastKey)
	}
}

// spyLimiter records the last key passed to Allow and always permits.
type spyLimiter struct {
	lastKey string
}

func (s *spyLimiter) Allow(_ context.Context, key string) (bool, error) {
	s.lastKey = key
	return true, nil
}
