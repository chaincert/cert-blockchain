package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/chaincertify/certd/x/certid/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) CreateProfile(ctx context.Context, msg *types.MsgCreateProfile) (*types.MsgCreateProfileResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	profile, err := k.Keeper.CreateProfile(sdkCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgCreateProfileResponse{Address: profile.Address}, nil
}

func (k msgServer) UpdateProfile(ctx context.Context, msg *types.MsgUpdateProfile) (*types.MsgUpdateProfileResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	profile, err := k.Keeper.UpdateProfile(sdkCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgUpdateProfileResponse{Address: profile.Address}, nil
}

func (k msgServer) AddCredential(ctx context.Context, msg *types.MsgAddCredential) (*types.MsgAddCredentialResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err := k.Keeper.AddCredential(sdkCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgAddCredentialResponse{CredentialUID: msg.AttestationUID}, nil
}

func (k msgServer) RemoveCredential(ctx context.Context, msg *types.MsgRemoveCredential) (*types.MsgRemoveCredentialResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err := k.Keeper.RemoveCredential(sdkCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgRemoveCredentialResponse{Success: true}, nil
}

func (k msgServer) VerifySocial(ctx context.Context, msg *types.MsgVerifySocial) (*types.MsgVerifySocialResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err := k.Keeper.VerifySocial(sdkCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgVerifySocialResponse{Verified: true, Platform: msg.Platform}, nil
}

func (k msgServer) RequestVerification(ctx context.Context, msg *types.MsgRequestVerification) (*types.MsgRequestVerificationResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err := k.Keeper.RequestVerification(sdkCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgRequestVerificationResponse{RequestID: msg.VerificationType, Status: "pending"}, nil
}

func (k msgServer) AwardBadge(ctx context.Context, msg *types.MsgAwardBadge) (*types.MsgAwardBadgeResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Check if sender is authority
	if msg.Authority != k.Keeper.GetAuthority() {
		return nil, types.ErrUnauthorized
	}
	err := k.Keeper.AwardBadge(sdkCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgAwardBadgeResponse{BadgeID: msg.BadgeName}, nil
}

func (k msgServer) RevokeBadge(ctx context.Context, msg *types.MsgRevokeBadge) (*types.MsgRevokeBadgeResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if msg.Authority != k.Keeper.GetAuthority() {
		return nil, types.ErrUnauthorized
	}
	err := k.Keeper.RevokeBadge(sdkCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgRevokeBadgeResponse{Success: true}, nil
}

func (k msgServer) UpdateTrustScore(ctx context.Context, msg *types.MsgUpdateTrustScore) (*types.MsgUpdateTrustScoreResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Check if sender is authority or authorized oracle
	if msg.Authority != k.Keeper.GetAuthority() && !k.Keeper.IsOracleAuthorized(sdkCtx, msg.Authority) {
		return nil, types.ErrUnauthorized
	}
	oldScore, _ := k.Keeper.GetTrustScore(sdkCtx, msg.User)
	err := k.Keeper.UpdateTrustScore(sdkCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgUpdateTrustScoreResponse{OldScore: oldScore, NewScore: msg.Score}, nil
}

func (k msgServer) SetVerificationStatus(ctx context.Context, msg *types.MsgSetVerificationStatus) (*types.MsgSetVerificationStatusResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if msg.Authority != k.Keeper.GetAuthority() {
		return nil, types.ErrUnauthorized
	}
	err := k.Keeper.SetVerificationStatus(sdkCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgSetVerificationStatusResponse{IsVerified: msg.IsVerified}, nil
}

func (k msgServer) AuthorizeOracle(ctx context.Context, msg *types.MsgAuthorizeOracle) (*types.MsgAuthorizeOracleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if msg.Authority != k.Keeper.GetAuthority() {
		return nil, types.ErrUnauthorized
	}
	err := k.Keeper.AuthorizeOracle(sdkCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgAuthorizeOracleResponse{Oracle: msg.Oracle}, nil
}

func (k msgServer) RevokeOracle(ctx context.Context, msg *types.MsgRevokeOracle) (*types.MsgRevokeOracleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if msg.Authority != k.Keeper.GetAuthority() {
		return nil, types.ErrUnauthorized
	}
	err := k.Keeper.RevokeOracle(sdkCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgRevokeOracleResponse{Success: true}, nil
}

func (k msgServer) RegisterHandle(ctx context.Context, msg *types.MsgRegisterHandle) (*types.MsgRegisterHandleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err := k.Keeper.RegisterHandle(sdkCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgRegisterHandleResponse{Handle: msg.Handle}, nil
}
