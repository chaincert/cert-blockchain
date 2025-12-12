package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/chaincertify/certd/x/attestation/types"
)

// CreateEncryptedAttestation creates a new encrypted attestation
// This implements Step 4 of the Encryption Flow per Whitepaper Section 3.2
func (k Keeper) CreateEncryptedAttestation(
	ctx sdk.Context,
	attester sdk.AccAddress,
	schemaUID string,
	ipfsCID string,
	encryptedDataHash string,
	recipients []sdk.AccAddress,
	encryptedSymmetricKeys map[string]string,
	revocable bool,
	expirationTime time.Time,
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

	// Validate IPFS CID format (basic validation)
	if len(ipfsCID) < 46 {
		return "", fmt.Errorf("invalid IPFS CID format")
	}

	// Validate encrypted data hash (should be hex-encoded SHA-256)
	if len(encryptedDataHash) != 64 {
		return "", fmt.Errorf("invalid encrypted data hash format (expected SHA-256 hex)")
	}

	// Generate attestation UID using encrypted data hash as part of data
	uid := types.GenerateUID(attester, schemaUID, ctx.BlockTime(), []byte(encryptedDataHash))

	// Build base attestation
	baseAttestation := types.Attestation{
		UID:             uid,
		SchemaUID:       schemaUID,
		Attester:        attester,
		Recipient:       recipients[0], // Primary recipient
		Time:            ctx.BlockTime(),
		ExpirationTime:  expirationTime,
		Revocable:       revocable,
		Data:            []byte(encryptedDataHash),
		AttestationType: types.AttestationTypeEncryptedMultiRecipient,
	}

	// If single recipient, use simpler type
	if len(recipients) == 1 {
		baseAttestation.AttestationType = types.AttestationTypeEncryptedFile
	}

	// Create encrypted attestation
	encryptedAttestation := types.EncryptedAttestation{
		Attestation:            baseAttestation,
		IPFSCID:                ipfsCID,
		EncryptedDataHash:      encryptedDataHash,
		Recipients:             recipients,
		EncryptedSymmetricKeys: encryptedSymmetricKeys,
	}

	// Serialize and store
	bz, err := json.Marshal(encryptedAttestation)
	if err != nil {
		return "", fmt.Errorf("failed to marshal encrypted attestation: %w", err)
	}

	store.Set(types.GetEncryptedAttestationKey(uid), bz)
	store.Set(types.GetAttestationKey(uid), bz) // Also store in main attestation store

	// Create indexes
	store.Set(types.GetAttestationByAttesterKey(attester, uid), []byte{1})
	store.Set(types.GetIPFSCIDIndexKey(ipfsCID), []byte(uid))
	store.Set(types.GetAttestationBySchemaKey(schemaUID, uid), []byte{1})

	// Index by each recipient
	for _, recipient := range recipients {
		store.Set(types.GetAttestationByRecipientKey(recipient, uid), []byte{1})
	}

	// Increment counts
	k.incrementAttestationCount(ctx)
	k.incrementEncryptedAttestationCount(ctx)

	k.Logger(ctx).Info("Encrypted attestation created",
		"uid", uid,
		"attester", attester.String(),
		"ipfs_cid", ipfsCID,
		"recipients_count", len(recipients),
	)

	return uid, nil
}

// GetEncryptedAttestation retrieves an encrypted attestation by UID
func (k Keeper) GetEncryptedAttestation(ctx sdk.Context, uid string) (*types.EncryptedAttestation, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetEncryptedAttestationKey(uid))
	if bz == nil {
		return nil, fmt.Errorf("encrypted attestation not found: %s", uid)
	}

	var attestation types.EncryptedAttestation
	if err := json.Unmarshal(bz, &attestation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal encrypted attestation: %w", err)
	}

	return &attestation, nil
}

// GetEncryptedAttestationByIPFSCID retrieves an encrypted attestation by IPFS CID
func (k Keeper) GetEncryptedAttestationByIPFSCID(ctx sdk.Context, cid string) (*types.EncryptedAttestation, error) {
	store := ctx.KVStore(k.storeKey)

	uidBz := store.Get(types.GetIPFSCIDIndexKey(cid))
	if uidBz == nil {
		return nil, fmt.Errorf("no attestation found for IPFS CID: %s", cid)
	}

	return k.GetEncryptedAttestation(ctx, string(uidBz))
}

// IsRecipientAuthorized checks if an address is authorized to access an encrypted attestation
// This implements access control per Whitepaper Section 3.3
func (k Keeper) IsRecipientAuthorized(ctx sdk.Context, uid string, address sdk.AccAddress) (bool, error) {
	attestation, err := k.GetEncryptedAttestation(ctx, uid)
	if err != nil {
		return false, err
	}

	// Check if address is in recipients list
	for _, recipient := range attestation.Recipients {
		if recipient.Equals(address) {
			return true, nil
		}
	}

	// Also allow attester to access
	if attestation.Attester.Equals(address) {
		return true, nil
	}

	return false, nil
}

// GetEncryptedKeyForRecipient retrieves the wrapped symmetric key for a specific recipient
func (k Keeper) GetEncryptedKeyForRecipient(ctx sdk.Context, uid string, recipient sdk.AccAddress) (string, error) {
	attestation, err := k.GetEncryptedAttestation(ctx, uid)
	if err != nil {
		return "", err
	}

	key, exists := attestation.EncryptedSymmetricKeys[recipient.String()]
	if !exists {
		return "", fmt.Errorf("no encrypted key found for recipient: %s", recipient.String())
	}

	return key, nil
}

