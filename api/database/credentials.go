package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Credential represents a verified credential
type Credential struct {
	ID             string    `json:"id"`
	UserAddress    string    `json:"user_address"`
	CredentialType string    `json:"credential_type"`
	AttestationUID string    `json:"attestation_uid"`
	Issuer         string    `json:"issuer"`
	Verified       bool      `json:"verified"`
	IssuedAt       time.Time `json:"issued_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// GetCredentialsByUser retrieves all credentials for a user
func (db *DB) GetCredentialsByUser(ctx context.Context, address string) ([]Credential, error) {
	query := `
		SELECT id, user_address, credential_type, attestation_uid, issuer, verified, issued_at, created_at
		FROM credentials
		WHERE user_address = $1
		ORDER BY created_at DESC
	`

	rows, err := db.conn.QueryContext(ctx, query, address)
	if err != nil {
		return nil, fmt.Errorf("failed to query credentials: %w", err)
	}
	defer rows.Close()

	var credentials []Credential
	for rows.Next() {
		var c Credential
		if err := rows.Scan(
			&c.ID, &c.UserAddress, &c.CredentialType, &c.AttestationUID,
			&c.Issuer, &c.Verified, &c.IssuedAt, &c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan credential: %w", err)
		}
		credentials = append(credentials, c)
	}

	return credentials, rows.Err()
}

// AddCredential adds a new credential to a user profile
func (db *DB) AddCredential(ctx context.Context, credential *Credential) error {
	// Verify max credentials per profile (50 per types_test.go)
	count, err := db.CountCredentials(ctx, credential.UserAddress)
	if err != nil {
		return err
	}
	if count >= 50 {
		return fmt.Errorf("maximum credentials (50) reached for user")
	}

	query := `
		INSERT INTO credentials (user_address, credential_type, attestation_uid, issuer, verified, issued_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	return db.conn.QueryRowContext(ctx, query,
		credential.UserAddress,
		credential.CredentialType,
		credential.AttestationUID,
		credential.Issuer,
		credential.Verified,
		credential.IssuedAt,
	).Scan(&credential.ID, &credential.CreatedAt)
}

// RemoveCredential removes a credential from a user profile
func (db *DB) RemoveCredential(ctx context.Context, userAddress, credentialID string) error {
	query := `DELETE FROM credentials WHERE id = $1 AND user_address = $2`

	result, err := db.conn.ExecContext(ctx, query, credentialID, userAddress)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// CountCredentials counts the number of credentials for a user
func (db *DB) CountCredentials(ctx context.Context, address string) (int, error) {
	query := `SELECT COUNT(*) FROM credentials WHERE user_address = $1`

	var count int
	err := db.conn.QueryRowContext(ctx, query, address).Scan(&count)
	return count, err
}

// SocialVerification represents a verified social account
type SocialVerification struct {
	ID          string     `json:"id"`
	UserAddress string     `json:"user_address"`
	Platform    string     `json:"platform"`
	Handle      string     `json:"handle"`
	ProofURL    string     `json:"proof_url"`
	Verified    bool       `json:"verified"`
	VerifiedAt  *time.Time `json:"verified_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// GetSocialVerifications retrieves all social verifications for a user
func (db *DB) GetSocialVerifications(ctx context.Context, address string) ([]SocialVerification, error) {
	query := `
		SELECT id, user_address, platform, handle, proof_url, verified, verified_at, created_at
		FROM social_verifications
		WHERE user_address = $1
	`

	rows, err := db.conn.QueryContext(ctx, query, address)
	if err != nil {
		return nil, fmt.Errorf("failed to query social verifications: %w", err)
	}
	defer rows.Close()

	var verifications []SocialVerification
	for rows.Next() {
		var v SocialVerification
		if err := rows.Scan(
			&v.ID, &v.UserAddress, &v.Platform, &v.Handle,
			&v.ProofURL, &v.Verified, &v.VerifiedAt, &v.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan verification: %w", err)
		}
		verifications = append(verifications, v)
	}

	return verifications, rows.Err()
}

// AddSocialVerification adds or updates a social verification
func (db *DB) AddSocialVerification(ctx context.Context, verification *SocialVerification) error {
	query := `
		INSERT INTO social_verifications (user_address, platform, handle, proof_url, verified, verified_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_address, platform) DO UPDATE SET
			handle = EXCLUDED.handle,
			proof_url = EXCLUDED.proof_url,
			verified = EXCLUDED.verified,
			verified_at = EXCLUDED.verified_at
		RETURNING id, created_at
	`

	return db.conn.QueryRowContext(ctx, query,
		verification.UserAddress,
		verification.Platform,
		verification.Handle,
		verification.ProofURL,
		verification.Verified,
		verification.VerifiedAt,
	).Scan(&verification.ID, &verification.CreatedAt)
}

