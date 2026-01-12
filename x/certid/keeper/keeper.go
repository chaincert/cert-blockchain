package keeper

import (
	"encoding/json"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/chaincertify/certd/x/certid/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	// Expected keepers
	accountKeeper     types.AccountKeeper
	bankKeeper        types.BankKeeper
	attestationKeeper types.AttestationKeeper
}

// NewKeeper creates a new CertID Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

// SetAttestationKeeper sets the attestation keeper (called after app initialization to avoid circular deps)
func (k *Keeper) SetAttestationKeeper(ak types.AttestationKeeper) {
	k.attestationKeeper = ak
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// CreateProfile creates a new CertID profile and attestation
func (k Keeper) CreateProfile(ctx sdk.Context, msg *types.MsgCreateProfile) (*types.CertID, error) {
	store := ctx.KVStore(k.storeKey)

	// Check if profile already exists
	key := types.GetProfileKey(msg.Creator)
	if store.Has(key) {
		return nil, types.ErrProfileAlreadyExists
	}

	// Create new profile
	profile := types.NewCertID(msg.Creator)
	profile.Name = msg.Name
	profile.Bio = msg.Bio
	profile.AvatarCID = msg.AvatarCID
	profile.PublicKey = msg.PublicKey
	profile.SocialLinks = msg.SocialLinks

	// Create attestation for the CertID profile if attestation keeper is available
	if k.attestationKeeper != nil {
		attester, err := sdk.AccAddressFromBech32(msg.Creator)
		if err != nil {
			return nil, err
		}

		// Encode profile data for attestation
		attestationData, err := k.encodeCertIDAttestationData(profile)
		if err != nil {
			k.Logger(ctx).Warn("Failed to encode attestation data", "error", err)
			// Continue without attestation - profile creation shouldn't fail
		} else {
			// Create self-attestation (attester == recipient for identity claim)
			attestationUID, err := k.attestationKeeper.CreateAttestation(
				ctx,
				attester,
				types.CertIDSchemaUID,
				attester, // Self-attestation: recipient is the same as attester
				time.Time{}, // No expiration for identity attestations
				false, // CertID attestations are not revocable
				"", // No reference UID
				attestationData,
			)
			if err != nil {
				k.Logger(ctx).Warn("Failed to create CertID attestation", "error", err)
				// Continue without attestation - profile creation shouldn't fail
			} else {
				// Store the attestation UID in the profile
				profile.AttestationUID = attestationUID
				k.Logger(ctx).Info("CertID attestation created", "address", msg.Creator, "attestationUID", attestationUID)
			}
		}
	}

	// Store profile
	bz, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}
	store.Set(key, bz)

	k.Logger(ctx).Info("Profile created", "address", msg.Creator)

	return profile, nil
}

// encodeCertIDAttestationData encodes CertID profile data for attestation
func (k Keeper) encodeCertIDAttestationData(profile *types.CertID) ([]byte, error) {
	// Create a simplified attestation payload
	attestationPayload := struct {
		Address     string `json:"address"`
		Handle      string `json:"handle"`
		Name        string `json:"name"`
		EntityType  uint8  `json:"entityType"`
		TrustScore  uint64 `json:"trustScore"`
		IsVerified  bool   `json:"isVerified"`
		CreatedAt   int64  `json:"createdAt"`
	}{
		Address:    profile.Address,
		Handle:     profile.Handle,
		Name:       profile.Name,
		EntityType: uint8(profile.EntityType),
		TrustScore: profile.TrustScore,
		IsVerified: profile.Verified,
		CreatedAt:  profile.CreatedAt.Unix(),
	}
	return json.Marshal(attestationPayload)
}

// UpdateProfile updates an existing CertID profile
func (k Keeper) UpdateProfile(ctx sdk.Context, msg *types.MsgUpdateProfile) (*types.CertID, error) {
	store := ctx.KVStore(k.storeKey)

	// Get existing profile
	key := types.GetProfileKey(msg.Creator)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrProfileNotFound
	}

	var profile types.CertID
	if err := json.Unmarshal(bz, &profile); err != nil {
		return nil, err
	}

	// Update fields if provided
	if msg.Name != "" {
		profile.Name = msg.Name
	}
	if msg.Bio != "" {
		profile.Bio = msg.Bio
	}
	if msg.AvatarCID != "" {
		profile.AvatarCID = msg.AvatarCID
	}
	if msg.PublicKey != "" {
		profile.PublicKey = msg.PublicKey
	}
	if msg.SocialLinks != nil {
		for k, v := range msg.SocialLinks {
			profile.SocialLinks[k] = v
		}
	}
	profile.UpdatedAt = time.Now()

	// Store updated profile
	bz, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}
	store.Set(key, bz)

	k.Logger(ctx).Info("Profile updated", "address", msg.Creator)

	return &profile, nil
}

