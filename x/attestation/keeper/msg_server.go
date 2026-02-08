package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/chaincertify/certd/x/attestation/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// RegisterSchema handles MsgRegisterSchema
func (k msgServer) RegisterSchema(goCtx context.Context, msg *types.MsgRegisterSchema) (*types.MsgRegisterSchemaResponse, error) {
	if goCtx == nil {
		return nil, nil
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	var resolver sdk.AccAddress
	if msg.Resolver != "" {
		resolver, err = sdk.AccAddressFromBech32(msg.Resolver)
		if err != nil {
			return nil, err
		}
	}

	uid, err := k.Keeper.RegisterSchema(ctx, creator, msg.Schema, resolver, msg.Revocable)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSchemaRegistered,
			sdk.NewAttribute(types.AttributeKeySchemaUID, uid),
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyRevocable, boolToString(msg.Revocable)),
		),
	)

	return &types.MsgRegisterSchemaResponse{
		Uid: uid,
	}, nil
}

// Attest handles MsgAttest for creating public attestations
func (k msgServer) Attest(goCtx context.Context, msg *types.MsgAttest) (*types.MsgAttestResponse, error) {
	if goCtx == nil {
		return nil, nil
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	attester, err := sdk.AccAddressFromBech32(msg.Attester)
	if err != nil {
		return nil, err
	}

	var recipient sdk.AccAddress
	if msg.Recipient != "" {
		recipient, err = sdk.AccAddressFromBech32(msg.Recipient)
		if err != nil {
			return nil, err
		}
	}

	var expirationTime time.Time
	if msg.ExpirationTime > 0 {
		expirationTime = time.Unix(msg.ExpirationTime, 0)
	}

	uid, err := k.Keeper.CreateAttestation(
		ctx,
		attester,
		msg.SchemaUID,
		recipient,
		expirationTime,
		msg.Revocable,
		msg.RefUID,
		msg.Data,
	)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAttestationCreated,
			sdk.NewAttribute(types.AttributeKeyAttestationUID, uid),
			sdk.NewAttribute(types.AttributeKeyAttester, msg.Attester),
			sdk.NewAttribute(types.AttributeKeySchemaUID, msg.SchemaUID),
			sdk.NewAttribute(types.AttributeKeyAttestationType, types.AttestationTypePublic),
		),
	)

	return &types.MsgAttestResponse{
		Uid: uid,
	}, nil
}

// Revoke handles MsgRevoke for revoking attestations
func (k msgServer) Revoke(goCtx context.Context, msg *types.MsgRevoke) (*types.MsgRevokeResponse, error) {
	if goCtx == nil {
		return nil, nil
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	revoker, err := sdk.AccAddressFromBech32(msg.Revoker)
	if err != nil {
		return nil, err
	}

	err = k.Keeper.RevokeAttestation(ctx, revoker, msg.UID)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAttestationRevoked,
			sdk.NewAttribute(types.AttributeKeyAttestationUID, msg.UID),
			sdk.NewAttribute(types.AttributeKeyRevoker, msg.Revoker),
		),
	)

	return &types.MsgRevokeResponse{}, nil
}

// CreateEncryptedAttestation handles MsgCreateEncryptedAttestation
// This is the core privacy feature per Whitepaper Section 3
func (k msgServer) CreateEncryptedAttestation(goCtx context.Context, msg *types.MsgCreateEncryptedAttestation) (*types.MsgCreateEncryptedAttestationResponse, error) {
	if goCtx == nil {
		return nil, nil
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	attester, err := sdk.AccAddressFromBech32(msg.Attester)
	if err != nil {
		return nil, err
	}

	// Parse recipients
	recipients := make([]sdk.AccAddress, len(msg.Recipients))
	for i, recipientStr := range msg.Recipients {
		recipient, err := sdk.AccAddressFromBech32(recipientStr)
		if err != nil {
			return nil, err
		}
		recipients[i] = recipient
	}

	var expirationTime time.Time
	if msg.ExpirationTime > 0 {
		expirationTime = time.Unix(msg.ExpirationTime, 0)
	}

	uid, err := k.Keeper.CreateEncryptedAttestation(
		ctx,
		attester,
		msg.SchemaUID,
		msg.IPFSCID,
		msg.EncryptedDataHash,
		recipients,
		msg.EncryptedSymmetricKeys,
		msg.Revocable,
		expirationTime,
	)
	if err != nil {
		return nil, err
	}

	// Emit event (without sensitive data)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEncryptedAttestationCreated,
			sdk.NewAttribute(types.AttributeKeyAttestationUID, uid),
			sdk.NewAttribute(types.AttributeKeyAttester, msg.Attester),
			sdk.NewAttribute(types.AttributeKeySchemaUID, msg.SchemaUID),
			sdk.NewAttribute(types.AttributeKeyIPFSCID, msg.IPFSCID),
			sdk.NewAttribute(types.AttributeKeyRecipientsCount, intToString(len(msg.Recipients))),
		),
	)

	return &types.MsgCreateEncryptedAttestationResponse{
		Uid: uid,
	}, nil
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func intToString(n int) string {
	return string(rune(n + '0'))
}

