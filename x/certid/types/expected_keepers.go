package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
	HasAccount(ctx sdk.Context, addr sdk.AccAddress) bool
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
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
	GetSchema(ctx sdk.Context, uid string) (*Schema, error)

	// RegisterSchema registers a new schema (used for genesis initialization)
	RegisterSchema(ctx sdk.Context, creator sdk.AccAddress, schema string, resolver sdk.AccAddress, revocable bool) (string, error)
}

// Schema represents an attestation schema (mirrors attestation/types.Schema)
type Schema struct {
	UID       string         `json:"uid"`
	Resolver  sdk.AccAddress `json:"resolver,omitempty"`
	Revocable bool           `json:"revocable"`
	Schema    string         `json:"schema"`
	Creator   sdk.AccAddress `json:"creator"`
}

