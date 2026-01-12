package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// KYCSession represents a Didit KYC verification session
type KYCSession struct {
	ID            string     `json:"id"`
	SessionID     string     `json:"session_id"`      // Didit session_id
	UserAddress   string     `json:"user_address"`    // CERT wallet address
	WorkflowID    string     `json:"workflow_id"`     // Didit workflow_id
	Status        string     `json:"status"`          // Not Started, In Progress, Approved, Declined, Abandoned
	SessionURL    string     `json:"session_url"`     // URL to redirect user to
	VendorData    string     `json:"vendor_data"`     // Our reference (user address)
	DecisionData  *string    `json:"decision_data"`   // JSON decision payload from webhook
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}

// KYC status constants matching Didit statuses
const (
	KYCStatusNotStarted = "Not Started"
	KYCStatusInProgress = "In Progress"
	KYCStatusInReview   = "In Review"
	KYCStatusApproved   = "Approved"
	KYCStatusDeclined   = "Declined"
	KYCStatusAbandoned  = "Abandoned"
	KYCStatusExpired    = "Expired"
)

// CreateKYCSession creates a new KYC session record
func (db *DB) CreateKYCSession(ctx context.Context, session *KYCSession) error {
	query := `
		INSERT INTO kyc_sessions (session_id, user_address, workflow_id, status, session_url, vendor_data)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return db.conn.QueryRowContext(ctx, query,
		session.SessionID,
		session.UserAddress,
		session.WorkflowID,
		session.Status,
		session.SessionURL,
		session.VendorData,
	).Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)
}

// GetKYCSessionBySessionID retrieves a KYC session by Didit session_id
func (db *DB) GetKYCSessionBySessionID(ctx context.Context, sessionID string) (*KYCSession, error) {
	query := `
		SELECT id, session_id, user_address, workflow_id, status, session_url, vendor_data, 
		       decision_data, created_at, updated_at, completed_at
		FROM kyc_sessions
		WHERE session_id = $1
	`
	var s KYCSession
	err := db.conn.QueryRowContext(ctx, query, sessionID).Scan(
		&s.ID, &s.SessionID, &s.UserAddress, &s.WorkflowID, &s.Status, &s.SessionURL,
		&s.VendorData, &s.DecisionData, &s.CreatedAt, &s.UpdatedAt, &s.CompletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get KYC session: %w", err)
	}
	return &s, nil
}

// GetKYCSessionByUserAddress retrieves the latest KYC session for a user
func (db *DB) GetKYCSessionByUserAddress(ctx context.Context, userAddress string) (*KYCSession, error) {
	query := `
		SELECT id, session_id, user_address, workflow_id, status, session_url, vendor_data, 
		       decision_data, created_at, updated_at, completed_at
		FROM kyc_sessions
		WHERE user_address = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var s KYCSession
	err := db.conn.QueryRowContext(ctx, query, userAddress).Scan(
		&s.ID, &s.SessionID, &s.UserAddress, &s.WorkflowID, &s.Status, &s.SessionURL,
		&s.VendorData, &s.DecisionData, &s.CreatedAt, &s.UpdatedAt, &s.CompletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get KYC session: %w", err)
	}
	return &s, nil
}

// UpdateKYCSessionStatus updates the status and optionally decision data
func (db *DB) UpdateKYCSessionStatus(ctx context.Context, sessionID, status string, decisionData *string) error {
	var query string
	var args []interface{}

	if status == KYCStatusApproved || status == KYCStatusDeclined {
		query = `
			UPDATE kyc_sessions 
			SET status = $1, decision_data = $2, updated_at = NOW(), completed_at = NOW()
			WHERE session_id = $3
		`
		args = []interface{}{status, decisionData, sessionID}
	} else {
		query = `
			UPDATE kyc_sessions 
			SET status = $1, decision_data = $2, updated_at = NOW()
			WHERE session_id = $3
		`
		args = []interface{}{status, decisionData, sessionID}
	}

	result, err := db.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update KYC session: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// HasApprovedKYC checks if a user has an approved KYC session
func (db *DB) HasApprovedKYC(ctx context.Context, userAddress string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM kyc_sessions WHERE user_address = $1 AND status = $2)`
	var exists bool
	err := db.conn.QueryRowContext(ctx, query, userAddress, KYCStatusApproved).Scan(&exists)
	return exists, err
}

