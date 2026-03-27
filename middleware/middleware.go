package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Amag1n3/rate-limiter/limiter"
)

// RateLimit returns an HTTP middleware that enforces the given limiter.
// An optional window duration sets the Retry-After header value on 429 responses.
// If omitted, Retry-After defaults to "1".
func RateLimit(l limiter.Limiter, window ...time.Duration) func(http.Handler) http.Handler {
	retryAfter := "1"
	if len(window) > 0 && window[0] > 0 {
		secs := int(window[0].Seconds())
		if secs < 1 {
			secs = 1
		}
		retryAfter = strconv.Itoa(secs)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := rateLimitKey(r)

			allowed, err := l.Allow(r.Context(), key)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", retryAfter)
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func rateLimitKey(r *http.Request) string {
	return r.URL.Path + ":" + clientIP(r)
}

func clientIP(r *http.Request) string {
	forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwardedFor != "" {
		return strings.TrimSpace(strings.Split(forwardedFor, ",")[0])
	}

	realIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}

	if remoteAddr := strings.TrimSpace(r.RemoteAddr); remoteAddr != "" {
		return remoteAddr
	}

	return "unknown"
}
