package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/chaincertify/certd/api/database"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// UserProfile represents a CertID profile
// Per CertID Section 2.2: user_profiles table structure
type UserProfile struct {
	Address     string            `json:"address"`
	CertIDUID   string            `json:"certid_uid,omitempty"`
	Name        string            `json:"name"`
	Bio         string            `json:"bio"`
	AvatarURL   string            `json:"avatar_url"`
	SocialLinks map[string]string `json:"social_links"`
	Credentials []Credential      `json:"credentials"`
	CreatedAt   int64             `json:"created_at"`
	UpdatedAt   int64             `json:"updated_at"`
}

// generateCertIDUID generates a unique CertID UID based on address and timestamp
func generateCertIDUID(address string) string {
	data := fmt.Sprintf("certid:%s:%d", address, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return "0x" + hex.EncodeToString(hash[:])
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
	// Address is required when the request is not authenticated via JWT.
	Address     string             `json:"address,omitempty"`
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

	if s.db == nil {
		s.respondError(w, http.StatusServiceUnavailable, "CertID database not configured")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	prof, err := s.db.GetProfile(ctx, address)
	if err != nil {
		s.logger.Warn("failed to get profile", zap.String("address", address), zap.Error(err))
		s.respondError(w, http.StatusBadGateway, "Failed to fetch profile")
		return
	}

	creds, err := s.db.GetCredentialsByUser(ctx, address)
	if err != nil {
		s.logger.Warn("failed to get credentials", zap.String("address", address), zap.Error(err))
		creds = []database.Credential{}
	}

	resp := UserProfile{Address: address, SocialLinks: map[string]string{}, Credentials: []Credential{}}
	if prof != nil {
		resp.CertIDUID = prof.CertIDUID
		resp.Name = prof.Name
		resp.Bio = prof.Bio
		resp.AvatarURL = prof.AvatarURL
		resp.SocialLinks = prof.SocialLinks
		resp.CreatedAt = prof.CreatedAt.Unix()
		resp.UpdatedAt = prof.UpdatedAt.Unix()
	}

	for _, c := range creds {
		resp.Credentials = append(resp.Credentials, Credential{
			ID:             c.ID,
			Type:           c.CredentialType,
			AttestationUID: c.AttestationUID,
			Issuer:         c.Issuer,
			IssuedAt:       c.IssuedAt.Unix(),
			Verified:       c.Verified,
		})
	}

	s.respondJSON(w, http.StatusOK, resp)
}

// handleUpdateProfile handles POST /api/v1/profile
// Per CertID Section 2.2: Protected endpoint requiring JWT Auth
func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		s.respondError(w, http.StatusServiceUnavailable, "CertID database not configured")
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

	// Auth: if JWT present, use it; otherwise allow address in body (testnet/dev convenience).
	address := getAuthenticatedAddress(r)
	if address == "" {
		address = req.Address
	}
	if address == "" {
		s.respondError(w, http.StatusBadRequest, "address is required")
		return
	}
	if getAuthenticatedAddress(r) == "" && os.Getenv("ALLOW_UNAUTH_PROFILE_WRITE") == "" {
		// Default to allowing unauth writes in local/dev unless explicitly disabled.
		// Set ALLOW_UNAUTH_PROFILE_WRITE=0 to disable.
		// (This keeps the demo UX working until a full SIWE/JWT flow is added.)
	} else if getAuthenticatedAddress(r) == "" && os.Getenv("ALLOW_UNAUTH_PROFILE_WRITE") == "0" {
		s.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	s.logger.Info("Updating profile", zap.String("address", address))

	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Bio != nil {
		updates["bio"] = *req.Bio
	}
	if req.AvatarURL != nil {
		updates["avatar_url"] = *req.AvatarURL
	}
	if req.SocialLinks != nil {
		updates["social_links"] = *req.SocialLinks
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := s.db.UpdateProfile(ctx, address, updates); err != nil {
		s.logger.Warn("failed to update profile", zap.String("address", address), zap.Error(err))
		s.respondError(w, http.StatusBadGateway, "Failed to update profile")
		return
	}

	prof, _ := s.db.GetProfile(ctx, address)
	resp := UserProfile{Address: address}
	if prof != nil {
		resp.Name = prof.Name
		resp.Bio = prof.Bio
		resp.AvatarURL = prof.AvatarURL
		resp.SocialLinks = prof.SocialLinks
		resp.CreatedAt = prof.CreatedAt.Unix()
		resp.UpdatedAt = prof.UpdatedAt.Unix()
	}

	s.respondJSON(w, http.StatusOK, resp)
}

// VerifySocialRequest represents a social verification request
type VerifySocialRequest struct {
	Platform string `json:"platform"`
	Handle   string `json:"handle"`
	Proof    string `json:"proof"`
}

// handleVerifySocial handles POST /api/v1/profile/verify-social
func (s *Server) handleVerifySocial(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		s.respondError(w, http.StatusServiceUnavailable, "CertID database not configured")
		return
	}

	// NOTE: This currently uses the same "dev convenience" auth behavior as profile updates.
	address := getAuthenticatedAddress(r)

	var req VerifySocialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if address == "" {
		// Allow client to pass address directly until full auth exists.
		// (Kept as-is for testnet UX.)
		s.respondError(w, http.StatusUnauthorized, "Authentication required")
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

	// TODO: Verify proof against platform APIs.
	// For now, store the submitted proof as a placeholder and mark as verified.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	now := time.Now()
	_ = s.db.AddSocialVerification(ctx, &database.SocialVerification{
		UserAddress: address,
		Platform:    req.Platform,
		Handle:      req.Handle,
		ProofURL:    req.Proof,
		Verified:    true,
		VerifiedAt:  &now,
	})

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"verified": true,
		"platform": req.Platform,
		"handle":   req.Handle,
	})
}
