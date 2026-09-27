package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type rateWindow struct {
	started time.Time
	count   int
}

type rateLimiter struct {
	mu      sync.Mutex
	windows map[string]rateWindow
}

func RateLimit(next http.Handler) http.Handler {
	limiter := &rateLimiter{windows: make(map[string]rateWindow)}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit, bucket := limitForPath(r.URL.Path)
		if limit == 0 || !limiter.allow(clientIP(r), bucket, limit) {
			if limit == 0 {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *rateLimiter) allow(client, bucket string, limit int) bool {
	now := time.Now()
	key := bucket + ":" + client
	l.mu.Lock()
	defer l.mu.Unlock()
	window := l.windows[key]
	if window.started.IsZero() || now.Sub(window.started) >= time.Minute {
		l.windows[key] = rateWindow{started: now, count: 1}
		return true
	}
	if window.count >= limit {
		return false
	}
	window.count++
	l.windows[key] = window
	return true
}

func limitForPath(path string) (int, string) {
	switch {
	case strings.HasPrefix(path, "/v1/auth/") || strings.HasPrefix(path, "/auth/"):
		return 30, "auth"
	case strings.HasPrefix(path, "/v1/compare"):
		return 60, "compare"
	case strings.HasPrefix(path, "/v1/sponsored/events") || strings.HasPrefix(path, "/v1/cart/events"):
		return 120, "events"
	default:
		return 0, ""
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}
