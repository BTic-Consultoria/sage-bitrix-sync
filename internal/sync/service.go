// internal/sync/service.go
package sync

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/BTic-Consultoria/sage-bitrix-sync/internal/bitrix"
	"github.com/BTic-Consultoria/sage-bitrix-sync/internal/config"
	"github.com/BTic-Consultoria/sage-bitrix-sync/internal/models"
	"github.com/BTic-Consultoria/sage-bitrix-sync/internal/repository"
)

// Service handles the complete synchronization process.
type Service struct {
	logger *log.Logger
}

// NewService creates a new sync service.
func NewService(logger *log.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

// SyncResult contains the results of a sync operation.
type SyncResult struct {
	ClientID        string    `json:"client_id"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Duration        string    `json:"duration"`
	SociosProcessed int       `json:"socios_processed"`
	SociosCreated   int       `json:"socios_created"`
	SociosUpdated   int       `json:"socios_updated"`
	SociosSkipped   int       `json:"socios_skipped"`
	Errors          []string  `json:"errors"`
	Success         bool      `json:"success"`
}

// SyncSocios performs the complete Sage → Bitrix24 sync for socios.
func (s *Service) SyncSocios(ctx context.Context, cfg *config.Config) (*SyncResult, error) {
	result := &SyncResult{
		ClientID:  cfg.Company.BitrixCode,
		StartTime: time.Now(),
		Errors:    make([]string, 0),
	}

	s.logger.Printf("🚀 Starting socios sync for client: %s", result.ClientID)

	// Step 1: Connect to Sage database.
	db, err := s.connectToSage(cfg)
	if err != nil {
		return s.completeResult(result, fmt.Errorf("failed to connect to Sage: %w", err))
	}
	defer db.Close()

	// Step 2: Create repositories and clients.
	socioRepo := repository.NewSocioRepository(db)
	bitrixClient := bitrix.NewClient(cfg.Bitrix.Endpoint, s.logger)

	// Step 3: Test Bitrix24 connection.
	if err := bitrixClient.TestConnection(ctx); err != nil {
		return s.completeResult(result, fmt.Errorf("failed to connect to Bitrix24: %w", err))
	}

	// Step 4: Get all socios from Sage.
	s.logger.Printf("📊 Fetching socios from Sage database...")
	sageSocios, err := socioRepo.GetAll(ctx)
	if err != nil {
		return s.completeResult(result, fmt.Errorf("failed to fetch socios from Sage: %w", err))
	}
	s.logger.Printf("✅ Found %d socios in Sage", len(sageSocios))

	// Step 5: Get existing socios from Bitrix24.
	s.logger.Printf("📊 Fetching existing socios from Bitrix24...")
	bitrixSocios, err := bitrixClient.ListSocios(ctx)
	if err != nil {
		return s.completeResult(result, fmt.Errorf("failed to fetch socios from Bitrix24: %w", err))
	}
	s.logger.Printf("✅ Found %d existing socios in Bitrix24", len(bitrixSocios))

	// Step 6: Synchronize socios.
	result.SociosProcessed = len(sageSocios)
	err = s.synchronizeSocios(ctx, bitrixClient, sageSocios, bitrixSocios, result)
	if err != nil {
		return s.completeResult(result, err)
	}

	// Step 7: Complete successfully.
	result.Success = true
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	s.logger.Printf("🎉 Sync completed successfully!")
	s.logger.Printf("   📊 Processed: %d socios", result.SociosProcessed)
	s.logger.Printf("   ✨ Created: %d socios", result.SociosCreated)
	s.logger.Printf("   📝 Updated: %d socios", result.SociosUpdated)
	s.logger.Printf("   ⏭️  Skipped: %d socios", result.SociosSkipped)
	s.logger.Printf("   ⏱️  Duration: %s", result.Duration)

	return result, nil
}

// synchronizeSocios implements the core sync logic.
func (s *Service) synchronizeSocios(ctx context.Context, bitrixClient *bitrix.Client, sageSocios []*models.Socio, bitrixSocios []bitrix.BitrixSocio, result *SyncResult) error {
	// Create a map of existing Bitrix socios by DNI for quick lookup.
	bitrixMap := make(map[string]*bitrix.BitrixSocio)
	for i := range bitrixSocios {
		if bitrixSocios[i].DNI != "" {
			bitrixMap[bitrixSocios[i].DNI] = &bitrixSocios[i]
		}
	}

	// Process each Sage socio.
	for _, sageSocio := range sageSocios {
		if sageSocio.DNI == "" {
			s.logger.Printf("⚠️  Skipping socio with empty DNI")
			result.SociosSkipped++
			continue
		}

		// Check if socio exists in Bitrix24.
		if bitrixSocio, exists := bitrixMap[sageSocio.DNI]; exists {
			// Socio exists - check if update is needed
			if bitrixClient.NeedsUpdate(bitrixSocio, sageSocio) {
				s.logger.Printf("📝 Updating socio: DNI=%s, Name=%s", sageSocio.DNI, sageSocio.RazonSocialEmpleado)

				err := bitrixClient.UpdateSocio(ctx, bitrixSocio.ID, sageSocio)
				if err != nil {
					errorMsg := fmt.Sprintf("Failed to update socio %s: %v", sageSocio.DNI, err)
					s.logger.Printf("❌ %s", errorMsg)
					result.Errors = append(result.Errors, errorMsg)
					continue
				}

				result.SociosUpdated++
			} else {
				s.logger.Printf("⏭️  Socio unchanged: DNI=%s", sageSocio.DNI)
				result.SociosSkipped++
			}
		} else {
			// Socio doesn't exist - create new one.
			s.logger.Printf("✨ Creating new socio: DNI=%s, Name=%s", sageSocio.DNI, sageSocio.RazonSocialEmpleado)

			err := bitrixClient.CreateSocio(ctx, sageSocio)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to create socio %s: %v", sageSocio.DNI, err)
				s.logger.Printf("❌ %s", errorMsg)
				result.Errors = append(result.Errors, errorMsg)
				continue
			}

			result.SociosCreated++
		}

		// Check for context cancellation.
		select {
		case <-ctx.Done():
			return fmt.Errorf("sync cancelled: %w", ctx.Err())
		default:
			// Continue processing.
		}
	}

	return nil
}

// connectToSage establishes connection to Sage database.
func (s *Service) connectToSage(cfg *config.Config) (*sql.DB, error) {
	connString := cfg.GetConnectionString()

	s.logger.Printf("🔌 Connecting to Sage database: %s@%s:%d/%s",
		cfg.SageDB.Username, cfg.SageDB.Host, cfg.SageDB.Port, cfg.SageDB.Database)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s.logger.Printf("✅ Connected to Sage database successfully")
	return db, nil
}

// completeResult helper to complete sync result with error.
func (s *Service) completeResult(result *SyncResult, err error) (*SyncResult, error) {
	result.Success = false
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	if err != nil {
		errorMsg := err.Error()
		result.Errors = append(result.Errors, errorMsg)
		s.logger.Printf("❌ Sync failed: %s", errorMsg)
	}

	return result, err
}
