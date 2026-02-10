package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CertIDKeeperI defines the expected interface for the CertID keeper
// Used for linking devices to CertID profiles
type CertIDKeeperI interface {
	// GetProfile retrieves a CertID profile by address
	GetProfile(ctx sdk.Context, address string) (interface{}, error)

	// UpdateProfileDevices updates the linked devices for a profile
	UpdateProfileDevices(ctx sdk.Context, address string, deviceIDs []string) error

	// GetTrustScore retrieves a user's trust score
	GetTrustScore(ctx sdk.Context, address string) (uint64, error)
}
