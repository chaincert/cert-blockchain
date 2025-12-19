package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/chaincertify/certd/api/database"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

// respondJSON sends a JSON response
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// respondError sends an error response
func (s *Server) respondError(w http.ResponseWriter, status int, message string) {
	s.respondJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Code:    status,
		Message: message,
	})
}

// handleHealth handles GET /api/v1/health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": getCurrentTimestamp(),
		"version":   "1.0.0",
	})
}

// handleCreateSchema handles POST /api/v1/schemas
func (s *Server) handleCreateSchema(w http.ResponseWriter, r *http.Request) {
	s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Schema    string `json:"schema"`
			Resolver  string `json:"resolver,omitempty"`
			Revocable bool   `json:"revocable"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.respondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Schema == "" {
			s.respondError(w, http.StatusBadRequest, "schema is required")
			return
		}

		creator := getAuthenticatedAddress(r)
		s.logger.Info("Creating schema",
			zap.String("creator", creator),
			zap.String("schema", req.Schema),
		)

		// TODO: Submit to blockchain

		s.respondJSON(w, http.StatusCreated, map[string]interface{}{
			"uid":       "0x" + generateUID(),
			"schema":    req.Schema,
			"revocable": req.Revocable,
			"creator":   creator,
		})
	})(w, r)
}

// handleGetSchema handles GET /api/v1/schemas/{uid}
func (s *Server) handleGetSchema(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	// TODO: Query blockchain

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"uid":       uid,
		"schema":    "",
		"revocable": true,
		"creator":   "",
	})
}

// handleCreateAttestation handles POST /api/v1/attestations
func (s *Server) handleCreateAttestation(w http.ResponseWriter, r *http.Request) {
	s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			SchemaUID      string `json:"schema_uid"`
			Recipient      string `json:"recipient"`
			Data           string `json:"data"`
			Revocable      bool   `json:"revocable"`
			ExpirationTime int64  `json:"expiration_time,omitempty"`
			RefUID         string `json:"ref_uid,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.respondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		attester := getAuthenticatedAddress(r)

		s.respondJSON(w, http.StatusCreated, map[string]interface{}{
			"uid":       "0x" + generateUID(),
			"attester":  attester,
			"recipient": req.Recipient,
			"timestamp": getCurrentTimestamp(),
		})
	})(w, r)
}

// handleGetAttestation handles GET /api/v1/attestations/{uid}
func (s *Server) handleGetAttestation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]
	if uid == "" {
		s.respondError(w, http.StatusBadRequest, "uid is required")
		return
	}

	// Best-effort: query the chain via certd.
	// Command: certd query attestation attestation <uid> --output json
	var raw map[string]any
	if err := s.execCertdQueryJSON(&raw, "attestation", "attestation", uid); err != nil {
		s.logger.Warn("failed to query attestation", zap.String("uid", uid), zap.Error(err))
		// Fallback to minimal response.
		s.respondJSON(w, http.StatusOK, map[string]any{"uid": uid})
		return
	}

	// Normalize common shapes for frontend convenience.
	if a, ok := raw["attestation"].(map[string]any); ok {
		out := map[string]any{"uid": uid}
		if v, ok := a["schema_uid"]; ok {
			out["schema"] = v
		}
		if v, ok := a["attester"]; ok {
			out["issuer"] = v
		}
		if v, ok := a["recipient"]; ok {
			out["recipient"] = v
		}
		if v, ok := a["time"]; ok {
			out["time"] = v
		}
		if v, ok := a["attestation_type"]; ok {
			out["type"] = v
			out["encrypted"] = v != "public" && v != ""
		}
		s.respondJSON(w, http.StatusOK, out)
		return
	}

	s.respondJSON(w, http.StatusOK, raw)
}

// handleGetAttestationsByAttester handles GET /api/v1/attestations/by-attester/{address}
func (s *Server) handleGetAttestationsByAttester(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	bech32Addr, err := toBech32Address(address)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	attestations, err := s.queryAttestationsByAttester(bech32Addr)
	if err != nil {
		s.logger.Warn("failed to query attestations by attester", zap.String("address", bech32Addr), zap.Error(err))
		s.respondError(w, http.StatusBadGateway, "Failed to query attestations")
		return
	}

	// Return a plain array for frontend convenience.
	s.respondJSON(w, http.StatusOK, attestations)
}

