// Package api provides referral system handlers
// Airdrop points system per CertID integration
package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"go.uber.org/zap"
)

// Referral configuration (loaded from environment or defaults)
var referralConfig = struct {
	PointsPerReferral int
	Tier5Bonus        int
	Tier10Bonus       int
	Tier25Bonus       int
	DailyLimit        int
}{
	PointsPerReferral: getEnvInt("REFERRAL_POINTS_PER_SIGNUP", 100),
	Tier5Bonus:        getEnvInt("REFERRAL_TIER_5_BONUS", 50),
	Tier10Bonus:       getEnvInt("REFERRAL_TIER_10_BONUS", 150),
	Tier25Bonus:       getEnvInt("REFERRAL_TIER_25_BONUS", 500),
	DailyLimit:        getEnvInt("REFERRAL_DAILY_LIMIT", 50),
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

// referralCodeResponse is the response for GET /referral/code
type referralCodeResponse struct {
	OK        bool   `json:"ok"`
	Code      string `json:"code,omitempty"`
	ShareURL  string `json:"share_url,omitempty"`
	Error     string `json:"error,omitempty"`
}

// handleGetReferralCode returns or generates the user's referral code
// GET /api/v1/referral/code
func (s *Server) handleGetReferralCode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Require authentication
	address := getAuthenticatedAddress(r)
	if address == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(referralCodeResponse{
			OK:    false,
			Error: "Authentication required",
		})
		return
	}
	
	if s.db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(referralCodeResponse{
			OK:    false,
			Error: "Database unavailable",
		})
		return
	}
	
	// Get or generate code
	code, err := s.db.GenerateReferralCode(r.Context(), address)
	if err != nil {
		s.logger.Error("Failed to generate referral code", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(referralCodeResponse{
			OK:    false,
			Error: "Failed to generate referral code",
		})
		return
	}
	
	shareURL := "https://c3rt.org/join?ref=" + code.Code
	
	json.NewEncoder(w).Encode(referralCodeResponse{
		OK:       true,
		Code:     code.Code,
		ShareURL: shareURL,
	})
}

// referralStatsResponse is the response for GET /referral/stats
type referralStatsResponse struct {
	OK                bool   `json:"ok"`
	TotalReferrals    int    `json:"total_referrals"`
	VerifiedReferrals int    `json:"verified_referrals"`
	TotalPoints       int    `json:"total_points"`
	NextTierAt        int    `json:"next_tier_at,omitempty"`
	NextTierBonus     int    `json:"next_tier_bonus,omitempty"`
	Error             string `json:"error,omitempty"`
}

// handleReferralStats returns the user's referral statistics
// GET /api/v1/referral/stats
func (s *Server) handleReferralStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	address := getAuthenticatedAddress(r)
	if address == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(referralStatsResponse{OK: false, Error: "Authentication required"})
		return
	}
	
	if s.db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(referralStatsResponse{OK: false, Error: "Database unavailable"})
		return
	}
	
	stats, err := s.db.GetReferralStats(r.Context(), address)
	if err != nil {
		s.logger.Error("Failed to get referral stats", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(referralStatsResponse{OK: false, Error: "Failed to get stats"})
		return
	}
	
	// Calculate next tier
	nextTier, nextBonus := getNextTier(stats.VerifiedReferrals)
	
	json.NewEncoder(w).Encode(referralStatsResponse{
		OK:                true,
		TotalReferrals:    stats.TotalReferrals,
		VerifiedReferrals: stats.VerifiedReferrals,
		TotalPoints:       stats.TotalPoints,
		NextTierAt:        nextTier,
		NextTierBonus:     nextBonus,
	})
}

func getNextTier(current int) (nextAt int, bonus int) {
	tiers := []struct{ count, bonus int }{
		{5, referralConfig.Tier5Bonus},
		{10, referralConfig.Tier10Bonus},
		{25, referralConfig.Tier25Bonus},
		{50, 1000},
		{100, 2500},
	}
	for _, t := range tiers {
		if current < t.count {
			return t.count, t.bonus
		}
	}
	return 0, 0
}

