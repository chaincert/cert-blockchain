package database

import (
	"context"
	"time"
)

// CreateSocialVerification creates a new verification code for a platform
// Stores the verification code in the 'handle' field
func (db *DB) CreateSocialVerification(ctx context.Context, address, platform, code string, _ time.Time) (*SocialVerification, error) {
	query := `
		INSERT INTO social_verifications (user_address, platform, handle)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_address, platform) DO UPDATE SET
			handle = $3,
			verified = FALSE,
			verified_at = NULL,
			proof_url = NULL,
			created_at = CURRENT_TIMESTAMP
		RETURNING id, user_address, platform, handle, COALESCE(proof_url, ''), verified, verified_at, created_at
	`
	var sv SocialVerification
	err := db.conn.QueryRowContext(ctx, query, address, platform, code).Scan(
		&sv.ID, &sv.UserAddress, &sv.Platform, &sv.Handle,
		&sv.ProofURL, &sv.Verified, &sv.VerifiedAt, &sv.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sv, nil
}

// GetSocialVerificationByAddressPlatform retrieves a verification for an address and platform
func (db *DB) GetSocialVerificationByAddressPlatform(ctx context.Context, address, platform string) (*SocialVerification, error) {
	query := `
		SELECT id, user_address, platform, handle, COALESCE(proof_url, ''), verified, verified_at, created_at
		FROM social_verifications
		WHERE user_address = $1 AND platform = $2
	`
	var sv SocialVerification
	err := db.conn.QueryRowContext(ctx, query, address, platform).Scan(
		&sv.ID, &sv.UserAddress, &sv.Platform, &sv.Handle,
		&sv.ProofURL, &sv.Verified, &sv.VerifiedAt, &sv.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sv, nil
}

// GetVerifiedSocialAccounts retrieves all verified social accounts for an address
func (db *DB) GetVerifiedSocialAccounts(ctx context.Context, address string) ([]SocialVerification, error) {
	query := `
		SELECT id, user_address, platform, handle, COALESCE(proof_url, ''), verified, verified_at, created_at
		FROM social_verifications
		WHERE user_address = $1 AND verified = TRUE
		ORDER BY verified_at DESC
	`
	rows, err := db.conn.QueryContext(ctx, query, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var verifications []SocialVerification
	for rows.Next() {
		var sv SocialVerification
		if err := rows.Scan(
			&sv.ID, &sv.UserAddress, &sv.Platform, &sv.Handle,
			&sv.ProofURL, &sv.Verified, &sv.VerifiedAt, &sv.CreatedAt,
		); err != nil {
			return nil, err
		}
		verifications = append(verifications, sv)
	}
	return verifications, nil
}

// MarkSocialVerificationComplete marks a verification as complete
func (db *DB) MarkSocialVerificationComplete(ctx context.Context, id string, postURL string) error {
	query := `
		UPDATE social_verifications
		SET verified = TRUE, verified_at = CURRENT_TIMESTAMP, proof_url = $2
		WHERE id = $1
	`
	_, err := db.conn.ExecContext(ctx, query, id, postURL)
	return err
}

// CountVerifiedSocialAccounts returns the count of verified social accounts for an address
func (db *DB) CountVerifiedSocialAccounts(ctx context.Context, address string) (int, error) {
	query := `SELECT COUNT(*) FROM social_verifications WHERE user_address = $1 AND verified = TRUE`
	var count int
	err := db.conn.QueryRowContext(ctx, query, address).Scan(&count)
	return count, err
}

