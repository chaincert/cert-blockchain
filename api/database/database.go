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
		SELECT address, name, bio, avatar_url, social_links, created_at, updated_at
		FROM user_profiles
		WHERE address = $1
	`

	var profile UserProfile
	var socialLinksJSON []byte

	err := db.conn.QueryRowContext(ctx, query, address).Scan(
		&profile.Address,
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

