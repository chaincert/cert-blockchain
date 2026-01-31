package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Sybil Resistance API Handlers
// Provides trust score validation for airdrop protection, governance, etc.

type TrustFactors struct {
	KYCVerified          bool    `json:"kyc_verified"`
	SocialVerifications  int     `json:"social_verifications"`
	OnChainActivity      int     `json:"onchain_activity"`
	AccountAgeMonths     int     `json:"account_age_months"`
	StakedAmount         float64 `json:"staked_amount"`
	AttestationsReceived int     `json:"attestations_received"`
}

type SybilCheckResponse struct {
	Address       string       `json:"address"`
	TrustScore    int          `json:"trust_score"`
	IsLikelyHuman bool         `json:"is_likely_human"`
	Factors       TrustFactors `json:"factors"`
	CheckedAt     time.Time    `json:"checked_at"`
}

type BatchCheckRequest struct {
	Addresses []string `json:"addresses"`
	Threshold int      `json:"threshold,omitempty"` // Optional threshold (default: 50)
}

type BatchCheckResponse struct {
	Results []SybilCheckResponse `json:"results"`
	Summary struct {
		Total      int `json:"total"`
		LikelyReal int `json:"likely_real"`
		Suspicious int `json:"suspicious"`
	} `json:"summary"`
}

// calculateSybilTrustScore computes a trust score (0-100) based on various factors
func calculateSybilTrustScore(factors TrustFactors) int {
	score := 0

	// KYC verification: +30 points
	if factors.KYCVerified {
		score += 30
	}

	// Social verifications: +10 per platform (max 40)
	score += sybilMin(factors.SocialVerifications*10, 40)

	// On-chain activity: +5 per significant interaction (max 20)
	score += sybilMin(factors.OnChainActivity*5, 20)

	// Account age: +1 per month (max 10)
	score += sybilMin(factors.AccountAgeMonths, 10)

	// Staked amount: +0.01 per CERT staked (max 20)
	stakedScore := int(factors.StakedAmount * 0.01)
	score += sybilMin(stakedScore, 20)

	// Attestations received: +2 per attestation (max 20)
	score += sybilMin(factors.AttestationsReceived*2, 20)

	// Cap at 100
	return sybilMin(score, 100)
}

// getTrustFactors retrieves all trust-related data for an address
func (s *Server) getTrustFactors(ctx context.Context, address string) TrustFactors {
	factors := TrustFactors{}

	// Count verified social accounts using existing method
	if s.db != nil {
		if count, err := s.db.CountVerifiedSocialAccounts(ctx, address); err == nil {
			factors.SocialVerifications = count
		}

		// Get profile age
		if prof, err := s.db.GetProfile(ctx, address); err == nil && prof != nil {
			monthsSinceCreation := int(time.Since(prof.CreatedAt).Hours() / 24 / 30)
			factors.AccountAgeMonths = monthsSinceCreation
		}

		// Count credentials/attestations
		if creds, err := s.db.GetCredentialsByUser(ctx, address); err == nil {
			factors.AttestationsReceived = len(creds)
			// Check if any credential is KYC type
			for _, c := range creds {
				if c.CredentialType == "KYC" || c.CredentialType == "KYC_L1" || c.CredentialType == "IDENTITY" {
					if c.Verified {
						factors.KYCVerified = true
					}
				}
			}
		}
	}

	// Mock values for on-chain data (TODO: implement blockchain queries)
	factors.OnChainActivity = 5
	factors.StakedAmount = 1000.0

	return factors
}

// handleSybilCheck returns trust score for a single address
func (s *Server) handleSybilCheck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	if address == "" {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "missing address"})
		return
	}

	factors := s.getTrustFactors(r.Context(), address)
	trustScore := calculateSybilTrustScore(factors)
	isLikelyHuman := trustScore >= 50 // Default threshold

	response := SybilCheckResponse{
		Address:       address,
		TrustScore:    trustScore,
		IsLikelyHuman: isLikelyHuman,
		Factors:       factors,
		CheckedAt:     time.Now(),
	}

	s.respondJSON(w, http.StatusOK, response)
}

// handleSybilBatchCheck returns trust scores for multiple addresses
func (s *Server) handleSybilBatchCheck(w http.ResponseWriter, r *http.Request) {
	var req BatchCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	// Set default threshold
	threshold := req.Threshold
	if threshold == 0 {
		threshold = 50
	}

	// Limit batch size
	if len(req.Addresses) > 100 {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "batch size limited to 100 addresses"})
		return
	}
	if len(req.Addresses) == 0 {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "no addresses provided"})
		return
	}

	response := BatchCheckResponse{
		Results: make([]SybilCheckResponse, 0, len(req.Addresses)),
	}

	// Process each address
	for _, address := range req.Addresses {
		factors := s.getTrustFactors(r.Context(), address)
		trustScore := calculateSybilTrustScore(factors)
		isLikelyHuman := trustScore >= threshold

		result := SybilCheckResponse{
			Address:       address,
			TrustScore:    trustScore,
			IsLikelyHuman: isLikelyHuman,
			Factors:       factors,
			CheckedAt:     time.Now(),
		}

		response.Results = append(response.Results, result)

		// Update summary
		response.Summary.Total++
		if isLikelyHuman {
			response.Summary.LikelyReal++
		} else {
			response.Summary.Suspicious++
		}
	}

	s.respondJSON(w, http.StatusOK, response)
}

// handleGetTrustScoreHistory returns historical trust score data
func (s *Server) handleGetTrustScoreHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	if address == "" {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "missing address"})
		return
	}

	factors := s.getTrustFactors(r.Context(), address)
	trustScore := calculateSybilTrustScore(factors)

	history := []map[string]interface{}{
		{
			"timestamp":   time.Now(),
			"trust_score": trustScore,
			"factors":     factors,
		},
	}

	s.respondJSON(w, http.StatusOK, history)
}

func sybilMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Ensure zap import is used
var _ = zap.String
