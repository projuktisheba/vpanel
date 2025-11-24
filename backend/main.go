package main

import (
	"context"
	"fmt"

	"github.com/projuktisheba/vpanel/backend/api"
	"github.com/projuktisheba/vpanel/backend/internal/services/vps"
)

// startup is called at application startup
func main() {
	// Start vps stats server in background
	go func() {
		vps.RunWebsocketServer()
	}()

	// Run backend server in main goroutine so the process stays alive
	ctx := context.Background()
	// Start backend server
	if err := api.RunServer(ctx); err != nil {
		fmt.Printf("Failed to start backend server: %v\n", err)
	}
}
