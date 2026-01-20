package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ante "github.com/evmos/evmos/v20/app/ante"
)

// NewAnteHandler returns an ante handler responsible for attempting to route an
// Ethereum or SDK transaction to an internal ante handler for performing
// transaction-level processing (e.g. fee payment, signature verification) before
// being passed onto it's respective handler.
func NewAnteHandler(options ante.HandlerOptions) (sdk.AnteHandler, error) {
	return ante.NewAnteHandler(options), nil
}

