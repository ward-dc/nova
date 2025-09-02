package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                     string
	RateLimitRequestsPerMin  int
	MaxWalletsPerRequest     int
	SolanaRPCEndpoint        string
	DragonflyAddr            string
	DragonflyPassword        string
	DragonflyDB              int
	MongoDBURI               string
	MongoDBDatabase          string
	MongoDBCollectionAPIKeys string
	APIKeyCacheTTL           int `json:"api_key_cache_ttl"`
	BalanceCacheTTL          int `json:"balance_cache_ttl"`
}

var AppConfig *Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, fallback to system environment variables")
	}

	AppConfig = &Config{
		Port:                     getEnvString("PORT", "8080"),
		RateLimitRequestsPerMin:  getEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 10),
		MaxWalletsPerRequest:     getEnvInt("MAX_WALLETS_PER_REQUEST", 50),
		SolanaRPCEndpoint:        getEnvString("SOLANA_RPC_ENDPOINT", "https://api.mainnet-beta.solana.com"),
		DragonflyAddr:            getEnvString("DRAGONFLY_ADDR", "localhost:6379"),
		DragonflyPassword:        getEnvString("DRAGONFLY_PASSWORD", ""),
		DragonflyDB:              getEnvInt("DRAGONFLY_DB", 0),
		MongoDBURI:               getEnvString("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBDatabase:          getEnvString("MONGODB_DATABASE", "nova_api"),
		MongoDBCollectionAPIKeys: getEnvString("MONGODB_COLLECTION_APIKEYS", "api_keys"),
		APIKeyCacheTTL:           getEnvInt("API_KEY_CACHE_TTL", 300),
		BalanceCacheTTL:          getEnvInt("BALANCE_CACHE_TTL", 300),
	}
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
