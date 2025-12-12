package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// AttestationKeyPrefix is the prefix for attestation store keys
	AttestationKeyPrefix = []byte{0x01}

	// SchemaKeyPrefix is the prefix for schema store keys
	SchemaKeyPrefix = []byte{0x02}

	// AttestationByAttesterPrefix indexes attestations by attester address
	AttestationByAttesterPrefix = []byte{0x03}

	// AttestationByRecipientPrefix indexes attestations by recipient address
	AttestationByRecipientPrefix = []byte{0x04}

	// AttestationBySchemaPrefix indexes attestations by schema UID
	AttestationBySchemaPrefix = []byte{0x05}

	// EncryptedAttestationKeyPrefix is the prefix for encrypted attestation store keys
	EncryptedAttestationKeyPrefix = []byte{0x06}

	// IPFSCIDIndexPrefix indexes encrypted attestations by IPFS CID
	IPFSCIDIndexPrefix = []byte{0x07}

	// AttestationCountKey stores the total attestation count
	AttestationCountKey = []byte{0x10}

	// EncryptedAttestationCountKey stores the encrypted attestation count
	EncryptedAttestationCountKey = []byte{0x11}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x20}
)

// GetAttestationKey returns the store key for an attestation by UID
func GetAttestationKey(uid string) []byte {
	return append(AttestationKeyPrefix, []byte(uid)...)
}

// GetSchemaKey returns the store key for a schema by UID
func GetSchemaKey(uid string) []byte {
	return append(SchemaKeyPrefix, []byte(uid)...)
}

// GetAttestationByAttesterKey returns the index key for attestations by attester
func GetAttestationByAttesterKey(attester sdk.AccAddress, uid string) []byte {
	key := append(AttestationByAttesterPrefix, attester.Bytes()...)
	return append(key, []byte(uid)...)
}

// GetAttestationByRecipientKey returns the index key for attestations by recipient
func GetAttestationByRecipientKey(recipient sdk.AccAddress, uid string) []byte {
	key := append(AttestationByRecipientPrefix, recipient.Bytes()...)
	return append(key, []byte(uid)...)
}

// GetAttestationBySchemaKey returns the index key for attestations by schema
func GetAttestationBySchemaKey(schemaUID string, uid string) []byte {
	key := append(AttestationBySchemaPrefix, []byte(schemaUID)...)
	return append(key, []byte(uid)...)
}

// GetEncryptedAttestationKey returns the store key for an encrypted attestation
func GetEncryptedAttestationKey(uid string) []byte {
	return append(EncryptedAttestationKeyPrefix, []byte(uid)...)
}

// GetIPFSCIDIndexKey returns the index key for encrypted attestations by IPFS CID
func GetIPFSCIDIndexKey(cid string) []byte {
	return append(IPFSCIDIndexPrefix, []byte(cid)...)
}

// GetAttestationIteratorPrefix returns the prefix for iterating all attestations
func GetAttestationIteratorPrefix() []byte {
	return AttestationKeyPrefix
}

// GetEncryptedAttestationIteratorPrefix returns the prefix for iterating encrypted attestations
func GetEncryptedAttestationIteratorPrefix() []byte {
	return EncryptedAttestationKeyPrefix
}

// GetAttestationsByAttesterIteratorPrefix returns the prefix for iterating by attester
func GetAttestationsByAttesterIteratorPrefix(attester sdk.AccAddress) []byte {
	return append(AttestationByAttesterPrefix, attester.Bytes()...)
}

// GetAttestationsByRecipientIteratorPrefix returns the prefix for iterating by recipient
func GetAttestationsByRecipientIteratorPrefix(recipient sdk.AccAddress) []byte {
	return append(AttestationByRecipientPrefix, recipient.Bytes()...)
}

// Uint64ToBytes converts uint64 to bytes
func Uint64ToBytes(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

// BytesToUint64 converts bytes to uint64
func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

