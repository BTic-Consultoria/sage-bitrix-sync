package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/arduriki/sage-bitrix-sync/internal/config"
	"github.com/arduriki/sage-bitrix-sync/internal/sync"
)

// ClientConfig represents configuration for a client
type ClientConfig struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	SageHost       string    `json:"sage_host"`
	SageDatabase   string    `json:"sage_database"`
	SageUsername   string    `json:"sage_username"`
	SagePassword   string    `json:"sage_password"`
	BitrixEndpoint string    `json:"bitrix_endpoint"`
	LastSync       time.Time `json:"last_sync"`
	Status         string    `json:"status"`
	SociosCount    int       `json:"socios_count"`
	SyncProgress   int       `json:"sync_progress"`
	IsSyncing      bool      `json:"is_syncing"`
	Enabled        bool      `json:"enabled"`
}

// SyncStatus represents current sync status
type SyncStatus struct {
	ClientID      string    `json:"client_id"`
	Status        string    `json:"status"`
	Progress      int       `json:"progress"`
	LastSync      time.Time `json:"last_sync"`
	SociosTotal   int       `json:"socios_total"`
	SociosCreated int       `json:"socios_created"`
	SociosUpdated int       `json:"socios_updated"`
	SociosSkipped int       `json:"socios_skipped"`
	Errors        []string  `json:"errors"`
	Duration      string    `json:"duration"`
}

// APIServer handles HTTP requests
type APIServer struct {
	clients    map[string]*ClientConfig
	syncStatus map[string]*SyncStatus
	logger     *log.Logger
}

func main() {
	logger := log.New(os.Stdout, "[API] ", log.LstdFlags|log.Lshortfile)

	// Initialize API server
	server := &APIServer{
		clients:    make(map[string]*ClientConfig),
		syncStatus: make(map[string]*SyncStatus),
		logger:     logger,
	}

	// Initialize with demo data
	server.initializeDemoData()

	// Setup routes
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/clients", server.listClients).Methods("GET")
	api.HandleFunc("/clients", server.createClient).Methods("POST")
	api.HandleFunc("/clients/{id}", server.getClient).Methods("GET")
	api.HandleFunc("/clients/{id}", server.updateClient).Methods("PUT")
	api.HandleFunc("/clients/{id}/sync", server.triggerSync).Methods("POST")
	api.HandleFunc("/clients/{id}/status", server.getSyncStatus).Methods("GET")
	api.HandleFunc("/clients/{id}/logs", server.getLogs).Methods("GET")
	api.HandleFunc("/stats", server.getStats).Methods("GET")

	// Health check - ✅ FIXED: using server instead of s
	router.HandleFunc("/health", server.healthCheck).Methods("GET")

	// Serve static files (dashboard)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	// Setup CORS
	corsHandler := handlers.CORS(
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}),
	)(router)

	// Start server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      corsHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Printf("🚀 API Server starting on :8080")
		logger.Printf("📊 Dashboard: http://localhost:8080")
		logger.Printf("🔧 API: http://localhost:8080/api/v1")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Printf("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Printf("✅ Server exited")
}

// initializeDemoData sets up demo clients for demonstration
func (s *APIServer) initializeDemoData() {
	s.clients["client-1"] = &ClientConfig{
		ID:             "client-1",
		Name:           "BTIC Consultoria",
		SageHost:       "SRVSAGE\\SAGEEXPRESS",
		SageDatabase:   "STANDARD",
		SageUsername:   "LOGIC",
		BitrixEndpoint: "https://bit24.bitrix24.eu/rest/2523/0lhk1imaxwik2lh5/",
		LastSync:       time.Now().Add(-2 * time.Minute),
		Status:         "active",
		SociosCount:    45,
		SyncProgress:   100,
		IsSyncing:      false,
		Enabled:        true,
	}

	s.clients["client-2"] = &ClientConfig{
		ID:             "client-2",
		Name:           "Demo Company A",
		SageHost:       "demo-sage-01",
		SageDatabase:   "DEMO_DB",
		SageUsername:   "demo_user",
		BitrixEndpoint: "https://demo-a.bitrix24.com/rest/123/webhook/",
		LastSync:       time.Now().Add(-15 * time.Minute),
		Status:         "idle",
		SociosCount:    67,
		SyncProgress:   100,
		IsSyncing:      false,
		Enabled:        true,
	}

	s.clients["client-3"] = &ClientConfig{
		ID:             "client-3",
		Name:           "Test Corp Ltd",
		SageHost:       "test-sage-db",
		SageDatabase:   "TEST_CORP",
		SageUsername:   "test_user",
		BitrixEndpoint: "https://testcorp.bitrix24.eu/rest/456/webhook/",
		LastSync:       time.Now().Add(-1 * time.Hour),
		Status:         "idle",
		SociosCount:    44,
		SyncProgress:   100,
		IsSyncing:      false,
		Enabled:        true,
	}
}

// HTTP Handlers
func (s *APIServer) listClients(w http.ResponseWriter, r *http.Request) {
	s.logger.Printf("📋 GET /api/v1/clients")

	var clients []*ClientConfig
	for _, client := range s.clients {
		clients = append(clients, client)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clients": clients,
		"total":   len(clients),
	})
}

func (s *APIServer) getClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["id"]

	s.logger.Printf("📋 GET /api/v1/clients/%s", clientID)

	client, exists := s.clients[clientID]
	if !exists {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}

