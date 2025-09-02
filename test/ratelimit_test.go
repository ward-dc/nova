package test

import (
	"net/http"
	"sync"
	"testing"

	"nova-api/config"
	"nova-api/middleware"
	"nova-api/models"

	"github.com/stretchr/testify/assert"
)

func TestIPRateLimiting(t *testing.T) {
	mockAuth := &MockAPIKeyValidator{}
	rateLimit := config.AppConfig.RateLimitRequestsPerMin
	testAPIKey := &models.APIKey{ID: "test-key", Note: "Test API Key"}
	mockAuth.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil).Maybe()

	server := CreateTestServerWithRateLimit(mockAuth)
	defer server.Close()

	request := models.BalanceRequest{
		Wallets: []string{"wallet1"},
	}

	successCount := 0
	rateLimitCount := 0

	for i := 0; i < rateLimit+1; i++ {
		resp := MakeAuthenticatedRequest(t, server, request, "valid-key")
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			successCount++
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitCount++
		}
	}

	assert.True(t, successCount > 0, "Should have some successful requests")
	assert.True(t, rateLimitCount > 0, "Should have some rate limited requests")
	assert.Equal(t, rateLimit+1, successCount+rateLimitCount, "All requests should be accounted for")
}

func TestFiveRequestsSameWallet(t *testing.T) {
	middleware.ResetRateLimiterForTesting()

	mockAuth := &MockAPIKeyValidator{}

	testAPIKey := &models.APIKey{ID: "test-key", Note: "Test API Key"}
	mockAuth.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil).Maybe()

	server := CreateTestServerWithRateLimit(mockAuth)
	defer server.Close()

	request := models.BalanceRequest{
		Wallets: []string{"same-wallet"},
	}

	successCount := 0
	rateLimitCount := 0

	for i := 0; i < 5; i++ {
		resp := MakeAuthenticatedRequest(t, server, request, "valid-key")
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			successCount++
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitCount++
		}
	}

	assert.True(t, successCount > 0, "Should have some successful requests")
}

func TestAuthAndRateLimitingIntegration(t *testing.T) {
	middleware.ResetRateLimiterForTesting()

	mockAuth := &MockAPIKeyValidator{}

	testAPIKey := &models.APIKey{ID: "test-key", Note: "Test API Key"}
	mockAuth.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil).Maybe()

	server := CreateTestServerWithRateLimit(mockAuth)
	defer server.Close()

	request := models.BalanceRequest{
		Wallets: []string{"wallet1"},
	}

	successWithAuth := 0
	rateLimitWithAuth := 0

	for i := 0; i < 8; i++ {
		resp := MakeAuthenticatedRequest(t, server, request, "valid-key")
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			successWithAuth++
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitWithAuth++
		}
	}

	assert.True(t, successWithAuth > 0, "Should have some successful authenticated requests")
}

func TestConcurrentRequests(t *testing.T) {
	mockAuth := &MockAPIKeyValidator{}

	rateLimit := config.AppConfig.RateLimitRequestsPerMin
	amountOfRequests := rateLimit + 1

	testAPIKey := &models.APIKey{ID: "test-key", Note: "Test API Key"}
	mockAuth.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil).Maybe()

	server := CreateTestServerWithRateLimit(mockAuth)
	defer server.Close()

	var wg sync.WaitGroup
	results := make(chan int, amountOfRequests)

	for i := 0; i < amountOfRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			request := models.BalanceRequest{
				Wallets: []string{"wallet"},
			}

			resp := MakeAuthenticatedRequest(t, server, request, "valid-key")
			defer resp.Body.Close()

			results <- resp.StatusCode
		}(i)
	}

	wg.Wait()
	close(results)

	successCount := 0
	rateLimitCount := 0

	for statusCode := range results {
		if statusCode == http.StatusOK {
			successCount++
		} else if statusCode == http.StatusTooManyRequests {
			rateLimitCount++
		}
	}

	assert.Equal(t, amountOfRequests, successCount+rateLimitCount, "All requests should be accounted for")
	assert.True(t, rateLimitCount > 0, "Should have rate limiting under concurrent load")
}
