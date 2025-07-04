package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/arduriki/sage-bitrix-sync/internal/bitrix"
	"github.com/arduriki/sage-bitrix-sync/internal/config"
	"github.com/arduriki/sage-bitrix-sync/internal/sync"
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

	// Step 2: First, let's discover what entity types are available
	fmt.Println("🔍 DISCOVERY MODE: Finding available Bitrix24 entity types...")
	fmt.Println("   This will help us determine the correct entity type for socios")
	fmt.Println()

	// Create Bitrix client for discovery
	bitrixClient := bitrix.NewClient(cfg.Bitrix.Endpoint, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Test connection first
	logger.Printf("🧪 Testing Bitrix24 connection...")
	if err := bitrixClient.TestConnection(ctx); err != nil {
		fmt.Printf("❌ Connection test failed: %v\n", err)
		fmt.Println("💡 But let's continue with discovery anyway...")
	}

	// Discovery phase
	fmt.Println("🔎 Phase 1: Discovering Smart Process entity types...")
	if err := bitrixClient.DiscoverEntityTypes(ctx); err != nil {
		fmt.Printf("⚠️  Smart Process discovery failed: %v\n", err)
	}
	fmt.Println()

	fmt.Println("🔎 Phase 2: Testing standard CRM entities...")
	if err := bitrixClient.TestStandardCRMEntities(ctx); err != nil {
		fmt.Printf("⚠️  Standard CRM test failed: %v\n", err)
	}
	fmt.Println()

	// Ask user what to do next
	fmt.Println("🎯 DISCOVERY COMPLETE!")
	fmt.Println()
	fmt.Println("Based on the results above:")
	fmt.Println("1. If you found a working entity type ID, update the constant in client.go")
	fmt.Println("2. If standard CRM entities work, we can modify the code to use those")
	fmt.Println("3. If nothing works, you may need to create a Smart Process in Bitrix24 first")
	fmt.Println()
	fmt.Println("💡 To proceed with sync testing:")
	fmt.Println("   - Update EntityTypeSocios in internal/bitrix/client.go")
	fmt.Println("   - Or we can modify the approach based on what works")
	fmt.Println()

	// Optional: Try the full sync if user wants to
	fmt.Print("🤔 Do you want to try the full sync anyway? (y/N): ")
	var response string
	fmt.Scanln(&response)
	
	if response == "y" || response == "Y" {
		fmt.Println()
		fmt.Println("🔄 Proceeding with full sync test...")
		
		// Step 3: Create sync service
		fmt.Println("🔧 Initializing sync service...")
		syncService := sync.NewService(logger)
		fmt.Println("✅ Sync service initialized")
		fmt.Println()

		// Step 4: Perform sync with timeout
		fmt.Println("🔄 Starting complete sync cycle...")
		fmt.Println("   This will:")
		fmt.Println("   1. Connect to your Sage database")
		fmt.Println("   2. Fetch all socios")
		fmt.Println("   3. Connect to Bitrix24")
		fmt.Println("   4. Sync socios to Bitrix24")
		fmt.Println()

		// Perform the sync
		result, err := syncService.SyncSocios(ctx, cfg)
		if err != nil {
			fmt.Printf("❌ Sync failed: %v\n", err)
			if result != nil {
				printSyncResult(result)
			}
			os.Exit(1)
		}

		// Display results
		fmt.Println()
		fmt.Println("🎉 Sync completed successfully!")
		printSyncResult(result)

		// Next steps
		fmt.Println()
		fmt.Println("🚀 Next steps:")
		fmt.Println("  1. Check your Bitrix24 account to verify the socios appeared")
		fmt.Println("  2. Try running the sync again to test updates")
		fmt.Println("  3. Ready to build the web API and multi-client support!")
		fmt.Println()
		fmt.Println("💡 Pro tip: Log into your Bitrix24 and check the CRM section!")
	} else {
		fmt.Println()
		fmt.Println("👍 No problem! Use the discovery results to:")
		fmt.Println("1. Update the entity type ID in the code")
		fmt.Println("2. Or let me know what entity types work and I'll help modify the approach")
	}
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