package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/arduriki/sage-bitrix-sync/internal/bitrix"
	"github.com/arduriki/sage-bitrix-sync/internal/config"
)

func main() {
	fmt.Println("🔍 Bitrix24 Debug Mode - Finding Your Socios!")
	fmt.Println("===============================================")

	// Create logger
	logger := log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Lshortfile)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("❌ Failed to load configuration:", err)
	}

	// Create Bitrix client
	bitrixClient := bitrix.NewClient(cfg.Bitrix.Endpoint, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	fmt.Println("🔍 Step 1: What is Entity Type 130?")
	if err := bitrixClient.DebugEntityType130(ctx); err != nil {
		fmt.Printf("❌ Debug failed: %v\n", err)
	}
	fmt.Println()

	fmt.Println("🔍 Step 2: What fields are available?")
	if err := bitrixClient.DebugCustomFields(ctx); err != nil {
		fmt.Printf("❌ Field check failed: %v\n", err)
	}
	fmt.Println()

	fmt.Println("🔍 Step 3: Search for our socios by DNI")
	if err := bitrixClient.SearchForOurSocios(ctx); err != nil {
		fmt.Printf("❌ Search failed: %v\n", err)
	}
	fmt.Println()

	fmt.Println("💡 Based on the results above:")
	fmt.Println("1. If socios are found → They exist but might be in a different Bitrix24 section")
	fmt.Println("2. If fields are missing → We need to create custom fields or use different entity")
	fmt.Println("3. If entity type 130 is wrong → We should try standard CRM entities")
	fmt.Println()
	fmt.Println("🔎 Check these Bitrix24 locations:")
	fmt.Println("  • CRM → Smart Processes")
	fmt.Println("  • CRM → Automation → Smart Process (Entity Type 130)")
	fmt.Println("  • CRM → Contacts (if we switch to standard entities)")
}
