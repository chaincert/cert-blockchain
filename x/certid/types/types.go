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

	// CertIDSchemaDefinition is the EAS schema definition for CertID profiles
	// Format: "address address,string handle,string name,uint8 entityType,uint64 trustScore,bool isVerified,uint256 createdAt"
	CertIDSchemaDefinition = "address address,string handle,string name,uint8 entityType,uint64 trustScore,bool isVerified,uint256 createdAt"

	// CertIDSchemaUID is a pre-computed schema UID for CertID attestations
	// This should be registered at genesis
	CertIDSchemaUID = "0xcertid_profile_schema_v1"
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

	// Handle is the unique handle (e.g., "alice.cert")
	Handle string `json:"handle,omitempty"`

	// Name is the display name
	Name string `json:"name,omitempty"`

	// Bio is a short biography
	Bio string `json:"bio,omitempty"`

	// AvatarCID is the IPFS CID of the avatar image
	AvatarCID string `json:"avatarCid,omitempty"`

	// MetadataURI is IPFS link to extended metadata
	MetadataURI string `json:"metadataUri,omitempty"`

	// AttestationUID is the UID of the on-chain attestation for this CertID
	AttestationUID string `json:"attestationUid,omitempty"`

	// PublicKey is the user's public key for encryption
	PublicKey string `json:"publicKey,omitempty"`

	// EntityType categorizes the profile (Individual, Institution, etc.)
	EntityType EntityType `json:"entityType"`

	// TrustScore is the dynamic reputation score (0-100)
	TrustScore uint64 `json:"trustScore"`

	// SocialLinks contains verified social media links
	SocialLinks map[string]string `json:"socialLinks,omitempty"`

	// Credentials contains verified credential attestation UIDs
	Credentials []string `json:"credentials,omitempty"`

	// Badges contains soulbound token badges (non-transferable)
	Badges map[string]*Badge `json:"badges,omitempty"`

	// CreatedAt is the creation timestamp
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is the last update timestamp
	UpdatedAt time.Time `json:"updatedAt"`

	// Verified indicates if the identity has been verified
	Verified bool `json:"verified"`

	// IsActive indicates if the profile is active
	IsActive bool `json:"isActive"`

	// VerificationLevel indicates the level of verification (0-3)
	VerificationLevel uint8 `json:"verificationLevel"`
}

// NewCertID creates a new CertID profile
func NewCertID(address string) *CertID {
	now := time.Now()
	return &CertID{
		Address:           address,
		EntityType:        EntityTypeIndividual,
		TrustScore:        0,
		SocialLinks:       make(map[string]string),
		Credentials:       []string{},
		Badges:            make(map[string]*Badge),
		CreatedAt:         now,
		UpdatedAt:         now,
		Verified:          false,
		IsActive:          true,
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

// EntityType represents the type of entity for a CertID profile
type EntityType uint8

const (
	EntityTypeIndividual   EntityType = 0
	EntityTypeInstitution  EntityType = 1
	EntityTypeSystemAdmin  EntityType = 2
	EntityTypeBot          EntityType = 3
)

// Standard badge identifiers (matching CertID.sol)
const (
	BadgeKYCL1    = "KYC_L1"
	BadgeKYCL2    = "KYC_L2"
	BadgeAcademic = "ACADEMIC_ISSUER"
	BadgeCreator  = "VERIFIED_CREATOR"
	BadgeGov      = "GOV_AGENCY"
	BadgeLegal    = "LEGAL_ENTITY"
	BadgeISO9001  = "ISO_9001_CERTIFIED"
)

// Badge represents a soulbound token (non-transferable badge)
type Badge struct {
	// ID is the unique identifier (hash of badge name)
	ID string `json:"id"`

	// Name is the human-readable badge name
	Name string `json:"name"`

	// Description is optional badge description
	Description string `json:"description,omitempty"`

	// AwardedAt is when the badge was awarded
	AwardedAt time.Time `json:"awardedAt"`

	// AwardedBy is the address that awarded the badge
	AwardedBy string `json:"awardedBy"`

	// IsRevoked indicates if the badge has been revoked
	IsRevoked bool `json:"isRevoked"`

	// RevokedAt is when the badge was revoked (if applicable)
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
}

// NewBadge creates a new Badge
func NewBadge(name, description, awardedBy string) *Badge {
	hash := sha256.Sum256([]byte(name))
	return &Badge{
		ID:          hex.EncodeToString(hash[:]),
		Name:        name,
		Description: description,
		AwardedAt:   time.Now(),
		AwardedBy:   awardedBy,
		IsRevoked:   false,
	}
}

// TrustScoreUpdate represents a trust score change event
type TrustScoreUpdate struct {
	Address   string    `json:"address"`
	OldScore  uint64    `json:"oldScore"`
	NewScore  uint64    `json:"newScore"`
	Reason    string    `json:"reason,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
	UpdatedBy string    `json:"updatedBy"`
}

// OracleAuthorization represents an authorized oracle
type OracleAuthorization struct {
	Address      string    `json:"address"`
	IsAuthorized bool      `json:"isAuthorized"`
	AuthorizedAt time.Time `json:"authorizedAt"`
	AuthorizedBy string    `json:"authorizedBy"`
}
