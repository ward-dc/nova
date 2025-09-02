package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type SolanaRPC struct {
	client *rpc.Client
}

func NewSolanaRPC(endpoint string) *SolanaRPC {
	client := rpc.New(endpoint)
	return &SolanaRPC{
		client: client,
	}
}

func (s *SolanaRPC) GetBalance(walletAddress string) (float64, error) {
	pubkey, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		return 0, fmt.Errorf("invalid wallet address: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	balance, err := s.client.GetBalance(ctx, pubkey, rpc.CommitmentFinalized)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %v", err)
	}

	// Convert lamports to SOL
	balanceSOL := float64(balance.Value) / 1_000_000_000
	return balanceSOL, nil
}
