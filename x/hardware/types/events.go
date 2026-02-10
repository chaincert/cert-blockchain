package types

// Event types for the hardware module
const (
	EventTypeDeviceRegistered    = "device_registered"
	EventTypeAttestationVerified = "attestation_verified"
	EventTypeDeviceLinked        = "device_linked"
	EventTypeDeviceSuspended     = "device_suspended"
	EventTypeDeviceReactivated   = "device_reactivated"
	EventTypeTrustScoreUpdated   = "device_trust_updated"

	AttributeKeyDeviceID       = "device_id"
	AttributeKeyManufacturer   = "manufacturer"
	AttributeKeyTEEType        = "tee_type"
	AttributeKeyOwner          = "owner"
	AttributeKeyTrustScore     = "trust_score"
	AttributeKeyCertIDAddress  = "certid_address"
	AttributeKeyReason         = "reason"
	AttributeKeyAttestationType = "attestation_type"
)
