package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// RegisterSchemaRequest represents the request to register a new schema
type RegisterSchemaRequest struct {
	Schema    string `json:"schema"`
	Resolver  string `json:"resolver,omitempty"`
	Revocable bool   `json:"revocable"`
	Creator   string `json:"creator"`
	Signature string `json:"signature"`
}

// SchemaResponse represents a schema response
type SchemaResponse struct {
	UID       string    `json:"uid"`
	Creator   string    `json:"creator"`
	Schema    string    `json:"schema"`
	Resolver  string    `json:"resolver,omitempty"`
	Revocable bool      `json:"revocable"`
	CreatedAt time.Time `json:"createdAt"`
}

// RegisterSchema handles POST /api/v1/schemas
func (h *Handler) RegisterSchema(w http.ResponseWriter, r *http.Request) {
	var req RegisterSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Schema == "" {
		respondError(w, http.StatusBadRequest, "Schema definition required")
		return
	}

	// TODO: Verify signature to authenticate creator

	// Generate schema UID (per EAS standard)
	uidData := fmt.Sprintf("%s%s%t", req.Schema, req.Resolver, req.Revocable)
	hash := sha256.Sum256([]byte(uidData))
	uid := "0x" + hex.EncodeToString(hash[:])

	// Check if schema already exists
	var existingUID string
	err := h.db.QueryRow(`SELECT uid FROM schemas WHERE uid = $1`, uid).Scan(&existingUID)
	if err == nil {
		respondError(w, http.StatusConflict, "Schema already exists")
		return
	}

	// Insert schema
	_, err = h.db.Exec(`
		INSERT INTO schemas (uid, creator, schema_definition, resolver, revocable)
		VALUES ($1, $2, $3, $4, $5)
	`, uid, req.Creator, req.Schema, req.Resolver, req.Revocable)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to register schema")
		return
	}

	respondJSON(w, http.StatusCreated, SchemaResponse{
		UID:       uid,
		Creator:   req.Creator,
		Schema:    req.Schema,
		Resolver:  req.Resolver,
		Revocable: req.Revocable,
		CreatedAt: time.Now(),
	})
}

// GetSchema handles GET /api/v1/schemas/{uid}
func (h *Handler) GetSchema(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	var resp SchemaResponse
	err := h.db.QueryRow(`
		SELECT uid, creator, schema_definition, resolver, revocable, created_at
		FROM schemas WHERE uid = $1
	`, uid).Scan(&resp.UID, &resp.Creator, &resp.Schema, &resp.Resolver, &resp.Revocable, &resp.CreatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "Schema not found")
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetAttestationsByAttester handles GET /api/v1/attestations/by-attester/{address}
func (h *Handler) GetAttestationsByAttester(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	rows, err := h.db.Query(`
		SELECT uid, schema_uid, attester, ipfs_cid, encrypted_data_hash, revocable, revoked, expiration_time, created_at
		FROM encrypted_attestations WHERE attester = $1
		ORDER BY created_at DESC
	`, address)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var attestations []EncryptedAttestationResponse
	for rows.Next() {
		var a EncryptedAttestationResponse
		if err := rows.Scan(&a.UID, &a.SchemaUID, &a.Attester, &a.IPFSCID, &a.EncryptedDataHash, &a.Revocable, &a.Revoked, &a.ExpirationTime, &a.CreatedAt); err == nil {
			attestations = append(attestations, a)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"attestations": attestations,
		"count":        len(attestations),
	})
}

// GetAttestationsByRecipient handles GET /api/v1/attestations/by-recipient/{address}
func (h *Handler) GetAttestationsByRecipient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	rows, err := h.db.Query(`
		SELECT ea.uid, ea.schema_uid, ea.attester, ea.ipfs_cid, ea.encrypted_data_hash, ea.revocable, ea.revoked, ea.expiration_time, ea.created_at
		FROM encrypted_attestations ea
		INNER JOIN attestation_recipients ar ON ea.uid = ar.attestation_uid
		WHERE ar.recipient = $1
		ORDER BY ea.created_at DESC
	`, address)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var attestations []EncryptedAttestationResponse
	for rows.Next() {
		var a EncryptedAttestationResponse
		if err := rows.Scan(&a.UID, &a.SchemaUID, &a.Attester, &a.IPFSCID, &a.EncryptedDataHash, &a.Revocable, &a.Revoked, &a.ExpirationTime, &a.CreatedAt); err == nil {
			attestations = append(attestations, a)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"attestations": attestations,
		"count":        len(attestations),
	})
}

