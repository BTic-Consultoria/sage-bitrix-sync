// cmd/test/main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/BTic-Consultoria/sage-bitrix-sync/internal/config"
	"github.com/BTic-Consultoria/sage-bitrix-sync/internal/repository"
)

func main() {
	fmt.Println("🏠 Sage-Bitrix Sync - Windows 11 Development Test")
	fmt.Println("=================================================")
	fmt.Println("Testing connection to SRVSAGE\\SAGEEXPRESS on port 64952...")
	fmt.Println()

	// Step 1: Load configuration
	fmt.Println("\n📋 Loading configuration from .env file...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("❌ Failed to load configuration:", err)
	}
	fmt.Printf("✅ Configuration loaded successfully\n")
	fmt.Printf("   🏢 Database: %s@%s:%d/%s\n", cfg.SageDB.Username, cfg.SageDB.Host, cfg.SageDB.Port, cfg.SageDB.Database)
	fmt.Printf("   🔗 Bitrix: %s\n", cfg.Bitrix.Endpoint)
	fmt.Printf("   📋 License: %s\n", cfg.License.ID)
	fmt.Printf("   🏭 Company Mapping: Bitrix '%s' ↔ Sage '%s'\n", cfg.Company.BitrixCode, cfg.Company.SageCode)
	fmt.Printf("   ⏱️  Sync Interval: %d minutes\n", cfg.Sync.IntervalMinutes)

	// Step 2: Connect to Sage database
	fmt.Println("\n🔌 Connecting to your Sage test database...")
	fmt.Printf("   Attempting connection to: %s:%d\n", cfg.SageDB.Host, cfg.SageDB.Port)
	db, err := connectToDatabase(cfg)
	if err != nil {
		fmt.Printf("❌ Database connection failed!\n")
		fmt.Printf("   Error: %v\n", err)
		fmt.Printf("\n🔧 Windows Troubleshooting tips:\n")
		fmt.Printf("   1. Check if SQL Server is running: services.msc → SQL Server (SAGEEXPRESS)\n")
		fmt.Printf("   2. Verify Windows Firewall allows port %d\n", cfg.SageDB.Port)
		fmt.Printf("   3. Test with SQL Server Management Studio first\n")
		fmt.Printf("   4. Check if you're on the same domain/network as SRVSAGE\n")
		fmt.Printf("   5. Verify your Tauri app can still connect\n")
		fmt.Printf("   6. Try 'ping SRVSAGE' from Command Prompt\n")
		log.Fatal("Cannot proceed without database connection")
	}
	defer db.Close()
	fmt.Println("✅ Database connection established successfully!")

	// Step 3: Test database connection
	fmt.Println("\n🧪 Testing database queries...")
	if err := testConnection(db); err != nil {
		fmt.Printf("❌ Database query test failed: %v\n", err)
		fmt.Printf("\n💡 This usually means:\n")
		fmt.Printf("   - Database connection works but permissions are limited\n")
		fmt.Printf("   - Need to check user permissions for SELECT queries\n")
		log.Fatal("Database connection test failed")
	}
	fmt.Println("✅ Database query test passed!")

	// Step 4: Test table existence
	fmt.Println("\n📋 Checking if required Sage tables exist...")
	tables := []string{"Personas", "SociosHistorico", "CargosFiscalHistorico"}
	for _, table := range tables {
		if err := checkTableExists(db, table); err != nil {
			fmt.Printf("⚠️  Warning: Table '%s' check failed: %v\n", table, err)
			fmt.Printf("   This might indicate schema differences between test and production\n")
		} else {
			fmt.Printf("✅ Table '%s' found\n", table)
		}
	}

	// Step 4: Create repository and fetch socios
	fmt.Println("\n👥 Fetching socios from Sage...")
	socioRepo := repository.NewSocioRepository(db)
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get count first
	count, err := socioRepo.Count(ctx)
	if err != nil {
		log.Printf("Warning: Could not get socio count: %v", err)
	} else {
		fmt.Printf("📊 Total socios in database: %d\n", count)
	}

	// Get all socios
	socios, err := socioRepo.GetAll(ctx)
	if err != nil {
		log.Fatal("Failed to fetch socios:", err)
	}

	fmt.Printf("✅ Successfully fetched %d socios\n", len(socios))

	// Step 5: Display sample results
	if len(socios) > 0 {
		fmt.Println("\n📋 Sample socios:")
		displayLimit := 5
		if len(socios) < displayLimit {
			displayLimit = len(socios)
		}

		for i := 0; i < displayLimit; i++ {
			socio := socios[i]
			fmt.Printf("  %d. DNI: %s | %s | Admin: %t | Participation: %.2f%%\n",
				i+1,
				socio.DNI,
				truncateString(socio.RazonSocialEmpleado, 30),
				socio.Administrador,
				socio.PorParticipacion)
		}

		if len(socios) > displayLimit {
			fmt.Printf("  ... and %d more socios\n", len(socios)-displayLimit)
		}
	} else {
		fmt.Println("⚠️  No socios found in the database")
	}

	// Step 6: Test individual lookup
	if len(socios) > 0 {
		fmt.Println("\n🔍 Testing individual socio lookup...")
		testDNI := socios[0].DNI
		socio, err := socioRepo.GetByDNI(ctx, testDNI)
		if err != nil {
			log.Printf("Warning: Failed to get socio by DNI: %v", err)
		} else if socio != nil {
			fmt.Printf("✅ Successfully found socio: %s\n", socio)
		} else {
			fmt.Printf("⚠️  Socio with DNI %s not found\n", testDNI)
		}
	}

	fmt.Println("\n🎉 Test completed successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Verify the socios data looks correct")
	fmt.Println("  2. Test with your actual Sage database")
	fmt.Println("  3. We'll add Bitrix24 integration next")
}

// connectToDatabase establishes connection to SQL Server
func connectToDatabase(cfg *config.Config) (*sql.DB, error) {
	connString := cfg.GetConnectionString()
	
	// Open database connection
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// testConnection verifies the database connection works
func testConnection(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Simple ping test
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Try a simple query
	var result int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("test query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected test query result: %d", result)
	}

	return nil
}

// checkTableExists verifies a table exists in the database
func checkTableExists(db *sql.DB, tableName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_NAME = @p1
	`
	var count int
	err := db.QueryRowContext(ctx, query, sql.Named("p1", tableName)).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("table '%s' not found", tableName)
	}

	return nil
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}