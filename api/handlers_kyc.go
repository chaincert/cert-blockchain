package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chaincertify/certd/api/database"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Didit API configuration - loaded from environment variables
type DiditConfig struct {
	APIKey        string
	WebhookSecret string
	WorkflowID    string
	BaseURL       string
}

func getDiditConfig() *DiditConfig {
	return &DiditConfig{
		APIKey:        os.Getenv("DIDIT_API_KEY"),
		WebhookSecret: os.Getenv("DIDIT_WEBHOOK_SECRET"),
		WorkflowID:    os.Getenv("DIDIT_WORKFLOW_ID"),
		BaseURL:       "https://verification.didit.me",
	}
}

// DiditSessionRequest is the request to create a Didit verification session
type DiditSessionRequest struct {
	WorkflowID string            `json:"workflow_id"`
	Callback   string            `json:"callback,omitempty"`
	VendorData string            `json:"vendor_data"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// DiditSessionResponse is the response from Didit session creation
type DiditSessionResponse struct {
	SessionID     string `json:"session_id"`
	SessionNumber int    `json:"session_number"`
	SessionToken  string `json:"session_token"`
	VendorData    string `json:"vendor_data"`
	Status        string `json:"status"`
	WorkflowID    string `json:"workflow_id"`
	URL           string `json:"url"`
}

// KYCStartRequest is the request from frontend to start KYC
type KYCStartRequest struct {
	CallbackURL string `json:"callback_url,omitempty"`
}

// KYCStartResponse is the response to frontend
type KYCStartResponse struct {
	SessionID  string `json:"session_id"`
	SessionURL string `json:"session_url"`
	Status     string `json:"status"`
}

// KYCStatusResponse is the response for KYC status check
type KYCStatusResponse struct {
	Status      string     `json:"status"`
	SessionID   string     `json:"session_id,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	HasKYC      bool       `json:"has_kyc"`
}

// handleStartKYC creates a new Didit KYC session for the authenticated user
// POST /api/v1/kyc/start
func (s *Server) handleStartKYC(w http.ResponseWriter, r *http.Request) {
	userAddress := getAuthenticatedAddress(r)
	if userAddress == "" {
		s.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Check if user already has approved KYC
	if s.db != nil {
		hasKYC, err := s.db.HasApprovedKYC(ctx, userAddress)
		if err != nil {
			s.logger.Error("Failed to check KYC status", zap.Error(err))
		}
		if hasKYC {
			s.respondError(w, http.StatusConflict, "KYC already approved")
			return
		}

		// Check for pending session
		existing, err := s.db.GetKYCSessionByUserAddress(ctx, userAddress)
		if err != nil {
			s.logger.Error("Failed to get existing KYC session", zap.Error(err))
		}
		if existing != nil && (existing.Status == database.KYCStatusInProgress || existing.Status == database.KYCStatusNotStarted) {
			// Return existing session URL
			s.respondJSON(w, http.StatusOK, KYCStartResponse{
				SessionID:  existing.SessionID,
				SessionURL: existing.SessionURL,
				Status:     existing.Status,
			})
			return
		}
	}

	// Get Didit config
	config := getDiditConfig()
	if config.APIKey == "" || config.WorkflowID == "" {
		s.respondError(w, http.StatusServiceUnavailable, "KYC service not configured")
		return
	}

	// Parse optional callback from request
	var req KYCStartRequest
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}

	// Create Didit session
	diditReq := DiditSessionRequest{
		WorkflowID: config.WorkflowID,
		VendorData: userAddress,
		Callback:   req.CallbackURL,
		Metadata: map[string]string{
			"platform": "CERT Blockchain",
			"address":  userAddress,
		},
	}

	reqBody, _ := json.Marshal(diditReq)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", config.BaseURL+"/v2/session/", bytes.NewReader(reqBody))
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to create request")
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", config.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		s.logger.Error("Didit API request failed", zap.Error(err))
		s.respondError(w, http.StatusBadGateway, "KYC service unavailable")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("Didit API error", zap.Int("status", resp.StatusCode), zap.String("body", string(body)))
		s.respondError(w, http.StatusBadGateway, "KYC service error")
		return
	}

	var diditResp DiditSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&diditResp); err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to parse KYC response")
		return
	}

	// Store session in database
	if s.db != nil {
		session := &database.KYCSession{
			SessionID:   diditResp.SessionID,
			UserAddress: userAddress,
			WorkflowID:  diditResp.WorkflowID,
			Status:      diditResp.Status,
			SessionURL:  diditResp.URL,
			VendorData:  userAddress,
		}
		if err := s.db.CreateKYCSession(ctx, session); err != nil {
			s.logger.Error("Failed to store KYC session", zap.Error(err))
			// Continue anyway - user can still verify
		}
	}

	s.logger.Info("KYC session created",
		zap.String("user", userAddress),
		zap.String("session_id", diditResp.SessionID),
	)

	s.respondJSON(w, http.StatusOK, KYCStartResponse{
		SessionID:  diditResp.SessionID,
		SessionURL: diditResp.URL,
		Status:     diditResp.Status,
	})
}

