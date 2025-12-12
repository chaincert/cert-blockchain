package certid

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/chaincertify/certd/x/certid/keeper"
	"github.com/chaincertify/certd/x/certid/types"
)

// GenesisState defines the CertID module's genesis state
type GenesisState struct {
	Params   Params         `json:"params" protobuf:"bytes,1,opt,name=params,proto3"`
	Profiles []types.CertID `json:"profiles" protobuf:"bytes,2,rep,name=profiles,proto3"`
}

// Proto interface implementations for GenesisState
func (gs *GenesisState) Reset()         { *gs = GenesisState{} }
func (gs *GenesisState) String() string { return "GenesisState" }
func (gs *GenesisState) ProtoMessage()  {}

// Params defines the parameters for the CertID module
type Params struct {
	// MaxNameLength is the maximum length for profile names
	MaxNameLength uint32 `json:"maxNameLength" protobuf:"varint,1,opt,name=max_name_length,proto3"`

	// MaxBioLength is the maximum length for profile bios
	MaxBioLength uint32 `json:"maxBioLength" protobuf:"varint,2,opt,name=max_bio_length,proto3"`

	// MaxSocialLinks is the maximum number of social links per profile
	MaxSocialLinks uint32 `json:"maxSocialLinks" protobuf:"varint,3,opt,name=max_social_links,proto3"`

	// MaxCredentials is the maximum number of credentials per profile
	MaxCredentials uint32 `json:"maxCredentials" protobuf:"varint,4,opt,name=max_credentials,proto3"`

	// ProfileCreationFee is the fee to create a profile (in ucert)
	ProfileCreationFee uint64 `json:"profileCreationFee" protobuf:"varint,5,opt,name=profile_creation_fee,proto3"`

	// VerificationFee is the fee for verification requests (in ucert)
	VerificationFee uint64 `json:"verificationFee" protobuf:"varint,6,opt,name=verification_fee,proto3"`
}

// Proto interface implementations for Params
func (p *Params) Reset()         { *p = Params{} }
func (p *Params) String() string { return "Params" }
func (p *Params) ProtoMessage()  {}

// DefaultParams returns default module parameters
func DefaultParams() Params {
	return Params{
		MaxNameLength:      100,
		MaxBioLength:       500,
		MaxSocialLinks:     10,
		MaxCredentials:     50,
		ProfileCreationFee: 1_000_000,  // 1 CERT
		VerificationFee:    10_000_000, // 10 CERT
	}
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:   DefaultParams(),
		Profiles: []types.CertID{},
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	// Validate params
	if gs.Params.MaxNameLength == 0 {
		return types.ErrInvalidName
	}
	if gs.Params.MaxBioLength == 0 {
		return types.ErrInvalidBio
	}

	// Validate profiles
	seen := make(map[string]bool)
	for _, profile := range gs.Profiles {
		if seen[profile.Address] {
			return types.ErrProfileAlreadyExists
		}
		seen[profile.Address] = true
	}

	return nil
}

// InitGenesis initializes the module's state from a genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState GenesisState) {
	// Set params
	// k.SetParams(ctx, genState.Params)

	// Import profiles
	for _, profile := range genState.Profiles {
		msg := &types.MsgCreateProfile{
			Creator:     profile.Address,
			Name:        profile.Name,
			Bio:         profile.Bio,
			AvatarCID:   profile.AvatarCID,
			PublicKey:   profile.PublicKey,
			SocialLinks: profile.SocialLinks,
		}
		k.CreateProfile(ctx, msg)
	}
}

// ExportGenesis exports the module's state to a genesis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *GenesisState {
	return &GenesisState{
		Params:   DefaultParams(),
		Profiles: []types.CertID{}, // TODO: Export all profiles
	}
}
