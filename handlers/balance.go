package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"nova-api/config"
	"nova-api/models"
)

type BalanceService interface {
	GetBalance(wallet string) (float64, error)
}

type BalanceHandler struct {
	balanceService BalanceService
}

func NewBalanceHandler(balanceService BalanceService) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
	}
}

func (bh *BalanceHandler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	var request models.BalanceRequest

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

	if len(request.Wallets) > config.AppConfig.MaxWalletsPerRequest {
		w.WriteHeader(http.StatusBadRequest)
		response := models.Response{
			Error: fmt.Sprintf("Too many wallets requested. Maximum %d wallets allowed per request", config.AppConfig.MaxWalletsPerRequest),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	balances := make([]models.WalletBalance, 0, len(request.Wallets))
	for _, wallet := range request.Wallets {
		balance, err := bh.balanceService.GetBalance(wallet)
		if err != nil {
			balances = append(balances, models.WalletBalance{
				Wallet: wallet,
				Error:  err.Error(),
			})
		} else {
			balances = append(balances, models.WalletBalance{
				Wallet:  wallet,
				Balance: balance,
			})
		}
	}

	response := models.Response{
		Data:    balances,
		Success: true,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
