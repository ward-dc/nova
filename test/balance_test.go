package test

import (
	"encoding/json"
	"net/http"
	"testing"

	"nova-api/config"
	"nova-api/models"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	config.Load()
	config.AppConfig.RateLimitRequestsPerMin = 5
	m.Run()
}

func TestSingleWalletBalance(t *testing.T) {
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
	assert.Empty(t, response.Error)

	balances, ok := response.Data.([]interface{})
	assert.True(t, ok)
	assert.Len(t, balances, 1)

	mockAuth.AssertExpectations(t)
}

func TestMultipleWalletsBalance(t *testing.T) {
	mockAuth := &MockAPIKeyValidator{}

	testAPIKey := &models.APIKey{ID: "test-key", Note: "Test API Key"}
	mockAuth.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil)

	server := CreateTestServer(mockAuth)
	defer server.Close()

	request := models.BalanceRequest{
		Wallets: []string{"wallet1", "wallet2", "wallet3"},
	}

	resp := MakeAuthenticatedRequest(t, server, request, "valid-key")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response models.Response
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	balances, ok := response.Data.([]interface{})
	assert.True(t, ok)
	assert.Len(t, balances, 3)

	mockAuth.AssertExpectations(t)
}

func TestComprehensiveScenario(t *testing.T) {
	mockAuth := &MockAPIKeyValidator{}

	testAPIKey := &models.APIKey{ID: "test-key", Note: "Test API Key"}
	mockAuth.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil).Times(3)

	server := CreateTestServer(mockAuth)
	defer server.Close()

	singleRequest := models.BalanceRequest{Wallets: []string{"wallet1"}}
	resp1 := MakeAuthenticatedRequest(t, server, singleRequest, "valid-key")
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	multiRequest := models.BalanceRequest{Wallets: []string{"wallet2", "wallet3", "wallet4"}}
	resp2 := MakeAuthenticatedRequest(t, server, multiRequest, "valid-key")
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	zeroRequest := models.BalanceRequest{Wallets: []string{"wallet5"}}
	resp3 := MakeAuthenticatedRequest(t, server, zeroRequest, "valid-key")
	defer resp3.Body.Close()
	assert.Equal(t, http.StatusOK, resp3.StatusCode)

	mockAuth.AssertExpectations(t)
}

func TestInvalidRequestPayloads(t *testing.T) {
	mockAuth := &MockAPIKeyValidator{}

	testAPIKey := &models.APIKey{ID: "test-key", Note: "Test API Key"}
	mockAuth.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil).Maybe()

	server := CreateTestServer(mockAuth)
	defer server.Close()

	emptyRequest := models.BalanceRequest{Wallets: []string{}}
	resp1 := MakeAuthenticatedRequest(t, server, emptyRequest, "valid-key")
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp1.StatusCode)

	manyWallets := make([]string, 51)
	for i := range manyWallets {
		manyWallets[i] = "wallet"
	}
	tooManyRequest := models.BalanceRequest{Wallets: manyWallets}
	resp2 := MakeAuthenticatedRequest(t, server, tooManyRequest, "valid-key")
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)

	mockAuth.AssertExpectations(t)
}
