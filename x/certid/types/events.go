package types

// Event types for the CertID module - Oracle Pattern
// These structured events enable WebSocket subscriptions from external relayers
// (e.g., certid-optimism Node.js adapter) instead of requiring polling.
const (
	// EventTypeTrustScoreUpdated is emitted when a trust score changes.
	// Primary event for the EVM Oracle/Relayer bridge pattern.
	EventTypeTrustScoreUpdated = "certid.v1.TrustScoreUpdated"

	// EventTypeProfileCreated is emitted when a new CertID profile is created.
	EventTypeProfileCreated = "certid.v1.ProfileCreated"

	// EventTypeProfileUpdated is emitted when a profile is updated.
	EventTypeProfileUpdated = "certid.v1.ProfileUpdated"

	// EventTypeVerificationStatusChanged is emitted when verification status changes.
	EventTypeVerificationStatusChanged = "certid.v1.VerificationStatusChanged"

	// EventTypeBadgeAwarded is emitted when a soulbound badge is awarded.
	EventTypeBadgeAwarded = "certid.v1.BadgeAwarded"

	// EventTypeBadgeRevoked is emitted when a badge is revoked.
	EventTypeBadgeRevoked = "certid.v1.BadgeRevoked"

	// EventTypeOracleAuthorized is emitted when an oracle is authorized.
	EventTypeOracleAuthorized = "certid.v1.OracleAuthorized"

	// EventTypeOracleRevoked is emitted when an oracle is revoked.
	EventTypeOracleRevoked = "certid.v1.OracleRevoked"
)

// Event attribute keys
const (
	AttributeKeyDID        = "did"
	AttributeKeyAddress    = "address"
	AttributeKeyScore      = "score"
	AttributeKeyOldScore   = "old_score"
	AttributeKeyTimestamp  = "timestamp"
	AttributeKeyProofHash  = "proof_hash"
	AttributeKeyAuthority  = "authority"
	AttributeKeyHandle     = "handle"
	AttributeKeyName       = "name"
	AttributeKeyIsVerified = "is_verified"
	AttributeKeyBadgeName  = "badge_name"
	AttributeKeyBadgeID    = "badge_id"
	AttributeKeyOracle     = "oracle"
	AttributeKeyUser       = "user"
)
