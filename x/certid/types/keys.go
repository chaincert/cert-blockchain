package types

// Key prefixes for the CertID module store
var (
	// ProfileKeyPrefix is the prefix for profile storage
	ProfileKeyPrefix = []byte{0x01}

	// VerificationKeyPrefix is the prefix for verification requests
	VerificationKeyPrefix = []byte{0x02}

	// CredentialKeyPrefix is the prefix for credentials
	CredentialKeyPrefix = []byte{0x03}

	// SocialVerificationKeyPrefix is the prefix for social verifications
	SocialVerificationKeyPrefix = []byte{0x04}

	// AddressToProfileKeyPrefix maps addresses to profile hashes
	AddressToProfileKeyPrefix = []byte{0x05}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x06}

	// HandleToAddressKeyPrefix maps handles to addresses
	HandleToAddressKeyPrefix = []byte{0x07}

	// OracleKeyPrefix is the prefix for authorized oracles
	OracleKeyPrefix = []byte{0x08}
)

// GetProfileKey returns the store key for a profile
func GetProfileKey(address string) []byte {
	return append(ProfileKeyPrefix, []byte(address)...)
}

// GetVerificationKey returns the store key for a verification request
func GetVerificationKey(address string, requestType string) []byte {
	key := append(VerificationKeyPrefix, []byte(address)...)
	return append(key, []byte(requestType)...)
}

// GetCredentialKey returns the store key for a credential
func GetCredentialKey(address string, credentialUID string) []byte {
	key := append(CredentialKeyPrefix, []byte(address)...)
	return append(key, []byte(credentialUID)...)
}

// GetSocialVerificationKey returns the store key for a social verification
func GetSocialVerificationKey(address string, platform string) []byte {
	key := append(SocialVerificationKeyPrefix, []byte(address)...)
	return append(key, []byte(platform)...)
}

// GetHandleToAddressKey returns the store key for handle-to-address mapping
func GetHandleToAddressKey(handle string) []byte {
	return append(HandleToAddressKeyPrefix, []byte(handle)...)
}

// GetOracleKey returns the store key for an oracle authorization
func GetOracleKey(oracle string) []byte {
	return append(OracleKeyPrefix, []byte(oracle)...)
}

