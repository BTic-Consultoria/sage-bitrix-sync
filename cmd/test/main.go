// cmd/test/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BTic-Consultoria/sage-bitrix-sync/internal/config"
	"github.com/BTic-Consultoria/sage-bitrix-sync/internal/sync"
)

func main() {
	fmt.Println("🚀 Sage-Bitrix Sync - Complete Integration Test")
	fmt.Println("===============================================")
	fmt.Println("Testing complete sync cycle: Sage → Bitrix24")
	fmt.Println()

	// Create logger
	logger := log.New(os.Stdout, "[SYNC] ", log.LstdFlags|log.Lshortfile)

	// Step 1: Load configuration
	fmt.Println("📋 Loading configuration from .env file...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("❌ Failed to load configuration:", err)
	}

	fmt.Printf("✅ Configuration loaded successfully\n")
	fmt.Printf("   🏢 Sage Database: %s@%s:%d/%s\n", cfg.SageDB.Username, cfg.SageDB.Host, cfg.SageDB.Port, cfg.SageDB.Database)
	fmt.Printf("   🔗 Bitrix24: %s\n", cfg.Bitrix.Endpoint)
	fmt.Printf("   📋 License: %s\n", cfg.License.ID)
	fmt.Printf("   🏭 Company Mapping: Bitrix '%s' ↔ Sage '%s'\n", cfg.Company.BitrixCode, cfg.Company.SageCode)
	fmt.Printf("   ⏱️  Sync Interval: %d minutes\n", cfg.Sync.IntervalMinutes)
	fmt.Println()

	// Step 2: Create sync service
	fmt.Println("🔧 Initializing sync service...")
	syncService := sync.NewService(logger)
	fmt.Println("✅ Sync service initialized")
	fmt.Println()

	// Step 3: Perform sync with timeout
	fmt.Println("🔄 Starting complete sync cycle...")
	fmt.Println("   This will:")
	fmt.Println("   1. Connect to your Sage database")
	fmt.Println("   2. Fetch all socios")
	fmt.Println("   3. Connect to Bitrix24")
	fmt.Println("   4. Sync socios to Bitrix24")
	fmt.Println()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Perform the sync
	result, err := syncService.SyncSocios(ctx, cfg)
	if err != nil {
		fmt.Printf("❌ Sync failed: %v\n", err)
		if result != nil {
			printSyncResult(result)
		}
		os.Exit(1)
	}

	// Step 4: Display results
	fmt.Println()
	fmt.Println("🎉 Sync completed successfully!")
	printSyncResult(result)

	// Step 5: Next steps
	fmt.Println()
	fmt.Println("🚀 Next steps:")
	fmt.Println("  1. Check your Bitrix24 account to verify the socios appeared")
	fmt.Println("  2. Try running the sync again to test updates")
	fmt.Println("  3. Ready to build the web API and multi-client support!")
	fmt.Println()
	fmt.Println("💡 Pro tip: Log into your Bitrix24 and check the CRM section!")
}

// printSyncResult displays detailed sync results
func printSyncResult(result *sync.SyncResult) {
	fmt.Println("📊 Sync Results:")
	fmt.Println("   ╭─────────────────────────────────────╮")
	fmt.Printf("   │ Client ID:       %-18s │\n", result.ClientID)
	fmt.Printf("   │ Duration:        %-18s │\n", result.Duration)
	fmt.Printf("   │ Success:         %-18v │\n", result.Success)
	fmt.Println("   ├─────────────────────────────────────┤")
	fmt.Printf("   │ Socios Processed: %-17d │\n", result.SociosProcessed)
	fmt.Printf("   │ Created:         %-18d │\n", result.SociosCreated)
	fmt.Printf("   │ Updated:         %-18d │\n", result.SociosUpdated)
	fmt.Printf("   │ Skipped:         %-18d │\n", result.SociosSkipped)
	fmt.Printf("   │ Errors:          %-18d │\n", len(result.Errors))
	fmt.Println("   ╰─────────────────────────────────────╯")

	if len(result.Errors) > 0 {
		fmt.Println()
		fmt.Println("⚠️  Errors encountered:")
		for i, err := range result.Errors {
			fmt.Printf("   %d. %s\n", i+1, err)
		}
	}

	if result.Success {
		fmt.Println()
		if result.SociosCreated > 0 {
			fmt.Printf("✨ %d new socios created in Bitrix24!\n", result.SociosCreated)
		}
		if result.SociosUpdated > 0 {
			fmt.Printf("📝 %d socios updated in Bitrix24!\n", result.SociosUpdated)
		}
		if result.SociosSkipped > 0 {
			fmt.Printf("⏭️  %d socios were already up-to-date\n", result.SociosSkipped)
		}
	}
}