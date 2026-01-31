package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// DID:web Support Handlers
// Implements W3C Decentralized Identifiers for enterprise integration

// DIDDocument represents a W3C DID Document
type DIDDocument struct {
	Context            interface{}          `json:"@context"`
	ID                 string               `json:"id"`
	Controller         string               `json:"controller,omitempty"`
	VerificationMethod []VerificationMethod `json:"verificationMethod"`
	Authentication     []interface{}        `json:"authentication"`
	AssertionMethod    []interface{}        `json:"assertionMethod,omitempty"`
	Service            []ServiceEndpoint    `json:"service,omitempty"`
	Created            *time.Time           `json:"created,omitempty"`
	Updated            *time.Time           `json:"updated,omitempty"`
}

// VerificationMethod represents a public key or verification method
type VerificationMethod struct {
	ID                  string `json:"id"`
	Type                string `json:"type"`
	Controller          string `json:"controller"`
	PublicKeyMultibase  string `json:"publicKeyMultibase,omitempty"`
	PublicKeyHex        string `json:"publicKeyHex,omitempty"`
	BlockchainAccountID string `json:"blockchainAccountId,omitempty"`
}

// ServiceEndpoint represents a service endpoint in DID document
type ServiceEndpoint struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
	Description     string `json:"description,omitempty"`
}

// WellKnownDIDConfig represents the .well-known/did-configuration.json
type WellKnownDIDConfig struct {
	Context    string   `json:"@context"`
	LinkedDIDs []string `json:"linked_dids"`
}

// handleGetWellKnownDID serves the .well-known/did.json configuration
func (s *Server) handleGetWellKnownDID(w http.ResponseWriter, r *http.Request) {
	config := WellKnownDIDConfig{
		Context: "https://identity.foundation/.well-known/did-configuration/v1",
		LinkedDIDs: []string{
			"did:web:c3rt.org",
		},
	}

	w.Header().Set("Content-Type", "application/did+json")
	s.respondJSON(w, http.StatusOK, config)
}

// handleGetDIDDocument returns the DID document for a specific address
func (s *Server) handleGetDIDDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	if address == "" {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "missing address"})
		return
	}

	// Build DID document using address only (profile data is optional enhancement)
	did := buildDIDDocument(address)

	w.Header().Set("Content-Type", "application/did+json")
	s.respondJSON(w, http.StatusOK, did)
}

// buildDIDDocument constructs a W3C-compliant DID document
func buildDIDDocument(address string) DIDDocument {
	didID := "did:web:c3rt.org:identity:" + address

	doc := DIDDocument{
		Context: []string{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/suites/secp256k1-2019/v1",
		},
		ID:         didID,
		Controller: didID,
	}

	// Add verification method (blockchain address as key)
	verificationMethodID := didID + "#key-1"
	doc.VerificationMethod = []VerificationMethod{
		{
			ID:                  verificationMethodID,
			Type:                "EcdsaSecp256k1VerificationKey2019",
			Controller:          didID,
			BlockchainAccountID: "cosmos:" + address, // CAIP-10 format
		},
	}

	// Set authentication methods
	doc.Authentication = []interface{}{verificationMethodID}
	doc.AssertionMethod = []interface{}{verificationMethodID}

	// Add service endpoints
	doc.Service = []ServiceEndpoint{
		{
			ID:              didID + "#certid-profile",
			Type:            "CertIDProfile",
			ServiceEndpoint: "https://c3rt.org/identity/" + address,
			Description:     "CertID decentralized identity profile",
		},
		{
			ID:              didID + "#trust-score",
			Type:            "TrustScoreAPI",
			ServiceEndpoint: "https://api.c3rt.org/api/v1/sybil/check/" + address,
			Description:     "Real-time trust score validation",
		},
		{
			ID:              didID + "#verifiable-credentials",
			Type:            "VerifiableCredentialRegistry",
			ServiceEndpoint: "https://api.c3rt.org/api/v1/identity/" + address + "/credentials",
			Description:     "Verified credentials and attestations",
		},
	}

	// Add creation/update timestamps
	now := time.Now()
	doc.Created = &now
	doc.Updated = &now

	return doc
}

// handleResolveDID resolves a DID to its document
func (s *Server) handleResolveDID(w http.ResponseWriter, r *http.Request) {
	did := r.URL.Query().Get("did")
	if did == "" {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "missing DID parameter"})
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"didDocument": map[string]string{
			"note": "Use GET /identity/{address}/did.json for full document",
		},
		"didResolutionMetadata": map[string]string{
			"contentType": "application/did+json",
		},
		"didDocumentMetadata": map[string]interface{}{
			"created": time.Now(),
			"updated": time.Now(),
		},
	})
}

// handleGetDIDVerifiablePresentation returns a verifiable presentation of identity
func (s *Server) handleGetDIDVerifiablePresentation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	if address == "" {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "missing address"})
		return
	}

	// Build verifiable presentation
	presentation := map[string]interface{}{
		"@context": []string{
			"https://www.w3.org/2018/credentials/v1",
		},
		"type":   []string{"VerifiablePresentation"},
		"holder": "did:web:c3rt.org:identity:" + address,
		"verifiableCredential": []map[string]interface{}{
			{
				"@context":     []string{"https://www.w3.org/2018/credentials/v1"},
				"type":         []string{"VerifiableCredential", "CertIDCredential"},
				"issuer":       "did:web:c3rt.org",
				"issuanceDate": time.Now().Format(time.RFC3339),
				"credentialSubject": map[string]interface{}{
					"id":           "did:web:c3rt.org:identity:" + address,
					"certidProfile": "https://c3rt.org/identity/" + address,
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/ld+json")
	s.respondJSON(w, http.StatusOK, presentation)
}

// handleExportDIDtoJSON allows downloading DID document as JSON file
func (s *Server) handleExportDIDtoJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	if address == "" {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "missing address"})
		return
	}

	did := buildDIDDocument(address)

	// Set download headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=did-"+address+".json")

	json.NewEncoder(w).Encode(did)
}
