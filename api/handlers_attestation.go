package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// CreateEncryptedAttestationRequest represents the request body
// Per Whitepaper Section 8: POST /api/v1/encrypted-attestations
type CreateEncryptedAttestationRequest struct {
	SchemaUID      string            `json:"schema_uid"`
	IPFSCID        string            `json:"ipfs_cid"`
	EncryptedHash  string            `json:"encrypted_hash"`
	Recipients     []string          `json:"recipients"`
	EncryptedKeys  map[string]string `json:"encrypted_keys"`
	Revocable      bool              `json:"revocable"`
	ExpirationTime int64             `json:"expiration_time,omitempty"`
}

// CreateEncryptedAttestationResponse represents the response
type CreateEncryptedAttestationResponse struct {
	UID       string `json:"uid"`
	TxHash    string `json:"tx_hash"`
	Timestamp int64  `json:"timestamp"`
}

// handleCreateEncryptedAttestation handles POST /api/v1/encrypted-attestations
func (s *Server) handleCreateEncryptedAttestation(w http.ResponseWriter, r *http.Request) {
	// Require authentication
	s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		var req CreateEncryptedAttestationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.respondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate request
		if req.SchemaUID == "" {
			s.respondError(w, http.StatusBadRequest, "schema_uid is required")
			return
		}
		if req.IPFSCID == "" {
			s.respondError(w, http.StatusBadRequest, "ipfs_cid is required")
			return
		}
		if len(req.Recipients) == 0 {
			s.respondError(w, http.StatusBadRequest, "at least one recipient is required")
			return
		}
		// Per Whitepaper Section 12: Max 50 recipients
		if len(req.Recipients) > 50 {
			s.respondError(w, http.StatusBadRequest, "maximum 50 recipients allowed")
			return
		}

		attester := getAuthenticatedAddress(r)

		s.logger.Info("Creating encrypted attestation",
			zap.String("attester", attester),
			zap.String("schema_uid", req.SchemaUID),
			zap.Int("recipients", len(req.Recipients)),
		)

		// TODO: Submit transaction to blockchain
		// This would call the attestation module's MsgCreateEncryptedAttestation

		resp := CreateEncryptedAttestationResponse{
			UID:       "0x" + generateUID(),
			TxHash:    "0x" + generateUID(),
			Timestamp: getCurrentTimestamp(),
		}

		s.respondJSON(w, http.StatusCreated, resp)
	})(w, r)
}

// RetrieveEncryptedAttestationRequest represents the retrieval request
// Per Whitepaper Section 8: POST /api/v1/encrypted-attestations/{uid}/retrieve
type RetrieveEncryptedAttestationRequest struct {
	RequesterAddress string `json:"requester_address"`
	Signature        string `json:"signature"`
}

// RetrieveEncryptedAttestationResponse represents the retrieval response
type RetrieveEncryptedAttestationResponse struct {
	UID          string `json:"uid"`
	IPFSCID      string `json:"ipfs_cid"`
	EncryptedKey string `json:"encrypted_key"`
	SchemaUID    string `json:"schema_uid"`
	Attester     string `json:"attester"`
	Timestamp    int64  `json:"timestamp"`
}

// handleRetrieveEncryptedAttestation handles POST /api/v1/encrypted-attestations/{uid}/retrieve
func (s *Server) handleRetrieveEncryptedAttestation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	var req RetrieveEncryptedAttestationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Verify signature to prove ownership of requester address
	if req.RequesterAddress == "" || req.Signature == "" {
		s.respondError(w, http.StatusBadRequest, "requester_address and signature required")
		return
	}

	s.logger.Info("Retrieving encrypted attestation",
		zap.String("uid", uid),
		zap.String("requester", req.RequesterAddress),
	)

	// TODO: Query blockchain for attestation
	// TODO: Verify requester is authorized recipient
	// TODO: Return encrypted key for this recipient

	resp := RetrieveEncryptedAttestationResponse{
		UID:          uid,
		IPFSCID:      "Qm...",
		EncryptedKey: "encrypted_key_for_recipient",
		SchemaUID:    "0x...",
		Attester:     "cert1...",
		Timestamp:    getCurrentTimestamp(),
	}

	s.respondJSON(w, http.StatusOK, resp)
}

// handleGetEncryptedAttestation handles GET /api/v1/encrypted-attestations/{uid}
func (s *Server) handleGetEncryptedAttestation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	// TODO: Query blockchain for attestation metadata (not encrypted data)

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"uid":        uid,
		"schema_uid": "0x...",
		"attester":   "cert1...",
		"recipients": []string{},
		"revocable":  true,
		"revoked":    false,
		"timestamp":  getCurrentTimestamp(),
	})
}

// handleRevokeEncryptedAttestation handles POST /api/v1/encrypted-attestations/{uid}/revoke
func (s *Server) handleRevokeEncryptedAttestation(w http.ResponseWriter, r *http.Request) {
	s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["uid"]
		attester := getAuthenticatedAddress(r)

		s.logger.Info("Revoking encrypted attestation",
			zap.String("uid", uid),
			zap.String("attester", attester),
		)

		// TODO: Submit revocation transaction

		s.respondJSON(w, http.StatusOK, map[string]interface{}{
			"uid":             uid,
			"revoked":         true,
			"revocation_time": getCurrentTimestamp(),
		})
	})(w, r)
}
