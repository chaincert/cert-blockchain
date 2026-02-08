package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/chaincertify/certd/x/certid/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

func (k queryServer) Profile(ctx context.Context, req *types.QueryProfileRequest) (*types.QueryProfileResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	profile, err := k.Keeper.GetProfile(sdkCtx, req.Address)
	if err != nil {
		return nil, err
	}
	return &types.QueryProfileResponse{Profile: profile}, nil
}

func (k queryServer) ProfileByHandle(ctx context.Context, req *types.QueryProfileByHandleRequest) (*types.QueryProfileByHandleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	profile, err := k.Keeper.GetProfileByHandle(sdkCtx, req.Handle)
	if err != nil {
		return nil, err
	}
	return &types.QueryProfileByHandleResponse{Profile: profile}, nil
}

func (k queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	// Simple params query - logic not fully implemented in keeper yet but placeholder
	return &types.QueryParamsResponse{Params: types.DefaultParams()}, nil
}
