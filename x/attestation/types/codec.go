package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterCodec registers the necessary types for Amino JSON serialization
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterSchema{}, "cert/attestation/MsgRegisterSchema", nil)
	cdc.RegisterConcrete(&MsgAttest{}, "cert/attestation/MsgAttest", nil)
	cdc.RegisterConcrete(&MsgRevoke{}, "cert/attestation/MsgRevoke", nil)
	cdc.RegisterConcrete(&MsgCreateEncryptedAttestation{}, "cert/attestation/MsgCreateEncryptedAttestation", nil)
}

// RegisterInterfaces registers the module types with the interface registry
// Each message type is registered separately to ensure unique TypeURLs
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// Register each message type individually with explicit TypeURL
	// The XXX_MessageName() method on each message provides the unique TypeURL
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterSchema{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgAttest{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRevoke{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateEncryptedAttestation{},
	)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(Amino)
	Amino.Seal()
}
