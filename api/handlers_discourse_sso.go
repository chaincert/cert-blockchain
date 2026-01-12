// Package api - Discourse SSO (DiscourseConnect) Handler
// Implements SSO integration for CERT Community Forum
// Reference: https://meta.discourse.org/t/discourseconnect-official-single-sign-on-for-discourse-sso/13045

package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// DiscourseSSOSecret should be set via environment variable
// Generate with: openssl rand -hex 32
var discourseSSOSecret = os.Getenv("DISCOURSE_SSO_SECRET")

// DiscourseUser represents user data for SSO
type DiscourseUser struct {
	ExternalID string // CertID address
	Email      string // User email (can be derived or provided)
	Username   string // Display name or handle
	Name       string // Full name
	AvatarURL  string // Avatar image URL
	Admin      bool   // Is admin
	Moderator  bool   // Is moderator
	Groups     string // Comma-separated group names
}

// handleDiscourseSSOLogin handles the SSO login flow from Discourse
// GET /api/v1/discourse/sso?sso=...&sig=...
func (s *Server) handleDiscourseSSOLogin(w http.ResponseWriter, r *http.Request) {
	ssoPayload := r.URL.Query().Get("sso")
	sig := r.URL.Query().Get("sig")

	if ssoPayload == "" || sig == "" {
		s.respondError(w, http.StatusBadRequest, "Missing sso or sig parameter")
		return
	}

	// Verify HMAC signature
	if !verifyDiscourseSignature(ssoPayload, sig) {
		s.logger.Warn("Invalid Discourse SSO signature")
		s.respondError(w, http.StatusForbidden, "Invalid signature")
		return
	}

	// Decode the payload to get nonce and return_sso_url
	decodedPayload, err := base64.StdEncoding.DecodeString(ssoPayload)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid payload encoding")
		return
	}

	params, err := url.ParseQuery(string(decodedPayload))
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid payload format")
		return
	}

	nonce := params.Get("nonce")
	returnURL := params.Get("return_sso_url")

	if nonce == "" || returnURL == "" {
		s.respondError(w, http.StatusBadRequest, "Missing nonce or return URL")
		return
	}

	// Check if user is authenticated (via JWT cookie or header)
	address := getAuthenticatedAddress(r)
	if address == "" {
		// Redirect to login page with return URL
		loginURL := fmt.Sprintf("https://c3rt.org/login?redirect=%s",
			url.QueryEscape(r.URL.String()))
		http.Redirect(w, r, loginURL, http.StatusFound)
		return
	}

	// Get user profile from database
	ctx := r.Context()
	profile, _ := s.db.GetProfile(ctx, address)

	// Build user data for Discourse
	user := DiscourseUser{
		ExternalID: address,
		Email:      fmt.Sprintf("%s@wallet.c3rt.org", discourseShortAddress(address)), // Pseudo-email from wallet
		Username:   discourseShortAddress(address),
		Name:       "",
		AvatarURL:  "",
		Admin:      false,
		Moderator:  false,
		Groups:     "cert_users",
	}

	// Populate from profile if available
	if profile != nil {
		if profile.Name != "" {
			user.Name = profile.Name
			user.Username = sanitizeUsername(profile.Name)
		}
		if profile.AvatarURL != "" {
			user.AvatarURL = profile.AvatarURL
		}
	}

	s.logger.Info("Discourse SSO login",
		zap.String("address", address),
		zap.String("username", user.Username))

	// Build response payload
	responsePayload := buildDiscoursePayload(nonce, user)

	// Sign and redirect back to Discourse
	encodedPayload := base64.StdEncoding.EncodeToString([]byte(responsePayload))
	responseSig := signDiscoursePayload(encodedPayload)

	redirectURL := fmt.Sprintf("%s?sso=%s&sig=%s",
		returnURL,
		url.QueryEscape(encodedPayload),
		url.QueryEscape(responseSig))

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// handleDiscourseLogout handles SSO logout
// POST /api/v1/discourse/logout
func (s *Server) handleDiscourseLogout(w http.ResponseWriter, r *http.Request) {
	// Clear session and redirect to main site
	http.Redirect(w, r, "https://c3rt.org", http.StatusFound)
}

// verifyDiscourseSignature verifies the HMAC-SHA256 signature from Discourse
func verifyDiscourseSignature(payload, signature string) bool {
	if discourseSSOSecret == "" {
		return false
	}
	expected := computeHMAC(payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// computeHMAC computes HMAC-SHA256 of the payload
func computeHMAC(payload string) string {
	h := hmac.New(sha256.New, []byte(discourseSSOSecret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// signDiscoursePayload signs the response payload
func signDiscoursePayload(payload string) string {
	return computeHMAC(payload)
}

// buildDiscoursePayload builds the SSO response payload
func buildDiscoursePayload(nonce string, user DiscourseUser) string {
	params := url.Values{}
	params.Set("nonce", nonce)
	params.Set("external_id", user.ExternalID)
	params.Set("email", user.Email)
	params.Set("username", user.Username)

	if user.Name != "" {
		params.Set("name", user.Name)
	}
	if user.AvatarURL != "" {
		params.Set("avatar_url", user.AvatarURL)
	}
	if user.Admin {
		params.Set("admin", "true")
	}
	if user.Moderator {
		params.Set("moderator", "true")
	}
	if user.Groups != "" {
		params.Set("add_groups", user.Groups)
	}

	// Suppress welcome message for returning users
	params.Set("suppress_welcome_message", "true")

	return params.Encode()
}

// discourseShortAddress returns a shortened wallet address for Discourse display
func discourseShortAddress(address string) string {
	if len(address) <= 12 {
		return address
	}
	return address[:6] + "..." + address[len(address)-4:]
}

// sanitizeUsername creates a valid Discourse username from a name
func sanitizeUsername(name string) string {
	// Discourse usernames: 3-20 chars, alphanumeric and underscores
	username := strings.ToLower(name)
	username = strings.ReplaceAll(username, " ", "_")

	// Remove non-alphanumeric chars except underscore
	var result strings.Builder
	for _, r := range username {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}

	username = result.String()

	// Ensure length constraints
	if len(username) < 3 {
		username = username + "_user"
	}
	if len(username) > 20 {
		username = username[:20]
	}

	return username
}

// RegisterDiscourseRoutes registers Discourse SSO routes
func (s *Server) RegisterDiscourseRoutes(router *mux.Router) {
	router.HandleFunc("/discourse/sso", s.handleDiscourseSSOLogin).Methods("GET")
	router.HandleFunc("/discourse/logout", s.handleDiscourseLogout).Methods("POST", "GET")
}

// Placeholder for storing SSO sessions (use Redis in production)
var ssoSessions = make(map[string]time.Time)

