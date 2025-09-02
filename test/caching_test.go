package test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"nova-api/models"

	"github.com/stretchr/testify/assert"
)

func TestCachingFunctionality(t *testing.T) {
	mockAuth := &MockAPIKeyValidator{}

	testAPIKey := &models.APIKey{ID: "test-key", Note: "Test API Key"}
	mockAuth.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil).Times(3)

	server := CreateTestServer(mockAuth)
	defer server.Close()

	request := models.BalanceRequest{
		Wallets: []string{"cached-wallet"},
	}

	for i := 0; i < 3; i++ {
		resp := MakeAuthenticatedRequest(t, server, request, "valid-key")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.Response
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		time.Sleep(10 * time.Millisecond)
	}

	mockAuth.AssertExpectations(t)
}

func TestBalanceServiceErrors(t *testing.T) {
	mockAuth := &MockAPIKeyValidator{}

	testAPIKey := &models.APIKey{ID: "test-key", Note: "Test API Key"}
	mockAuth.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil)

	server := CreateTestServer(mockAuth)
	defer server.Close()

	request := models.BalanceRequest{
		Wallets: []string{"error-wallet"},
	}

	resp := MakeAuthenticatedRequest(t, server, request, "valid-key")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response models.Response
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockAuth.AssertExpectations(t)
}
