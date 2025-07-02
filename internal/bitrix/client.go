package bitrix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BTic-Consultoria/sage-bitrix-sync/internal/models"
)

// Client handles Bitrix24 API operations using only standard library.
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *log.Logger
}

// NewClient creates a new Bitrix24 client.
func NewClient(webhookURL string, logger *log.Logger) *Client {
	// Clean up the webhook URL to get base URL
	baseURL := strings.TrimSuffix(webhookURL, "/")

	// Create HTTP client with sensible defaults.
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		logger:     logger,
	}
}

// BitrixSocio represents a socio in Bitrix24 format.
type BitrixSocio struct {
	ID                  int    `json:"id,omitempty"`
	Title               string `json:"title"`
	EntityTypeID        int    `json:"entityTypeId"`
	DNI                 string `json:"ufCrm55Dni"`
	Cargo               string `json:"ufCrm55Cargo"`
	Administrador       string `json:"ufCrm55Admin"` // "Y" or "N"
	Participacion       string `json:"ufCrm55Participacion"`
	RazonSocialEmpleado string `json:"ufCrm55RazonSocial"`
}

// BitrixResponse represents Bitrix24 API response.
type BitrixResponse struct {
	Result interface{} `json:"result"`
	Error  *struct {
		ErrorCode        string `json:"error"`
		ErrorDescription string `json:"error_description"`
	} `json:"error"`
}

// BitrixListResponse represents list response from Bitrix24.
type BitrixListResponse struct {
	Result *struct {
		Items []BitrixSocio `json:"items"`
		Total int           `json:"total"`
	} `json:"result"`
	Error *struct {
		ErrorCode        string `json:"error"`
		ErrorDescription string `json:"error_description"`
	} `json:"error"`
}

// Constants for Bitrix24
const EntityTypeSocios = 55 // This might need adjustment based on your Bitrix24 config.