// handleGetAttestationsByRecipient handles GET /api/v1/attestations/by-recipient/{address}
func (s *Server) handleGetAttestationsByRecipient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	bech32Addr, err := toBech32Address(address)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	attestations, err := s.queryAttestationsByRecipient(bech32Addr)
	if err != nil {
		s.logger.Warn("failed to query attestations by recipient", zap.String("address", bech32Addr), zap.Error(err))
		s.respondError(w, http.StatusBadGateway, "Failed to query attestations")
		return
	}

	// Return a plain array for frontend convenience.
	s.respondJSON(w, http.StatusOK, attestations)
}

// handleAddCredential handles POST /api/v1/profile/credentials
func (s *Server) handleAddCredential(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		s.respondError(w, http.StatusServiceUnavailable, "CertID database not configured")
		return
	}

	// NOTE: until a full auth flow exists, we allow non-JWT writes unless disabled.
	address := getAuthenticatedAddress(r)
	if address == "" && os.Getenv("ALLOW_UNAUTH_PROFILE_WRITE") == "0" {
		s.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req struct {
		UserAddress    string `json:"user_address"`
		CredentialType string `json:"credential_type"`
		AttestationUID string `json:"attestation_uid"`
		Issuer         string `json:"issuer"`
		Verified       bool   `json:"verified"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.UserAddress == "" {
		req.UserAddress = address
	}
	if req.UserAddress == "" {
		s.respondError(w, http.StatusBadRequest, "user_address is required")
		return
	}
	if req.CredentialType == "" || req.AttestationUID == "" || req.Issuer == "" {
		s.respondError(w, http.StatusBadRequest, "credential_type, attestation_uid, and issuer are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	c := &database.Credential{
		UserAddress:    req.UserAddress,
		CredentialType: req.CredentialType,
		AttestationUID: req.AttestationUID,
		Issuer:         req.Issuer,
		Verified:       req.Verified,
		IssuedAt:       time.Now(),
	}
	if err := s.db.AddCredential(ctx, c); err != nil {
		s.logger.Warn("failed to add credential", zap.String("address", req.UserAddress), zap.Error(err))
		s.respondError(w, http.StatusBadGateway, "Failed to add credential")
		return
	}

	s.respondJSON(w, http.StatusCreated, c)
}

// handleRemoveCredential handles DELETE /api/v1/profile/credentials/{id}
func (s *Server) handleRemoveCredential(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		s.respondError(w, http.StatusServiceUnavailable, "CertID database not configured")
		return
	}

	address := getAuthenticatedAddress(r)
	if address == "" && os.Getenv("ALLOW_UNAUTH_PROFILE_WRITE") == "0" {
		s.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		s.respondError(w, http.StatusBadRequest, "id is required")
		return
	}

	// For now, user can pass the user_address in a header when unauth.
	userAddress := address
	if userAddress == "" {
		userAddress = r.Header.Get("X-User-Address")
	}
	if userAddress == "" {
		s.respondError(w, http.StatusBadRequest, "user address is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := s.db.RemoveCredential(ctx, userAddress, id); err != nil {
		s.respondError(w, http.StatusNotFound, "credential not found")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// handleGetStats handles GET /api/v1/stats
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_attestations":           0,
		"total_schemas":                0,
		"total_encrypted_attestations": 0,
		"total_profiles":               0,
	})
}

// handleGetProposals handles GET /api/v1/governance/proposals
func (s *Server) handleGetProposals(w http.ResponseWriter, r *http.Request) {
	// Return empty proposals list for now
	// TODO: Query blockchain for active governance proposals
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"proposals": []interface{}{},
		"total":     0,
	})
}

// generateUID generates a random UID
func generateUID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getCurrentTimestamp returns current Unix timestamp
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
