package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"nova-api/config"
	"nova-api/middleware"
	"nova-api/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAPIKeyValidator struct {
	mock.Mock
}

func (m *MockAPIKeyValidator) ValidateAPIKey(key string) (*models.APIKey, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.APIKey), args.Error(1)
}

type MockBalanceHandler struct {
	mock.Mock
}

func (m *MockBalanceHandler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	var request models.BalanceRequest

	maxWallets := config.AppConfig.MaxWalletsPerRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := models.Response{
			Error: "Invalid JSON payload",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	if len(request.Wallets) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		response := models.Response{
			Error: "Wallets array cannot be empty",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	if len(request.Wallets) > maxWallets {
		w.WriteHeader(http.StatusBadRequest)
		response := models.Response{
			Error: "Too many wallets requested. Maximum " + strconv.Itoa(maxWallets) + " wallets allowed per request",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	balances := make([]models.WalletBalance, 0, len(request.Wallets))
	for _, wallet := range request.Wallets {
		balances = append(balances, models.WalletBalance{
			Wallet:  wallet,
			Balance: 1.5,
		})
	}

	response := models.Response{
		Data:    balances,
		Success: true,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreateTestServer(validator *MockAPIKeyValidator) *httptest.Server {
	balanceHandler := &MockBalanceHandler{}

	router := mux.NewRouter()
	router.Use(middleware.CORSMiddleware)

	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.APIKeyAuth(validator))
	api.HandleFunc("/get-balance", balanceHandler.GetBalanceHandler).Methods("POST")

	return httptest.NewServer(router)
}

func CreateTestServerWithRateLimit(validator *MockAPIKeyValidator) *httptest.Server {
	balanceHandler := &MockBalanceHandler{}

	router := mux.NewRouter()
	router.Use(middleware.RateLimitMiddleware)
	router.Use(middleware.CORSMiddleware)

	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.APIKeyAuth(validator))
	api.HandleFunc("/get-balance", balanceHandler.GetBalanceHandler).Methods("POST")

	return httptest.NewServer(router)
}

func MakeAuthenticatedRequest(t *testing.T, server *httptest.Server, payload interface{}, apiKey string) *http.Response {
	jsonData, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", server.URL+"/api/get-balance", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Token", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	return resp
}

func MakeUnauthenticatedRequest(t *testing.T, server *httptest.Server, payload interface{}) *http.Response {
	jsonData, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", server.URL+"/api/get-balance", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	return resp
}
