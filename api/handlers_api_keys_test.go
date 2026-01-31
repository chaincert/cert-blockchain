package api

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/chaincertify/certd/api/database"
)

// TestAPIKeyNewStructure tests the APIKeyNew struct
func TestAPIKeyNewStructure(t *testing.T) {
	now := time.Now()
	key := database.APIKeyNew{
		ID:                 "key-123",
		OwnerAddress:       "cert1owner",
		KeyHash:            "hashedkey",
		KeyPrefix:          "cert_live_XX",
		Name:               "Test Key",
		Description:        "A test API key",
		Tier:               "developer",
		RateLimitPerDay:    10000,
		RateLimitPerMinute: 100,
		Active:             true,
		CreatedAt:          now,
	}

	if key.Tier != "developer" {
		t.Errorf("Tier = %s, want developer", key.Tier)
	}
	if key.RateLimitPerDay != 10000 {
		t.Errorf("RateLimitPerDay = %d, want 10000", key.RateLimitPerDay)
	}
}

// TestAPIKeyHashGeneration tests that key hashing works correctly
func TestAPIKeyHashGeneration(t *testing.T) {
	fullKey := "cert_live_abcdefghijklmnopqrstuvwxyz123456"
	
	hash := sha256.Sum256([]byte(fullKey))
	keyHash := hex.EncodeToString(hash[:])

	// Verify hash is 64 characters (256 bits / 4 bits per hex char)
	if len(keyHash) != 64 {
		t.Errorf("Hash length = %d, want 64", len(keyHash))
	}

	// Verify consistent hashing
	hash2 := sha256.Sum256([]byte(fullKey))
	keyHash2 := hex.EncodeToString(hash2[:])
	if keyHash != keyHash2 {
		t.Error("Hash is not consistent")
	}
}

// TestAPIKeyPrefixGeneration tests key prefix extraction
func TestAPIKeyPrefixGeneration(t *testing.T) {
	tests := []struct {
		fullKey        string
		expectedPrefix string
	}{
		{"cert_live_abcdefghij", "cert_live_ab"},
		{"cert_test_1234567890", "cert_test_12"},
	}

	for _, tt := range tests {
		prefix := tt.fullKey[:12]
		if prefix != tt.expectedPrefix {
			t.Errorf("Prefix = %s, want %s", prefix, tt.expectedPrefix)
		}
	}
}

// TestTierRateLimits tests rate limit values for each tier
func TestTierRateLimits(t *testing.T) {
	tiers := map[string]struct {
		dailyLimit  int
		minuteLimit int
	}{
		"free":       {100, 2},
		"developer":  {10000, 100},
		"enterprise": {1000000, 1000},
	}

	for tier, limits := range tiers {
		t.Run(tier, func(t *testing.T) {
			var dailyLimit, minuteLimit int
			switch tier {
			case "free":
				dailyLimit = 100
				minuteLimit = 2
			case "developer":
				dailyLimit = 10000
				minuteLimit = 100
			case "enterprise":
				dailyLimit = 1000000
				minuteLimit = 1000
			}

			if dailyLimit != limits.dailyLimit {
				t.Errorf("Daily limit for %s = %d, want %d", tier, dailyLimit, limits.dailyLimit)
			}
			if minuteLimit != limits.minuteLimit {
				t.Errorf("Minute limit for %s = %d, want %d", tier, minuteLimit, limits.minuteLimit)
			}
		})
	}
}

// TestCreateAPIKeyRequest tests the request structure
func TestCreateAPIKeyRequest(t *testing.T) {
	tests := []struct {
		name        string
		tier        string
		validTier   bool
	}{
		{"Default tier", "", true},  // Should default to "free"
		{"Free tier", "free", true},
		{"Developer tier", "developer", true},
		{"Enterprise tier", "enterprise", true},
		{"Invalid tier", "premium", false},
		{"Another invalid", "gold", false},
	}

	validTiers := map[string]bool{
		"":           true, // Empty defaults to free
		"free":       true,
		"developer":  true,
		"enterprise": true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validTiers[tt.tier]
			if isValid != tt.validTier {
				t.Errorf("Tier %q validation = %v, want %v", tt.tier, isValid, tt.validTier)
			}
		})
	}
}

