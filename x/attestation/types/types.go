package types

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "attestation"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_attestation"
)

// Attestation types per Whitepaper Section 3.4
const (
	AttestationTypePublic                    = "public"
	AttestationTypeEncryptedFile             = "encrypted_file"
	AttestationTypeEncryptedMultiRecipient   = "encrypted_multi_recipient"
	AttestationTypeEncryptedBusinessDocument = "encrypted_business_document"
)

// Attestation represents a generic on-chain attestation (EAS compatible)
// Per Whitepaper Section 2.2 and 3
type Attestation struct {
	// UID is the unique identifier for this attestation
	UID string `json:"uid" protobuf:"bytes,1,opt,name=uid,proto3"`

	// SchemaUID references the schema this attestation follows
	SchemaUID string `json:"schema_uid" protobuf:"bytes,2,opt,name=schema_uid,proto3"`

	// Attester is the address that created this attestation
	Attester sdk.AccAddress `json:"attester" protobuf:"bytes,3,opt,name=attester,proto3"`

	// Recipient is the primary recipient of this attestation
	Recipient sdk.AccAddress `json:"recipient,omitempty" protobuf:"bytes,4,opt,name=recipient,proto3"`

	// Time is the timestamp when the attestation was created
	Time time.Time `json:"time" protobuf:"bytes,5,opt,name=time,proto3,stdtime"`

	// ExpirationTime is when the attestation expires (0 = never)
	ExpirationTime time.Time `json:"expiration_time,omitempty" protobuf:"bytes,6,opt,name=expiration_time,proto3,stdtime"`

	// RevocationTime is when the attestation was revoked (0 = not revoked)
	RevocationTime time.Time `json:"revocation_time,omitempty" protobuf:"bytes,7,opt,name=revocation_time,proto3,stdtime"`

	// Revocable indicates if this attestation can be revoked
	Revocable bool `json:"revocable" protobuf:"varint,8,opt,name=revocable,proto3"`

	// RefUID is a reference to another attestation
	RefUID string `json:"ref_uid,omitempty" protobuf:"bytes,9,opt,name=ref_uid,proto3"`

	// Data contains the attestation data (encoded)
	Data []byte `json:"data" protobuf:"bytes,10,opt,name=data,proto3"`

	// AttestationType distinguishes public vs encrypted attestations
	AttestationType string `json:"attestation_type" protobuf:"bytes,11,opt,name=attestation_type,proto3"`
}

// Proto interface implementations for Attestation
func (a *Attestation) Reset()         { *a = Attestation{} }
func (a *Attestation) String() string { return a.UID }
func (a *Attestation) ProtoMessage()  {}

// XXX_MessageName returns the fully qualified protobuf message name
func (*Attestation) XXX_MessageName() string { return "cert.attestation.v1.Attestation" }

// EncryptedAttestation extends Attestation with encryption-specific fields
// Per Whitepaper Section 3.2 and 3.4
type EncryptedAttestation struct {
	Attestation

	// IPFSCID is the Content Identifier for the encrypted file on IPFS
	IPFSCID string `json:"ipfs_cid" protobuf:"bytes,12,opt,name=ipfs_cid,proto3"`

	// EncryptedDataHash is the SHA-256 hash of the encrypted ciphertext
	EncryptedDataHash string `json:"encrypted_data_hash" protobuf:"bytes,13,opt,name=encrypted_data_hash,proto3"`

	// Recipients is the list of authorized recipients
	Recipients []sdk.AccAddress `json:"recipients" protobuf:"bytes,14,rep,name=recipients,proto3"`

	// EncryptedSymmetricKeys contains ECIES-wrapped AES keys for each recipient
	// Map of recipient address -> wrapped key (hex encoded)
	EncryptedSymmetricKeys map[string]string `json:"encrypted_symmetric_keys" protobuf:"bytes,15,rep,name=encrypted_symmetric_keys,proto3"`
}

// Proto interface implementations for EncryptedAttestation
func (ea *EncryptedAttestation) Reset()         { *ea = EncryptedAttestation{} }
func (ea *EncryptedAttestation) String() string { return ea.UID }
func (ea *EncryptedAttestation) ProtoMessage()  {}

// XXX_MessageName returns the fully qualified protobuf message name
func (*EncryptedAttestation) XXX_MessageName() string {
	return "cert.attestation.v1.EncryptedAttestation"
}

