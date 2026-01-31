package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chaincertify/certd/api/database"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// API Key handlers

type createAPIKeyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Tier        string `json:"tier"`
}

type createAPIKeyResponse struct {
	Key    string               `json:"key"` // Full key returned only once
	APIKey *database.APIKeyNew  `json:"api_key"`
}

// handleCreateAPIKey creates a new API key for the authenticated user
func (s *Server) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	address, ok := r.Context().Value("address").(string)
	if !ok || address == "" {
		s.respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req createAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	// Validate tier
	if req.Tier == "" {
		req.Tier = "free"
	}
	if req.Tier != "free" && req.Tier != "developer" && req.Tier != "enterprise" {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid tier"})
		return
	}

	// Generate random API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		s.logger.Error("failed to generate random key", zap.Error(err))
		s.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate key"})
		return
	}
	
	// Create key in format: cert_live_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
	fullKey := "cert_live_" + hex.EncodeToString(keyBytes)
	keyPrefix := fullKey[:12] // cert_live_XX
	
	// Hash the key for storage
	hash := sha256.Sum256([]byte(fullKey))
	keyHash := hex.EncodeToString(hash[:])

	// Set rate limits based on tier
	var dailyLimit, minuteLimit int
	switch req.Tier {
	case "free":
		dailyLimit = 100
		minuteLimit = 2
	case "developer":
		dailyLimit = 10000
		minuteLimit = 100
	case "enterprise":
		dailyLimit = 1000000
		minuteLimit = 1000
	}

	// Create API key in database
	apiKey := &database.APIKeyNew{
		OwnerAddress:       address,
		KeyHash:            keyHash,
		KeyPrefix:          keyPrefix,
		Name:               req.Name,
		Description:        req.Description,
		Tier:               req.Tier,
		RateLimitPerDay:    dailyLimit,
		RateLimitPerMinute: minuteLimit,
		Active:             true,
	}

	if err := s.db.CreateAPIKeyNew(r.Context(), apiKey); err != nil {
		s.logger.Error("failed to create API key", zap.Error(err))
		s.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create key"})
		return
	}

	s.respondJSON(w, http.StatusOK, createAPIKeyResponse{
		Key:    fullKey, // Return full key only this once
		APIKey: apiKey,
	})
}

// handleListAPIKeys lists all API keys for the authenticated user
func (s *Server) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	address, ok := r.Context().Value("address").(string)
	if !ok || address == "" {
		s.respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	keys, err := s.db.ListAPIKeysByOwner(r.Context(), address)
	if err != nil {
		s.logger.Error("failed to list API keys", zap.Error(err))
		s.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list keys"})
		return
	}

	s.respondJSON(w, http.StatusOK, keys)
}

// handleRevokeAPIKey revokes an API key
func (s *Server) handleRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	address, ok := r.Context().Value("address").(string)
	if !ok || address == "" {
		s.respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	keyID := chi.URLParam(r, "keyId")
	if keyID == "" {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "missing key ID"})
		return
	}

	if err := s.db.RevokeAPIKey(r.Context(), keyID, address); err != nil {
		s.logger.Error("failed to revoke API key", zap.Error(err))
		s.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to revoke key"})
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

// handleGetAPIKeyUsage gets usage statistics for an API key
func (s *Server) handleGetAPIKeyUsage(w http.ResponseWriter, r *http.Request) {
	address, ok := r.Context().Value("address").(string)
	if !ok || address == "" {
		s.respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	keyID := chi.URLParam(r, "keyId")
	if keyID == "" {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "missing key ID"})
		return
	}

	// Get daily summaries for the last 30 days
	summaries, err := s.db.GetUsageSummary(r.Context(), keyID, "day", 30)
	if err != nil {
		s.logger.Error("failed to get usage summary", zap.Error(err))
		s.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get usage"})
		return
	}

	s.respondJSON(w, http.StatusOK, summaries)
}

// handleGetAPITiers returns available API tiers
func (s *Server) handleGetAPITiers(w http.ResponseWriter, r *http.Request) {
	tiers, err := s.db.GetAPITiers(r.Context())
	if err != nil {
		s.logger.Error("failed to get API tiers", zap.Error(err))
		s.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get tiers"})
		return
	}

	s.respondJSON(w, http.StatusOK, tiers)
}

// Rate limit middleware

// apiKeyResponseWriter wraps http.ResponseWriter to capture status code
type apiKeyResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *apiKeyResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// rateLimitMiddleware checks API key rate limits
func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// No API key provided - allow but count as anonymous
			next.ServeHTTP(w, r)
			return
		}

		// Hash the provided key
		hash := sha256.Sum256([]byte(apiKey))
		keyHash := hex.EncodeToString(hash[:])

		// Look up the key
		key, err := s.db.GetAPIKeyByHash(r.Context(), keyHash)
		if err != nil {
			s.logger.Error("failed to get API key", zap.Error(err))
			s.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		if key == nil {
			s.respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid API key"})
			return
		}

		// Check if key is expired
		if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
			s.respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "API key expired"})
			return
		}

		// Check rate limits
		allowed, err := s.db.CheckRateLimit(r.Context(), key.ID, key.RateLimitPerDay, key.RateLimitPerMinute)
		if err != nil {
			s.logger.Error("failed to check rate limit", zap.Error(err))
			s.respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		if !allowed {
			s.respondJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "rate limit exceeded",
				"tier":  key.Tier,
				"daily_limit": fmt.Sprintf("%d", key.RateLimitPerDay),
			})
			return
		}

		// Track request start time for response time tracking
		startTime := time.Now()

		// Add API key info to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "api_key_id", key.ID)
		ctx = context.WithValue(ctx, "api_key_tier", key.Tier)

		// Create response writer wrapper to capture status code
		rw := &apiKeyResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call next handler
		next.ServeHTTP(rw, r.WithContext(ctx))

		// Track usage asynchronously
		go func() {
			responseTimeMs := int(time.Since(startTime).Milliseconds())
			if err := s.db.IncrementAPIUsage(context.Background(), key.ID, rw.statusCode, responseTimeMs); err != nil {
				s.logger.Error("failed to increment API usage", zap.Error(err))
			}
			// Update last used timestamp
			if err := s.db.UpdateAPIKeyLastUsed(context.Background(), key.ID); err != nil {
				s.logger.Error("failed to update last used", zap.Error(err))
			}
		}()
	})
}
