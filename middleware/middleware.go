package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/Amag1n3/rate-limiter/limiter"
)

func RateLimit(l limiter.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := rateLimitKey(r)

			allowed, err := l.Allow(r.Context(), key)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", "1")
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
