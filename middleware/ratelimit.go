package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	"nova-api/config"
	"nova-api/data"
	"nova-api/models"
)

type RateLimitEntry struct {
	Count     int
	ExpiresAt time.Time
}

var rateLimiter = data.NewMemoryCache()

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		rateLimit := config.AppConfig.RateLimitRequestsPerMin

		if !allowRequest(ip, rateLimit) {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Header().Set("Content-Type", "application/json")
			response := models.Response{
				Error:   "Rate limit exceeded.",
				Success: false,
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func allowRequest(ip string, rateLimit int) bool {
	key := "ratelimit:" + ip

	if cached, found := rateLimiter.Get(key); found {
		if entry, ok := cached.(*RateLimitEntry); ok {
			if entry.Count >= rateLimit {
				return false
			}

			entry.Count++
			rateLimiter.Set(key, entry, time.Minute)
			return true
		}
	}

	entry := &RateLimitEntry{
		Count:     1,
		ExpiresAt: time.Now().Add(time.Minute),
	}
	rateLimiter.Set(key, entry, time.Minute)
	return true
}

func ResetRateLimiterForTesting() {
	rateLimiter = data.NewMemoryCache()
}
