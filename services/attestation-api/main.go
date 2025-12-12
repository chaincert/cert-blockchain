package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/chaincertify/certd/services/attestation-api/handlers"
	"github.com/chaincertify/certd/services/attestation-api/middleware"
)

// CERT Encrypted Attestation Service API
// Per Whitepaper Section 8 - Custom API Endpoints

func main() {
	// Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		rpcURL = "http://localhost:8545"
	}

	ipfsURL := os.Getenv("IPFS_URL")
	if ipfsURL == "" {
		ipfsURL = "http://localhost:5001"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://localhost:5432/cert_attestations?sslmode=disable"
	}

	// Initialize handlers
	h, err := handlers.NewHandler(rpcURL, ipfsURL, dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize handlers: %v", err)
	}
	defer h.Close()

	// Setup router
	r := mux.NewRouter()

	// Apply global middleware
	r.Use(middleware.CORS)
	r.Use(middleware.Logging)
	r.Use(middleware.RateLimit)

	// API v1 routes per Whitepaper Section 8
	api := r.PathPrefix("/api/v1").Subrouter()

	// Encrypted Attestation endpoints
	api.HandleFunc("/encrypted-attestations", h.CreateEncryptedAttestation).Methods("POST", "OPTIONS")
	api.HandleFunc("/encrypted-attestations/{uid}", h.GetEncryptedAttestation).Methods("GET", "OPTIONS")
	api.HandleFunc("/encrypted-attestations/{uid}/retrieve", h.RetrieveEncryptedData).Methods("POST", "OPTIONS")
	api.HandleFunc("/encrypted-attestations/{uid}/revoke", h.RevokeAttestation).Methods("POST", "OPTIONS")

	// Schema endpoints
	api.HandleFunc("/schemas", h.RegisterSchema).Methods("POST", "OPTIONS")
	api.HandleFunc("/schemas/{uid}", h.GetSchema).Methods("GET", "OPTIONS")

	// Query endpoints
	api.HandleFunc("/attestations/by-attester/{address}", h.GetAttestationsByAttester).Methods("GET", "OPTIONS")
	api.HandleFunc("/attestations/by-recipient/{address}", h.GetAttestationsByRecipient).Methods("GET", "OPTIONS")

	// CertID Profile endpoints (Whitepaper CertID Section)
	api.HandleFunc("/profile/{address}", h.GetProfile).Methods("GET", "OPTIONS")
	api.HandleFunc("/profile", h.UpdateProfile).Methods("POST", "OPTIONS")

	// Auth endpoints
	api.HandleFunc("/auth/challenge", h.GetAuthChallenge).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/verify", h.VerifySignature).Methods("POST", "OPTIONS")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"cert-attestation-api"}`))
	}).Methods("GET")

	// Start server
	log.Printf("CERT Encrypted Attestation Service starting on port %s", port)
	log.Printf("RPC URL: %s", rpcURL)
	log.Printf("IPFS URL: %s", ipfsURL)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

