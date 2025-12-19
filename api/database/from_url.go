package database

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// NewFromURL creates a DB from a standard Postgres URL.
//
// Supported formats:
// - postgres://user:pass@host:5432/dbname?sslmode=disable
// - postgresql://user:pass@host:5432/dbname?sslmode=disable
func NewFromURL(databaseURL string, logger *zap.Logger) (*DB, error) {
	databaseURL = strings.TrimSpace(databaseURL)
	if databaseURL == "" {
		return nil, fmt.Errorf("database url is empty")
	}

	u, err := url.Parse(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database url: %w", err)
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return nil, fmt.Errorf("unsupported database scheme: %s", u.Scheme)
	}

	cfg := DefaultConfig()
	cfg.Host = u.Hostname()

	if portStr := u.Port(); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = p
		}
	}

	if u.User != nil {
		cfg.User = u.User.Username()
		if pw, ok := u.User.Password(); ok {
			cfg.Password = pw
		}
	}

	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName != "" {
		cfg.DBName = dbName
	}

	if ssl := strings.TrimSpace(u.Query().Get("sslmode")); ssl != "" {
		cfg.SSLMode = ssl
	}

	return New(cfg, logger)
}
