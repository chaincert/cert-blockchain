package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	hardwaretypes "github.com/chaincertify/certd/x/hardware/types"
	"github.com/chaincertify/certd/x/trustscore/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods
// for the trustscore module's state.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	// Authority is the module authority address (for governance actions)
	authority string
}

// NewKeeper creates a new TrustScore Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
	}
}

// GetAuthority returns the module's authority address
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// CalculateHumanityScore delegates to the deterministic scoring algorithm
// in x/hardware/keeper/scoring.go. The trustscore module is responsible for
// storing and querying scores; the algorithm itself lives in x/hardware.
func (k Keeper) CalculateAndStoreHumanityScore(ctx sdk.Context, address string, factors hardwaretypes.HumanityFactors) (hardwaretypes.HumanityResult, error) {
	// Use the deterministic algorithm from x/hardware/keeper
	result := CalculateHumanityScore(factors)

	// Store the result
	store := ctx.KVStore(k.storeKey)
	scoreKey := types.GetScoreKey(address)

	bz, err := json.Marshal(result)
	if err != nil {
		return result, fmt.Errorf("failed to marshal humanity result: %w", err)
	}
	store.Set(scoreKey, bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeScoreCalculated,
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyScore, fmt.Sprintf("%d", result.Score)),
			sdk.NewAttribute(types.AttributeKeyVerifiedHuman, fmt.Sprintf("%t", result.IsVerifiedHuman)),
		),
	)

	k.Logger(ctx).Info("humanity score calculated",
		"address", address,
		"score", result.Score,
		"is_verified_human", result.IsVerifiedHuman,
	)

	return result, nil
}

// GetHumanityScore retrieves a stored humanity score for an address
func (k Keeper) GetHumanityScore(ctx sdk.Context, address string) (*hardwaretypes.HumanityResult, error) {
	store := ctx.KVStore(k.storeKey)
	scoreKey := types.GetScoreKey(address)

	bz := store.Get(scoreKey)
	if bz == nil {
		return nil, types.ErrScoreNotFound
	}

	var result hardwaretypes.HumanityResult
	if err := json.Unmarshal(bz, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal humanity result: %w", err)
	}

	return &result, nil
}
