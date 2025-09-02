package data

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"nova-api/config"
	"nova-api/models"
)

type APIKeyValidator interface {
	ValidateAPIKey(key string) (*models.APIKey, error)
}

type MongoService struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
	cache      *MemoryCache
}

func NewMongoService() (*MongoService, error) {
	service := &MongoService{
		cache: NewMemoryCache(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.AppConfig.MongoDBURI))
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		return nil, fmt.Errorf("MongoDB connection failed: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Printf("Failed to ping MongoDB: %v", err)
		return nil, fmt.Errorf("MongoDB connection failed: %w", err)
	}

	database := client.Database(config.AppConfig.MongoDBDatabase)
	collection := database.Collection(config.AppConfig.MongoDBCollectionAPIKeys)

	service.client = client
	service.database = database
	service.collection = collection

	log.Printf("Connected to MongoDB: %s", config.AppConfig.MongoDBDatabase)

	return service, nil
}

func (ms *MongoService) ValidateAPIKey(key string) (*models.APIKey, error) {
	if config.AppConfig.APIKeyCacheTTL > 0 {
		if cached, found := ms.cache.Get(key); found {
			if apiKey, ok := cached.(*models.APIKey); ok {
				return apiKey, nil
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var apiKey models.APIKey

	// Convert string key to ObjectID as MongoDB stores keys as ObjectIDs and we use the id field as the key
	objectID, err := primitive.ObjectIDFromHex(key)
	if err != nil {
		return nil, fmt.Errorf("invalid API key format: %w", err)
	}

	filter := bson.M{"_id": objectID}

	opts := options.FindOne().SetProjection(bson.M{"_id": 1})
	err = ms.collection.FindOne(ctx, filter, opts).Decode(&apiKey)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("invalid API key")
		}
		return nil, fmt.Errorf("failed to validate API key: %w", err)
	}

	cacheEnabled := config.AppConfig.APIKeyCacheTTL > 0
	if cacheEnabled {
		ttl := time.Duration(config.AppConfig.APIKeyCacheTTL) * time.Second
		ms.cache.Set(key, &apiKey, ttl)
	}
	return &apiKey, nil
}

// Close closes the MongoDB connection
func (ms *MongoService) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return ms.client.Disconnect(ctx)
}

// Ping tests the MongoDB connection
func (ms *MongoService) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return ms.client.Ping(ctx, nil)
}
