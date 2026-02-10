package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HardwareKeeperI defines the expected interface for the hardware keeper
// Used for querying device trust data when calculating humanity scores
type HardwareKeeperI interface {
	// GetDevice retrieves a device by ID
	GetDevice(ctx sdk.Context, deviceID string) (interface{}, error)

	// GetDevicesByOwner returns all devices owned by an address
	GetDevicesByOwner(ctx sdk.Context, owner string) interface{}
}
