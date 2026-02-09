package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenerateProof creates a deterministic proof hash for a trust score update.
// The proof combines the DID/address, score, and block context to produce a
// hash that the Stylus/EVM contract can verify on-chain.
//
// Proof structure: SHA256(did + score + blockHeight + blockTime)
// This ensures uniqueness per block and is resistant to replay attacks.
func (k Keeper) GenerateProof(ctx sdk.Context, did string, score uint64) string {
	blockHeight := ctx.BlockHeight()
	blockTime := ctx.BlockTime().Unix()

	// Deterministic proof payload
	payload := fmt.Sprintf("%s:%d:%d:%d", did, score, blockHeight, blockTime)
	hash := sha256.Sum256([]byte(payload))
	return "0x" + hex.EncodeToString(hash[:])
}

// GenerateProfileProof creates a proof hash for profile state changes.
// Used for profile creation/update events relayed to EVM chains.
func (k Keeper) GenerateProfileProof(ctx sdk.Context, address string) string {
	blockHeight := ctx.BlockHeight()
	blockTime := ctx.BlockTime().Unix()

	payload := fmt.Sprintf("profile:%s:%d:%d", address, blockHeight, blockTime)
	hash := sha256.Sum256([]byte(payload))
	return "0x" + hex.EncodeToString(hash[:])
}

// GenerateBadgeProof creates a proof hash for badge award/revoke events.
func (k Keeper) GenerateBadgeProof(ctx sdk.Context, user string, badgeID string) string {
	blockHeight := ctx.BlockHeight()
	blockTime := ctx.BlockTime().Unix()

	payload := fmt.Sprintf("badge:%s:%s:%d:%d", user, badgeID, blockHeight, blockTime)
	hash := sha256.Sum256([]byte(payload))
	return "0x" + hex.EncodeToString(hash[:])
}
