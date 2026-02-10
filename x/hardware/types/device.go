package types

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "hardware"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// TEEType represents supported Trusted Execution Environment types
// Per Whitepaper v3.0 - Mobile-First, Edge-Native Identity
type TEEType string

const (
	// TEETypeTrustZone - ARM TrustZone for IoT devices and Android
	TEETypeTrustZone TEEType = "ARM_TRUSTZONE"

	// TEETypeSecureEnclave - Apple Secure Enclave for iOS devices
	TEETypeSecureEnclave TEEType = "APPLE_SECURE_ENCLAVE"

	// TEETypeSGX - Intel SGX for server environments (future)
	TEETypeSGX TEEType = "INTEL_SGX"

	// TEETypeSEV - AMD SEV for cloud environments (future)
	TEETypeSEV TEEType = "AMD_SEV"
)

// SupportedTEETypes returns the list of currently supported TEE types
// Prioritizing ARM TrustZone and Apple Secure Enclave per grant requirements
func SupportedTEETypes() []TEEType {
	return []TEEType{
		TEETypeTrustZone,
		TEETypeSecureEnclave,
	}
}

// Device represents a hardware device with TEE capabilities
// Per Whitepaper v3.0 Section 3.1: Hardware-Anchored Identity
type Device struct {
	// DeviceID is a unique identifier derived from TEE attestation
	DeviceID string `json:"device_id"`

	// SerialNumber is the manufacturer's serial number
	SerialNumber string `json:"serial_number,omitempty"`

	// Manufacturer identifies the device manufacturer
	Manufacturer string `json:"manufacturer"`

	// Model identifies the device model
	Model string `json:"model,omitempty"`

	// TEEType specifies the type of TEE (TrustZone, SecureEnclave, etc.)
	TEEType TEEType `json:"tee_type"`

	// PublicKey is the hardware-bound public key from the TEE
	PublicKey []byte `json:"public_key"`

	// OwnerAddress is the CertID address that owns this device
	OwnerAddress string `json:"owner_address"`

	// TrustScore is the device-specific trust score (0-100)
	// Calculated from: Uptime, DataQuality, AttestationHistory
	TrustScore uint64 `json:"trust_score"`

	// Uptime is the percentage of time device has been online (0-100)
	Uptime float64 `json:"uptime"`

	// DataQuality is the congruence score from cross-device validation (0-100)
	DataQuality float64 `json:"data_quality"`

	// AttestationCount tracks successful attestations
	AttestationCount uint64 `json:"attestation_count"`

	// LastAttestAt is the timestamp of last successful attestation
	LastAttestAt time.Time `json:"last_attest_at"`

	// RegisteredAt is the timestamp when device was registered
	RegisteredAt time.Time `json:"registered_at"`

	// IsActive indicates if device is currently active
	IsActive bool `json:"is_active"`

	// IsSuspended indicates if device has been flagged for suspicious activity
	IsSuspended bool `json:"is_suspended"`

	// SuspensionReason provides context for suspension
	SuspensionReason string `json:"suspension_reason,omitempty"`
}

// TEEAttestation represents a cryptographic proof from a TEE
// This proves the device is genuine hardware, not emulated
type TEEAttestation struct {
	// DeviceID identifies the attesting device
	DeviceID string `json:"device_id"`

	// AttestationData is the raw TEE attestation blob
	// For TrustZone: ARM attestation token
	// For SecureEnclave: Apple DeviceCheck/App Attest token
	AttestationData []byte `json:"attestation_data"`

	// ManufacturerSig is the manufacturer's cryptographic signature
	ManufacturerSig []byte `json:"manufacturer_sig,omitempty"`

	// Nonce is the challenge nonce to prevent replay attacks
	Nonce []byte `json:"nonce"`

	// Timestamp when attestation was generated
	Timestamp time.Time `json:"timestamp"`

	// AttestationType indicates the attestation context
	AttestationType AttestationType `json:"attestation_type"`

	// Verified indicates if this attestation has been verified
	Verified bool `json:"verified"`

	// VerifiedAt is when verification completed
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}

// AttestationType indicates the context of an attestation
type AttestationType string

const (
	// AttestationTypeInitial is for device registration
	AttestationTypeInitial AttestationType = "initial"

	// AttestationTypePeriodic is for regular heartbeat attestations
	AttestationTypePeriodic AttestationType = "periodic"

	// AttestationTypeChallenge is for on-demand challenge-response
	AttestationTypeChallenge AttestationType = "challenge"

	// AttestationTypeBoot is for secure boot verification
	AttestationTypeBoot AttestationType = "boot"
)

// HumanityScore represents aggregated anti-Sybil metrics
// Per Whitepaper v3.0 Section 2: "The Sybil Problem"
type HumanityScore struct {
	// Address is the CertID address
	Address string `json:"address"`

	// Score is the overall humanity score (0-100)
	Score uint64 `json:"score"`

	// DeviceCount is number of verified devices linked
	DeviceCount uint64 `json:"device_count"`

	// AverageDeviceTrust is the average trust score of linked devices
	AverageDeviceTrust float64 `json:"average_device_trust"`

	// GeoDispersion measures geographic diversity of devices
	GeoDispersion float64 `json:"geo_dispersion"`

	// UsagePatternScore measures human-like usage patterns
	UsagePatternScore float64 `json:"usage_pattern_score"`

	// LastUpdated is when score was last calculated
	LastUpdated time.Time `json:"last_updated"`
}

// NewDevice creates a new Device instance
func NewDevice(deviceID, manufacturer string, teeType TEEType, publicKey []byte, owner sdk.AccAddress) *Device {
	now := time.Now()
	return &Device{
		DeviceID:         deviceID,
		Manufacturer:     manufacturer,
		TEEType:          teeType,
		PublicKey:        publicKey,
		OwnerAddress:     owner.String(),
		TrustScore:       0, // Starts at 0, increases with successful attestations
		Uptime:           0,
		DataQuality:      0,
		AttestationCount: 0,
		RegisteredAt:     now,
		LastAttestAt:     now,
		IsActive:         true,
		IsSuspended:      false,
	}
}

// GenerateDeviceID generates a unique device ID from TEE public key
func GenerateDeviceID(publicKey []byte, teeType TEEType) string {
	data := append(publicKey, []byte(teeType)...)
	hash := sha256.Sum256(data)
	return "dev_" + hex.EncodeToString(hash[:16])
}

// CalculateTrustScore computes device trust score from metrics
// Formula per Whitepaper: Trust = (Uptime × 0.3) + (DataQuality × 0.5) + (AttestationBonus × 0.2)
func (d *Device) CalculateTrustScore() uint64 {
	// Uptime component (30%)
	uptimeComponent := d.Uptime * 0.3

	// Data quality component (50%)
	qualityComponent := d.DataQuality * 0.5

	// Attestation bonus based on count (20%, max 20 points)
	attestBonus := float64(d.AttestationCount)
	if attestBonus > 100 {
		attestBonus = 100
	}
	attestComponent := attestBonus * 0.2

	score := uptimeComponent + qualityComponent + attestComponent
	if score > 100 {
		score = 100
	}

	return uint64(score)
}

// ValidateTEEType checks if the TEE type is supported
func ValidateTEEType(teeType TEEType) bool {
	for _, supported := range SupportedTEETypes() {
		if teeType == supported {
			return true
		}
	}
	return false
}
