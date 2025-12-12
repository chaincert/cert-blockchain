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

// CreateEncryptedAttestationRequest represents the request body
// Per Whitepaper Section 3.2 - Encrypted Attestation Flow
type CreateEncryptedAttestationRequest struct {
	SchemaUID         string            `json:"schemaUID"`
	IPFSCID           string            `json:"ipfsCID"`
	EncryptedDataHash string            `json:"encryptedDataHash"`
	Recipients        []RecipientKey    `json:"recipients"`
	Revocable         bool              `json:"revocable"`
	ExpirationTime    *time.Time        `json:"expirationTime,omitempty"`
	Signature         string            `json:"signature"`
}

// RecipientKey represents a recipient and their encrypted symmetric key
type RecipientKey struct {
	Address      string `json:"address"`
	EncryptedKey string `json:"encryptedKey"`
}

// EncryptedAttestationResponse represents the response
type EncryptedAttestationResponse struct {
	UID               string         `json:"uid"`
	SchemaUID         string         `json:"schemaUID"`
	Attester          string         `json:"attester"`
	IPFSCID           string         `json:"ipfsCID"`
	EncryptedDataHash string         `json:"encryptedDataHash"`
	Recipients        []string       `json:"recipients"`
	Revocable         bool           `json:"revocable"`
	Revoked           bool           `json:"revoked"`
	ExpirationTime    *time.Time     `json:"expirationTime,omitempty"`
	CreatedAt         time.Time      `json:"createdAt"`
}

// CreateEncryptedAttestation handles POST /api/v1/encrypted-attestations
// Implements Step 4 of Whitepaper Section 3.2 - On-Chain Anchoring
func (h *Handler) CreateEncryptedAttestation(w http.ResponseWriter, r *http.Request) {
	var req CreateEncryptedAttestationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request per Whitepaper Section 12 constraints
	if len(req.Recipients) == 0 {
		respondError(w, http.StatusBadRequest, "At least one recipient required")
		return
	}
	if len(req.Recipients) > 50 {
		respondError(w, http.StatusBadRequest, "Maximum 50 recipients allowed")
		return
	}
	if len(req.IPFSCID) < 46 {
		respondError(w, http.StatusBadRequest, "Invalid IPFS CID")
		return
	}
	if len(req.EncryptedDataHash) != 64 {
		respondError(w, http.StatusBadRequest, "Invalid encrypted data hash (must be 32 bytes hex)")
		return
	}

	// TODO: Verify signature to get attester address
	attester := "cert1..." // Placeholder - would be extracted from signature verification

	// Generate UID
	uidData := fmt.Sprintf("%s%s%d%s", req.SchemaUID, attester, time.Now().UnixNano(), req.EncryptedDataHash)
	hash := sha256.Sum256([]byte(uidData))
	uid := "0x" + hex.EncodeToString(hash[:])

	// Store in database
	tx, err := h.db.Begin()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer tx.Rollback()

	// Insert attestation
	_, err = tx.Exec(`
		INSERT INTO encrypted_attestations (uid, schema_uid, attester, ipfs_cid, encrypted_data_hash, revocable, expiration_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, uid, req.SchemaUID, attester, req.IPFSCID, req.EncryptedDataHash, req.Revocable, req.ExpirationTime)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create attestation")
		return
	}

	// Insert recipients
	for _, recipient := range req.Recipients {
		_, err = tx.Exec(`
			INSERT INTO attestation_recipients (attestation_uid, recipient, encrypted_key)
			VALUES ($1, $2, $3)
		`, uid, recipient.Address, recipient.EncryptedKey)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to add recipient")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	// Return response
	recipients := make([]string, len(req.Recipients))
	for i, r := range req.Recipients {
		recipients[i] = r.Address
	}

	respondJSON(w, http.StatusCreated, EncryptedAttestationResponse{
		UID:               uid,
		SchemaUID:         req.SchemaUID,
		Attester:          attester,
		IPFSCID:           req.IPFSCID,
		EncryptedDataHash: req.EncryptedDataHash,
		Recipients:        recipients,
		Revocable:         req.Revocable,
		Revoked:           false,
		ExpirationTime:    req.ExpirationTime,
		CreatedAt:         time.Now(),
	})
}

// GetEncryptedAttestation handles GET /api/v1/encrypted-attestations/{uid}
func (h *Handler) GetEncryptedAttestation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	var resp EncryptedAttestationResponse
	err := h.db.QueryRow(`
		SELECT uid, schema_uid, attester, ipfs_cid, encrypted_data_hash, revocable, revoked, expiration_time, created_at
		FROM encrypted_attestations WHERE uid = $1
	`, uid).Scan(&resp.UID, &resp.SchemaUID, &resp.Attester, &resp.IPFSCID, &resp.EncryptedDataHash, &resp.Revocable, &resp.Revoked, &resp.ExpirationTime, &resp.CreatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "Attestation not found")
		return
	}

	// Get recipients
	rows, err := h.db.Query(`SELECT recipient FROM attestation_recipients WHERE attestation_uid = $1`, uid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get recipients")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var recipient string
		if err := rows.Scan(&recipient); err == nil {
			resp.Recipients = append(resp.Recipients, recipient)
		}
	}

	respondJSON(w, http.StatusOK, resp)
}

// RetrieveEncryptedData handles POST /api/v1/encrypted-attestations/{uid}/retrieve
// Implements Step 5 of Whitepaper Section 3.2 - Retrieval & Decryption
func (h *Handler) RetrieveEncryptedData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	var req struct {
		Requester string `json:"requester"`
		Signature string `json:"signature"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// TODO: Verify signature to authenticate requester

	// Check if requester is authorized
	var encryptedKey string
	err := h.db.QueryRow(`
		SELECT encrypted_key FROM attestation_recipients 
		WHERE attestation_uid = $1 AND recipient = $2
	`, uid, req.Requester).Scan(&encryptedKey)
	if err != nil {
		respondError(w, http.StatusForbidden, "Not authorized to access this attestation")
		return
	}

	// Get attestation details
	var ipfsCID string
	err = h.db.QueryRow(`SELECT ipfs_cid FROM encrypted_attestations WHERE uid = $1 AND revoked = false`, uid).Scan(&ipfsCID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Attestation not found or revoked")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"ipfsCID":      ipfsCID,
		"encryptedKey": encryptedKey,
	})
}

// RevokeAttestation handles POST /api/v1/encrypted-attestations/{uid}/revoke
func (h *Handler) RevokeAttestation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	var req struct {
		Attester  string `json:"attester"`
		Signature string `json:"signature"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// TODO: Verify signature and check attester ownership

	result, err := h.db.Exec(`
		UPDATE encrypted_attestations SET revoked = true, updated_at = CURRENT_TIMESTAMP
		WHERE uid = $1 AND attester = $2 AND revocable = true
	`, uid, req.Attester)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to revoke attestation")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusBadRequest, "Attestation not found, not owned by attester, or not revocable")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

