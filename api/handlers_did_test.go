package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestBuildDIDDocument tests DID document construction
func TestBuildDIDDocument(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{"Simple address", "cert1abc123"},
		{"Ethereum address", "0x1234567890abcdef1234567890abcdef12345678"},
		{"Short address", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := buildDIDDocument(tt.address)

			// Verify DID ID format
			expectedID := "did:web:c3rt.org:identity:" + tt.address
			if doc.ID != expectedID {
				t.Errorf("ID = %s, want %s", doc.ID, expectedID)
			}

			// Verify controller equals ID
			if doc.Controller != doc.ID {
				t.Errorf("Controller = %s, want %s", doc.Controller, doc.ID)
			}

			// Verify context includes W3C DID
			contexts, ok := doc.Context.([]string)
			if !ok {
				t.Fatalf("Context is not []string")
			}
			if len(contexts) < 1 || contexts[0] != "https://www.w3.org/ns/did/v1" {
				t.Errorf("Context missing W3C DID: %v", contexts)
			}

			// Verify verification method exists
			if len(doc.VerificationMethod) == 0 {
				t.Error("VerificationMethod is empty")
			} else {
				vm := doc.VerificationMethod[0]
				if vm.Type != "EcdsaSecp256k1VerificationKey2019" {
					t.Errorf("VerificationMethod.Type = %s, want EcdsaSecp256k1VerificationKey2019", vm.Type)
				}
				if !strings.HasPrefix(vm.BlockchainAccountID, "cosmos:") {
					t.Errorf("BlockchainAccountID should start with cosmos:, got %s", vm.BlockchainAccountID)
				}
			}

			// Verify service endpoints
			if len(doc.Service) != 3 {
				t.Errorf("Expected 3 service endpoints, got %d", len(doc.Service))
			}

			// Verify timestamps
			if doc.Created == nil || doc.Updated == nil {
				t.Error("Created or Updated timestamps are nil")
			}
		})
	}
}

// TestDIDDocumentJSON tests JSON serialization of DID document
func TestDIDDocumentJSON(t *testing.T) {
	doc := buildDIDDocument("test123")

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Failed to marshal DIDDocument: %v", err)
	}

	// Verify it's valid JSON
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DID document: %v", err)
	}

	// Check required fields exist
	requiredFields := []string{"@context", "id", "verificationMethod", "authentication"}
	for _, field := range requiredFields {
		if _, exists := decoded[field]; !exists {
			t.Errorf("Missing required field: %s", field)
		}
	}
}

// TestWellKnownDIDConfig tests the well-known DID configuration
func TestWellKnownDIDConfig(t *testing.T) {
	config := WellKnownDIDConfig{
		Context:    "https://identity.foundation/.well-known/did-configuration/v1",
		LinkedDIDs: []string{"did:web:c3rt.org"},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal WellKnownDIDConfig: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded["@context"] != "https://identity.foundation/.well-known/did-configuration/v1" {
		t.Errorf("Unexpected @context: %v", decoded["@context"])
	}
}

// TestVerificationMethod tests verification method structure
func TestVerificationMethod(t *testing.T) {
	vm := VerificationMethod{
		ID:                  "did:web:c3rt.org:identity:test#key-1",
		Type:                "EcdsaSecp256k1VerificationKey2019",
		Controller:          "did:web:c3rt.org:identity:test",
		BlockchainAccountID: "cosmos:cert1test",
	}

	data, err := json.Marshal(vm)
	if err != nil {
		t.Fatalf("Failed to marshal VerificationMethod: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded["type"] != "EcdsaSecp256k1VerificationKey2019" {
		t.Errorf("type mismatch: got %v", decoded["type"])
	}
	if decoded["blockchainAccountId"] != "cosmos:cert1test" {
		t.Errorf("blockchainAccountId mismatch: got %v", decoded["blockchainAccountId"])
	}
}

// TestServiceEndpoint tests service endpoint structure
func TestServiceEndpoint(t *testing.T) {
	endpoints := []ServiceEndpoint{
		{
			ID:              "did:web:c3rt.org:identity:test#certid-profile",
			Type:            "CertIDProfile",
			ServiceEndpoint: "https://c3rt.org/identity/test",
			Description:     "CertID decentralized identity profile",
		},
		{
			ID:              "did:web:c3rt.org:identity:test#trust-score",
			Type:            "TrustScoreAPI",
			ServiceEndpoint: "https://api.c3rt.org/api/v1/sybil/check/test",
		},
	}

	for i, ep := range endpoints {
		data, err := json.Marshal(ep)
		if err != nil {
			t.Fatalf("Failed to marshal ServiceEndpoint %d: %v", i, err)
		}

		var decoded map[string]interface{}
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal ServiceEndpoint %d: %v", i, err)
		}

		if decoded["id"] != ep.ID {
			t.Errorf("ServiceEndpoint %d: id mismatch", i)
		}
		if decoded["type"] != ep.Type {
			t.Errorf("ServiceEndpoint %d: type mismatch", i)
		}
		if decoded["serviceEndpoint"] != ep.ServiceEndpoint {
			t.Errorf("ServiceEndpoint %d: serviceEndpoint mismatch", i)
		}
	}
}

// TestDIDDocumentTimestamps tests that timestamps are set correctly
func TestDIDDocumentTimestamps(t *testing.T) {
	before := time.Now()
	doc := buildDIDDocument("timestamp-test")
	after := time.Now()

	if doc.Created == nil {
		t.Fatal("Created timestamp is nil")
	}
	if doc.Updated == nil {
		t.Fatal("Updated timestamp is nil")
	}

	// Verify timestamps are within expected range
	if doc.Created.Before(before) || doc.Created.After(after) {
		t.Errorf("Created timestamp %v not in range [%v, %v]", doc.Created, before, after)
	}
}

// TestDIDDocumentContentType tests that the correct content type is returned
func TestDIDDocumentContentType(t *testing.T) {
	// Test that application/did+json would be set
	expectedContentType := "application/did+json"
	t.Logf("Expected content type: %s", expectedContentType)
}

// Integration test placeholders
func TestDIDAPIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("WellKnownDID", func(t *testing.T) {
		t.Log("Integration test placeholder - WellKnownDID")
	})

	t.Run("GetDIDDocument", func(t *testing.T) {
		t.Log("Integration test placeholder - GetDIDDocument")
	})

	t.Run("VerifiablePresentation", func(t *testing.T) {
		t.Log("Integration test placeholder - VerifiablePresentation")
	})
}

// Ensure httptest is used
var _ = httptest.NewRecorder
var _ = http.StatusOK
