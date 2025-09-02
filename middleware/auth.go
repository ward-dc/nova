package middleware

import (
	"encoding/json"
	"net/http"

	"nova-api/models"
)

type APIKeyValidator interface {
	ValidateAPIKey(key string) (*models.APIKey, error)
}

func APIKeyAuth(validator APIKeyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-Token")
			if apiKey == "" {
				http.Error(w, "Missing X-Token header", http.StatusUnauthorized)
				return
			}

			_, err := validator.ValidateAPIKey(apiKey)
			if err != nil {
				response := models.Response{
					Error:   "Invalid API key",
					Success: false,
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(response)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