// handleGetKYCStatus returns the KYC status for the authenticated user
// GET /api/v1/kyc/status
func (s *Server) handleGetKYCStatus(w http.ResponseWriter, r *http.Request) {
	userAddress := getAuthenticatedAddress(r)
	if userAddress == "" {
		s.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if s.db == nil {
		s.respondJSON(w, http.StatusOK, KYCStatusResponse{
			Status: "unknown",
			HasKYC: false,
		})
		return
	}

	// Check for approved KYC
	hasKYC, _ := s.db.HasApprovedKYC(ctx, userAddress)

	// Get latest session
	session, err := s.db.GetKYCSessionByUserAddress(ctx, userAddress)
	if err != nil {
		s.logger.Error("Failed to get KYC session", zap.Error(err))
	}

	resp := KYCStatusResponse{
		Status: "none",
		HasKYC: hasKYC,
	}

	if session != nil {
		resp.Status = session.Status
		resp.SessionID = session.SessionID
		resp.CompletedAt = session.CompletedAt
	}

	s.respondJSON(w, http.StatusOK, resp)
}

// DiditWebhookPayload represents the webhook payload from Didit
type DiditWebhookPayload struct {
	SessionID   string                 `json:"session_id"`
	Status      string                 `json:"status"`
	WebhookType string                 `json:"webhook_type"`
	CreatedAt   int64                  `json:"created_at"`
	Timestamp   int64                  `json:"timestamp"`
	WorkflowID  string                 `json:"workflow_id"`
	VendorData  string                 `json:"vendor_data"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
	Decision    map[string]interface{} `json:"decision,omitempty"`
}

// handleKYCWebhook processes webhooks from Didit
// POST /api/v1/kyc/webhook
func (s *Server) handleKYCWebhook(w http.ResponseWriter, r *http.Request) {
	config := getDiditConfig()
	if config.WebhookSecret == "" {
		s.logger.Error("Webhook secret not configured")
		http.Error(w, "Webhook not configured", http.StatusServiceUnavailable)
		return
	}

	// Read raw body for signature verification
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("Failed to read webhook body", zap.Error(err))
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Get signature and timestamp headers
	signature := r.Header.Get("X-Signature")
	timestamp := r.Header.Get("X-Timestamp")

	if signature == "" || timestamp == "" {
		s.logger.Warn("Missing webhook signature headers")
		http.Error(w, "Missing signature headers", http.StatusUnauthorized)
		return
	}

	// Validate timestamp (within 5 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		http.Error(w, "Invalid timestamp", http.StatusUnauthorized)
		return
	}
	currentTime := time.Now().Unix()
	if abs(currentTime-ts) > 300 {
		s.logger.Warn("Webhook timestamp too old", zap.Int64("timestamp", ts), zap.Int64("current", currentTime))
		http.Error(w, "Request timestamp is stale", http.StatusUnauthorized)
		return
	}

	// Verify HMAC signature
	mac := hmac.New(sha256.New, []byte(config.WebhookSecret))
	mac.Write(rawBody)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		s.logger.Warn("Invalid webhook signature",
			zap.String("expected", expectedSignature),
			zap.String("received", signature),
		)
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse webhook payload
	var payload DiditWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		s.logger.Error("Failed to parse webhook payload", zap.Error(err))
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	s.logger.Info("KYC webhook received",
		zap.String("session_id", payload.SessionID),
		zap.String("status", payload.Status),
		zap.String("webhook_type", payload.WebhookType),
		zap.String("vendor_data", payload.VendorData),
	)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Store decision data as JSON if present
	var decisionJSON *string
	if payload.Decision != nil {
		data, _ := json.Marshal(payload.Decision)
		str := string(data)
		decisionJSON = &str
	}

	// Update session status in database
	if s.db != nil {
		if err := s.db.UpdateKYCSessionStatus(ctx, payload.SessionID, payload.Status, decisionJSON); err != nil {
			s.logger.Error("Failed to update KYC session", zap.Error(err))
		}

		// If approved, add KYC_L1 credential/badge to user
		if payload.Status == database.KYCStatusApproved {
			userAddress := strings.ToLower(payload.VendorData)
			if userAddress != "" {
				credential := &database.Credential{
					UserAddress:    userAddress,
					CredentialType: "KYC_L1",
					AttestationUID: "kyc_didit_" + payload.SessionID,
					Issuer:         "didit.me",
					Verified:       true,
					IssuedAt:       time.Now(),
				}
				if err := s.db.AddCredential(ctx, credential); err != nil {
					s.logger.Error("Failed to add KYC credential", zap.Error(err), zap.String("user", userAddress))
				} else {
					s.logger.Info("KYC_L1 badge awarded", zap.String("user", userAddress))
				}
			}
		}
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Webhook processed"})
}

// abs returns absolute value of int64
func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// handleGetKYCSession returns details of a specific KYC session
// GET /api/v1/kyc/session/{sessionId}
func (s *Server) handleGetKYCSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	userAddress := getAuthenticatedAddress(r)
	if userAddress == "" {
		s.respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if s.db == nil {
		s.respondError(w, http.StatusServiceUnavailable, "Database not available")
		return
	}

	session, err := s.db.GetKYCSessionBySessionID(ctx, sessionID)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to get session")
		return
	}

	if session == nil {
		s.respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	// Verify user owns this session
	if !strings.EqualFold(session.UserAddress, userAddress) {
		s.respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	s.respondJSON(w, http.StatusOK, session)
}

