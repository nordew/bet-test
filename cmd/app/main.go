package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nordew/bet-test/internal/client"
	"github.com/nordew/bet-test/internal/service"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file")
	}
}

func main() {
	apiBURL := flag.String(
		"apiBURL", 
		os.Getenv("API_B_URL"), 
		"URL for API B",
	)
	flag.Parse()

	if *apiBURL == "" {
		log.Fatal("API_B_URL not set")
	}

	log.Printf("API B URL: %s", *apiBURL)

	apiClient := client.NewAPIClient()
	dispatcherService := service.NewDispatcher(apiClient, *apiBURL)

	ctx := context.Background()
	if err := dispatcherService.ProcessAndDispatchUsers(ctx); err != nil {
		log.Fatalf("User processing failed: %v", err)
	}

	log.Println("Service finished.")
} 