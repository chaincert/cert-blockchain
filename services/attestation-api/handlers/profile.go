package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// CertIDProfile represents a user profile in the CertID system
// Per Whitepaper CertID Section
type CertIDProfile struct {
	Address     string            `json:"address"`
	Name        string            `json:"name,omitempty"`
	Bio         string            `json:"bio,omitempty"`
	AvatarURL   string            `json:"avatarUrl,omitempty"`
	SocialLinks map[string]string `json:"socialLinks,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// UpdateProfileRequest represents the request to update a profile
type UpdateProfileRequest struct {
	Address     string            `json:"address"`
	Name        string            `json:"name,omitempty"`
	Bio         string            `json:"bio,omitempty"`
	AvatarURL   string            `json:"avatarUrl,omitempty"`
	SocialLinks map[string]string `json:"socialLinks,omitempty"`
	Signature   string            `json:"signature"`
}

// GetProfile handles GET /api/v1/profile/{address}
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	var profile CertIDProfile
	var socialLinksJSON []byte

	err := h.db.QueryRow(`
		SELECT address, name, bio, avatar_url, social_links, created_at, updated_at
		FROM user_profiles WHERE address = $1
	`, address).Scan(&profile.Address, &profile.Name, &profile.Bio, &profile.AvatarURL, &socialLinksJSON, &profile.CreatedAt, &profile.UpdatedAt)

	if err == sql.ErrNoRows {
		// Return empty profile for non-existent addresses
		respondJSON(w, http.StatusOK, CertIDProfile{
			Address:     address,
			SocialLinks: make(map[string]string),
		})
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Parse social links JSON
	if len(socialLinksJSON) > 0 {
		json.Unmarshal(socialLinksJSON, &profile.SocialLinks)
	}

	respondJSON(w, http.StatusOK, profile)
}

// UpdateProfile handles POST /api/v1/profile
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Address == "" {
		respondError(w, http.StatusBadRequest, "Address required")
		return
	}

	// TODO: Verify signature to authenticate profile owner

	// Convert social links to JSON
	socialLinksJSON, _ := json.Marshal(req.SocialLinks)

	// Upsert profile
	_, err := h.db.Exec(`
		INSERT INTO user_profiles (address, name, bio, avatar_url, social_links, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
		ON CONFLICT (address) DO UPDATE SET
			name = EXCLUDED.name,
			bio = EXCLUDED.bio,
			avatar_url = EXCLUDED.avatar_url,
			social_links = EXCLUDED.social_links,
			updated_at = CURRENT_TIMESTAMP
	`, req.Address, req.Name, req.Bio, req.AvatarURL, socialLinksJSON)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "updated",
		"address": req.Address,
	})
}

// AuthChallenge represents an authentication challenge
type AuthChallenge struct {
	Challenge string    `json:"challenge"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// GetAuthChallenge handles GET /api/v1/auth/challenge
func (h *Handler) GetAuthChallenge(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		respondError(w, http.StatusBadRequest, "Address required")
		return
	}

	// Generate challenge message
	challenge := AuthChallenge{
		Challenge: "Sign this message to authenticate with CERT Blockchain.\n\n" +
			"Address: " + address + "\n" +
			"Timestamp: " + time.Now().UTC().Format(time.RFC3339),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	respondJSON(w, http.StatusOK, challenge)
}

// VerifySignatureRequest represents the signature verification request
type VerifySignatureRequest struct {
	Address   string `json:"address"`
	Challenge string `json:"challenge"`
	Signature string `json:"signature"`
}

// VerifySignature handles POST /api/v1/auth/verify
func (h *Handler) VerifySignature(w http.ResponseWriter, r *http.Request) {
	var req VerifySignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// TODO: Implement actual signature verification using secp256k1
	// This would verify that the signature was created by the private key
	// corresponding to the provided address

	// For now, return success (placeholder)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"verified": true,
		"address":  req.Address,
	})
}

