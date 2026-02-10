package types

// Store key prefixes for the hardware module
var (
	// DeviceKeyPrefix is the prefix for device storage
	// Format: DeviceKeyPrefix | DeviceID -> Device
	DeviceKeyPrefix = []byte{0x01}

	// AttestationKeyPrefix is the prefix for attestation storage
	// Format: AttestationKeyPrefix | DeviceID | Timestamp -> TEEAttestation
	AttestationKeyPrefix = []byte{0x02}

	// OwnerDeviceIndexPrefix indexes devices by owner address
	// Format: OwnerDeviceIndexPrefix | OwnerAddress | DeviceID -> nil
	OwnerDeviceIndexPrefix = []byte{0x03}

	// HumanityScoreKeyPrefix is the prefix for humanity score storage
	// Format: HumanityScoreKeyPrefix | Address -> HumanityScore
	HumanityScoreKeyPrefix = []byte{0x04}

	// PendingChallengePrefix stores pending attestation challenges
	// Format: PendingChallengePrefix | DeviceID -> Challenge
	PendingChallengePrefix = []byte{0x05}
)

// GetDeviceKey returns the store key for a device
func GetDeviceKey(deviceID string) []byte {
	return append(DeviceKeyPrefix, []byte(deviceID)...)
}

// GetAttestationKey returns the store key for an attestation
func GetAttestationKey(deviceID string, timestamp int64) []byte {
	key := append(AttestationKeyPrefix, []byte(deviceID)...)
	// Append timestamp as big-endian bytes for proper ordering
	tsBytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		tsBytes[i] = byte(timestamp & 0xff)
		timestamp >>= 8
	}
	return append(key, tsBytes...)
}

// GetOwnerDeviceIndexKey returns the index key for owner->device mapping
func GetOwnerDeviceIndexKey(ownerAddress, deviceID string) []byte {
	key := append(OwnerDeviceIndexPrefix, []byte(ownerAddress)...)
	key = append(key, []byte("/")...)
	return append(key, []byte(deviceID)...)
}

// GetHumanityScoreKey returns the store key for a humanity score
func GetHumanityScoreKey(address string) []byte {
	return append(HumanityScoreKeyPrefix, []byte(address)...)
}

// GetPendingChallengeKey returns the store key for a pending challenge
func GetPendingChallengeKey(deviceID string) []byte {
	return append(PendingChallengePrefix, []byte(deviceID)...)
}