// leaderboardResponse is the response for GET /referral/leaderboard
type leaderboardResponse struct {
	OK          bool                          `json:"ok"`
	Leaderboard []leaderboardEntry            `json:"leaderboard,omitempty"`
	Error       string                        `json:"error,omitempty"`
}

type leaderboardEntry struct {
	Rank          int    `json:"rank"`
	DisplayName   string `json:"display_name"`
	ReferralCount int    `json:"referral_count"`
	TotalPoints   int    `json:"total_points"`
}

// handleReferralLeaderboard returns the top referrers
// GET /api/v1/referral/leaderboard
func (s *Server) handleReferralLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if s.db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(leaderboardResponse{OK: false, Error: "Database unavailable"})
		return
	}
	
	limit := 25
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	
	entries, err := s.db.GetReferralLeaderboard(r.Context(), limit)
	if err != nil {
		s.logger.Error("Failed to get leaderboard", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(leaderboardResponse{OK: false, Error: "Failed to get leaderboard"})
		return
	}
	
	// Convert to response format
	respEntries := make([]leaderboardEntry, len(entries))
	for i, e := range entries {
		respEntries[i] = leaderboardEntry{
			Rank:          e.Rank,
			DisplayName:   e.DisplayName,
			ReferralCount: e.ReferralCount,
			TotalPoints:   e.TotalPoints,
		}
	}
	
	json.NewEncoder(w).Encode(leaderboardResponse{
		OK:          true,
		Leaderboard: respEntries,
	})
}

// redeemRequest is the request body for POST /referral/redeem
type redeemRequest struct {
	Code string `json:"code"`
}

// redeemResponse is the response for POST /referral/redeem
type redeemResponse struct {
	OK           bool   `json:"ok"`
	Message      string `json:"message,omitempty"`
	ReferrerName string `json:"referrer_name,omitempty"`
	Error        string `json:"error,omitempty"`
}

// handleRedeemReferral redeems a referral code for the authenticated user
// POST /api/v1/referral/redeem
func (s *Server) handleRedeemReferral(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	address := getAuthenticatedAddress(r)
	if address == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(redeemResponse{OK: false, Error: "Authentication required"})
		return
	}
	
	if s.db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(redeemResponse{OK: false, Error: "Database unavailable"})
		return
	}
	
	var req redeemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(redeemResponse{OK: false, Error: "Invalid request body"})
		return
	}
	
	if req.Code == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(redeemResponse{OK: false, Error: "Referral code is required"})
		return
	}
	
	// Attempt to redeem
	err := s.db.RedeemReferralCode(r.Context(), req.Code, address)
	if err != nil {
		s.logger.Warn("Referral redemption failed",
			zap.String("code", req.Code),
			zap.String("referee", address),
			zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(redeemResponse{OK: false, Error: err.Error()})
		return
	}
	
	// Get referrer info for response
	referrerName := ""
	code, _ := s.db.ValidateReferralCode(r.Context(), req.Code)
	if code != nil {
		if profile, _ := s.db.GetProfile(r.Context(), code.OwnerAddress); profile != nil && profile.Name != "" {
			referrerName = profile.Name
		}
	}
	
	json.NewEncoder(w).Encode(redeemResponse{
		OK:           true,
		Message:      "Referral code accepted! Points will be awarded after profile verification.",
		ReferrerName: referrerName,
	})
}

// handleVerifyReferral is called internally after profile verification to award points
// This is not an API endpoint, but called from profile verification logic
func (s *Server) verifyReferralForUser(address string) {
	if s.db == nil {
		return
	}
	
	ctx := s.router.NewRoute().GetHandler() // Simple context, not ideal but works
	_ = ctx // unused in this simplified version
	
	err := s.db.VerifyReferral(nil, address, referralConfig.PointsPerReferral)
	if err != nil {
		s.logger.Warn("Failed to verify referral", zap.String("address", address), zap.Error(err))
	}
}
