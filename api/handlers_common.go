package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

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

	s.respondJSON(w, http.StatusOK, map[string]interface{}{"uid": uid})
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
	s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		s.respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
	})(w, r)
}

// handleRemoveCredential handles DELETE /api/v1/profile/credentials/{id}
func (s *Server) handleRemoveCredential(w http.ResponseWriter, r *http.Request) {
	s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		s.respondJSON(w, http.StatusOK, map[string]string{"status": "removed"})
	})(w, r)
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
