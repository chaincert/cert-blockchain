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
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
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

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// CreateProfile creates a new CertID profile
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

	// Store profile
	bz, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}
	store.Set(key, bz)

	k.Logger(ctx).Info("Profile created", "address", msg.Creator)

	return profile, nil
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

