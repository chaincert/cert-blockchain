// Package api provides developer API key management handlers
package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/chaincertify/certd/api/database"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// generateAPIKey creates a new random API key
func generateAPIKey() (fullKey, keyHash, keyPrefix string, err error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", "", err
	}

	// Create the full key with prefix
	fullKey = "cert_" + hex.EncodeToString(bytes)

	// Hash the key for storage
	hash := sha256.Sum256([]byte(fullKey))
	keyHash = hex.EncodeToString(hash[:])

	// Store prefix for display (cert_xxxx...xxxx)
	keyPrefix = fullKey[:12]

	return fullKey, keyHash, keyPrefix, nil
}

// handleGetApiKeys returns all API keys for the authenticated user
func (s *Server) handleGetApiKeys(w http.ResponseWriter, r *http.Request) {
	// Get address from JWT token (using proper context key)
	addressStr := getAuthenticatedAddress(r)
	if addressStr == "" {
		// Fallback to query param for dev mode
		addressStr = r.URL.Query().Get("address")
		if addressStr == "" {
			s.respondError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
	}

	if s.db == nil {
		// Return empty array if no database
		s.respondJSON(w, http.StatusOK, []interface{}{})
		return
	}

	keys, err := s.db.GetAPIKeys(r.Context(), addressStr)
	if err != nil {
		s.logger.Error("Failed to get API keys", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to retrieve API keys")
		return
	}

	if keys == nil {
		keys = []database.APIKey{}
	}

	s.respondJSON(w, http.StatusOK, keys)
}

// handleCreateApiKey creates a new API key
func (s *Server) handleCreateApiKey(w http.ResponseWriter, r *http.Request) {
	// Get address from JWT token (using proper context key)
	addressStr := getAuthenticatedAddress(r)
	if addressStr == "" {
		// Fallback to query param for dev mode
		addressStr = r.URL.Query().Get("address")
		if addressStr == "" {
			s.respondError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
	}

	// Parse request body
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		req.Name = "Default Key"
	}

	// Generate new API key
	fullKey, keyHash, keyPrefix, err := generateAPIKey()
	if err != nil {
		s.logger.Error("Failed to generate API key", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to generate API key")
		return
	}

	if s.db == nil {
		// Return mock response if no database
		s.respondJSON(w, http.StatusCreated, map[string]interface{}{
			"id":         "mock-id",
			"name":       req.Name,
			"key":        fullKey,
			"key_prefix": keyPrefix,
			"rate_limit": 1000,
			"message":    "Save this key - it won't be shown again!",
		})
		return
	}

	// Store in database
	apiKey, err := s.db.CreateAPIKey(r.Context(), addressStr, keyHash, keyPrefix, req.Name)
	if err != nil {
		s.logger.Error("Failed to create API key", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to create API key")
		return
	}

	// Return the full key only on creation
	s.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         apiKey.ID,
		"name":       apiKey.Name,
		"key":        fullKey,
		"key_prefix": apiKey.KeyPrefix,
		"rate_limit": apiKey.RateLimit,
		"created_at": apiKey.CreatedAt,
		"message":    "Save this key - it won't be shown again!",
	})
}

// handleDeleteApiKey deletes an API key
func (s *Server) handleDeleteApiKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID := vars["keyId"]

	// Get address from JWT token (using proper context key)
	addressStr := getAuthenticatedAddress(r)
	if addressStr == "" {
		// Fallback to query param for dev mode
		addressStr = r.URL.Query().Get("address")
		if addressStr == "" {
			s.respondError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
	}

	if s.db == nil {
		s.respondJSON(w, http.StatusOK, map[string]string{"message": "API key deleted"})
		return
	}

	err := s.db.DeleteAPIKey(r.Context(), keyID, addressStr)
	if err != nil {
		s.logger.Error("Failed to delete API key", zap.Error(err))
		s.respondError(w, http.StatusNotFound, "API key not found")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]string{"message": "API key deleted"})
}

// handleGetApiUsage returns usage statistics
func (s *Server) handleGetApiUsage(w http.ResponseWriter, r *http.Request) {
	// Get address from JWT token (using proper context key)
	addressStr := getAuthenticatedAddress(r)
	if addressStr == "" {
		// Fallback to query param for dev mode
		addressStr = r.URL.Query().Get("address")
		if addressStr == "" {
			s.respondError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
	}

	if s.db == nil {
		s.respondJSON(w, http.StatusOK, map[string]interface{}{
			"total_requests":    0,
			"requests_today":    0,
			"requests_this_week": 0,
			"avg_response_ms":   0,
		})
		return
	}

	stats, err := s.db.GetAPIUsage(r.Context(), addressStr)
	if err != nil {
		s.logger.Error("Failed to get API usage", zap.Error(err))
		s.respondError(w, http.StatusInternalServerError, "Failed to retrieve usage statistics")
		return
	}

	s.respondJSON(w, http.StatusOK, stats)
}

