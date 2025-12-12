package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestRegisterSchemaRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     RegisterSchemaRequest
		expectValid bool
		expectError string
	}{
		{
			name: "valid schema request",
			request: RegisterSchemaRequest{
				Schema:    "string name, uint256 age, bool verified",
				Resolver:  "",
				Revocable: true,
				Creator:   "cert1abc123...",
				Signature: "0xsignature",
			},
			expectValid: true,
		},
		{
			name: "valid schema with resolver",
			request: RegisterSchemaRequest{
				Schema:    "bytes32 attestationUID, string comment",
				Resolver:  "0x1234567890123456789012345678901234567890",
				Revocable: true,
				Creator:   "cert1abc123...",
				Signature: "0xsignature",
			},
			expectValid: true,
		},
		{
			name: "empty schema",
			request: RegisterSchemaRequest{
				Schema:    "",
				Revocable: true,
				Creator:   "cert1abc123...",
				Signature: "0xsignature",
			},
			expectValid: false,
			expectError: "Schema definition required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errMsg := validateSchemaRequest(tt.request)
			if valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v", tt.expectValid, valid)
			}
			if !tt.expectValid && errMsg != tt.expectError {
				t.Errorf("expected error=%q, got error=%q", tt.expectError, errMsg)
			}
		})
	}
}

// validateSchemaRequest validates the schema registration request
func validateSchemaRequest(req RegisterSchemaRequest) (bool, string) {
	if req.Schema == "" {
		return false, "Schema definition required"
	}
	return true, ""
}

func TestSchemaResponse_JSON(t *testing.T) {
	now := time.Now()

	resp := SchemaResponse{
		UID:       "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		Creator:   "cert1creator...",
		Schema:    "string name, uint256 age, bool verified",
		Resolver:  "0x1234567890123456789012345678901234567890",
		Revocable: true,
		CreatedAt: now,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var decoded SchemaResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.UID != resp.UID {
		t.Errorf("expected UID=%s, got UID=%s", resp.UID, decoded.UID)
	}
	if decoded.Schema != resp.Schema {
		t.Errorf("expected Schema=%s, got Schema=%s", resp.Schema, decoded.Schema)
	}
	if decoded.Revocable != resp.Revocable {
		t.Errorf("expected Revocable=%v, got Revocable=%v", resp.Revocable, decoded.Revocable)
	}
}

func TestSchemaUID_Generation(t *testing.T) {
	// Test that the same schema produces the same UID
	schema1 := "string name, uint256 age"
	resolver1 := ""
	revocable1 := true

	uid1 := generateSchemaUID(schema1, resolver1, revocable1)
	uid2 := generateSchemaUID(schema1, resolver1, revocable1)

	if uid1 != uid2 {
		t.Error("same schema should produce same UID")
	}

	// Test that different schemas produce different UIDs
	schema2 := "string name, uint256 balance"
	uid3 := generateSchemaUID(schema2, resolver1, revocable1)

	if uid1 == uid3 {
		t.Error("different schemas should produce different UIDs")
	}

	// Test that UID format is valid
	if len(uid1) != 66 || uid1[:2] != "0x" {
		t.Errorf("invalid UID format: %s", uid1)
	}
}

// generateSchemaUID generates a deterministic schema UID
func generateSchemaUID(schema, resolver string, revocable bool) string {
	uidData := fmt.Sprintf("%s%s%t", schema, resolver, revocable)
	hash := sha256.Sum256([]byte(uidData))
	return "0x" + hex.EncodeToString(hash[:])
}
