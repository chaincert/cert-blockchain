package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/chaincertify/certd/x/attestation/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// Schema queries a schema by UID
func (k queryServer) Schema(goCtx context.Context, req *types.QuerySchemaRequest) (*types.QuerySchemaResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	schema, err := k.Keeper.GetSchema(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	return &types.QuerySchemaResponse{
		Schema: schema,
	}, nil
}

// Attestation queries an attestation by UID
func (k queryServer) Attestation(goCtx context.Context, req *types.QueryAttestationRequest) (*types.QueryAttestationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	attestation, err := k.Keeper.GetAttestation(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	return &types.QueryAttestationResponse{
		Attestation: attestation,
	}, nil
}

// AttestationsByAttester queries all attestations by an attester
func (k queryServer) AttestationsByAttester(goCtx context.Context, req *types.QueryAttestationsByAttesterRequest) (*types.QueryAttestationsByAttesterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	attester, err := sdk.AccAddressFromBech32(req.Attester)
	if err != nil {
		return nil, err
	}

	attestations, err := k.Keeper.GetAttestationsByAttester(ctx, attester)
	if err != nil {
		return nil, err
	}

	return &types.QueryAttestationsByAttesterResponse{
		Attestations: attestations,
	}, nil
}

// AttestationsByRecipient queries all attestations for a recipient
func (k queryServer) AttestationsByRecipient(goCtx context.Context, req *types.QueryAttestationsByRecipientRequest) (*types.QueryAttestationsByRecipientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	recipient, err := sdk.AccAddressFromBech32(req.Recipient)
	if err != nil {
		return nil, err
	}

	attestations, err := k.Keeper.GetAttestationsByRecipient(ctx, recipient)
	if err != nil {
		return nil, err
	}

	return &types.QueryAttestationsByRecipientResponse{
		Attestations: attestations,
	}, nil
}

// EncryptedAttestation queries an encrypted attestation with access control
func (k queryServer) EncryptedAttestation(goCtx context.Context, req *types.QueryEncryptedAttestationRequest) (*types.QueryEncryptedAttestationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	attestation, err := k.Keeper.GetEncryptedAttestation(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	response := &types.QueryEncryptedAttestationResponse{
		Attestation: attestation,
	}

	// If requester is provided, check authorization and include encrypted key
	if req.Requester != "" {
		requester, err := sdk.AccAddressFromBech32(req.Requester)
		if err != nil {
			return nil, err
		}

		authorized, err := k.Keeper.IsRecipientAuthorized(ctx, req.Uid, requester)
		if err != nil {
			return nil, err
		}

		if authorized {
			encryptedKey, err := k.Keeper.GetEncryptedKeyForRecipient(ctx, req.Uid, requester)
			if err == nil {
				response.EncryptedKey = encryptedKey
			}
		}
	}

	return response, nil
}

// Stats returns attestation statistics
func (k queryServer) Stats(goCtx context.Context, req *types.QueryStatsRequest) (*types.QueryStatsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryStatsResponse{
		TotalAttestations:          k.Keeper.GetAttestationCount(ctx),
		TotalEncryptedAttestations: k.Keeper.GetEncryptedAttestationCount(ctx),
		TotalSchemas:               k.Keeper.GetSchemaCount(ctx),
	}, nil
}

