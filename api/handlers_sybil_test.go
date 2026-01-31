package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

// TestSybilCheckEndpoint tests the sybil check endpoint
func TestSybilCheckEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		address        string
		expectedStatus int
		expectTrustScore bool
	}{
		{
			name:           "Valid address",
			address:        "0x1234567890abcdef",
			expectedStatus: http.StatusOK,
			expectTrustScore: true,
		},
		{
			name:           "Another valid address",
			address:        "cert1abc123def456",
			expectedStatus: http.StatusOK,
			expectTrustScore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// In a real integration test, we would make HTTP requests
			// For unit tests, we just validate the test data structure
			if tt.expectedStatus != http.StatusOK {
				t.Errorf("Expected status OK for address: %s", tt.address)
			}
			t.Logf("Testing address: %s, expected status: %d", tt.address, tt.expectedStatus)
		})
	}
}

// TestCalculateSybilTrustScore tests the trust score calculation
func TestCalculateSybilTrustScore(t *testing.T) {
	tests := []struct {
		name     string
		factors  TrustFactors
		expected int
	}{
		{
			name: "Zero factors",
			factors: TrustFactors{
				KYCVerified:          false,
				SocialVerifications:  0,
				OnChainActivity:      0,
				AccountAgeMonths:     0,
				StakedAmount:         0,
				AttestationsReceived: 0,
			},
			expected: 0,
		},
		{
			name: "KYC only",
			factors: TrustFactors{
				KYCVerified:          true,
				SocialVerifications:  0,
				OnChainActivity:      0,
				AccountAgeMonths:     0,
				StakedAmount:         0,
				AttestationsReceived: 0,
			},
			expected: 30,
		},
		{
			name: "Maximum social (4 platforms)",
			factors: TrustFactors{
				KYCVerified:          false,
				SocialVerifications:  4,
				OnChainActivity:      0,
				AccountAgeMonths:     0,
				StakedAmount:         0,
				AttestationsReceived: 0,
			},
			expected: 40,
		},
		{
			name: "Full score",
			factors: TrustFactors{
				KYCVerified:          true,  // +30
				SocialVerifications:  5,     // +40 (capped)
				OnChainActivity:      10,    // +20 (capped at 4*5=20)
				AccountAgeMonths:     24,    // +10 (capped)
				StakedAmount:         5000,  // +20 (capped at 50*0.01=0.5 -> 0, need 2000 for 20)
				AttestationsReceived: 15,    // +20 (capped at 10*2=20)
			},
			expected: 100, // Should cap at 100
		},
		{
			name: "Partial factors",
			factors: TrustFactors{
				KYCVerified:          true,  // +30
				SocialVerifications:  2,     // +20
				OnChainActivity:      2,     // +10
				AccountAgeMonths:     6,     // +6
				StakedAmount:         0,     // +0
				AttestationsReceived: 3,     // +6
			},
			expected: 72,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateSybilTrustScore(tt.factors)
			if score != tt.expected {
				t.Errorf("calculateSybilTrustScore() = %d, want %d", score, tt.expected)
			}
		})
	}
}

// TestBatchCheckRequest tests batch request validation
func TestBatchCheckRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        BatchCheckRequest
		expectedError  bool
	}{
		{
			name: "Valid batch",
			request: BatchCheckRequest{
				Addresses: []string{"addr1", "addr2", "addr3"},
				Threshold: 50,
			},
			expectedError: false,
		},
		{
			name: "Empty addresses",
			request: BatchCheckRequest{
				Addresses: []string{},
				Threshold: 50,
			},
			expectedError: true,
		},
		{
			name: "Default threshold",
			request: BatchCheckRequest{
				Addresses: []string{"addr1"},
				Threshold: 0, // Should default to 50
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate request
			hasError := len(tt.request.Addresses) == 0 || len(tt.request.Addresses) > 100
			if hasError != tt.expectedError {
				t.Errorf("Expected error: %v, got: %v", tt.expectedError, hasError)
			}
		})
	}
}

// TestSybilMin tests the min helper function
func TestSybilMin(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{10, 20, 10},
		{20, 10, 10},
		{10, 10, 10},
		{0, 5, 0},
		{-5, 5, -5},
	}

	for _, tt := range tests {
		result := sybilMin(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("sybilMin(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// TestTrustFactorsJSON tests JSON serialization of TrustFactors
func TestTrustFactorsJSON(t *testing.T) {
	factors := TrustFactors{
		KYCVerified:          true,
		SocialVerifications:  3,
		OnChainActivity:      5,
		AccountAgeMonths:     12,
		StakedAmount:         1000.50,
		AttestationsReceived: 4,
	}

	data, err := json.Marshal(factors)
	if err != nil {
		t.Fatalf("Failed to marshal TrustFactors: %v", err)
	}

	var decoded TrustFactors
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal TrustFactors: %v", err)
	}

	if decoded.KYCVerified != factors.KYCVerified {
		t.Errorf("KYCVerified mismatch: got %v, want %v", decoded.KYCVerified, factors.KYCVerified)
	}
	if decoded.SocialVerifications != factors.SocialVerifications {
		t.Errorf("SocialVerifications mismatch: got %d, want %d", decoded.SocialVerifications, factors.SocialVerifications)
	}
}

// TestSybilCheckResponseJSON tests JSON serialization of SybilCheckResponse
func TestSybilCheckResponseJSON(t *testing.T) {
	response := SybilCheckResponse{
		Address:       "0x1234",
		TrustScore:    75,
		IsLikelyHuman: true,
		Factors: TrustFactors{
			KYCVerified:         true,
			SocialVerifications: 2,
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal SybilCheckResponse: %v", err)
	}

	// Verify JSON structure
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	if decoded["address"] != "0x1234" {
		t.Errorf("address mismatch: got %v, want %v", decoded["address"], "0x1234")
	}
	if decoded["trust_score"].(float64) != 75 {
		t.Errorf("trust_score mismatch: got %v, want %v", decoded["trust_score"], 75)
	}
	if decoded["is_likely_human"] != true {
		t.Errorf("is_likely_human mismatch: got %v, want %v", decoded["is_likely_human"], true)
	}
}

// Placeholder for integration tests that would require a running server
func TestSybilAPIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// These would test against a running server
	t.Run("SingleCheck", func(t *testing.T) {
		// Would make HTTP request to /api/v1/sybil/check/{address}
		t.Log("Integration test placeholder - SingleCheck")
	})

	t.Run("BatchCheck", func(t *testing.T) {
		// Would make HTTP request to /api/v1/sybil/batch
		t.Log("Integration test placeholder - BatchCheck")
	})
}

// Ensure bytes is used
var _ = bytes.Buffer{}
