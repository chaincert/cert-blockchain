package types

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "certid"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MinUsernameLength is the minimum length for a username
	MinUsernameLength = 3

	// MaxUsernameLength is the maximum length for a username
	MaxUsernameLength = 32
)

// Params defines the parameters for the certid module
type Params struct {
	// MaxUsernameLength is the maximum length for a username
	MaxUsernameLength uint32 `json:"max_username_length"`

	// MaxDisplayNameLength is the maximum length for a display name
	MaxDisplayNameLength uint32 `json:"max_display_name_length"`

	// MaxBioLength is the maximum length for a bio
	MaxBioLength uint32 `json:"max_bio_length"`

	// MaxCredentials is the maximum number of credentials per profile
	MaxCredentials uint32 `json:"max_credentials"`

	// RegistrationFee is the fee required to register a CertID
	RegistrationFee sdk.Coin `json:"registration_fee"`
}

// DefaultParams returns the default parameters for the certid module
func DefaultParams() Params {
	return Params{
		MaxUsernameLength:    32,
		MaxDisplayNameLength: 100,
		MaxBioLength:         500,
		MaxCredentials:       50,
		RegistrationFee:      sdk.NewCoin("ucert", math.ZeroInt()),
	}
}

// ValidateUsername validates a username string
func ValidateUsername(username string) error {
	if len(username) == 0 {
		return ErrInvalidUsername.Wrap("username cannot be empty")
	}

	if len(username) < MinUsernameLength {
		return ErrInvalidUsername.Wrapf("username must be at least %d characters", MinUsernameLength)
	}

	if len(username) > MaxUsernameLength {
		return ErrInvalidUsername.Wrapf("username must be at most %d characters", MaxUsernameLength)
	}

	// Username must start with a letter and contain only lowercase letters, numbers, and underscores
	validUsername := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	if !validUsername.MatchString(username) {
		return ErrInvalidUsername.Wrap("username must start with a letter and contain only lowercase letters, numbers, and underscores")
	}

	return nil
}

// SupportedSocialPlatforms returns the list of supported social platforms
func SupportedSocialPlatforms() []string {
	return []string{
		"twitter",
		"github",
		"linkedin",
		"discord",
		"telegram",
		"facebook",
		"instagram",
		"youtube",
	}
}

// ValidCredentialTypes returns the list of valid credential types
func ValidCredentialTypes() []string {
	return []string{
		"education",
		"employment",
		"certification",
		"identity",
		"membership",
		"achievement",
		"license",
	}
}

// CertID represents a decentralized identity profile
// Per Whitepaper CertID Section
type CertID struct {
	// Address is the blockchain address (primary key)
	Address string `json:"address"`

	// Name is the display name
	Name string `json:"name,omitempty"`

	// Bio is a short biography
	Bio string `json:"bio,omitempty"`

	// AvatarCID is the IPFS CID of the avatar image
	AvatarCID string `json:"avatarCid,omitempty"`

	// PublicKey is the user's public key for encryption
	PublicKey string `json:"publicKey,omitempty"`

	// SocialLinks contains verified social media links
	SocialLinks map[string]string `json:"socialLinks,omitempty"`

	// Credentials contains verified credential attestation UIDs
	Credentials []string `json:"credentials,omitempty"`

	// CreatedAt is the creation timestamp
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is the last update timestamp
	UpdatedAt time.Time `json:"updatedAt"`

	// Verified indicates if the identity has been verified
	Verified bool `json:"verified"`

	// VerificationLevel indicates the level of verification (0-3)
	VerificationLevel uint8 `json:"verificationLevel"`
}

// NewCertID creates a new CertID profile
func NewCertID(address string) *CertID {
	now := time.Now()
	return &CertID{
		Address:           address,
		SocialLinks:       make(map[string]string),
		Credentials:       []string{},
		CreatedAt:         now,
		UpdatedAt:         now,
		Verified:          false,
		VerificationLevel: 0,
	}
}

// GenerateCertIDHash generates a unique hash for the CertID
func GenerateCertIDHash(address string, timestamp time.Time) string {
	data := address + timestamp.String()
	hash := sha256.Sum256([]byte(data))
	return "0x" + hex.EncodeToString(hash[:])
}

// VerificationRequest represents a request to verify a CertID
type VerificationRequest struct {
	Address          string     `json:"address"`
	RequestType      string     `json:"requestType"` // "email", "social", "document"
	VerificationData string     `json:"verificationData"`
	Status           string     `json:"status"` // "pending", "approved", "rejected"
	CreatedAt        time.Time  `json:"createdAt"`
	ProcessedAt      *time.Time `json:"processedAt,omitempty"`
}

// SocialVerification represents a verified social media account
type SocialVerification struct {
	Platform    string    `json:"platform"` // "twitter", "github", "linkedin"
	Handle      string    `json:"handle"`
	Verified    bool      `json:"verified"`
	VerifiedAt  time.Time `json:"verifiedAt"`
	ProofTxHash string    `json:"proofTxHash,omitempty"`
}

// Credential represents a verified credential attached to a CertID
type Credential struct {
	UID            string     `json:"uid"`
	Type           string     `json:"type"` // "education", "employment", "certification"
	Issuer         string     `json:"issuer"`
	Title          string     `json:"title"`
	Description    string     `json:"description,omitempty"`
	IssuedAt       time.Time  `json:"issuedAt"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	AttestationUID string     `json:"attestationUid"`
}

// VerificationLevels defines the verification level requirements
var VerificationLevels = map[uint8]string{
	0: "Unverified",
	1: "Basic (Email verified)",
	2: "Standard (Email + Social verified)",
	3: "Premium (Email + Social + Document verified)",
}

// GetVerificationLevelName returns the name for a verification level
func GetVerificationLevelName(level uint8) string {
	if name, ok := VerificationLevels[level]; ok {
		return name
	}
	return "Unknown"
}
