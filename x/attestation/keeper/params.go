package keeper

import (
	"encoding/json"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/chaincertify/certd/x/attestation/types"
)

// SetParams sets the attestation module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(params)
	store.Set(types.ParamsKey, bz)
}

// GetParams returns the attestation module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}
	var params types.Params
	json.Unmarshal(bz, &params)
	return params
}

// GetSchemaCount returns the total number of schemas
func (k Keeper) GetSchemaCount(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.SchemaKeyPrefix)
	defer iterator.Close()

	var count uint64
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return count
}

// GetAllSchemas returns all registered schemas
func (k Keeper) GetAllSchemas(ctx sdk.Context) []types.Schema {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.SchemaKeyPrefix)
	defer iterator.Close()

	var schemas []types.Schema
	for ; iterator.Valid(); iterator.Next() {
		var schema types.Schema
		if err := json.Unmarshal(iterator.Value(), &schema); err == nil {
			schemas = append(schemas, schema)
		}
	}
	return schemas
}

// GetAllAttestations returns all attestations
func (k Keeper) GetAllAttestations(ctx sdk.Context) []types.Attestation {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.AttestationKeyPrefix)
	defer iterator.Close()

	var attestations []types.Attestation
	for ; iterator.Valid(); iterator.Next() {
		var attestation types.Attestation
		if err := json.Unmarshal(iterator.Value(), &attestation); err == nil {
			attestations = append(attestations, attestation)
		}
	}
	return attestations
}

// GetAllEncryptedAttestations returns all encrypted attestations
func (k Keeper) GetAllEncryptedAttestations(ctx sdk.Context) []types.EncryptedAttestation {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.EncryptedAttestationKeyPrefix)
	defer iterator.Close()

	var attestations []types.EncryptedAttestation
	for ; iterator.Valid(); iterator.Next() {
		var attestation types.EncryptedAttestation
		if err := json.Unmarshal(iterator.Value(), &attestation); err == nil {
			attestations = append(attestations, attestation)
		}
	}
	return attestations
}

// ImportAttestation imports an attestation during genesis
func (k Keeper) ImportAttestation(ctx sdk.Context, attestation types.Attestation) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(attestation)
	store.Set(types.GetAttestationKey(attestation.UID), bz)
	k.incrementAttestationCount(ctx)
}

// ImportEncryptedAttestation imports an encrypted attestation during genesis
func (k Keeper) ImportEncryptedAttestation(ctx sdk.Context, attestation types.EncryptedAttestation) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(attestation)
	store.Set(types.GetEncryptedAttestationKey(attestation.UID), bz)
	store.Set(types.GetAttestationKey(attestation.UID), bz)
	k.incrementAttestationCount(ctx)
	k.incrementEncryptedAttestationCount(ctx)
}
