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
const EntityTypeSocios = 130

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

// TestConnection verifies the Bitrix24 connection works
// Using a simpler endpoint that requires fewer permissions
func (c *Client) TestConnection(ctx context.Context) error {
	c.logger.Printf("🧪 Testing Bitrix24 connection to: %s", c.baseURL)

	// Option 1: Try a simple CRM method instead of user.current
	var result BitrixResponse
	testBody := map[string]interface{}{
		"entityTypeId": EntityTypeSocios,
		"start":        0,
		"limit":        1, // Just get 1 record to test
	}

	err := c.doJSONRequest(ctx, "/crm.item.list", testBody, &result)
	if err != nil {
		// If CRM method also fails, try the simplest possible test
		c.logger.Printf("⚠️  CRM test failed, trying basic connection test...")
		return c.testBasicConnection(ctx)
	}

	// Check for API errors
	if err := c.checkBitrixError(&result); err != nil {
		return fmt.Errorf("Bitrix24 API test failed: %w", err)
	}

	c.logger.Printf("✅ Bitrix24 connection successful!")
	return nil
}

// testBasicConnection tries the most basic connection test
func (c *Client) testBasicConnection(ctx context.Context) error {
	c.logger.Printf("🔍 Testing basic Bitrix24 connectivity...")

	// Try a very simple GET request to see if the webhook responds at all
	var result map[string]interface{}

	// Some webhooks respond to simple requests without specific methods
	err := c.doGETRequest(ctx, "/", &result)
	if err != nil {
		// Even if this fails, if we get a proper HTTP response, the webhook is working
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "405") {
			c.logger.Printf("✅ Webhook is responding (got expected 404/405)")
			return nil
		}
		return fmt.Errorf("webhook not responding: %w", err)
	}

	c.logger.Printf("✅ Basic Bitrix24 connection working!")
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

// DiscoverEntityTypes tries to discover available CRM entity types
func (c *Client) DiscoverEntityTypes(ctx context.Context) error {
	c.logger.Printf("🔍 Discovering available Bitrix24 entity types...")

	// Try to get available entity types
	var result BitrixResponse
	err := c.doJSONRequest(ctx, "/crm.enum.entitytype", map[string]interface{}{}, &result)
	if err != nil {
		c.logger.Printf("⚠️  Could not get entity types via API: %v", err)
		return c.tryCommonEntityTypes(ctx)
	}

	c.logger.Printf("✅ Entity types discovery result: %+v", result)
	return nil
}

// tryCommonEntityTypes tests common entity type IDs
func (c *Client) tryCommonEntityTypes(ctx context.Context) error {
	c.logger.Printf("🔍 Testing common entity type IDs...")

	// Common entity type IDs for different Bitrix24 setups
	commonEntityTypes := []int{
		// Smart Process IDs (most common)
		128, 130, 132, 134, 136, 138, 140, 142, 144, 146, 148, 150,

		// Some installations use lower numbers
		1, 2, 3, 4, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60,

		// Try some higher numbers
		100, 102, 104, 106, 108, 110, 112, 114, 116, 118, 120, 122, 124, 126,
	}

	for _, entityTypeID := range commonEntityTypes {
		c.logger.Printf("🧪 Testing entity type ID: %d", entityTypeID)

		testBody := map[string]interface{}{
			"entityTypeId": entityTypeID,
			"start":        0,
			"limit":        1,
		}

		var result BitrixListResponse
		err := c.doJSONRequest(ctx, "/crm.item.list", testBody, &result)

		if err != nil {
			if strings.Contains(err.Error(), "ENTITY_TYPE_NOT_SUPPORTED") {
				c.logger.Printf("❌ Entity type %d not supported", entityTypeID)
				continue
			}
			c.logger.Printf("⚠️  Entity type %d error: %v", entityTypeID, err)
			continue
		}

		// Check for API errors
		if err := c.checkBitrixError(&result); err != nil {
			if strings.Contains(err.Error(), "ENTITY_TYPE_NOT_SUPPORTED") {
				c.logger.Printf("❌ Entity type %d not supported", entityTypeID)
				continue
			}
			c.logger.Printf("⚠️  Entity type %d API error: %v", entityTypeID, err)
			continue
		}

		// Success! This entity type works
		c.logger.Printf("✅ FOUND WORKING ENTITY TYPE: %d", entityTypeID)
		c.logger.Printf("   Response: %+v", result)

		// If we found any items, show them
		if result.Result != nil && len(result.Result.Items) > 0 {
			c.logger.Printf("   Sample item: %+v", result.Result.Items[0])
		}

		return nil
	}

	c.logger.Printf("❌ No working entity types found. You may need to:")
	c.logger.Printf("   1. Create a Smart Process in Bitrix24 first")
	c.logger.Printf("   2. Use standard CRM entities (contacts, companies)")
	c.logger.Printf("   3. Check your webhook permissions")

	return fmt.Errorf("no supported entity types found")
}

