package handlers

import (
	"encoding/json"
	"testing"
	"time"
)

// MockDB implements a mock database for testing
type MockDB struct {
	attestations map[string]EncryptedAttestationResponse
	recipients   map[string][]RecipientKey
}

func NewMockDB() *MockDB {
	return &MockDB{
		attestations: make(map[string]EncryptedAttestationResponse),
		recipients:   make(map[string][]RecipientKey),
	}
}

func TestCreateEncryptedAttestationRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateEncryptedAttestationRequest
		expectValid bool
		expectError string
	}{
		{
			name: "valid request",
			request: CreateEncryptedAttestationRequest{
				SchemaUID:         "0x" + "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				IPFSCID:           "QmYwAPJzv5CZsnAzt8auVZRn5W7x8Hd8fH6FS6NVQP3fSw",
				EncryptedDataHash: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				Recipients: []RecipientKey{
					{Address: "0x1234567890123456789012345678901234567890", EncryptedKey: "0xencryptedkey"},
				},
				Revocable: true,
				Signature: "0xsignature",
			},
			expectValid: true,
		},
		{
			name: "empty recipients",
			request: CreateEncryptedAttestationRequest{
				SchemaUID:         "0x" + "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				IPFSCID:           "QmYwAPJzv5CZsnAzt8auVZRn5W7x8Hd8fH6FS6NVQP3fSw",
				EncryptedDataHash: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				Recipients:        []RecipientKey{},
				Revocable:         true,
				Signature:         "0xsignature",
			},
			expectValid: false,
			expectError: "At least one recipient required",
		},
		{
			name: "too many recipients (>50)",
			request: CreateEncryptedAttestationRequest{
				SchemaUID:         "0x" + "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				IPFSCID:           "QmYwAPJzv5CZsnAzt8auVZRn5W7x8Hd8fH6FS6NVQP3fSw",
				EncryptedDataHash: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				Recipients:        make([]RecipientKey, 51),
				Revocable:         true,
				Signature:         "0xsignature",
			},
			expectValid: false,
			expectError: "Maximum 50 recipients allowed",
		},
		{
			name: "invalid IPFS CID (too short)",
			request: CreateEncryptedAttestationRequest{
				SchemaUID:         "0x" + "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				IPFSCID:           "Qm123",
				EncryptedDataHash: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				Recipients: []RecipientKey{
					{Address: "0x1234567890123456789012345678901234567890", EncryptedKey: "0xencryptedkey"},
				},
				Revocable: true,
				Signature: "0xsignature",
			},
			expectValid: false,
			expectError: "Invalid IPFS CID",
		},
		{
			name: "invalid encrypted data hash (wrong length)",
			request: CreateEncryptedAttestationRequest{
				SchemaUID:         "0x" + "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				IPFSCID:           "QmYwAPJzv5CZsnAzt8auVZRn5W7x8Hd8fH6FS6NVQP3fSw",
				EncryptedDataHash: "tooshort",
				Recipients: []RecipientKey{
					{Address: "0x1234567890123456789012345678901234567890", EncryptedKey: "0xencryptedkey"},
				},
				Revocable: true,
				Signature: "0xsignature",
			},
			expectValid: false,
			expectError: "Invalid encrypted data hash (must be 32 bytes hex)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errMsg := validateAttestationRequest(tt.request)
			if valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v", tt.expectValid, valid)
			}
			if !tt.expectValid && errMsg != tt.expectError {
				t.Errorf("expected error=%q, got error=%q", tt.expectError, errMsg)
			}
		})
	}
}

// validateAttestationRequest validates the request per whitepaper constraints
func validateAttestationRequest(req CreateEncryptedAttestationRequest) (bool, string) {
	if len(req.Recipients) == 0 {
		return false, "At least one recipient required"
	}
	if len(req.Recipients) > 50 {
		return false, "Maximum 50 recipients allowed"
	}
	if len(req.IPFSCID) < 46 {
		return false, "Invalid IPFS CID"
	}
	if len(req.EncryptedDataHash) != 64 {
		return false, "Invalid encrypted data hash (must be 32 bytes hex)"
	}
	return true, ""
}

func TestEncryptedAttestationResponse_JSON(t *testing.T) {
	now := time.Now()
	expTime := now.Add(24 * time.Hour)

	resp := EncryptedAttestationResponse{
		UID:               "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		SchemaUID:         "0xschema",
		Attester:          "cert1...",
		IPFSCID:           "QmYwAPJzv5CZsnAzt8auVZRn5W7x8Hd8fH6FS6NVQP3fSw",
		EncryptedDataHash: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		Recipients:        []string{"0x1234567890123456789012345678901234567890"},
		Revocable:         true,
		Revoked:           false,
		ExpirationTime:    &expTime,
		CreatedAt:         now,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var decoded EncryptedAttestationResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.UID != resp.UID {
		t.Errorf("expected UID=%s, got UID=%s", resp.UID, decoded.UID)
	}
	if decoded.Revocable != resp.Revocable {
		t.Errorf("expected Revocable=%v, got Revocable=%v", resp.Revocable, decoded.Revocable)
	}
}
