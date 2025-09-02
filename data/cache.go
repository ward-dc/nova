package data

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"nova-api/config"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client *redis.Client
}

func NewCacheService(addr, password string, db int) *CacheService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	log.Printf("Connected to Redis: %s", addr)

	return &CacheService{
		client: rdb,
	}
}

func (c *CacheService) GetBalance(walletAddress string) (float64, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("balance:%s", walletAddress)

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to get from cache: %v", err)
	}

	var balance float64
	if err := json.Unmarshal([]byte(val), &balance); err != nil {
		return 0, false, fmt.Errorf("failed to unmarshal cached balance: %v", err)
	}

	return balance, true, nil
}

func (c *CacheService) SetBalance(walletAddress string, balance float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("balance:%s", walletAddress)

	balanceBytes, err := json.Marshal(balance)
	if err != nil {
		return fmt.Errorf("failed to marshal balance: %w", err)
	}

	ttl := time.Duration(config.AppConfig.BalanceCacheTTL) * time.Second
	err = c.client.Set(ctx, key, balanceBytes, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache balance: %w", err)
	}

	return nil
}

func (c *CacheService) GetAPIKey(key string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cacheKey := fmt.Sprintf("apikey:%s", key)

	val, err := c.client.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return "", false, nil // Cache miss
	}
	if err != nil {
		return "", false, fmt.Errorf("failed to get cached API key: %w", err)
	}

	return val, true, nil
}

func (c *CacheService) SetAPIKey(key string, ttlSeconds int) error {
	// If ttlSeconds is less than or equal to 0, We assume caching is disabled
	if ttlSeconds <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cacheKey := fmt.Sprintf("apikey:%s", key)
	ttl := time.Duration(ttlSeconds) * time.Second

	err := c.client.Set(ctx, cacheKey, key, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache API key: %w", err)
	}

	return nil
}

func (c *CacheService) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (c *CacheService) Set(key, value string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.client.Set(ctx, key, value, ttl).Err()
}
func (c *CacheService) Close() error {
	return c.client.Close()
}

func (c *CacheService) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.client.Ping(ctx).Err()
}
