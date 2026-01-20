
package app

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MockDistributionKeeper implements utils.DistributionKeeper for chain configurations
// that do not have the distribution module (like CERT currently).
type MockDistributionKeeper struct{}

func (m MockDistributionKeeper) WithdrawDelegationRewards(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error) {
	// No-op: return no coins, no error
	return sdk.Coins{}, nil
}
