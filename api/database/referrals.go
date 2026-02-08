// Package database provides referral-related database operations
package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ReferralCode represents a user's referral code
type ReferralCode struct {
	Code          string     `json:"code"`
	OwnerAddress  string     `json:"owner_address"`
	UsesRemaining int        `json:"uses_remaining"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Referral represents a referral relationship
type Referral struct {
	ID              string     `json:"id"`
	ReferrerAddress string     `json:"referrer_address"`
	RefereeAddress  string     `json:"referee_address"`
	ReferralCode    string     `json:"referral_code"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty"`
}

// ReferralStats holds referral statistics for a user
type ReferralStats struct {
	TotalReferrals   int `json:"total_referrals"`
	VerifiedReferrals int `json:"verified_referrals"`
	TotalPoints      int `json:"total_points"`
	Rank             int `json:"rank,omitempty"`
}

// LeaderboardEntry represents a leaderboard row
type LeaderboardEntry struct {
	Rank          int    `json:"rank"`
	Address       string `json:"address"`
	DisplayName   string `json:"display_name,omitempty"`
	ReferralCount int    `json:"referral_count"`
	TotalPoints   int    `json:"total_points"`
}

// generateCode creates a random 8-character alphanumeric code
func generateCode() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(bytes)), nil
}

// GetReferralCode retrieves or generates a referral code for a user
func (db *DB) GetReferralCode(ctx context.Context, address string) (*ReferralCode, error) {
	query := `
		SELECT code, owner_address, uses_remaining, expires_at, created_at
		FROM referral_codes
		WHERE owner_address = $1
	`
	
	var rc ReferralCode
	err := db.conn.QueryRowContext(ctx, query, address).Scan(
		&rc.Code, &rc.OwnerAddress, &rc.UsesRemaining, &rc.ExpiresAt, &rc.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get referral code: %w", err)
	}
	
	return &rc, nil
}

// GenerateReferralCode creates a new referral code for a user
func (db *DB) GenerateReferralCode(ctx context.Context, address string) (*ReferralCode, error) {
	// Check if user already has a code
	existing, err := db.GetReferralCode(ctx, address)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	
	// Generate new code with collision retry
	var code string
	for attempts := 0; attempts < 5; attempts++ {
		code, err = generateCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate code: %w", err)
		}
		
		query := `
			INSERT INTO referral_codes (code, owner_address)
			VALUES ($1, $2)
			RETURNING code, owner_address, uses_remaining, expires_at, created_at
		`
		
		var rc ReferralCode
		err = db.conn.QueryRowContext(ctx, query, code, address).Scan(
			&rc.Code, &rc.OwnerAddress, &rc.UsesRemaining, &rc.ExpiresAt, &rc.CreatedAt,
		)
		if err == nil {
			return &rc, nil
		}
		// If duplicate code, retry
		if !strings.Contains(err.Error(), "duplicate") {
			return nil, fmt.Errorf("failed to create referral code: %w", err)
		}
	}
	
	return nil, fmt.Errorf("failed to generate unique code after 5 attempts")
}

// ValidateReferralCode checks if a code is valid for redemption
func (db *DB) ValidateReferralCode(ctx context.Context, code string) (*ReferralCode, error) {
	query := `
		SELECT code, owner_address, uses_remaining, expires_at, created_at
		FROM referral_codes
		WHERE code = $1
		  AND (uses_remaining = -1 OR uses_remaining > 0)
		  AND (expires_at IS NULL OR expires_at > NOW())
	`
	
	var rc ReferralCode
	err := db.conn.QueryRowContext(ctx, query, strings.ToUpper(code)).Scan(
		&rc.Code, &rc.OwnerAddress, &rc.UsesRemaining, &rc.ExpiresAt, &rc.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to validate code: %w", err)
	}
	
	return &rc, nil
}

// RedeemReferralCode records a referral (called during new user signup)
func (db *DB) RedeemReferralCode(ctx context.Context, code, refereeAddress string) error {
	// Validate code
	rc, err := db.ValidateReferralCode(ctx, code)
	if err != nil {
		return err
	}
	if rc == nil {
		return fmt.Errorf("invalid or expired referral code")
	}
	
	// Prevent self-referral
	if rc.OwnerAddress == refereeAddress {
		return fmt.Errorf("self-referral not allowed")
	}
	
	// Check if referee was already referred
	var exists bool
	err = db.conn.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM referrals WHERE referee_address = $1)",
		refereeAddress,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check existing referral: %w", err)
	}
	if exists {
		return fmt.Errorf("user has already been referred")
	}
	
	// Create referral record
	_, err = db.conn.ExecContext(ctx, `
		INSERT INTO referrals (referrer_address, referee_address, referral_code, status)
		VALUES ($1, $2, $3, 'pending')
	`, rc.OwnerAddress, refereeAddress, rc.Code)
	if err != nil {
		return fmt.Errorf("failed to create referral: %w", err)
	}
	
	// Decrement uses if limited
	if rc.UsesRemaining > 0 {
		_, err = db.conn.ExecContext(ctx,
			"UPDATE referral_codes SET uses_remaining = uses_remaining - 1 WHERE code = $1",
			rc.Code,
		)
		if err != nil {
			db.logger.Warn("Failed to decrement referral code uses", 
				zap.String("code", rc.Code), zap.Error(err))
		}
	}
	
	return nil
}

