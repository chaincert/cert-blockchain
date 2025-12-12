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

