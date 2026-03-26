package middleware

import (
	"net"
	"net/http"
	"rate_limiter/limiter"
)

func RateLimit(l limiter.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, _ := net.SplitHostPort(r.RemoteAddr)
			key := host

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
