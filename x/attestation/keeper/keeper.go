package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/chaincertify/certd/x/attestation/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods
// for the various parts of the attestation state machine
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	memKey   storetypes.StoreKey

	// Authority is the address capable of executing governance proposals
	authority string
}

// NewKeeper creates a new attestation Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey, memKey storetypes.StoreKey,
	authority string,
) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		memKey:    memKey,
		authority: authority,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the authority address for governance
func (k Keeper) GetAuthority() string {
	return k.authority
}

// RegisterSchema registers a new attestation schema
func (k Keeper) RegisterSchema(ctx sdk.Context, creator sdk.AccAddress, schema string, resolver sdk.AccAddress, revocable bool) (string, error) {
	store := ctx.KVStore(k.storeKey)

	// Generate schema UID
	schemaUID := types.GenerateSchemaUID(schema, resolver, revocable)

	// Check if schema already exists
	if store.Has(types.GetSchemaKey(schemaUID)) {
		return "", fmt.Errorf("schema already exists: %s", schemaUID)
	}

	// Create schema object
	schemaObj := types.Schema{
		UID:       schemaUID,
		Resolver:  resolver,
		Revocable: revocable,
		Schema:    schema,
		Creator:   creator,
	}

	// Serialize and store
	bz, err := json.Marshal(schemaObj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}

	store.Set(types.GetSchemaKey(schemaUID), bz)

	k.Logger(ctx).Info("Schema registered", "uid", schemaUID, "creator", creator.String())

	return schemaUID, nil
}

// GetSchema retrieves a schema by UID
func (k Keeper) GetSchema(ctx sdk.Context, uid string) (*types.Schema, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetSchemaKey(uid))
	if bz == nil {
		return nil, fmt.Errorf("schema not found: %s", uid)
	}

	var schema types.Schema
	if err := json.Unmarshal(bz, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &schema, nil
}

// CreateAttestation creates a new public attestation
func (k Keeper) CreateAttestation(
	ctx sdk.Context,
	attester sdk.AccAddress,
	schemaUID string,
	recipient sdk.AccAddress,
	expirationTime time.Time,
	revocable bool,
	refUID string,
	data []byte,
) (string, error) {
	store := ctx.KVStore(k.storeKey)

	// Verify schema exists
	schema, err := k.GetSchema(ctx, schemaUID)
	if err != nil {
		return "", err
	}

	// Check revocability against schema
	if revocable && !schema.Revocable {
		return "", fmt.Errorf("schema does not allow revocable attestations")
	}

	// Generate attestation UID
	uid := types.GenerateUID(attester, schemaUID, ctx.BlockTime(), data)

	// Create attestation
	attestation := types.Attestation{
		UID:             uid,
		SchemaUID:       schemaUID,
		Attester:        attester,
		Recipient:       recipient,
		Time:            ctx.BlockTime(),
		ExpirationTime:  expirationTime,
		Revocable:       revocable,
		RefUID:          refUID,
		Data:            data,
		AttestationType: types.AttestationTypePublic,
	}

	// Serialize and store
	bz, err := json.Marshal(attestation)
	if err != nil {
		return "", fmt.Errorf("failed to marshal attestation: %w", err)
	}

	store.Set(types.GetAttestationKey(uid), bz)

	// Create indexes
	store.Set(types.GetAttestationByAttesterKey(attester, uid), []byte{1})
	if len(recipient) > 0 {
		store.Set(types.GetAttestationByRecipientKey(recipient, uid), []byte{1})
	}
	store.Set(types.GetAttestationBySchemaKey(schemaUID, uid), []byte{1})

	// Increment attestation count
	k.incrementAttestationCount(ctx)

	k.Logger(ctx).Info("Attestation created", "uid", uid, "attester", attester.String(), "type", "public")

	return uid, nil
}