// TestStandardCRMEntities tries standard CRM entities with cleaner output
func (c *Client) TestStandardCRMEntities(ctx context.Context) error {
	c.logger.Printf("🔍 Testing standard CRM entities...")

	// Standard CRM entities
	standardEntities := map[string]string{
		"crm.contact.list": "Contacts",
		"crm.company.list": "Companies",
		"crm.lead.list":    "Leads",
		"crm.deal.list":    "Deals",
	}

	for method, name := range standardEntities {
		c.logger.Printf("🧪 Testing %s (%s)...", name, method)

		testBody := map[string]interface{}{
			"start": 0,
			"limit": 1, // Just 1 record to avoid verbose output!
		}

		var result map[string]interface{}
		err := c.doJSONRequest(ctx, "/"+method, testBody, &result)

		if err != nil {
			c.logger.Printf("❌ %s failed: %v", name, err)
			continue
		}

		// Extract just the count, not all the data!
		if resultData, ok := result["result"].([]interface{}); ok {
			c.logger.Printf("✅ %s works! Found %d records", name, len(resultData))
		} else {
			c.logger.Printf("✅ %s works! (Response format: %T)", name, result["result"])
		}

		// Show total count if available
		if total, exists := result["total"]; exists {
			c.logger.Printf("   📊 Total %s in system: %v", name, total)
		}
	}

	return nil
}

// DebugEntityType130 shows what entity type 130 actually contains
func (c *Client) DebugEntityType130(ctx context.Context) error {
	c.logger.Printf("🔍 Debugging what Entity Type 130 actually is...")

	// Get list of items in entity type 130
	testBody := map[string]interface{}{
		"entityTypeId": 130,
		"start":        0,
		"limit":        10, // Get a few items to see the structure
	}

	var result BitrixListResponse
	err := c.doJSONRequest(ctx, "/crm.item.list", testBody, &result)
	if err != nil {
		return fmt.Errorf("failed to debug entity type 130: %w", err)
	}

	if err := c.checkBitrixError(&result); err != nil {
		return err
	}

	c.logger.Printf("✅ Entity Type 130 contents:")
	if result.Result != nil {
		c.logger.Printf("   📊 Total items: %d", result.Result.Total)
		c.logger.Printf("   📝 Items returned: %d", len(result.Result.Items))
		
		if len(result.Result.Items) > 0 {
			c.logger.Printf("   📋 Sample item structure:")
			for i, item := range result.Result.Items {
				if i >= 2 { // Only show first 2 items
					break
				}
				c.logger.Printf("      Item %d: ID=%d, Title='%s'", i+1, item.ID, item.Title)
				c.logger.Printf("              DNI='%s', Cargo='%s'", item.DNI, item.Cargo)
				c.logger.Printf("              Admin='%s', Participation='%s'", item.Administrador, item.Participacion)
			}
		}
	}

	return nil
}

// DebugCustomFields checks what fields are available in entity type 130
func (c *Client) DebugCustomFields(ctx context.Context) error {
	c.logger.Printf("🔍 Checking available fields in Entity Type 130...")

	// Try to get field definitions
	testBody := map[string]interface{}{
		"entityTypeId": 130,
	}

	var result map[string]interface{}
	err := c.doJSONRequest(ctx, "/crm.item.fields", testBody, &result)
	if err != nil {
		c.logger.Printf("⚠️  Could not get field definitions: %v", err)
		return nil // Don't fail, just continue
	}

	c.logger.Printf("✅ Available fields in Entity Type 130:")
	if fields, ok := result["result"].(map[string]interface{}); ok {
		for fieldName, fieldInfo := range fields {
			if fieldName[:5] == "ufCrm" { // Only show custom fields
				c.logger.Printf("   📝 %s: %v", fieldName, fieldInfo)
			}
		}
	}

	return nil
}

// SearchForOurSocios looks for the socios we just created
func (c *Client) SearchForOurSocios(ctx context.Context) error {
	c.logger.Printf("🔍 Searching for our recently created socios...")

	// Search by DNI in entity type 130
	testDNIs := []string{"123456789A", "345345332C", "99999999R", "B65799900"}
	
	for _, dni := range testDNIs {
		c.logger.Printf("🔎 Searching for DNI: %s", dni)
		
		// Search with filter
		searchBody := map[string]interface{}{
			"entityTypeId": 130,
			"filter": map[string]interface{}{
				"ufCrm55Dni": dni,
			},
		}

		var result BitrixListResponse
		err := c.doJSONRequest(ctx, "/crm.item.list", searchBody, &result)
		if err != nil {
			c.logger.Printf("❌ Search failed for %s: %v", dni, err)
			continue
		}

		if err := c.checkBitrixError(&result); err != nil {
			c.logger.Printf("❌ Search error for %s: %v", dni, err)
			continue
		}

		if result.Result != nil && len(result.Result.Items) > 0 {
			item := result.Result.Items[0]
			c.logger.Printf("✅ FOUND %s: ID=%d, Title='%s'", dni, item.ID, item.Title)
		} else {
			c.logger.Printf("❌ NOT FOUND: %s", dni)
		}
	}

	return nil
}