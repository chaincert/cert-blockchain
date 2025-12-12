package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterCodec registers the necessary types for the module
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateProfile{}, "certid/CreateProfile", nil)
	cdc.RegisterConcrete(&MsgUpdateProfile{}, "certid/UpdateProfile", nil)
	cdc.RegisterConcrete(&MsgAddCredential{}, "certid/AddCredential", nil)
	cdc.RegisterConcrete(&MsgVerifySocial{}, "certid/VerifySocial", nil)
}

// RegisterInterfaces registers the module's interface types
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateProfile{},
		&MsgUpdateProfile{},
		&MsgAddCredential{},
		&MsgVerifySocial{},
	)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(amino)
	amino.Seal()
}

// MsgServer is the server API for Msg service
type MsgServer interface {
	CreateProfile(ctx sdk.Context, msg *MsgCreateProfile) (*MsgCreateProfileResponse, error)
	UpdateProfile(ctx sdk.Context, msg *MsgUpdateProfile) (*MsgUpdateProfileResponse, error)
	AddCredential(ctx sdk.Context, msg *MsgAddCredential) (*MsgAddCredentialResponse, error)
	VerifySocial(ctx sdk.Context, msg *MsgVerifySocial) (*MsgVerifySocialResponse, error)
}

// Response types
type MsgCreateProfileResponse struct {
	Address string `json:"address"`
}

type MsgUpdateProfileResponse struct {
	Address string `json:"address"`
}

type MsgAddCredentialResponse struct {
	CredentialUID string `json:"credentialUid"`
}

type MsgVerifySocialResponse struct {
	Verified bool   `json:"verified"`
	Platform string `json:"platform"`
}