// BusinessDocumentAttestation extends EncryptedAttestation for enterprise use
// Per Whitepaper Section 3.4 - EncryptedBusinessDocumentAttestation schema
type BusinessDocumentAttestation struct {
	EncryptedAttestation

	// BusinessID is the enterprise identifier
	BusinessID string `json:"business_id" protobuf:"bytes,16,opt,name=business_id,proto3"`

	// DocumentCategory classifies the document type
	DocumentCategory string `json:"document_category" protobuf:"bytes,17,opt,name=document_category,proto3"`

	// ValidUntil is the document validity expiration
	ValidUntil time.Time `json:"valid_until,omitempty" protobuf:"bytes,18,opt,name=valid_until,proto3,stdtime"`
}

// Proto interface implementations for BusinessDocumentAttestation
func (bda *BusinessDocumentAttestation) Reset()         { *bda = BusinessDocumentAttestation{} }
func (bda *BusinessDocumentAttestation) String() string { return bda.UID }
func (bda *BusinessDocumentAttestation) ProtoMessage()  {}

// XXX_MessageName returns the fully qualified protobuf message name
func (*BusinessDocumentAttestation) XXX_MessageName() string {
	return "cert.attestation.v1.BusinessDocumentAttestation"
}

// Schema represents an EAS schema definition
type Schema struct {
	// UID is the unique identifier for this schema
	UID string `json:"uid" protobuf:"bytes,1,opt,name=uid,proto3"`

	// Resolver is an optional resolver contract address
	Resolver sdk.AccAddress `json:"resolver,omitempty" protobuf:"bytes,2,opt,name=resolver,proto3"`

	// Revocable indicates if attestations using this schema can be revoked
	Revocable bool `json:"revocable" protobuf:"varint,3,opt,name=revocable,proto3"`

	// Schema is the ABI-encoded schema definition
	Schema string `json:"schema" protobuf:"bytes,4,opt,name=schema,proto3"`

	// Creator is the address that registered this schema
	Creator sdk.AccAddress `json:"creator" protobuf:"bytes,5,opt,name=creator,proto3"`
}

// Proto interface implementations for Schema
func (s *Schema) Reset()         { *s = Schema{} }
func (s *Schema) String() string { return s.UID }
func (s *Schema) ProtoMessage()  {}

// XXX_MessageName returns the fully qualified protobuf message name
func (*Schema) XXX_MessageName() string { return "cert.attestation.v1.Schema" }

// GenerateUID generates a unique identifier for an attestation
func GenerateUID(attester sdk.AccAddress, schemaUID string, timestamp time.Time, data []byte) string {
	combined := append(attester.Bytes(), []byte(schemaUID)...)
	combined = append(combined, []byte(timestamp.String())...)
	combined = append(combined, data...)
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:])
}

// GenerateSchemaUID generates a unique identifier for a schema
func GenerateSchemaUID(schema string, resolver sdk.AccAddress, revocable bool) string {
	combined := []byte(schema)
	combined = append(combined, resolver.Bytes()...)
	if revocable {
		combined = append(combined, 1)
	} else {
		combined = append(combined, 0)
	}
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:])
}

// Params defines the parameters for the attestation module
type Params struct {
	// MaxRecipientsPerAttestation is the maximum number of recipients per encrypted attestation
	MaxRecipientsPerAttestation uint32 `json:"max_recipients_per_attestation" protobuf:"varint,1,opt,name=max_recipients_per_attestation,proto3"`

	// MaxEncryptedFileSize is the maximum size in bytes for encrypted files
	MaxEncryptedFileSize uint64 `json:"max_encrypted_file_size" protobuf:"varint,2,opt,name=max_encrypted_file_size,proto3"`

	// AttestationFee is the fee for creating an attestation (optional)
	AttestationFee sdk.Coins `json:"attestation_fee" protobuf:"bytes,3,rep,name=attestation_fee,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins"`
}

// Proto interface implementations for Params
func (p *Params) Reset()         { *p = Params{} }
func (p *Params) String() string { return "Params" }
func (p *Params) ProtoMessage()  {}

// XXX_MessageName returns the fully qualified protobuf message name for TypeURL registration
func (*Params) XXX_MessageName() string { return "cert.attestation.v1.Params" }

// DefaultParams returns default module parameters per Whitepaper Section 12
func DefaultParams() Params {
	return Params{
		MaxRecipientsPerAttestation: 50,                // Whitepaper Section 12
		MaxEncryptedFileSize:        100 * 1024 * 1024, // 100 MB - Whitepaper Section 12
		AttestationFee:              sdk.NewCoins(),    // No fee by default
	}
}
