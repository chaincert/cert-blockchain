package types

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	attestationtypes "github.com/chaincertify/certd/x/attestation/types"
)

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

// AttestationKeeper defines the expected attestation keeper interface
// Used to create attestations when CertID profiles are created
type AttestationKeeper interface {
	// CreateAttestation creates a new public attestation
	CreateAttestation(
		ctx sdk.Context,
		attester sdk.AccAddress,
		schemaUID string,
		recipient sdk.AccAddress,
		expirationTime time.Time,
		revocable bool,
		refUID string,
		data []byte,
	) (string, error)

	// GetSchema retrieves a schema by UID (used to verify CertID schema exists)
	GetSchema(ctx sdk.Context, uid string) (*attestationtypes.Schema, error)

	// RegisterSchema registers a new schema (used for genesis initialization)
	RegisterSchema(ctx sdk.Context, creator sdk.AccAddress, schema string, resolver sdk.AccAddress, revocable bool) (string, error)
}