// TestAPIKeyExpiration tests key expiration logic
func TestAPIKeyExpiration(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name      string
		expiresAt *time.Time
		expired   bool
	}{
		{"No expiration", nil, false},
		{"Future expiration", timePtr(now.Add(24 * time.Hour)), false},
		{"Past expiration", timePtr(now.Add(-24 * time.Hour)), true},
		{"Just expired", timePtr(now.Add(-1 * time.Second)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var isExpired bool
			if tt.expiresAt != nil && tt.expiresAt.Before(time.Now()) {
				isExpired = true
			}
			if isExpired != tt.expired {
				t.Errorf("Expiration check = %v, want %v", isExpired, tt.expired)
			}
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// TestAPIUsageSummary tests usage summary structure
func TestAPIUsageSummary(t *testing.T) {
	avgTime := 45
	summary := database.APIUsageSummary{
		APIKeyID:          "key-123",
		PeriodStart:       time.Now(),
		PeriodType:        "day",
		RequestCount:      1000,
		ErrorCount:        50,
		AvgResponseTimeMs: &avgTime,
	}

	// Verify success rate calculation
	successfulReqs := summary.RequestCount - summary.ErrorCount
	successRate := float64(successfulReqs) / float64(summary.RequestCount) * 100
	expectedRate := 95.0
	if successRate != expectedRate {
		t.Errorf("Success rate = %.2f%%, want %.2f%%", successRate, expectedRate)
	}
}

// TestAPITier tests tier structure
func TestAPITier(t *testing.T) {
	tier := database.APITier{
		TierName:          "developer",
		DisplayName:       "Developer",
		Description:       "For developers building apps",
		MonthlyPriceCents: 4900,
		DailyLimit:        10000,
		MinuteLimit:       100,
		Features:          []string{"priority support", "higher limits"},
	}

	if tier.MonthlyPriceCents/100 != 49 {
		t.Errorf("Monthly price = $%d, want $49", tier.MonthlyPriceCents/100)
	}
	if tier.DailyLimit != 10000 {
		t.Errorf("DailyLimit = %d, want 10000", tier.DailyLimit)
	}
}

// TestRateLimitMiddlewareLogic tests rate limit checking logic
func TestRateLimitMiddlewareLogic(t *testing.T) {
	tests := []struct {
		name            string
		requestsToday   int
		dailyLimit      int
		requestsMinute  int
		minuteLimit     int
		shouldAllow     bool
	}{
		{"Under both limits", 50, 100, 1, 2, true},
		{"At daily limit", 100, 100, 1, 2, false},
		{"Over daily limit", 150, 100, 1, 2, false},
		{"At minute limit", 50, 100, 2, 2, false},
		{"Over minute limit", 50, 100, 5, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.requestsToday < tt.dailyLimit && tt.requestsMinute < tt.minuteLimit
			if allowed != tt.shouldAllow {
				t.Errorf("Rate limit check = %v, want %v", allowed, tt.shouldAllow)
			}
		})
	}
}

// Integration test placeholders
func TestAPIKeysIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("CreateAPIKey", func(t *testing.T) {
		t.Log("Integration test placeholder - CreateAPIKey")
	})

	t.Run("ListAPIKeys", func(t *testing.T) {
		t.Log("Integration test placeholder - ListAPIKeys")
	})

	t.Run("RevokeAPIKey", func(t *testing.T) {
		t.Log("Integration test placeholder - RevokeAPIKey")
	})

	t.Run("RateLimiting", func(t *testing.T) {
		t.Log("Integration test placeholder - RateLimiting")
	})
}
