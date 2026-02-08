// Package api provides referral system tests
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chaincertify/certd/api/database"
	"go.uber.org/zap"
)

// TestReferralCodeGeneration tests that referral codes are generated correctly
func TestReferralCodeGeneration(t *testing.T) {
	// Skip if no test database
	db := setupTestDB(t)
	if db == nil {
		t.Skip("No test database available")
	}
	defer db.Close()

	ctx := context.Background()
	address := "cert1testuser12345"

	// Generate code
	code, err := db.GenerateReferralCode(ctx, address)
	if err != nil {
		t.Fatalf("Failed to generate referral code: %v", err)
	}

	if code == nil {
		t.Fatal("Expected code, got nil")
	}

	if len(code.Code) != 8 {
		t.Errorf("Expected 8-char code, got %d chars: %s", len(code.Code), code.Code)
	}

	if code.OwnerAddress != address {
		t.Errorf("Expected owner %s, got %s", address, code.OwnerAddress)
	}

	// Verify idempotency
	code2, err := db.GenerateReferralCode(ctx, address)
	if err != nil {
		t.Fatalf("Second generate failed: %v", err)
	}

	if code.Code != code2.Code {
		t.Error("Same user should get same code")
	}
}

// TestReferralCodeRedemption tests the referral code redemption flow
func TestReferralCodeRedemption(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("No test database available")
	}
	defer db.Close()

	ctx := context.Background()
	referrer := "cert1referrer123"
	referee := "cert1referee456"

	// Generate code for referrer
	code, err := db.GenerateReferralCode(ctx, referrer)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// Redeem code for referee
	err = db.RedeemReferralCode(ctx, code.Code, referee)
	if err != nil {
		t.Fatalf("Failed to redeem code: %v", err)
	}

	// Verify stats updated
	stats, err := db.GetReferralStats(ctx, referrer)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalReferrals != 1 {
		t.Errorf("Expected 1 referral, got %d", stats.TotalReferrals)
	}
}

// TestSelfReferralBlocked tests that users cannot refer themselves
func TestSelfReferralBlocked(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("No test database available")
	}
	defer db.Close()

	ctx := context.Background()
	user := "cert1selfref123"

	code, err := db.GenerateReferralCode(ctx, user)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// Try to redeem own code
	err = db.RedeemReferralCode(ctx, code.Code, user)
	if err == nil {
		t.Error("Expected error for self-referral, got nil")
	}

	if err.Error() != "self-referral not allowed" {
		t.Errorf("Expected 'self-referral not allowed', got: %v", err)
	}
}

// TestDuplicateRefereeBlocked tests that users can only be referred once
func TestDuplicateRefereeBlocked(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("No test database available")
	}
	defer db.Close()

	ctx := context.Background()
	referrer1 := "cert1ref1"
	referrer2 := "cert1ref2"
	referee := "cert1dup"

	code1, _ := db.GenerateReferralCode(ctx, referrer1)
	code2, _ := db.GenerateReferralCode(ctx, referrer2)

	// First redemption should succeed
	err := db.RedeemReferralCode(ctx, code1.Code, referee)
	if err != nil {
		t.Fatalf("First redemption failed: %v", err)
	}

	// Second redemption should fail
	err = db.RedeemReferralCode(ctx, code2.Code, referee)
	if err == nil {
		t.Error("Expected error for duplicate referee, got nil")
	}
}

// TestLeaderboard tests the leaderboard query
func TestLeaderboard(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("No test database available")
	}
	defer db.Close()

	ctx := context.Background()

	// Get leaderboard (may be empty)
	entries, err := db.GetReferralLeaderboard(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to get leaderboard: %v", err)
	}

	// Just verify it returns without error
	_ = entries
}

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *database.DB {
	// Use test database URL from environment
	// This will skip tests if no test DB is available
	logger := zap.NewNop()
	db, err := database.NewFromURL("postgres://cert:cert@localhost:5432/certid_test?sslmode=disable", logger)
	if err != nil {
		return nil
	}
	return db
}

// TestReferralAPIEndpoints tests the HTTP handlers
func TestReferralAPIEndpoints(t *testing.T) {
	t.Run("leaderboard_returns_json", func(t *testing.T) {
		// Create test server
		logger := zap.NewNop()
		config := DefaultConfig()
		server := NewServer(config, logger)

		// Make request to leaderboard (public endpoint)
		req := httptest.NewRequest("GET", "/api/v1/referral/leaderboard", nil)
		rec := httptest.NewRecorder()

		server.router.ServeHTTP(rec, req)

		// Should return JSON even without DB
		if rec.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json, got %s", rec.Header().Get("Content-Type"))
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}
	})

	t.Run("code_requires_auth", func(t *testing.T) {
		logger := zap.NewNop()
		config := DefaultConfig()
		server := NewServer(config, logger)

		req := httptest.NewRequest("GET", "/api/v1/referral/code", nil)
		rec := httptest.NewRecorder()

		server.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", rec.Code)
		}
	})

	t.Run("redeem_requires_code", func(t *testing.T) {
		logger := zap.NewNop()
		config := DefaultConfig()
		server := NewServer(config, logger)

		// Create request with empty body
		req := httptest.NewRequest("POST", "/api/v1/referral/redeem", bytes.NewReader([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.router.ServeHTTP(rec, req)

		// Should fail (either auth or validation)
		if rec.Code == http.StatusOK {
			t.Error("Expected error response for empty code")
		}
	})
}
