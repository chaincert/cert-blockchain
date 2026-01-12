package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Supported social platforms
var supportedPlatforms = map[string]bool{
	"twitter":  true,
	"linkedin": true,
	"facebook": true,
}

// Platform display names
var platformNames = map[string]string{
	"twitter":  "X (Twitter)",
	"linkedin": "LinkedIn",
	"facebook": "Facebook",
}

type socialGenerateRequest struct {
	Platform string `json:"platform"`
}

type socialGenerateResponse struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Platform string `json:"platform"`
	Error    string `json:"error,omitempty"`
}

type socialVerifyRequest struct {
	Platform string `json:"platform"`
	PostURL  string `json:"post_url"`
}

type socialVerifyResponse struct {
	OK       bool   `json:"ok"`
	Platform string `json:"platform,omitempty"`
	Error    string `json:"error,omitempty"`
}

type socialAccountResponse struct {
	Platform   string `json:"platform"`
	Name       string `json:"name"`
	PostURL    string `json:"post_url,omitempty"`
	VerifiedAt int64  `json:"verified_at"`
}

type socialStatusResponse struct {
	Accounts []socialAccountResponse `json:"accounts"`
}

// handleSocialGenerate generates a verification code for a social platform
// POST /api/v1/social/generate
func (s *Server) handleSocialGenerate(w http.ResponseWriter, r *http.Request) {
	address := getAuthenticatedAddress(r)
	if address == "" {
		s.respondJSON(w, http.StatusUnauthorized, socialGenerateResponse{Error: "authentication required"})
		return
	}

	var req socialGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondJSON(w, http.StatusBadRequest, socialGenerateResponse{Error: "invalid request"})
		return
	}

	platform := strings.ToLower(req.Platform)
	if !supportedPlatforms[platform] {
		s.respondJSON(w, http.StatusBadRequest, socialGenerateResponse{Error: "unsupported platform"})
		return
	}

	// Generate unique code
	codeBytes := make([]byte, 8)
	if _, err := rand.Read(codeBytes); err != nil {
		s.respondJSON(w, http.StatusInternalServerError, socialGenerateResponse{Error: "failed to generate code"})
		return
	}
	code := fmt.Sprintf("C3RT-%s", strings.ToUpper(hex.EncodeToString(codeBytes)))

	// Store in database (expires in 24 hours)
	expiresAt := time.Now().Add(24 * time.Hour)
	ctx := r.Context()
	_, err := s.db.CreateSocialVerification(ctx, address, platform, code, expiresAt)
	if err != nil {
		s.logger.Error("failed to create social verification", zap.Error(err))
		s.respondJSON(w, http.StatusInternalServerError, socialGenerateResponse{Error: "failed to create verification"})
		return
	}

	// Generate promotional message
	message := fmt.Sprintf("I just verified my identity on @C3RT_org - the decentralized identity and attestation platform built on blockchain. Join the future of verifiable credentials! 🔐\n\nVerification code: %s\n\nhttps://c3rt.org", code)

	s.respondJSON(w, http.StatusOK, socialGenerateResponse{
		Code:     code,
		Message:  message,
		Platform: platform,
	})
}

