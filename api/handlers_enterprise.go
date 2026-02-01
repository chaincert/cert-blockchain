package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// enterpriseContactRequest represents the contact form submission
type enterpriseContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	UseCase string `json:"useCase"`
	Message string `json:"message"`
}

// enterpriseContactResponse represents the response
type enterpriseContactResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// handleEnterpriseContact handles enterprise sales contact form submissions
func (s *Server) handleEnterpriseContact(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req enterpriseContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondJSON(w, http.StatusBadRequest, enterpriseContactResponse{
			Error: "Invalid request body",
		})
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.Name) == "" {
		s.respondJSON(w, http.StatusBadRequest, enterpriseContactResponse{
			Error: "Name is required",
		})
		return
	}
	if strings.TrimSpace(req.Email) == "" || !strings.Contains(req.Email, "@") {
		s.respondJSON(w, http.StatusBadRequest, enterpriseContactResponse{
			Error: "Valid email is required",
		})
		return
	}
	if strings.TrimSpace(req.Company) == "" {
		s.respondJSON(w, http.StatusBadRequest, enterpriseContactResponse{
			Error: "Company name is required",
		})
		return
	}

	// Log the contact request
	s.logger.Info("Enterprise contact received",
		zap.String("name", req.Name),
		zap.String("email", req.Email),
		zap.String("company", req.Company),
		zap.String("useCase", req.UseCase),
		zap.Time("timestamp", time.Now()),
	)

	// Store in database if available
	if s.db != nil {
		if err := s.db.StoreEnterpriseContact(r.Context(), req.Name, req.Email, req.Company, req.UseCase, req.Message); err != nil {
			s.logger.Warn("Failed to store enterprise contact", zap.Error(err))
			// Don't fail the request, just log the error
		}
	}

	// Format notification message
	notificationMsg := fmt.Sprintf(
		"🏢 New Enterprise Inquiry\n\n"+
			"**Name:** %s\n"+
			"**Email:** %s\n"+
			"**Company:** %s\n"+
			"**Use Case:** %s\n"+
			"**Message:** %s\n"+
			"**Time:** %s",
		req.Name, req.Email, req.Company, req.UseCase, req.Message,
		time.Now().Format(time.RFC3339),
	)

	// Log for email notification (in production, send actual email)
	s.logger.Info("Enterprise notification", zap.String("notification", notificationMsg))

	s.respondJSON(w, http.StatusOK, enterpriseContactResponse{
		OK:      true,
		Message: "Thank you for your interest! Our team will contact you within 24 hours.",
	})
}
