package types

import (
	"cosmossdk.io/errors"
)

// Module errors
var (
	// ErrInvalidParams is returned when module parameters are invalid
	ErrInvalidParams = errors.Register(ModuleName, 1, "invalid params")

	// ErrSchemaNotFound is returned when a schema is not found
	ErrSchemaNotFound = errors.Register(ModuleName, 2, "schema not found")

	// ErrAttestationNotFound is returned when an attestation is not found
	ErrAttestationNotFound = errors.Register(ModuleName, 3, "attestation not found")

	// ErrUnauthorized is returned when an action is not authorized
	ErrUnauthorized = errors.Register(ModuleName, 4, "unauthorized")

	// ErrAttestationNotRevocable is returned when trying to revoke a non-revocable attestation
	ErrAttestationNotRevocable = errors.Register(ModuleName, 5, "attestation is not revocable")

	// ErrAttestationAlreadyRevoked is returned when attestation is already revoked
	ErrAttestationAlreadyRevoked = errors.Register(ModuleName, 6, "attestation already revoked")

	// ErrAttestationExpired is returned when attestation has expired
	ErrAttestationExpired = errors.Register(ModuleName, 7, "attestation has expired")

	// ErrDuplicateSchema is returned when a schema already exists
	ErrDuplicateSchema = errors.Register(ModuleName, 8, "schema already exists")

	// ErrInvalidIPFSCID is returned when IPFS CID is invalid
	ErrInvalidIPFSCID = errors.Register(ModuleName, 9, "invalid IPFS CID")

	// ErrInvalidEncryptedHash is returned when encrypted data hash is invalid
	ErrInvalidEncryptedHash = errors.Register(ModuleName, 10, "invalid encrypted data hash")

	// ErrTooManyRecipients is returned when too many recipients are specified
	ErrTooManyRecipients = errors.Register(ModuleName, 11, "too many recipients")

	// ErrMissingEncryptedKey is returned when an encrypted key is missing for a recipient
	ErrMissingEncryptedKey = errors.Register(ModuleName, 12, "missing encrypted key for recipient")

	// ErrRecipientNotAuthorized is returned when recipient is not authorized to access attestation
	ErrRecipientNotAuthorized = errors.Register(ModuleName, 13, "recipient not authorized to access attestation")

	// ErrInvalidSchemaFormat is returned when schema format is invalid
	ErrInvalidSchemaFormat = errors.Register(ModuleName, 14, "invalid schema format")
)

