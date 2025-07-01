package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/BTic-Consultoria/sage-bitrix-sync/internal/models"
)

// SocioRepository handles database operations for Socio entities.
type SocioRepository struct {
	db *sql.DB
}

// NewSocioRepository creates a new repository instance.
func NewSocioRepository(db *sql.DB) *SocioRepository {
	return &SocioRepository{
		db: db,
	}
}

// GetAll retrieves all socios from the Sage database.
func (r *SocioRepository) GetAll(ctx context.Context) ([]*models.Socio, error) {
	query := `
	SELECT
		CodigoEmpresa,
		PorParticipacion,
		Administrador,
		CargoAdministrador,
		DNI,
		RazonSocialEmpleado
	FROM Socios
	WHERE DNI IS NOT NULL AND DNI != ''
	ORDER BY DNI
	`

	// Execute query with context for timeout control.
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query socios: %w", err)
	}
	defer rows.Close()

	var socios []*models.Socio

	for rows.Next() {
		socio := &models.Socio{}

		// Scan row data into struct.
		err := socio.ScanFromDB(rows)
		if err != nil {
			log.Printf("Warning: failed to scan socio row: %v", err)
			continue // skip invalid rows but continue processing.
		}

		// Only valid socios.
		if socio.IsValid() {
			socios = append(socios, socio)
		}
	}

	// Check for iteration errors.
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over socio rows: %w", err)
	}

	return socios, nil
}

// GetByDNI retrieves a specific socio by DNI.
func (r *SocioRepository) GetByDNI(ctx context.Context, dni string) (*models.Socio, error) {
	if dni == "" {
		return nil, fmt.Errorf("DNI cannot be empty")
	}

	query := `
	SELECT
		CodigoEmpresa,
		PorParticipacion,
		Administrador,
		CargoAdministrador,
		DNI,
		RazonSocialEmpleado
	FROM Socios
	WHERE DNI = ?
	`

	row := r.db.QueryRowContext(ctx, query, dni)

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
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get socio by DNI %s: %w", dni, err)
	}

	return socio, nil
}

// GetAllExcept retrieves all socios except those with specified DNIs.
func (r *SocioRepository) GetAllExcept(ctx context.Context, excludeDNIs []string) ([]*models.Socio, error) {
	// If no exclusions, return all
	if len(excludeDNIs) == 0 {
		return r.GetAll(ctx)
	}

	// Build placeholders for the IN clause
	placeholders := ""
	args := make([]interface{}, len(excludeDNIs))
	for i, dni := range excludeDNIs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
		args[i] = dni
	}

	query := fmt.Sprintf(`
	SELECT
		CodigoEmpresa,
		PorParticipacion,
		Administrador,
		CargoAdministrador,
		DNI,
		RazonSocialEmpleado
	FROM Socios
	WHERE DNI IS NOT NULL
		AND DNI != ''
		AND DNI NOT IN (%s)
	ORDER BY DNI
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
	query := "SELECT COUNT(*) FROM Socios WHERE DNI IS NOT NULL AND DNI != ''"

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