// VerifyReferral marks a referral as verified and awards points
func (db *DB) VerifyReferral(ctx context.Context, refereeAddress string, pointsPerReferral int) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Get referral and update status
	var referralID, referrerAddress string
	err = tx.QueryRowContext(ctx, `
		UPDATE referrals 
		SET status = 'verified', verified_at = NOW()
		WHERE referee_address = $1 AND status = 'pending'
		RETURNING id, referrer_address
	`, refereeAddress).Scan(&referralID, &referrerAddress)
	
	if err == sql.ErrNoRows {
		return nil // No pending referral
	}
	if err != nil {
		return fmt.Errorf("failed to verify referral: %w", err)
	}
	
	// Award points to referrer
	_, err = tx.ExecContext(ctx, `
		INSERT INTO referral_points (user_address, points, reason, reference_id)
		VALUES ($1, $2, 'referral', $3)
	`, referrerAddress, pointsPerReferral, referralID)
	if err != nil {
		return fmt.Errorf("failed to award points: %w", err)
	}
	
	// Check for tier bonuses
	var verifiedCount int
	err = tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM referrals WHERE referrer_address = $1 AND status = 'verified'",
		referrerAddress,
	).Scan(&verifiedCount)
	if err == nil {
		bonusPoints := calculateTierBonus(verifiedCount)
		if bonusPoints > 0 {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO referral_points (user_address, points, reason, metadata)
				VALUES ($1, $2, 'tier_bonus', $3)
			`, referrerAddress, bonusPoints, fmt.Sprintf(`{"tier_count": %d}`, verifiedCount))
			if err != nil {
				db.logger.Warn("Failed to award tier bonus", zap.Error(err))
			}
		}
	}
	
	return tx.Commit()
}

// calculateTierBonus returns bonus points for reaching referral milestones
func calculateTierBonus(count int) int {
	switch count {
	case 5:
		return 50
	case 10:
		return 150
	case 25:
		return 500
	case 50:
		return 1000
	case 100:
		return 2500
	default:
		return 0
	}
}

// GetReferralStats returns referral statistics for a user
func (db *DB) GetReferralStats(ctx context.Context, address string) (*ReferralStats, error) {
	stats := &ReferralStats{}
	
	// Get referral counts
	err := db.conn.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'verified') as verified
		FROM referrals
		WHERE referrer_address = $1
	`, address).Scan(&stats.TotalReferrals, &stats.VerifiedReferrals)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get referral counts: %w", err)
	}
	
	// Get total points
	err = db.conn.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(points), 0) FROM referral_points WHERE user_address = $1",
		address,
	).Scan(&stats.TotalPoints)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get points: %w", err)
	}
	
	return stats, nil
}

// GetReferralLeaderboard returns the top referrers
func (db *DB) GetReferralLeaderboard(ctx context.Context, limit int) ([]LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	
	query := `
		SELECT 
			r.referrer_address,
			COALESCE(p.name, CONCAT(LEFT(r.referrer_address, 8), '...')) as display_name,
			COUNT(*) as referral_count,
			COALESCE(SUM(rp.points), 0) as total_points
		FROM referrals r
		LEFT JOIN user_profiles p ON p.address = r.referrer_address
		LEFT JOIN referral_points rp ON rp.user_address = r.referrer_address
		WHERE r.status = 'verified'
		GROUP BY r.referrer_address, p.name
		ORDER BY total_points DESC, referral_count DESC
		LIMIT $1
	`
	
	rows, err := db.conn.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer rows.Close()
	
	var entries []LeaderboardEntry
	rank := 1
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.Address, &e.DisplayName, &e.ReferralCount, &e.TotalPoints); err != nil {
			return nil, err
		}
		e.Rank = rank
		// Privacy: mask address
		if len(e.Address) > 12 {
			e.Address = e.Address[:8] + "..." + e.Address[len(e.Address)-4:]
		}
		entries = append(entries, e)
		rank++
	}
	
	return entries, rows.Err()
}

// AddReferralPoints adds points to a user's account
func (db *DB) AddReferralPoints(ctx context.Context, address string, points int, reason string) error {
	_, err := db.conn.ExecContext(ctx, `
		INSERT INTO referral_points (user_address, points, reason)
		VALUES ($1, $2, $3)
	`, address, points, reason)
	return err
}
