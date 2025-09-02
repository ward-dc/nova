package models

// Response represents the API response structure
type Response struct {
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Success bool        `json:"success"`
}

// BalanceRequest represents the request structure for balance queries
type BalanceRequest struct {
	Wallets []string `json:"wallets"`
}

// WalletBalance represents a single wallet's balance information
type WalletBalance struct {
	Wallet  string  `json:"wallet"`
	Balance float64 `json:"balance,omitempty"`
	Error   string  `json:"error,omitempty"`
}

type APIKey struct {
	ID string `bson:"_id,omitempty" json:"id"`
	// Decided to add the note so we know what the API key is for
	// This is not used in the code but can be useful for tracking
	Note string `bson:"note" json:"note"`
}
