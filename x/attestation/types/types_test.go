package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/chaincertify/certd/x/attestation/types"
)

func TestGenerateUID(t *testing.T) {
	attester := sdk.AccAddress("attester1")
	schemaUID := "0x1234567890abcdef"
	timestamp := time.Now()
	data := []byte(`{"test":"data"}`)

	uid1 := types.GenerateUID(attester, schemaUID, timestamp, data)
	uid2 := types.GenerateUID(attester, schemaUID, timestamp, data)

	// Same inputs should produce same UID
	require.Equal(t, uid1, uid2)

	// Different inputs should produce different UIDs
	uid3 := types.GenerateUID(attester, schemaUID, timestamp.Add(time.Second), data)
	require.NotEqual(t, uid1, uid3)

	// UID should be a valid hex string (64 chars for SHA-256)
	require.Len(t, uid1, 64)
}

func TestGenerateSchemaUID(t *testing.T) {
	schema := "string name, uint256 age"
	resolver := sdk.AccAddress{}
	revocable := true

	uid1 := types.GenerateSchemaUID(schema, resolver, revocable)
	uid2 := types.GenerateSchemaUID(schema, resolver, revocable)

	// Same inputs should produce same UID
	require.Equal(t, uid1, uid2)

	// Different revocable flag should produce different UID
	uid3 := types.GenerateSchemaUID(schema, resolver, false)
	require.NotEqual(t, uid1, uid3)

	// UID should be a valid hex string
	require.Len(t, uid1, 64)
}

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()

	// Per Whitepaper Section 12
	require.Equal(t, uint32(50), params.MaxRecipientsPerAttestation)
	require.Equal(t, uint64(100*1024*1024), params.MaxEncryptedFileSize)
	require.True(t, params.AttestationFee.IsZero())
}

func TestAttestationTypes(t *testing.T) {
	// Verify attestation type constants
	require.Equal(t, "public", types.AttestationTypePublic)
	require.Equal(t, "encrypted_file", types.AttestationTypeEncryptedFile)
	require.Equal(t, "encrypted_multi_recipient", types.AttestationTypeEncryptedMultiRecipient)
	require.Equal(t, "encrypted_business_document", types.AttestationTypeEncryptedBusinessDocument)
}

func TestModuleConstants(t *testing.T) {
	require.Equal(t, "attestation", types.ModuleName)
	require.Equal(t, "attestation", types.StoreKey)
	require.Equal(t, "attestation", types.RouterKey)
}

