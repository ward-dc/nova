package data

import (
	"log"
	"sync"

	"nova-api/rpc"
)

type BalanceService struct {
	rpcClient     *rpc.SolanaRPC
	cacheService  *CacheService
	walletMutexes map[string]*sync.Mutex
	mutexMapLock  sync.RWMutex
}

func NewBalanceService(rpcEndpoint, dragonflyAddr, dragonflyPassword string, dragonflyDB int) *BalanceService {
	return &BalanceService{
		rpcClient:     rpc.NewSolanaRPC(rpcEndpoint),
		cacheService:  NewCacheService(dragonflyAddr, dragonflyPassword, dragonflyDB),
		walletMutexes: make(map[string]*sync.Mutex),
	}
}

func (bs *BalanceService) getWalletMutex(walletAddress string) *sync.Mutex {
	bs.mutexMapLock.RLock()
	mutex, exists := bs.walletMutexes[walletAddress]
	bs.mutexMapLock.RUnlock()

	if exists {
		return mutex
	}

	bs.mutexMapLock.Lock()
	defer bs.mutexMapLock.Unlock()

	if mutex, exists := bs.walletMutexes[walletAddress]; exists {
		return mutex
	}

	mutex = &sync.Mutex{}
	bs.walletMutexes[walletAddress] = mutex
	return mutex
}

func (bs *BalanceService) GetBalance(walletAddress string) (float64, error) {
	walletMutex := bs.getWalletMutex(walletAddress)
	walletMutex.Lock()
	defer walletMutex.Unlock()

	if balance, found, err := bs.cacheService.GetBalance(walletAddress); err != nil {
		log.Printf("Cache error for wallet %s: %v", walletAddress, err)
	} else if found {
		return balance, nil
	}

	balance, err := bs.rpcClient.GetBalance(walletAddress)
	if err != nil {
		return 0, err
	}

	if err := bs.cacheService.SetBalance(walletAddress, balance); err != nil {
		log.Printf("Failed to cache balance for wallet %s: %v", walletAddress, err)
	}

	return balance, nil
}

func (bs *BalanceService) Close() error {
	return bs.cacheService.Close()
}
func (bs *BalanceService) Ping() error {
	return bs.cacheService.Ping()
}
