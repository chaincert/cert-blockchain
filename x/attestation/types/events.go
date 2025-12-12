package types

// Event types for the attestation module
const (
	EventTypeSchemaRegistered           = "schema_registered"
	EventTypeAttestationCreated         = "attestation_created"
	EventTypeAttestationRevoked         = "attestation_revoked"
	EventTypeEncryptedAttestationCreated = "encrypted_attestation_created"
)

// Attribute keys for attestation events
const (
	AttributeKeySchemaUID       = "schema_uid"
	AttributeKeyAttestationUID  = "attestation_uid"
	AttributeKeyAttester        = "attester"
	AttributeKeyRecipient       = "recipient"
	AttributeKeyCreator         = "creator"
	AttributeKeyRevoker         = "revoker"
	AttributeKeyRevocable       = "revocable"
	AttributeKeyAttestationType = "attestation_type"
	AttributeKeyIPFSCID         = "ipfs_cid"
	AttributeKeyRecipientsCount = "recipients_count"
)

