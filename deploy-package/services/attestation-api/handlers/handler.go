package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	_ "github.com/lib/pq"
)

// Handler contains all HTTP handlers for the attestation API.
//
// NOTE: This deploy-package copy exists so the deploy artifacts can be built/tested
// with `go test ./...` from the module root without type-check failures.
// The primary implementation lives in `cert-blockchain/services/attestation-api/handlers`.
type Handler struct {
	rpcURL  string
	ipfsURL string
	db      *sql.DB
}

// NewHandler creates a new Handler instance.
func NewHandler(rpcURL, ipfsURL, dbURL string) (*Handler, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Initialize database schema
	if err := initDB(db); err != nil {
		return nil, err
	}

	return &Handler{
		rpcURL:  rpcURL,
		ipfsURL: ipfsURL,
		db:      db,
	}, nil
}

// Close closes database connections.
func (h *Handler) Close() error {
	return h.db.Close()
}

// initDB initializes the database schema.
func initDB(db *sql.DB) error {
	schema := `
	-- Encrypted attestations table
	CREATE TABLE IF NOT EXISTS encrypted_attestations (
		uid VARCHAR(66) PRIMARY KEY,
		schema_uid VARCHAR(66) NOT NULL,
		attester VARCHAR(42) NOT NULL,
		ipfs_cid VARCHAR(100) NOT NULL,
		encrypted_data_hash VARCHAR(66) NOT NULL,
		revocable BOOLEAN DEFAULT true,
		revoked BOOLEAN DEFAULT false,
		expiration_time TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Recipients table for multi-recipient attestations
	CREATE TABLE IF NOT EXISTS attestation_recipients (
		id SERIAL PRIMARY KEY,
		attestation_uid VARCHAR(66) NOT NULL REFERENCES encrypted_attestations(uid),
		recipient VARCHAR(42) NOT NULL,
		encrypted_key TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(attestation_uid, recipient)
	);

	-- Schemas table
	CREATE TABLE IF NOT EXISTS schemas (
		uid VARCHAR(66) PRIMARY KEY,
		creator VARCHAR(42) NOT NULL,
		schema_definition TEXT NOT NULL,
		resolver VARCHAR(42),
		revocable BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- User profiles table (CertID)
	CREATE TABLE IF NOT EXISTS user_profiles (
		address VARCHAR(42) PRIMARY KEY,
		name VARCHAR(255),
		bio TEXT,
		avatar_url TEXT,
		social_links JSONB,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create indexes
	CREATE INDEX IF NOT EXISTS idx_attestations_attester ON encrypted_attestations(attester);
	CREATE INDEX IF NOT EXISTS idx_attestations_ipfs_cid ON encrypted_attestations(ipfs_cid);
	CREATE INDEX IF NOT EXISTS idx_recipients_recipient ON attestation_recipients(recipient);
	CREATE INDEX IF NOT EXISTS idx_recipients_attestation ON attestation_recipients(attestation_uid);
	`

	_, err := db.Exec(schema)
	return err
}

// respondJSON sends a JSON response.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// respondError sends an error response.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