// GetProfile retrieves a CertID profile by address
func (k Keeper) GetProfile(ctx sdk.Context, address string) (*types.CertID, error) {
	store := ctx.KVStore(k.storeKey)

	key := types.GetProfileKey(address)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrProfileNotFound
	}

	var profile types.CertID
	if err := json.Unmarshal(bz, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// AddCredential adds a credential to a profile
func (k Keeper) AddCredential(ctx sdk.Context, msg *types.MsgAddCredential) error {
	store := ctx.KVStore(k.storeKey)

	// Get profile
	profile, err := k.GetProfile(ctx, msg.Creator)
	if err != nil {
		return err
	}

	// Check if credential already exists
	for _, cred := range profile.Credentials {
		if cred == msg.AttestationUID {
			return types.ErrCredentialAlreadyExists
		}
	}

	// Add credential
	profile.Credentials = append(profile.Credentials, msg.AttestationUID)
	profile.UpdatedAt = time.Now()

	// Store updated profile
	bz, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	store.Set(types.GetProfileKey(msg.Creator), bz)

	k.Logger(ctx).Info("Credential added", "address", msg.Creator, "attestationUID", msg.AttestationUID)

	return nil
}

// VerifySocial verifies a social media account
func (k Keeper) VerifySocial(ctx sdk.Context, msg *types.MsgVerifySocial) error {
	store := ctx.KVStore(k.storeKey)

	// Get profile
	profile, err := k.GetProfile(ctx, msg.Creator)
	if err != nil {
		return err
	}

	// TODO: Implement actual social verification logic
	// This would involve verifying the proof (e.g., a signed message posted on the platform)

	// Update social links
	profile.SocialLinks[msg.Platform] = msg.Handle
	profile.UpdatedAt = time.Now()

	// Update verification level if applicable
	if profile.VerificationLevel < 2 {
		profile.VerificationLevel = 2
	}

	// Store updated profile
	bz, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	store.Set(types.GetProfileKey(msg.Creator), bz)

	// Store social verification record
	verification := types.SocialVerification{
		Platform:   msg.Platform,
		Handle:     msg.Handle,
		Verified:   true,
		VerifiedAt: time.Now(),
	}
	verificationBz, _ := json.Marshal(verification)
	store.Set(types.GetSocialVerificationKey(msg.Creator, msg.Platform), verificationBz)

	k.Logger(ctx).Info("Social verified", "address", msg.Creator, "platform", msg.Platform)

	return nil
}

// AwardBadge awards a soulbound badge to a user (SBT - non-transferable)
func (k Keeper) AwardBadge(ctx sdk.Context, msg *types.MsgAwardBadge) error {
	store := ctx.KVStore(k.storeKey)

	// Get profile
	profile, err := k.GetProfile(ctx, msg.User)
	if err != nil {
		return err
	}

	if !profile.IsActive {
		return types.ErrProfileNotActive
	}

	// Create badge
	badge := types.NewBadge(msg.BadgeName, msg.Description, msg.Authority)

	// Check if badge already awarded
	if profile.Badges == nil {
		profile.Badges = make(map[string]*types.Badge)
	}
	if _, exists := profile.Badges[badge.ID]; exists {
		return types.ErrBadgeAlreadyAwarded
	}

	// Award badge
	profile.Badges[badge.ID] = badge
	profile.UpdatedAt = time.Now()

	// Store updated profile
	bz, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	store.Set(types.GetProfileKey(msg.User), bz)

	k.Logger(ctx).Info("Badge awarded", "user", msg.User, "badge", msg.BadgeName)

	return nil
}

// RevokeBadge revokes a badge from a user
func (k Keeper) RevokeBadge(ctx sdk.Context, msg *types.MsgRevokeBadge) error {
	store := ctx.KVStore(k.storeKey)

	// Get profile
	profile, err := k.GetProfile(ctx, msg.User)
	if err != nil {
		return err
	}

	// Find badge by name
	badge := types.NewBadge(msg.BadgeName, "", "")
	if profile.Badges == nil || profile.Badges[badge.ID] == nil {
		return types.ErrBadgeNotFound
	}

	// Mark badge as revoked
	now := time.Now()
	profile.Badges[badge.ID].IsRevoked = true
	profile.Badges[badge.ID].RevokedAt = &now
	profile.UpdatedAt = now

	// Store updated profile
	bz, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	store.Set(types.GetProfileKey(msg.User), bz)

	k.Logger(ctx).Info("Badge revoked", "user", msg.User, "badge", msg.BadgeName)

	return nil
}

// UpdateTrustScore updates a user's trust score
func (k Keeper) UpdateTrustScore(ctx sdk.Context, msg *types.MsgUpdateTrustScore) error {
	store := ctx.KVStore(k.storeKey)

	// Get profile
	profile, err := k.GetProfile(ctx, msg.User)
	if err != nil {
		return err
	}

	if !profile.IsActive {
		return types.ErrProfileNotActive
	}

	oldScore := profile.TrustScore
	profile.TrustScore = msg.Score
	profile.UpdatedAt = time.Now()

	// Store updated profile
	bz, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	store.Set(types.GetProfileKey(msg.User), bz)

	k.Logger(ctx).Info("Trust score updated", "user", msg.User, "oldScore", oldScore, "newScore", msg.Score)

	return nil
}

// SetVerificationStatus sets verification status for a profile
func (k Keeper) SetVerificationStatus(ctx sdk.Context, msg *types.MsgSetVerificationStatus) error {
	store := ctx.KVStore(k.storeKey)

	// Get profile
	profile, err := k.GetProfile(ctx, msg.User)
	if err != nil {
		return err
	}

	profile.Verified = msg.IsVerified
	profile.UpdatedAt = time.Now()

	// Store updated profile
	bz, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	store.Set(types.GetProfileKey(msg.User), bz)

	k.Logger(ctx).Info("Verification status changed", "user", msg.User, "isVerified", msg.IsVerified)

	return nil
}

// HasBadge checks if a user has a specific badge
func (k Keeper) HasBadge(ctx sdk.Context, address, badgeName string) (bool, error) {
	profile, err := k.GetProfile(ctx, address)
	if err != nil {
		return false, err
	}

	badge := types.NewBadge(badgeName, "", "")
	if profile.Badges == nil {
		return false, nil
	}

	b, exists := profile.Badges[badge.ID]
	if !exists {
		return false, nil
	}

	return !b.IsRevoked, nil
}

// GetTrustScore retrieves a user's trust score
func (k Keeper) GetTrustScore(ctx sdk.Context, address string) (uint64, error) {
	profile, err := k.GetProfile(ctx, address)
	if err != nil {
		return 0, err
	}
	return profile.TrustScore, nil
}

// RemoveCredential removes a credential from a profile
func (k Keeper) RemoveCredential(ctx sdk.Context, msg *types.MsgRemoveCredential) error {
	store := ctx.KVStore(k.storeKey)

	// Get profile
	profile, err := k.GetProfile(ctx, msg.Creator)
	if err != nil {
		return err
	}

	// Find and remove credential
	found := false
	credentials := make([]string, 0, len(profile.Credentials))
	for _, cred := range profile.Credentials {
		if cred == msg.AttestationUID {
			found = true
			continue
		}
		credentials = append(credentials, cred)
	}

	if !found {
		return types.ErrCredentialNotFound
	}

	profile.Credentials = credentials
	profile.UpdatedAt = time.Now()

	// Store updated profile
	bz, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	store.Set(types.GetProfileKey(msg.Creator), bz)

	k.Logger(ctx).Info("Credential removed", "address", msg.Creator, "attestationUID", msg.AttestationUID)

	return nil
}

// RegisterHandle registers a unique handle for a profile
func (k Keeper) RegisterHandle(ctx sdk.Context, msg *types.MsgRegisterHandle) error {
	store := ctx.KVStore(k.storeKey)

	// Check if profile exists
	profile, err := k.GetProfile(ctx, msg.Creator)
	if err != nil {
		return err
	}

	// Check if handle is already taken
	handleKey := types.GetHandleToAddressKey(msg.Handle)
	if store.Has(handleKey) {
		return types.ErrHandleAlreadyTaken
	}

	// If profile already has a handle, remove the old mapping
	if profile.Handle != "" {
		oldHandleKey := types.GetHandleToAddressKey(profile.Handle)
		store.Delete(oldHandleKey)
	}

	// Update profile with new handle
	profile.Handle = msg.Handle
	profile.UpdatedAt = time.Now()

	// Store profile
	bz, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	store.Set(types.GetProfileKey(msg.Creator), bz)

	// Store handle-to-address mapping
	store.Set(handleKey, []byte(msg.Creator))

	k.Logger(ctx).Info("Handle registered", "address", msg.Creator, "handle", msg.Handle)

	return nil
}

// GetProfileByHandle retrieves a CertID profile by handle
func (k Keeper) GetProfileByHandle(ctx sdk.Context, handle string) (*types.CertID, error) {
	store := ctx.KVStore(k.storeKey)

	// Get address from handle
	handleKey := types.GetHandleToAddressKey(handle)
	addressBz := store.Get(handleKey)
	if addressBz == nil {
		return nil, types.ErrProfileNotFound
	}

	return k.GetProfile(ctx, string(addressBz))
}

// AuthorizeOracle authorizes an oracle to perform operations
func (k Keeper) AuthorizeOracle(ctx sdk.Context, msg *types.MsgAuthorizeOracle) error {
	store := ctx.KVStore(k.storeKey)

	auth := types.OracleAuthorization{
		Address:      msg.Oracle,
		IsAuthorized: true,
		AuthorizedAt: time.Now(),
		AuthorizedBy: msg.Authority,
	}

	bz, err := json.Marshal(auth)
	if err != nil {
		return err
	}
	store.Set(types.GetOracleKey(msg.Oracle), bz)

	k.Logger(ctx).Info("Oracle authorized", "oracle", msg.Oracle, "by", msg.Authority)

	return nil
}

// RevokeOracle revokes an oracle's authorization
func (k Keeper) RevokeOracle(ctx sdk.Context, msg *types.MsgRevokeOracle) error {
	store := ctx.KVStore(k.storeKey)

	oracleKey := types.GetOracleKey(msg.Oracle)
	if !store.Has(oracleKey) {
		return types.ErrOracleNotAuthorized
	}

	store.Delete(oracleKey)

	k.Logger(ctx).Info("Oracle revoked", "oracle", msg.Oracle, "by", msg.Authority)

	return nil
}

// IsOracleAuthorized checks if an oracle is authorized
func (k Keeper) IsOracleAuthorized(ctx sdk.Context, oracle string) bool {
	store := ctx.KVStore(k.storeKey)
	oracleKey := types.GetOracleKey(oracle)
	bz := store.Get(oracleKey)
	if bz == nil {
		return false
	}

	var auth types.OracleAuthorization
	if err := json.Unmarshal(bz, &auth); err != nil {
		return false
	}
	return auth.IsAuthorized
}

// RequestVerification creates a verification request
func (k Keeper) RequestVerification(ctx sdk.Context, msg *types.MsgRequestVerification) error {
	store := ctx.KVStore(k.storeKey)

	// Check if profile exists
	if _, err := k.GetProfile(ctx, msg.Creator); err != nil {
		return err
	}

	request := types.VerificationRequest{
		Address:          msg.Creator,
		RequestType:      msg.VerificationType,
		VerificationData: msg.VerificationData,
		Status:           "pending",
		CreatedAt:        time.Now(),
	}

	bz, err := json.Marshal(request)
	if err != nil {
		return err
	}
	store.Set(types.GetVerificationKey(msg.Creator, msg.VerificationType), bz)

	k.Logger(ctx).Info("Verification requested", "address", msg.Creator, "type", msg.VerificationType)

	return nil
}

// GetAllProfiles returns all CertID profiles (for genesis export)
func (k Keeper) GetAllProfiles(ctx sdk.Context) []types.CertID {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ProfileKeyPrefix)
	defer iterator.Close()

	var profiles []types.CertID
	for ; iterator.Valid(); iterator.Next() {
		var profile types.CertID
		if err := json.Unmarshal(iterator.Value(), &profile); err != nil {
			continue
		}
		profiles = append(profiles, profile)
	}

	return profiles
}

