package keeper

import (
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/chaincertify/certd/x/attestation/types"
)

// GetAttestation retrieves an attestation by UID
func (k Keeper) GetAttestation(ctx sdk.Context, uid string) (*types.Attestation, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetAttestationKey(uid))
	if bz == nil {
		return nil, fmt.Errorf("attestation not found: %s", uid)
	}

	var attestation types.Attestation
	if err := json.Unmarshal(bz, &attestation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attestation: %w", err)
	}

	return &attestation, nil
}

// RevokeAttestation revokes an existing attestation
func (k Keeper) RevokeAttestation(ctx sdk.Context, revoker sdk.AccAddress, uid string) error {
	store := ctx.KVStore(k.storeKey)

	attestation, err := k.GetAttestation(ctx, uid)
	if err != nil {
		return err
	}

	// Check if attestation is revocable
	if !attestation.Revocable {
		return fmt.Errorf("attestation is not revocable")
	}

	// Check if already revoked
	if !attestation.RevocationTime.IsZero() {
		return fmt.Errorf("attestation already revoked at: %s", attestation.RevocationTime.String())
	}

	// Only attester can revoke
	if !attestation.Attester.Equals(revoker) {
		return fmt.Errorf("only the attester can revoke this attestation")
	}

	// Set revocation time
	attestation.RevocationTime = ctx.BlockTime()

	// Update store
	bz, err := json.Marshal(attestation)
	if err != nil {
		return fmt.Errorf("failed to marshal revoked attestation: %w", err)
	}

	store.Set(types.GetAttestationKey(uid), bz)

	k.Logger(ctx).Info("Attestation revoked", "uid", uid, "revoker", revoker.String())

	return nil
}

// GetAttestationsByAttester returns all attestations created by an attester
func (k Keeper) GetAttestationsByAttester(ctx sdk.Context, attester sdk.AccAddress) ([]types.Attestation, error) {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetAttestationsByAttesterIteratorPrefix(attester)

	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var attestations []types.Attestation
	for ; iterator.Valid(); iterator.Next() {
		// Extract UID from key (key = prefix + attester + uid)
		key := iterator.Key()
		uid := string(key[len(prefix):])

		attestation, err := k.GetAttestation(ctx, uid)
		if err != nil {
			continue // Skip invalid attestations
		}
		attestations = append(attestations, *attestation)
	}

	return attestations, nil
}

// GetAttestationsByRecipient returns all attestations where address is a recipient
func (k Keeper) GetAttestationsByRecipient(ctx sdk.Context, recipient sdk.AccAddress) ([]types.Attestation, error) {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetAttestationsByRecipientIteratorPrefix(recipient)

	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var attestations []types.Attestation
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		uid := string(key[len(prefix):])

		attestation, err := k.GetAttestation(ctx, uid)
		if err != nil {
			continue
		}
		attestations = append(attestations, *attestation)
	}

	return attestations, nil
}

// IsAttestationExpired checks if an attestation has expired
func (k Keeper) IsAttestationExpired(ctx sdk.Context, uid string) (bool, error) {
	attestation, err := k.GetAttestation(ctx, uid)
	if err != nil {
		return false, err
	}

	if attestation.ExpirationTime.IsZero() {
		return false, nil // Never expires
	}

	return ctx.BlockTime().After(attestation.ExpirationTime), nil
}

// IsAttestationRevoked checks if an attestation has been revoked
func (k Keeper) IsAttestationRevoked(ctx sdk.Context, uid string) (bool, error) {
	attestation, err := k.GetAttestation(ctx, uid)
	if err != nil {
		return false, err
	}

	return !attestation.RevocationTime.IsZero(), nil
}

// GetAttestationCount returns the total number of attestations
func (k Keeper) GetAttestationCount(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.AttestationCountKey)
	if bz == nil {
		return 0
	}
	return types.BytesToUint64(bz)
}

// GetEncryptedAttestationCount returns the number of encrypted attestations
func (k Keeper) GetEncryptedAttestationCount(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.EncryptedAttestationCountKey)
	if bz == nil {
		return 0
	}
	return types.BytesToUint64(bz)
}

// incrementAttestationCount increments the total attestation count
func (k Keeper) incrementAttestationCount(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	count := k.GetAttestationCount(ctx)
	store.Set(types.AttestationCountKey, types.Uint64ToBytes(count+1))
}

// incrementEncryptedAttestationCount increments the encrypted attestation count
func (k Keeper) incrementEncryptedAttestationCount(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	count := k.GetEncryptedAttestationCount(ctx)
	store.Set(types.EncryptedAttestationCountKey, types.Uint64ToBytes(count+1))
}
