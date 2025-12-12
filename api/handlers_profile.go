package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// UserProfile represents a CertID profile
// Per CertID Section 2.2: user_profiles table structure
type UserProfile struct {
	Address     string            `json:"address"`
	Name        string            `json:"name"`
	Bio         string            `json:"bio"`
	AvatarURL   string            `json:"avatar_url"`
	SocialLinks map[string]string `json:"social_links"`
	Credentials []Credential      `json:"credentials"`
	CreatedAt   int64             `json:"created_at"`
	UpdatedAt   int64             `json:"updated_at"`
}

// Credential represents a verified credential linked to a profile
type Credential struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	AttestationUID string `json:"attestation_uid"`
	Issuer         string `json:"issuer"`
	IssuedAt       int64  `json:"issued_at"`
	Verified       bool   `json:"verified"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Name        *string            `json:"name,omitempty"`
	Bio         *string            `json:"bio,omitempty"`
	AvatarURL   *string            `json:"avatar_url,omitempty"`
	SocialLinks *map[string]string `json:"social_links,omitempty"`
}

// handleGetProfile handles GET /api/v1/profile/{address}
// Per CertID Section 2.2: Public endpoint
func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	if address == "" {
		s.respondError(w, http.StatusBadRequest, "address is required")
		return
	}

	s.logger.Info("Getting profile", zap.String("address", address))

	// TODO: Query PostgreSQL user_profiles table
	profile := UserProfile{
		Address:     address,
		Name:        "",
		Bio:         "",
		AvatarURL:   "",
		SocialLinks: map[string]string{},
		Credentials: []Credential{},
		CreatedAt:   0,
		UpdatedAt:   0,
	}

	s.respondJSON(w, http.StatusOK, profile)
}

// handleUpdateProfile handles POST /api/v1/profile
// Per CertID Section 2.2: Protected endpoint requiring JWT Auth
func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		// Extract UserAddressKey from JWT (set by authMiddleware)
		address := getAuthenticatedAddress(r)
		if address == "" {
			s.respondError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		var req UpdateProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.respondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate bio length (max 500 chars per types_test.go)
		if req.Bio != nil && len(*req.Bio) > 500 {
			s.respondError(w, http.StatusBadRequest, "Bio must be 500 characters or less")
			return
		}

		// Validate name length (max 100 chars)
		if req.Name != nil && len(*req.Name) > 100 {
			s.respondError(w, http.StatusBadRequest, "Name must be 100 characters or less")
			return
		}

		s.logger.Info("Updating profile",
			zap.String("address", address),
		)

		// TODO: Update PostgreSQL user_profiles table

		profile := UserProfile{
			Address:   address,
			UpdatedAt: getCurrentTimestamp(),
		}
		if req.Name != nil {
			profile.Name = *req.Name
		}
		if req.Bio != nil {
			profile.Bio = *req.Bio
		}
		if req.AvatarURL != nil {
			profile.AvatarURL = *req.AvatarURL
		}
		if req.SocialLinks != nil {
			profile.SocialLinks = *req.SocialLinks
		}

		s.respondJSON(w, http.StatusOK, profile)
	})(w, r)
}

// VerifySocialRequest represents a social verification request
type VerifySocialRequest struct {
	Platform string `json:"platform"`
	Handle   string `json:"handle"`
	Proof    string `json:"proof"`
}

// handleVerifySocial handles POST /api/v1/profile/verify-social
func (s *Server) handleVerifySocial(w http.ResponseWriter, r *http.Request) {
	s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		address := getAuthenticatedAddress(r)

		var req VerifySocialRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.respondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate supported platforms
		supportedPlatforms := map[string]bool{
			"twitter":  true,
			"github":   true,
			"linkedin": true,
			"discord":  true,
		}
		if !supportedPlatforms[req.Platform] {
			s.respondError(w, http.StatusBadRequest, "Unsupported platform")
			return
		}

		s.logger.Info("Verifying social account",
			zap.String("address", address),
			zap.String("platform", req.Platform),
			zap.String("handle", req.Handle),
		)

		// TODO: Verify proof and update profile

		s.respondJSON(w, http.StatusOK, map[string]interface{}{
			"verified": true,
			"platform": req.Platform,
			"handle":   req.Handle,
		})
	})(w, r)
}