func (s *APIServer) createClient(w http.ResponseWriter, r *http.Request) {
	s.logger.Printf("➕ POST /api/v1/clients")

	var client ClientConfig
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if client.ID == "" {
		client.ID = fmt.Sprintf("client-%d", time.Now().Unix())
	}

	// Set defaults
	client.Status = "idle"
	client.SyncProgress = 0
	client.IsSyncing = false
	client.Enabled = true
	client.LastSync = time.Time{} // Zero time

	s.clients[client.ID] = &client

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(client)
}

func (s *APIServer) updateClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["id"]

	s.logger.Printf("📝 PUT /api/v1/clients/%s", clientID)

	client, exists := s.clients[clientID]
	if !exists {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	var updates ClientConfig
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update fields (keep ID and sync status)
	updates.ID = client.ID
	updates.IsSyncing = client.IsSyncing
	updates.SyncProgress = client.SyncProgress

	s.clients[clientID] = &updates

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updates)
}

func (s *APIServer) triggerSync(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["id"]

	s.logger.Printf("🚀 POST /api/v1/clients/%s/sync", clientID)

	client, exists := s.clients[clientID]
	if !exists {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	if client.IsSyncing {
		http.Error(w, "Sync already in progress", http.StatusConflict)
		return
	}

	// Start sync in background
	go s.performSync(clientID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Sync started",
		"status":  "syncing",
	})
}

func (s *APIServer) getSyncStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["id"]

	status, exists := s.syncStatus[clientID]
	if !exists {
		status = &SyncStatus{
			ClientID: clientID,
			Status:   "idle",
			Progress: 0,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *APIServer) getLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["id"]

	s.logger.Printf("📊 GET /api/v1/clients/%s/logs", clientID)

	// Mock logs for demo
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-5 * time.Minute),
			"level":     "INFO",
			"message":   "Sync completed successfully",
			"details":   "Created: 2, Updated: 3, Skipped: 40",
		},
		{
			"timestamp": time.Now().Add(-1 * time.Hour),
			"level":     "INFO",
			"message":   "Sync started",
			"details":   "Found 45 socios in Sage database",
		},
		{
			"timestamp": time.Now().Add(-2 * time.Hour),
			"level":     "INFO",
			"message":   "Connected to Sage database",
			"details":   fmt.Sprintf("Host: %s", s.clients[clientID].SageHost),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs": logs,
	})
}

func (s *APIServer) getStats(w http.ResponseWriter, r *http.Request) {
	s.logger.Printf("📊 GET /api/v1/stats")

	totalClients := len(s.clients)
	totalSocios := 0
	syncingCount := 0

	for _, client := range s.clients {
		totalSocios += client.SociosCount
		if client.IsSyncing {
			syncingCount++
		}
	}

	stats := map[string]interface{}{
		"total_clients":   totalClients,
		"total_socios":    totalSocios,
		"syncing_count":   syncingCount,
		"sync_jobs_today": 24, // Mock data
		"uptime":          "99.8%",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ✅ FIXED: Added the missing healthCheck method
func (s *APIServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"version": "1.0.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// performSync simulates/performs actual sync
func (s *APIServer) performSync(clientID string) {
	client := s.clients[clientID]
	if client == nil {
		return
	}

	s.logger.Printf("🔄 Starting sync for client: %s", client.Name)

	// Update status
	client.IsSyncing = true
	client.Status = "syncing"
	client.SyncProgress = 0

	// Initialize sync status
	s.syncStatus[clientID] = &SyncStatus{
		ClientID: clientID,
		Status:   "syncing",
		Progress: 0,
	}

	// For demo: simulate progress
	for progress := 0; progress <= 100; progress += 10 {
		time.Sleep(500 * time.Millisecond)
		client.SyncProgress = progress
		s.syncStatus[clientID].Progress = progress

		if progress == 100 {
			break
		}
	}

	// Complete sync
	client.IsSyncing = false
	client.Status = "active"
	client.LastSync = time.Now()
	client.SyncProgress = 100

	s.syncStatus[clientID].Status = "completed"
	s.syncStatus[clientID].Progress = 100
	s.syncStatus[clientID].LastSync = time.Now()

	s.logger.Printf("✅ Sync completed for client: %s", client.Name)

	// TODO: Integrate with your existing sync.Service:
		syncService := sync.NewService(s.logger)

		// ✅ FIXED: Use CompanyMappingConfig instead of CompanyConfig
		cfg := &config.Config{
			SageDB: config.SageDBConfig{
				Host:     client.SageHost,
				Database: client.SageDatabase,
				Username: client.SageUsername,
				Password: client.SagePassword,
			},
			Bitrix: config.BitrixConfig{
				Endpoint: client.BitrixEndpoint,
			},
			Company: config.CompanyMappingConfig{  // ✅ Correct struct name
				BitrixCode: "auto",
				SageCode:   "1",
			},
			License: config.LicenseConfig{
				ID: "multi-client-saas",
			},
		}

		result, err := syncService.SyncSocios(context.Background(), cfg)
		if err != nil {
			s.logger.Printf("❌ Sync failed for %s: %v", client.Name, err)
			client.Status = "error"
			s.syncStatus[clientID].Status = "error"
			s.syncStatus[clientID].Errors = append(s.syncStatus[clientID].Errors, err.Error())
			return
		}

		client.SociosCount = result.SociosProcessed
		s.syncStatus[clientID].SociosTotal = result.SociosProcessed
		s.syncStatus[clientID].SociosCreated = result.SociosCreated
		s.syncStatus[clientID].SociosUpdated = result.SociosUpdated
		s.syncStatus[clientID].SociosSkipped = result.SociosSkipped
		s.syncStatus[clientID].Duration = result.Duration

}
