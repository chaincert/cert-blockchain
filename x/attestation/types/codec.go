package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
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

	// Register all attestation types with gogoproto for gRPC marshaling.
	// This enables the gRPC codec to marshal/unmarshal our manually-defined types.
	registerProtoTypes()
}

// registerProtoTypes registers all attestation module types with gogoproto's registry.
// This is required for gRPC services to properly marshal/unmarshal messages
// when using manually-defined types instead of protoc-generated code.
func registerProtoTypes() {
	// Core data types
	proto.RegisterType((*Attestation)(nil), "cert.attestation.v1.Attestation")
	proto.RegisterType((*EncryptedAttestation)(nil), "cert.attestation.v1.EncryptedAttestation")
	proto.RegisterType((*Schema)(nil), "cert.attestation.v1.Schema")
	proto.RegisterType((*Params)(nil), "cert.attestation.v1.Params")

	// Query request/response types
	proto.RegisterType((*QuerySchemaRequest)(nil), "cert.attestation.v1.QuerySchemaRequest")
	proto.RegisterType((*QuerySchemaResponse)(nil), "cert.attestation.v1.QuerySchemaResponse")
	proto.RegisterType((*QueryAttestationRequest)(nil), "cert.attestation.v1.QueryAttestationRequest")
	proto.RegisterType((*QueryAttestationResponse)(nil), "cert.attestation.v1.QueryAttestationResponse")
	proto.RegisterType((*QueryAttestationsByAttesterRequest)(nil), "cert.attestation.v1.QueryAttestationsByAttesterRequest")
	proto.RegisterType((*QueryAttestationsByAttesterResponse)(nil), "cert.attestation.v1.QueryAttestationsByAttesterResponse")
	proto.RegisterType((*QueryAttestationsByRecipientRequest)(nil), "cert.attestation.v1.QueryAttestationsByRecipientRequest")
	proto.RegisterType((*QueryAttestationsByRecipientResponse)(nil), "cert.attestation.v1.QueryAttestationsByRecipientResponse")
	proto.RegisterType((*QueryEncryptedAttestationRequest)(nil), "cert.attestation.v1.QueryEncryptedAttestationRequest")
	proto.RegisterType((*QueryEncryptedAttestationResponse)(nil), "cert.attestation.v1.QueryEncryptedAttestationResponse")
	proto.RegisterType((*QueryStatsRequest)(nil), "cert.attestation.v1.QueryStatsRequest")
	proto.RegisterType((*QueryStatsResponse)(nil), "cert.attestation.v1.QueryStatsResponse")

	// Message types (Tx)
	proto.RegisterType((*MsgRegisterSchema)(nil), "cert.attestation.v1.MsgRegisterSchema")
	proto.RegisterType((*MsgRegisterSchemaResponse)(nil), "cert.attestation.v1.MsgRegisterSchemaResponse")
	proto.RegisterType((*MsgAttest)(nil), "cert.attestation.v1.MsgAttest")
	proto.RegisterType((*MsgAttestResponse)(nil), "cert.attestation.v1.MsgAttestResponse")
	proto.RegisterType((*MsgRevoke)(nil), "cert.attestation.v1.MsgRevoke")
	proto.RegisterType((*MsgRevokeResponse)(nil), "cert.attestation.v1.MsgRevokeResponse")
	proto.RegisterType((*MsgCreateEncryptedAttestation)(nil), "cert.attestation.v1.MsgCreateEncryptedAttestation")
	proto.RegisterType((*MsgCreateEncryptedAttestationResponse)(nil), "cert.attestation.v1.MsgCreateEncryptedAttestationResponse")
}