// handleSocialVerify verifies a social media post contains the verification code
// POST /api/v1/social/verify
func (s *Server) handleSocialVerify(w http.ResponseWriter, r *http.Request) {
	address := getAuthenticatedAddress(r)
	if address == "" {
		s.respondJSON(w, http.StatusUnauthorized, socialVerifyResponse{Error: "authentication required"})
		return
	}

	var req socialVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondJSON(w, http.StatusBadRequest, socialVerifyResponse{Error: "invalid request"})
		return
	}

	platform := strings.ToLower(req.Platform)
	if !supportedPlatforms[platform] {
		s.respondJSON(w, http.StatusBadRequest, socialVerifyResponse{Error: "unsupported platform"})
		return
	}

	if req.PostURL == "" {
		s.respondJSON(w, http.StatusBadRequest, socialVerifyResponse{Error: "post_url is required"})
		return
	}

	// Validate URL format for platform
	if !isValidPlatformURL(platform, req.PostURL) {
		s.respondJSON(w, http.StatusBadRequest, socialVerifyResponse{Error: "invalid URL for this platform"})
		return
	}

	ctx := r.Context()

	// Get pending verification
	sv, err := s.db.GetSocialVerificationByAddressPlatform(ctx, address, platform)
	if err != nil {
		s.respondJSON(w, http.StatusBadRequest, socialVerifyResponse{Error: "no pending verification found. Please generate a code first."})
		return
	}

	if sv.Verified {
		s.respondJSON(w, http.StatusBadRequest, socialVerifyResponse{Error: "already verified"})
		return
	}

	// Check if code is older than 24 hours
	if time.Since(sv.CreatedAt) > 24*time.Hour {
		s.respondJSON(w, http.StatusBadRequest, socialVerifyResponse{Error: "verification code expired. Please generate a new one."})
		return
	}

	// Fetch the post content and check for the code (stored in Handle field)
	found, err := fetchAndVerifyPost(req.PostURL, sv.Handle)
	if err != nil {
		s.logger.Warn("failed to fetch post", zap.String("url", req.PostURL), zap.Error(err))
		s.respondJSON(w, http.StatusBadRequest, socialVerifyResponse{Error: "could not fetch post. Make sure the post is public."})
		return
	}

	if !found {
		s.respondJSON(w, http.StatusBadRequest, socialVerifyResponse{Error: "verification code not found in the post"})
		return
	}

	// Mark as verified
	if err := s.db.MarkSocialVerificationComplete(ctx, sv.ID, req.PostURL); err != nil {
		s.logger.Error("failed to mark verification complete", zap.Error(err))
		s.respondJSON(w, http.StatusInternalServerError, socialVerifyResponse{Error: "failed to complete verification"})
		return
	}

	s.logger.Info("social verification complete", zap.String("address", address), zap.String("platform", platform))
	s.respondJSON(w, http.StatusOK, socialVerifyResponse{OK: true, Platform: platform})
}

// handleSocialStatus returns verified social accounts for an address
// GET /api/v1/social/{address}
func (s *Server) handleSocialStatus(w http.ResponseWriter, r *http.Request) {
	vars := muxVars(r)
	address := strings.ToLower(vars["address"])
	if address == "" {
		s.respondJSON(w, http.StatusBadRequest, socialStatusResponse{})
		return
	}

	ctx := r.Context()
	verifications, err := s.db.GetVerifiedSocialAccounts(ctx, address)
	if err != nil {
		s.respondJSON(w, http.StatusOK, socialStatusResponse{Accounts: []socialAccountResponse{}})
		return
	}

	accounts := make([]socialAccountResponse, 0, len(verifications))
	for _, v := range verifications {
		acc := socialAccountResponse{
			Platform: v.Platform,
			Name:     platformNames[v.Platform],
			PostURL:  v.ProofURL,
		}
		if v.VerifiedAt != nil {
			acc.VerifiedAt = v.VerifiedAt.Unix()
		}
		accounts = append(accounts, acc)
	}

	s.respondJSON(w, http.StatusOK, socialStatusResponse{Accounts: accounts})
}

// isValidPlatformURL validates that the URL matches the expected platform
func isValidPlatformURL(platform, url string) bool {
	switch platform {
	case "twitter":
		// x.com or twitter.com status URLs
		return strings.Contains(url, "twitter.com/") || strings.Contains(url, "x.com/")
	case "linkedin":
		return strings.Contains(url, "linkedin.com/")
	case "facebook":
		return strings.Contains(url, "facebook.com/")
	}
	return false
}

// fetchAndVerifyPost fetches a URL and checks if it contains the verification code
func fetchAndVerifyPost(url, code string) (bool, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	// Set user agent to avoid blocks
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; C3RT-Bot/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("status %d", resp.StatusCode)
	}

	// Read body (limit to 1MB)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return false, err
	}

	// Check if code is present in the page
	bodyStr := string(body)
	return strings.Contains(bodyStr, code), nil
}

// muxVars helper to get route variables
func muxVars(r *http.Request) map[string]string {
	// Import gorilla/mux Vars if using mux, or use chi/pat
	// For now, simple path parsing
	vars := make(map[string]string)
	// Try to extract from mux if available
	if v := r.Context().Value("vars"); v != nil {
		if m, ok := v.(map[string]string); ok {
			return m
		}
	}
	// Fallback: parse path manually
	path := r.URL.Path
	// Expected: /api/v1/social/{address}
	re := regexp.MustCompile(`/api/v1/social/([^/]+)$`)
	if matches := re.FindStringSubmatch(path); len(matches) > 1 {
		vars["address"] = matches[1]
	}
	return vars
}

