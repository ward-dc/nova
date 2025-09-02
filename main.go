package main

import (
	"fmt"
	"log"
	"net/http"

	"nova-api/config"
	"nova-api/data"
	"nova-api/handlers"
	"nova-api/middleware"

	"github.com/gorilla/mux"
)

func main() {
	config.Load()

	balanceService := data.NewBalanceService(
		config.AppConfig.SolanaRPCEndpoint,
		config.AppConfig.DragonflyAddr,
		config.AppConfig.DragonflyPassword,
		config.AppConfig.DragonflyDB,
	)
	defer balanceService.Close()

	mongoService, err := data.NewMongoService()
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB service: %v", err)
	}
	defer mongoService.Close()

	balanceHandler := handlers.NewBalanceHandler(balanceService)

	router := mux.NewRouter()

	router.Use(middleware.RateLimitMiddleware)
	router.Use(middleware.CORSMiddleware)

	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.APIKeyAuth(mongoService))
	api.HandleFunc("/get-balance", balanceHandler.GetBalanceHandler).Methods("POST")

	fmt.Printf("API Server starting on port %s\n", config.AppConfig.Port)

	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Port, router))
}
