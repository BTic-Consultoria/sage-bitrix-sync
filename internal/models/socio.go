package models

import (
	"database/sql"
	"strconv"
	"time"
)

// Socio represents a partner/stakeholder in the Sage system.
type Socio struct {
	CodigoEmpresa       int     `json:"codigo_empresa" db:"CodigoEmpresa"`
	PorParticipacion    float64 `json:"por_participacion" db:"PorParticipacion"`
	Administrador       bool    `json:"administrador" db:"Administrador"`
	CargoAdministrador  string  `json:"cargo_administrdor" db:"CargoAdministrador"`
	DNI                 string  `json:"dni" db:"DNI"`
	RazonSocialEmpleado string  `json:"razon_social_empleado" db:"RazonSocialEmpleado"`

	CreatedAt *time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type BitrixSocio struct {
	ID           int        `json:"id"`
	Title        string     `json:"title"`
	CreatedTime  *time.Time `json:"createdTime,omitempty"`
	UpdatedTime  *time.Time `json:"updatedTIme,omitempty"`
	CategoryId   *int       `json:"categoryId,omitempty"`
	EntityTypeId *int       `json:"entityTypeId,omitempty"`

	// ufCrm55* fields
	DNI                 string `json:"ufCrm55Dni"`
	Cargo               string `json:"ufCrm55Cargo"`
	Administrador       string `json:"ufCrm55Admin"`
	Participacion       string `json:"ufCrm55Participacion"`
	RazonSocialEmpleado string `json:"ufCrm55RazonSocial"`
}

// FromSageSocio converts a Sage Socio to BitrixSocio format.
func (bs *BitrixSocio) FromSageSocio(socio *Socio) {
	// Convert boolean to Y/N
	admin := "N"
	if socio.Administrador {
		admin = "Y"
	}

	// Set default cargo if empty
	cargo := socio.CargoAdministrador
	if cargo == "" {
		cargo = "No especificado"
	}

	// Build title from RazonSocialEmpleado if available, otherwise use DNI.
	title := socio.RazonSocialEmpleado
	if title == "" {
		title = socio.DNI
	}

	*bs = BitrixSocio{
		Title:               title,
		DNI:                 socio.DNI,
		Cargo:               cargo,
		Administrador:       admin,
		Participacion:       formatFloat(socio.PorParticipacion),
		RazonSocialEmpleado: socio.RazonSocialEmpleado,
	}
}

// ToSageSocio converts a BitrixSocio back to Sage Socio format.
func (bs *BitrixSocio) ToSageSocio() *Socio {
	// Parse participation percentage.
	participacion := parseFloat(bs.Participacion)

	return &Socio{
		DNI:                 bs.DNI,
		PorParticipacion:    participacion,
		Administrador:       bs.Administrador == "Y",
		CargoAdministrador:  bs.Cargo,
		RazonSocialEmpleado: bs.RazonSocialEmpleado,
	}
}

// NeedsUpdate checks if the Bitrix socio needs to be updated with Sage data.
func (bs *BitrixSocio) NeedsUpdate(sageSocio *Socio) bool {
	// Convert Sage socio to Bitrix format for comparison.
	newBitrix := &BitrixSocio{}
	newBitrix.FromSageSocio(sageSocio)

	// Compare key field.
	return bs.Cargo != newBitrix.Cargo ||
		bs.Administrador != newBitrix.Administrador ||
		bs.Participacion != newBitrix.Participacion ||
		bs.RazonSocialEmpleado != newBitrix.RazonSocialEmpleado
}

// IsValid checks if a Socio has required fields.
func (s *Socio) IsValid() bool {
	return s.DNI != ""
}

// String returns a string representation of the Socio.
func (s *Socio) String() string {
	return "Socio{DNI: " + s.DNI + ", RazonSocial: " + s.RazonSocialEmpleado + "}"
}

// formatFloat converts float64 to string for Bitrix API.
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// parseFloat converts string to float64, returns 0 if invalid.
func parseFloat(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return 0.0
}

// ScanFromDB scans database row into Socio struct
// this helps with sql.Rows.Scan() when reading from database.
func (s *Socio) ScanFromDB(rows *sql.Rows) error {
	return rows.Scan(
		&s.CodigoEmpresa,
		&s.PorParticipacion,
		&s.Administrador,
		&s.CargoAdministrador,
		&s.DNI,
		&s.RazonSocialEmpleado,
	)
}
