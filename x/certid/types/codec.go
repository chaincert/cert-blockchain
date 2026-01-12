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
	cdc.RegisterConcrete(&MsgRemoveCredential{}, "certid/RemoveCredential", nil)
	cdc.RegisterConcrete(&MsgVerifySocial{}, "certid/VerifySocial", nil)
	cdc.RegisterConcrete(&MsgRequestVerification{}, "certid/RequestVerification", nil)
	cdc.RegisterConcrete(&MsgAwardBadge{}, "certid/AwardBadge", nil)
	cdc.RegisterConcrete(&MsgRevokeBadge{}, "certid/RevokeBadge", nil)
	cdc.RegisterConcrete(&MsgUpdateTrustScore{}, "certid/UpdateTrustScore", nil)
	cdc.RegisterConcrete(&MsgSetVerificationStatus{}, "certid/SetVerificationStatus", nil)
	cdc.RegisterConcrete(&MsgAuthorizeOracle{}, "certid/AuthorizeOracle", nil)
	cdc.RegisterConcrete(&MsgRevokeOracle{}, "certid/RevokeOracle", nil)
	cdc.RegisterConcrete(&MsgRegisterHandle{}, "certid/RegisterHandle", nil)
}

// RegisterInterfaces registers the module's interface types
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateProfile{},
		&MsgUpdateProfile{},
		&MsgAddCredential{},
		&MsgRemoveCredential{},
		&MsgVerifySocial{},
		&MsgRequestVerification{},
		&MsgAwardBadge{},
		&MsgRevokeBadge{},
		&MsgUpdateTrustScore{},
		&MsgSetVerificationStatus{},
		&MsgAuthorizeOracle{},
		&MsgRevokeOracle{},
		&MsgRegisterHandle{},
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
	RemoveCredential(ctx sdk.Context, msg *MsgRemoveCredential) (*MsgRemoveCredentialResponse, error)
	VerifySocial(ctx sdk.Context, msg *MsgVerifySocial) (*MsgVerifySocialResponse, error)
	RequestVerification(ctx sdk.Context, msg *MsgRequestVerification) (*MsgRequestVerificationResponse, error)
	AwardBadge(ctx sdk.Context, msg *MsgAwardBadge) (*MsgAwardBadgeResponse, error)
	RevokeBadge(ctx sdk.Context, msg *MsgRevokeBadge) (*MsgRevokeBadgeResponse, error)
	UpdateTrustScore(ctx sdk.Context, msg *MsgUpdateTrustScore) (*MsgUpdateTrustScoreResponse, error)
	SetVerificationStatus(ctx sdk.Context, msg *MsgSetVerificationStatus) (*MsgSetVerificationStatusResponse, error)
	AuthorizeOracle(ctx sdk.Context, msg *MsgAuthorizeOracle) (*MsgAuthorizeOracleResponse, error)
	RevokeOracle(ctx sdk.Context, msg *MsgRevokeOracle) (*MsgRevokeOracleResponse, error)
	RegisterHandle(ctx sdk.Context, msg *MsgRegisterHandle) (*MsgRegisterHandleResponse, error)
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

type MsgRemoveCredentialResponse struct {
	Success bool `json:"success"`
}

type MsgVerifySocialResponse struct {
	Verified bool   `json:"verified"`
	Platform string `json:"platform"`
}

type MsgRequestVerificationResponse struct {
	RequestID string `json:"requestId"`
	Status    string `json:"status"`
}

type MsgAwardBadgeResponse struct {
	BadgeID string `json:"badgeId"`
}

type MsgRevokeBadgeResponse struct {
	Success bool `json:"success"`
}

type MsgUpdateTrustScoreResponse struct {
	OldScore uint64 `json:"oldScore"`
	NewScore uint64 `json:"newScore"`
}

type MsgSetVerificationStatusResponse struct {
	IsVerified bool `json:"isVerified"`
}

type MsgAuthorizeOracleResponse struct {
	Oracle string `json:"oracle"`
}

type MsgRevokeOracleResponse struct {
	Success bool `json:"success"`
}

type MsgRegisterHandleResponse struct {
	Handle string `json:"handle"`
}
