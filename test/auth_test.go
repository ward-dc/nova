package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"nova-api/models"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticationFailures(t *testing.T) {
	mockAuth := &MockAPIKeyValidator{}

	mockAuth.On("ValidateAPIKey", "invalid-key").Return(nil, fmt.Errorf("invalid API key"))

	server := CreateTestServer(mockAuth)
	defer server.Close()

	request := models.BalanceRequest{
		Wallets: []string{"wallet1"},
	}

	resp := MakeAuthenticatedRequest(t, server, request, "invalid-key")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	resp2 := MakeUnauthenticatedRequest(t, server, request)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp2.StatusCode)

	mockAuth.AssertExpectations(t)
}

func TestValidAuthentication(t *testing.T) {
	mockAuth := &MockAPIKeyValidator{}

	testAPIKey := &models.APIKey{ID: "test-key", Note: "Test API Key"}
	mockAuth.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil)

	server := CreateTestServer(mockAuth)
	defer server.Close()

	request := models.BalanceRequest{
		Wallets: []string{"wallet1"},
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
