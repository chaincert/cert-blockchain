// Package database provides PostgreSQL database access for CertID
// Per CertID Section 2.2: user_profiles database
package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// DB wraps the database connection
type DB struct {
	conn   *sql.DB
	logger *zap.Logger
}

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DefaultConfig returns default database configuration
func DefaultConfig() *Config {
	return &Config{
		Host:    "localhost",
		Port:    5432,
		User:    "cert",
		DBName:  "certid",
		SSLMode: "disable",
	}
}

// New creates a new database connection
func New(config *Config, logger *zap.Logger) (*DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL database",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.DBName),
	)

	return &DB{conn: conn, logger: logger}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// UserProfile represents a CertID profile in the database
type UserProfile struct {
	Address     string            `json:"address"`
	CertIDUID   string            `json:"certid_uid"`
	Name        string            `json:"name"`
	Bio         string            `json:"bio"`
	AvatarURL   string            `json:"avatar_url"`
	SocialLinks map[string]string `json:"social_links"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// GetProfile retrieves a user profile by address
func (db *DB) GetProfile(ctx context.Context, address string) (*UserProfile, error) {
	query := `
		SELECT address, COALESCE(certid_uid, ''), name, bio, avatar_url, social_links, created_at, updated_at
		FROM user_profiles
		WHERE address = $1
	`

	var profile UserProfile
	var socialLinksJSON []byte

	err := db.conn.QueryRowContext(ctx, query, address).Scan(
		&profile.Address,
		&profile.CertIDUID,
		&profile.Name,
		&profile.Bio,
		&profile.AvatarURL,
		&socialLinksJSON,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	if err := json.Unmarshal(socialLinksJSON, &profile.SocialLinks); err != nil {
		profile.SocialLinks = make(map[string]string)
	}

	return &profile, nil
}

// CreateProfile creates a new user profile
func (db *DB) CreateProfile(ctx context.Context, profile *UserProfile) error {
	socialLinksJSON, err := json.Marshal(profile.SocialLinks)
	if err != nil {
		return fmt.Errorf("failed to marshal social links: %w", err)
	}

	query := `
		INSERT INTO user_profiles (address, name, bio, avatar_url, social_links)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (address) DO UPDATE SET
			name = EXCLUDED.name,
			bio = EXCLUDED.bio,
			avatar_url = EXCLUDED.avatar_url,
			social_links = EXCLUDED.social_links
		RETURNING created_at, updated_at
	`

	return db.conn.QueryRowContext(ctx, query,
		profile.Address,
		profile.Name,
		profile.Bio,
		profile.AvatarURL,
		socialLinksJSON,
	).Scan(&profile.CreatedAt, &profile.UpdatedAt)
}

// UpdateProfile updates an existing user profile
func (db *DB) UpdateProfile(ctx context.Context, address string, updates map[string]interface{}) error {
	// Build dynamic update query
	// For simplicity, using a full update approach
	profile, err := db.GetProfile(ctx, address)
	if err != nil {
		return err
	}
	if profile == nil {
		profile = &UserProfile{Address: address, SocialLinks: make(map[string]string)}
	}

	if name, ok := updates["name"].(string); ok {
		profile.Name = name
	}
	if bio, ok := updates["bio"].(string); ok {
		profile.Bio = bio
	}
	if avatarURL, ok := updates["avatar_url"].(string); ok {
		profile.AvatarURL = avatarURL
	}
	if socialLinks, ok := updates["social_links"].(map[string]string); ok {
		profile.SocialLinks = socialLinks
	}

	return db.CreateProfile(ctx, profile)
}

// Transaction represents a transaction record for the explorer
type Transaction struct {
	Hash          string                 `json:"hash"`
	Status        string                 `json:"status"`
	BlockNumber   int64                  `json:"block_number"`
	Timestamp     time.Time              `json:"timestamp"`
	FromAddress   string                 `json:"from"`
	ToAddress     string                 `json:"to"`
	ValueCert     string                 `json:"value_cert"`
	GasLimit      int64                  `json:"gas_limit"`
	GasUsed       int64                  `json:"gas_used"`
	GasPrice      int64                  `json:"gas_price"`
	TxFee         int64                  `json:"tx_fee"`
	InputData     string                 `json:"input_data"`
	EcosystemType string                 `json:"ecosystem_type"`
	CertHash      string                 `json:"cert_hash,omitempty"`
	Metadata      string                 `json:"metadata,omitempty"`
	DecodedParams map[string]interface{} `json:"decoded_params,omitempty"`
}

// GetTransaction retrieves a transaction by hash
func (db *DB) GetTransaction(ctx context.Context, hash string) (*Transaction, error) {
	query := `
		SELECT hash, status, block_number, timestamp, from_address, to_address,
			   value_cert, gas_limit, gas_used, gas_price, tx_fee, input_data,
			   ecosystem_type, cert_hash, metadata, decoded_params
		FROM transactions
		WHERE hash = $1
	`

	var tx Transaction
	var decodedParamsJSON []byte
	var certHash, metadata sql.NullString

	err := db.conn.QueryRowContext(ctx, query, hash).Scan(
		&tx.Hash,
		&tx.Status,
		&tx.BlockNumber,
		&tx.Timestamp,
		&tx.FromAddress,
		&tx.ToAddress,
		&tx.ValueCert,
		&tx.GasLimit,
		&tx.GasUsed,
		&tx.GasPrice,
		&tx.TxFee,
		&tx.InputData,
		&tx.EcosystemType,
		&certHash,
		&metadata,
		&decodedParamsJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if certHash.Valid {
		tx.CertHash = certHash.String
	}
	if metadata.Valid {
		tx.Metadata = metadata.String
	}
	if err := json.Unmarshal(decodedParamsJSON, &tx.DecodedParams); err != nil {
		tx.DecodedParams = make(map[string]interface{})
	}

	return &tx, nil
}

// SaveTransaction stores a transaction in the database
func (db *DB) SaveTransaction(ctx context.Context, tx *Transaction) error {
	decodedParamsJSON, err := json.Marshal(tx.DecodedParams)
	if err != nil {
		decodedParamsJSON = []byte("{}")
	}

	query := `
		INSERT INTO transactions (
			hash, status, block_number, timestamp, from_address, to_address,
			value_cert, gas_limit, gas_used, gas_price, tx_fee, input_data,
			ecosystem_type, cert_hash, metadata, decoded_params
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (hash) DO UPDATE SET
			status = EXCLUDED.status,
			gas_used = EXCLUDED.gas_used,
			decoded_params = EXCLUDED.decoded_params
	`

	_, err = db.conn.ExecContext(ctx, query,
		tx.Hash,
		tx.Status,
		tx.BlockNumber,
		tx.Timestamp,
		tx.FromAddress,
		tx.ToAddress,
		tx.ValueCert,
		tx.GasLimit,
		tx.GasUsed,
		tx.GasPrice,
		tx.TxFee,
		tx.InputData,
		tx.EcosystemType,
		sql.NullString{String: tx.CertHash, Valid: tx.CertHash != ""},
		sql.NullString{String: tx.Metadata, Valid: tx.Metadata != ""},
		decodedParamsJSON,
	)

	return err
}

// GetAddressLabel retrieves a label for an address
func (db *DB) GetAddressLabel(ctx context.Context, address string) (string, error) {
	query := `SELECT label FROM address_labels WHERE address = $1`
	var label string
	err := db.conn.QueryRowContext(ctx, query, address).Scan(&label)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return label, err
}

// APIKey represents a developer API key
type APIKey struct {
	ID           string     `json:"id"`
	OwnerAddress string     `json:"owner_address"`
	KeyPrefix    string     `json:"key_prefix"`
	Name         string     `json:"name"`
	RateLimit    int        `json:"rate_limit"`
	Active       bool       `json:"active"`
	TotalReqs    int64      `json:"total_requests"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// GetAPIKeys retrieves all API keys for an owner
func (db *DB) GetAPIKeys(ctx context.Context, ownerAddress string) ([]APIKey, error) {
	query := `
		SELECT id, owner_address, key_prefix, name, rate_limit, active,
		       total_requests, last_used_at, created_at, expires_at
		FROM api_keys
		WHERE owner_address = $1
		ORDER BY created_at DESC`

	rows, err := db.conn.QueryContext(ctx, query, ownerAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		err := rows.Scan(&k.ID, &k.OwnerAddress, &k.KeyPrefix, &k.Name, &k.RateLimit,
			&k.Active, &k.TotalReqs, &k.LastUsedAt, &k.CreatedAt, &k.ExpiresAt)
		if err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// CreateAPIKey creates a new API key
func (db *DB) CreateAPIKey(ctx context.Context, ownerAddress, keyHash, keyPrefix, name string) (*APIKey, error) {
	query := `
		INSERT INTO api_keys (owner_address, key_hash, key_prefix, name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, owner_address, key_prefix, name, rate_limit, active,
		          total_requests, last_used_at, created_at, expires_at`

	var k APIKey
	err := db.conn.QueryRowContext(ctx, query, ownerAddress, keyHash, keyPrefix, name).Scan(
		&k.ID, &k.OwnerAddress, &k.KeyPrefix, &k.Name, &k.RateLimit,
		&k.Active, &k.TotalReqs, &k.LastUsedAt, &k.CreatedAt, &k.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

// DeleteAPIKey deletes an API key
func (db *DB) DeleteAPIKey(ctx context.Context, keyID, ownerAddress string) error {
	query := `DELETE FROM api_keys WHERE id = $1 AND owner_address = $2`
	result, err := db.conn.ExecContext(ctx, query, keyID, ownerAddress)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ValidateAPIKey validates an API key hash and returns the key info if valid
func (db *DB) ValidateAPIKey(ctx context.Context, keyHash string) (*APIKey, error) {
	query := `
		SELECT id, owner_address, key_prefix, name, rate_limit, active,
		       total_requests, last_used_at, created_at, expires_at
		FROM api_keys
		WHERE key_hash = $1 AND active = true
		  AND (expires_at IS NULL OR expires_at > NOW())`

	var k APIKey
	err := db.conn.QueryRowContext(ctx, query, keyHash).Scan(
		&k.ID, &k.OwnerAddress, &k.KeyPrefix, &k.Name, &k.RateLimit,
		&k.Active, &k.TotalReqs, &k.LastUsedAt, &k.CreatedAt, &k.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

// IncrementAPIKeyUsage increments the request count for an API key
func (db *DB) IncrementAPIKeyUsage(ctx context.Context, keyID string) error {
	query := `UPDATE api_keys SET total_requests = total_requests + 1, last_used_at = NOW() WHERE id = $1`
	_, err := db.conn.ExecContext(ctx, query, keyID)
	return err
}

// APIUsageStats holds usage statistics
type APIUsageStats struct {
	TotalRequests   int64 `json:"total_requests"`
	RequestsToday   int64 `json:"requests_today"`
	RequestsThisWeek int64 `json:"requests_this_week"`
	AvgResponseMs   int   `json:"avg_response_ms"`
}

// GetAPIUsage retrieves usage statistics for an owner
func (db *DB) GetAPIUsage(ctx context.Context, ownerAddress string) (*APIUsageStats, error) {
	stats := &APIUsageStats{}

	// Get total requests across all keys
	query := `SELECT COALESCE(SUM(total_requests), 0) FROM api_keys WHERE owner_address = $1`
	db.conn.QueryRowContext(ctx, query, ownerAddress).Scan(&stats.TotalRequests)

	return stats, nil
}

// FaucetTransaction represents a faucet disbursement
type FaucetTransaction struct {
	ID        string    `json:"id"`
	TxHash    string    `json:"tx_hash"`
	Recipient string    `json:"recipient_address"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// RecordFaucetTransaction stores a faucet transaction in the database
func (db *DB) RecordFaucetTransaction(ctx context.Context, txHash, recipient string, amount int64) error {
	query := `
		INSERT INTO faucet_transactions (tx_hash, recipient_address, amount, status)
		VALUES ($1, $2, $3, 'completed')
		ON CONFLICT (tx_hash) DO NOTHING`
	_, err := db.conn.ExecContext(ctx, query, txHash, recipient, amount)
	return err
}

// GetFaucetBalance returns the total amount received by an address from the faucet (in ucert)
func (db *DB) GetFaucetBalance(ctx context.Context, address string) (int64, error) {
	query := `SELECT COALESCE(SUM(amount), 0) FROM faucet_transactions
	          WHERE recipient_address = $1 AND status = 'completed'`
	var total int64
	err := db.conn.QueryRowContext(ctx, query, address).Scan(&total)
	return total, err
}

// GetFaucetTransactions returns faucet transactions for an address
func (db *DB) GetFaucetTransactions(ctx context.Context, address string, limit int) ([]FaucetTransaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query := `SELECT id, tx_hash, recipient_address, amount, status, created_at
	          FROM faucet_transactions
	          WHERE recipient_address = $1
	          ORDER BY created_at DESC LIMIT $2`
	rows, err := db.conn.QueryContext(ctx, query, address, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []FaucetTransaction
	for rows.Next() {
		var tx FaucetTransaction
		if err := rows.Scan(&tx.ID, &tx.TxHash, &tx.Recipient, &tx.Amount, &tx.Status, &tx.CreatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, rows.Err()
}