// doJSONRequest performs a JSON POST request and handles common patterns.
func (c *Client) doJSONRequest(ctx context.Context, endpoint string, requestBody interface{}, response interface{}) error {
	// 1. Marshal request body to JSON.
	var jsonData []byte
	var err error

	if requestBody != nil {
		jsonData, err = json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	// 2. Create HTTP request.
	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 3. Set headers.
	req.Header.Set("Content-Type", "application/json")

	// 4. Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// 5. Check status code.
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// 6. Parse response.
	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// doGETRequest performs a GET request for simple endpoints.
func (c *Client) doGETRequest(ctx context.Context, endpoint string, response interface{}) error {
	// 1. Create HTTP request.
	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 2. Execute request.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// 3. Check status code.
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// 4. Parse response.
	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// checkBitrixError checks for Bitrix24 API errors in the response.
func (c *Client) checkBitrixError(response interface{}) error {
	// Use type assertion to check for error fields
	switch r := response.(type) {
	case *BitrixResponse:
		if r.Error != nil && r.Error.ErrorCode != "" {
			return fmt.Errorf("Bitrix24 API error: %s - %s", r.Error.ErrorCode, r.Error.ErrorDescription)
		}
	case *BitrixListResponse:
		if r.Error != nil && r.Error.ErrorCode != "" {
			return fmt.Errorf("Bitrix24 API error: %s - %s", r.Error.ErrorCode, r.Error.ErrorDescription)
		}
	}
	return nil
}

// TestConnection verifies the Bitrix24 connection works.
func (c *Client) TestConnection(ctx context.Context) error {
	c.logger.Printf("🧪 Testing Bitrix24 connection to: %s", c.baseURL)

	var result BitrixResponse
	err := c.doGETRequest(ctx, "/user.current", &result)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	// Check for API errors.
	if err := c.checkBitrixError(&result); err != nil {
		return err
	}

	c.logger.Printf("✅ Bitrix24 connection successful!")
	return nil
}

// ListSocios retrieves all existing socios from Bitrix24.
func (c *Client) ListSocios(ctx context.Context) ([]BitrixSocio, error) {
	c.logger.Printf("📥 Fetching existing socios from Bitrix24...")

	// Prepare request.
	requestBody := map[string]interface{}{
		"entityTypeId": EntityTypeSocios,
	}

	// Execute request.
	var result BitrixListResponse
	err := c.doJSONRequest(ctx, "/crm.item.list", requestBody, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to list socios: %w", err)
	}

	// Check for API errors.
	if err := c.checkBitrixError(&result); err != nil {
		return nil, err
	}

	// Handle nil result.
	if result.Result == nil {
		c.logger.Printf("✅ Found 0 existing socios in Bitrix24")
		return []BitrixSocio{}, nil
	}

	c.logger.Printf("✅ Found %d existing socios in Bitrix24", len(result.Result.Items))
	return result.Result.Items, nil
}

// CreateSocio creates a new socio in Bitrix24.
func (c *Client) CreateSocio(ctx context.Context, socio *models.Socio) error {
	bitrixSocio := c.convertSageToBitrix(socio)
	c.logger.Printf("📤 Creating socio in Bitrix24: DNI=%s, Name=%s", socio.DNI, socio.RazonSocialEmpleado)

	// Prepare request.
	requestBody := map[string]interface{}{
		"entityTypeId": EntityTypeSocios,
		"fields":       c.convertToFields(bitrixSocio),
	}

	// Execute request.
	var result BitrixResponse
	err := c.doJSONRequest(ctx, "/crm.item.add", requestBody, &result)
	if err != nil {
		return fmt.Errorf("failed to create socio: %w", err)
	}

	// Check for API errors.
	if err := c.checkBitrixError(&result); err != nil {
		return err
	}

	c.logger.Printf("✅ Successfully created socio: DNI=%s", socio.DNI)
	return nil
}

// UpdateSocio updates an existing socio in Bitrix24.
func (c *Client) UpdateSocio(ctx context.Context, bitrixID int, socio *models.Socio) error {
	bitrixSocio := c.convertSageToBitrix(socio)
	c.logger.Printf("📝 Updating socio in Bitrix24: ID=%d, DNI=%s", bitrixID, socio.DNI)

	// Prepare request.
	requestBody := map[string]interface{}{
		"id":     bitrixID,
		"fields": c.convertToFields(bitrixSocio),
	}

	// Execute request.
	var result BitrixResponse
	err := c.doJSONRequest(ctx, "/crm.item.update", requestBody, &result)
	if err != nil {
		return fmt.Errorf("failed to update socio: %w", err)
	}

	// Check for API errors.
	if err := c.checkBitrixError(&result); err != nil {
		return err
	}

	c.logger.Printf("✅ Successfully updated socio: DNI=%s", socio.DNI)
	return nil
}

// convertSageToBitrix converts a Sage Socio to Bitrix24 format.
func (c *Client) convertSageToBitrix(socio *models.Socio) *BitrixSocio {
	// Convert boolean to Y/N string.
	admin := "N"
	if socio.Administrador {
		admin = "Y"
	}

	// Set default cargo if empty.
	cargo := socio.CargoAdministrador
	if cargo == "" {
		cargo = "No especificado"
	}

	// Build title from RazonSocialEmpleado if available, otherwise use DNI.
	title := socio.RazonSocialEmpleado
	if title == "" {
		title = socio.DNI
	}

	// Format participation as string.
	participacion := strconv.FormatFloat(socio.PorParticipacion, 'f', 2, 64)

	return &BitrixSocio{
		Title:               title,
		EntityTypeID:        EntityTypeSocios,
		DNI:                 socio.DNI,
		Cargo:               cargo,
		Administrador:       admin,
		Participacion:       participacion,
		RazonSocialEmpleado: socio.RazonSocialEmpleado,
	}
}

// convertToFields converts BitrixSocio to fields map for API requests.
func (c *Client) convertToFields(bitrixSocio *BitrixSocio) map[string]interface{} {
	return map[string]interface{}{
		"title":                bitrixSocio.Title,
		"ufCrm55Dni":           bitrixSocio.DNI,
		"ufCrm55Cargo":         bitrixSocio.Cargo,
		"ufCrm55Admin":         bitrixSocio.Administrador,
		"ufCrm55Participacion": bitrixSocio.Participacion,
		"ufCrm55RazonSocial":   bitrixSocio.RazonSocialEmpleado,
	}
}

// NeedsUpdate checks if a Bitrix socio needs to be updated with Sage data.
func (c *Client) NeedsUpdate(bitrixSocio *BitrixSocio, sageSocio *models.Socio) bool {
	expectedBitrix := c.convertSageToBitrix(sageSocio)

	return bitrixSocio.Cargo != expectedBitrix.Cargo ||
		bitrixSocio.Administrador != expectedBitrix.Administrador ||
		bitrixSocio.Participacion != expectedBitrix.Participacion ||
		bitrixSocio.RazonSocialEmpleado != expectedBitrix.RazonSocialEmpleado
}

// FindSocioByDNI finds a Bitrix socio by DNI.
func (c *Client) FindSocioByDNI(socios []BitrixSocio, dni string) *BitrixSocio {
	for _, socio := range socios {
		if socio.DNI == dni {
			return &socio
		}
	}
	return nil
}
