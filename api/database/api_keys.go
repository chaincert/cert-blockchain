package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// APIKeyNew represents an API key for accessing the CertID API (extended version)
type APIKeyNew struct {
	ID                   string     `json:"id"`
	OwnerAddress         string     `json:"owner_address"`
	KeyHash              string     `json:"-"` // Never expose the hash
	KeyPrefix            string     `json:"key_prefix"`
	Name                 string     `json:"name"`
	Description          string     `json:"description"`
	Tier                 string     `json:"tier"`
	RateLimitPerDay      int        `json:"rate_limit_per_day"`
	RateLimitPerMinute   int        `json:"rate_limit_per_minute"`
	Active               bool       `json:"active"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty"`
	StripeCustomerID     *string    `json:"stripe_customer_id,omitempty"`
	BillingEmail         *string    `json:"billing_email,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	LastUsedAt           *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt            *time.Time `json:"expires_at,omitempty"`
}

// APITier represents a pricing tier
type APITier struct {
	TierName          string    `json:"tier_name"`
	DisplayName       string    `json:"display_name"`
	Description       string    `json:"description"`
	DailyLimit        int       `json:"daily_limit"`
	MinuteLimit       int       `json:"minute_limit"`
	MonthlyPriceCents int       `json:"monthly_price_cents"`
	Features          []string  `json:"features"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// APIUsageSummary represents aggregated usage statistics
type APIUsageSummary struct {
	ID                int64     `json:"id"`
	APIKeyID          string    `json:"api_key_id"`
	PeriodType        string    `json:"period_type"`
	PeriodStart       time.Time `json:"period_start"`
	RequestCount      int       `json:"request_count"`
	ErrorCount        int       `json:"error_count"`
	AvgResponseTimeMs *int      `json:"avg_response_time_ms,omitempty"`
	LastUpdated       time.Time `json:"last_updated"`
}

// CreateAPIKeyNew creates a new API key with tier support
func (db *DB) CreateAPIKeyNew(ctx context.Context, key *APIKeyNew) error {
	query := `
		INSERT INTO api_keys (
			owner_address, key_hash, key_prefix, name, description,
			tier, rate_limit_per_day, rate_limit_per_minute, active,
			stripe_subscription_id, stripe_customer_id, billing_email, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at`

	return db.conn.QueryRowContext(ctx, query,
		key.OwnerAddress, key.KeyHash, key.KeyPrefix, key.Name, key.Description,
		key.Tier, key.RateLimitPerDay, key.RateLimitPerMinute, key.Active,
		key.StripeSubscriptionID, key.StripeCustomerID, key.BillingEmail, key.ExpiresAt,
	).Scan(&key.ID, &key.CreatedAt)
}

// GetAPIKeyByHash retrieves an API key by its hash
func (db *DB) GetAPIKeyByHash(ctx context.Context, keyHash string) (*APIKeyNew, error) {
	query := `
		SELECT id, owner_address, key_hash, key_prefix, name, COALESCE(description, ''),
			tier, rate_limit_per_day, rate_limit_per_minute, active,
			stripe_subscription_id, stripe_customer_id, billing_email,
			created_at, last_used_at, expires_at
		FROM api_keys
		WHERE key_hash = $1 AND active = true`

	key := &APIKeyNew{}
	err := db.conn.QueryRowContext(ctx, query, keyHash).Scan(
		&key.ID, &key.OwnerAddress, &key.KeyHash, &key.KeyPrefix, &key.Name, &key.Description,
		&key.Tier, &key.RateLimitPerDay, &key.RateLimitPerMinute, &key.Active,
		&key.StripeSubscriptionID, &key.StripeCustomerID, &key.BillingEmail,
		&key.CreatedAt, &key.LastUsedAt, &key.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return key, err
}

// ListAPIKeysByOwner lists all API keys for a given owner
func (db *DB) ListAPIKeysByOwner(ctx context.Context, ownerAddress string) ([]*APIKeyNew, error) {
	query := `
		SELECT id, owner_address, key_prefix, name, COALESCE(description, ''),
			tier, rate_limit_per_day, rate_limit_per_minute, active,
			created_at, last_used_at, expires_at
		FROM api_keys
		WHERE owner_address = $1
		ORDER BY created_at DESC`

	rows, err := db.conn.QueryContext(ctx, query, ownerAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*APIKeyNew
	for rows.Next() {
		key := &APIKeyNew{}
		err := rows.Scan(
			&key.ID, &key.OwnerAddress, &key.KeyPrefix, &key.Name, &key.Description,
			&key.Tier, &key.RateLimitPerDay, &key.RateLimitPerMinute, &key.Active,
			&key.CreatedAt, &key.LastUsedAt, &key.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

// UpdateAPIKeyLastUsed updates the last_used_at timestamp
func (db *DB) UpdateAPIKeyLastUsed(ctx context.Context, keyID string) error {
	query := `UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := db.conn.ExecContext(ctx, query, keyID)
	return err
}

// RevokeAPIKey deactivates an API key
func (db *DB) RevokeAPIKey(ctx context.Context, keyID, ownerAddress string) error {
	query := `UPDATE api_keys SET active = false WHERE id = $1 AND owner_address = $2`
	_, err := db.conn.ExecContext(ctx, query, keyID, ownerAddress)
	return err
}

// CheckRateLimit checks if an API key has exceeded its rate limits
func (db *DB) CheckRateLimit(ctx context.Context, keyID string, dailyLimit, minuteLimit int) (bool, error) {
	var allowed bool
	query := `SELECT check_rate_limit($1, $2, $3)`
	err := db.conn.QueryRowContext(ctx, query, keyID, dailyLimit, minuteLimit).Scan(&allowed)
	return allowed, err
}

// IncrementAPIUsage increments the usage counters for an API key
func (db *DB) IncrementAPIUsage(ctx context.Context, keyID string, statusCode, responseTimeMs int) error {
	// Call both day and minute increment
	query := `
		SELECT increment_api_usage_summary($1, 'day', date_trunc('day', CURRENT_TIMESTAMP), $2, $3);
		SELECT increment_api_usage_summary($1, 'minute', date_trunc('minute', CURRENT_TIMESTAMP), $2, $3);
	`
	_, err := db.conn.ExecContext(ctx, query, keyID, statusCode, responseTimeMs)
	return err
}

// GetAPITiers retrieves all available API tiers
func (db *DB) GetAPITiers(ctx context.Context) ([]*APITier, error) {
	query := `
		SELECT tier_name, display_name, description, daily_limit, minute_limit,
			monthly_price_cents, features, created_at, updated_at
		FROM api_tiers
		ORDER BY monthly_price_cents ASC`

	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiers []*APITier
	for rows.Next() {
		tier := &APITier{}
		var featuresJSON []byte
		err := rows.Scan(
			&tier.TierName, &tier.DisplayName, &tier.Description,
			&tier.DailyLimit, &tier.MinuteLimit, &tier.MonthlyPriceCents,
			&featuresJSON, &tier.CreatedAt, &tier.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		// Parse features JSON
		if len(featuresJSON) > 0 {
			json.Unmarshal(featuresJSON, &tier.Features)
		}
		tiers = append(tiers, tier)
	}
	return tiers, nil
}

// GetUsageSummary gets usage statistics for an API key
func (db *DB) GetUsageSummary(ctx context.Context, keyID string, periodType string, limit int) ([]*APIUsageSummary, error) {
	query := `
		SELECT id, api_key_id, period_type, period_start, request_count,
			error_count, avg_response_time_ms, last_updated
		FROM api_usage_summary
		WHERE api_key_id = $1 AND period_type = $2
		ORDER BY period_start DESC
		LIMIT $3`

	rows, err := db.conn.QueryContext(ctx, query, keyID, periodType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*APIUsageSummary
	for rows.Next() {
		summary := &APIUsageSummary{}
		err := rows.Scan(
			&summary.ID, &summary.APIKeyID, &summary.PeriodType, &summary.PeriodStart,
			&summary.RequestCount, &summary.ErrorCount, &summary.AvgResponseTimeMs,
			&summary.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}
