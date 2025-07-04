package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/microsoft/go-mssqldb" // SQL Server driver
	"github.com/arduriki/sage-bitrix-sync/internal/models"
)

// SocioRepository handles database operations for Socio entities
// This is similar to your SocioRepository class in .NET
type SocioRepository struct {
	db *sql.DB
}

// NewSocioRepository creates a new repository instance
// In Go, we use constructor functions instead of constructors
func NewSocioRepository(db *sql.DB) *SocioRepository {
	return &SocioRepository{
		db: db,
	}
}

// GetAll retrieves all socios from the Sage database
// This matches your actual C# query with the proper JOINs
func (r *SocioRepository) GetAll(ctx context.Context) ([]*models.Socio, error) {
	// This query matches your actual Sage database structure from SocioRepository.cs
	query := `
		SELECT 
			sh.CodigoEmpresa,
			sh.PorParticipacion,
			cfh.Administrador,
			cfh.CargoAdministrador,
			p.Dni as DNI,
			p.RazonSocialEmpleado
		FROM 
			Personas p
			INNER JOIN SociosHistorico sh ON p.GuidPersona = sh.GuidPersona
			INNER JOIN CargosFiscalHistorico cfh ON p.GuidPersona = cfh.GuidPersona
		WHERE 
			p.Dni IS NOT NULL AND p.Dni != ''
		ORDER BY p.Dni
	`

	// Execute query with context for timeout control
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query socios: %w", err)
	}
	defer rows.Close() // Always close rows when done

	var socios []*models.Socio

	// Iterate through results
	for rows.Next() {
		socio := &models.Socio{}

		// Scan row data into struct
		err := socio.ScanFromDB(rows)
		if err != nil {
			log.Printf("Warning: failed to scan socio row: %v", err)
			continue // Skip invalid rows but continue processing
		}

		// Only add valid socios
		if socio.IsValid() {
			socios = append(socios, socio)
		}
	}

	// Check for iteration errors
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over socio rows: %w", err)
	}

	return socios, nil
}

// GetByDNI retrieves a specific socio by DNI
// This matches your actual database structure with proper JOINs
func (r *SocioRepository) GetByDNI(ctx context.Context, dni string) (*models.Socio, error) {
	if dni == "" {
		return nil, fmt.Errorf("DNI cannot be empty")
	}

	query := `
		SELECT 
			sh.CodigoEmpresa,
			sh.PorParticipacion,
			cfh.Administrador,
			cfh.CargoAdministrador,
			p.Dni as DNI,
			p.RazonSocialEmpleado
		FROM 
			Personas p
			INNER JOIN SociosHistorico sh ON p.GuidPersona = sh.GuidPersona
			INNER JOIN CargosFiscalHistorico cfh ON p.GuidPersona = cfh.GuidPersona
		WHERE 
			p.Dni = @p1
	`

	row := r.db.QueryRowContext(ctx, query, sql.Named("p1", dni))

	socio := &models.Socio{}
	err := row.Scan(
		&socio.CodigoEmpresa,
		&socio.PorParticipacion,
		&socio.Administrador,
		&socio.CargoAdministrador,
		&socio.DNI,
		&socio.RazonSocialEmpleado,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found, but not an error
		}
		return nil, fmt.Errorf("failed to get socio by DNI %s: %w", dni, err)
	}

	return socio, nil
}

// GetAllExcept retrieves all socios except those with specified DNIs
// This is equivalent to your GetAllExcept() method in .NET
func (r *SocioRepository) GetAllExcept(ctx context.Context, excludeDNIs []string) ([]*models.Socio, error) {
	if len(excludeDNIs) == 0 {
		return r.GetAll(ctx) // If no exclusions, return all
	}

	// Build placeholders for the IN clause using SQL Server syntax
	placeholders := ""
	args := make([]interface{}, len(excludeDNIs))
	for i, dni := range excludeDNIs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("@p%d", i+1)
		args[i] = sql.Named(fmt.Sprintf("p%d", i+1), dni)
	}

	query := fmt.Sprintf(`
		SELECT 
			sh.CodigoEmpresa,
			sh.PorParticipacion,
			cfh.Administrador,
			cfh.CargoAdministrador,
			p.Dni as DNI,
			p.RazonSocialEmpleado
		FROM 
			Personas p
			INNER JOIN SociosHistorico sh ON p.GuidPersona = sh.GuidPersona
			INNER JOIN CargosFiscalHistorico cfh ON p.GuidPersona = cfh.GuidPersona
		WHERE 
			p.Dni IS NOT NULL 
			AND p.Dni != ''
			AND p.Dni NOT IN (%s)
		ORDER BY p.Dni
	`, placeholders)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query socios excluding DNIs: %w", err)
	}
	defer rows.Close()

	var socios []*models.Socio

	for rows.Next() {
		socio := &models.Socio{}
		err := socio.ScanFromDB(rows)
		if err != nil {
			log.Printf("Warning: failed to scan socio row: %v", err)
			continue
		}

		if socio.IsValid() {
			socios = append(socios, socio)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over socio rows: %w", err)
	}

	return socios, nil
}

// Count returns the total number of socios in the database
func (r *SocioRepository) Count(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM Personas p
			INNER JOIN SociosHistorico sh ON p.GuidPersona = sh.GuidPersona
			INNER JOIN CargosFiscalHistorico cfh ON p.GuidPersona = cfh.GuidPersona
		WHERE p.Dni IS NOT NULL AND p.Dni != ''
	`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count socios: %w", err)
	}

	return count, nil
}

// Close closes the database connection
func (r *SocioRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
