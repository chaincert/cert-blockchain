package attestation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/chaincertify/certd/x/attestation/keeper"
	"github.com/chaincertify/certd/x/attestation/types"
)

// GenesisState defines the attestation module's genesis state
type GenesisState struct {
	// Params defines module parameters
	Params types.Params `json:"params" protobuf:"bytes,1,opt,name=params,proto3"`

	// Schemas contains pre-deployed schemas
	Schemas []types.Schema `json:"schemas" protobuf:"bytes,2,rep,name=schemas,proto3"`

	// Attestations contains any genesis attestations
	Attestations []types.Attestation `json:"attestations" protobuf:"bytes,3,rep,name=attestations,proto3"`

	// EncryptedAttestations contains any genesis encrypted attestations
	EncryptedAttestations []types.EncryptedAttestation `json:"encrypted_attestations" protobuf:"bytes,4,rep,name=encrypted_attestations,proto3"`
}

// Proto interface implementations for GenesisState
func (gs *GenesisState) Reset()         { *gs = GenesisState{} }
func (gs *GenesisState) String() string { return "GenesisState" }
func (gs *GenesisState) ProtoMessage()  {}

// XXX_MessageName returns the fully qualified protobuf message name for TypeURL registration
func (*GenesisState) XXX_MessageName() string { return "cert.attestation.v1.GenesisState" }

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:                types.DefaultParams(),
		Schemas:               GetDefaultSchemas(),
		Attestations:          []types.Attestation{},
		EncryptedAttestations: []types.EncryptedAttestation{},
	}
}

// GetDefaultSchemas returns the pre-deployed EAS schemas per Whitepaper Section 3.4
func GetDefaultSchemas() []types.Schema {
	return []types.Schema{
		{
			// EncryptedFileAttestation schema per Whitepaper Section 3.4
			UID:       "0x1", // Will be generated properly
			Revocable: true,
			Schema:    "string ipfsCID, bytes32 encryptedDataHash, address recipient, bytes encryptedSymmetricKey, uint256 timestamp",
		},
		{
			// EncryptedMultiRecipientAttestation schema per Whitepaper Section 3.4
			UID:       "0x2",
			Revocable: true,
			Schema:    "string ipfsCID, bytes32 encryptedDataHash, address[] recipients, bytes[] encryptedSymmetricKeys, bool revocable",
		},
		{
			// EncryptedBusinessDocumentAttestation schema per Whitepaper Section 3.4
			UID:       "0x3",
			Revocable: true,
			Schema:    "string ipfsCID, bytes32 encryptedDataHash, address[] recipients, bytes[] encryptedSymmetricKeys, string businessID, string documentCategory, uint256 validUntil",
		},
		{
			// Public attestation schema
			UID:       "0x4",
			Revocable: true,
			Schema:    "bytes32 dataHash, string metadata, uint256 timestamp",
		},
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	// Validate params
	if gs.Params.MaxRecipientsPerAttestation == 0 {
		return types.ErrInvalidParams
	}
	if gs.Params.MaxEncryptedFileSize == 0 {
		return types.ErrInvalidParams
	}

	// Validate schemas
	schemaUIDs := make(map[string]bool)
	for _, schema := range gs.Schemas {
		if schemaUIDs[schema.UID] {
			return types.ErrDuplicateSchema
		}
		schemaUIDs[schema.UID] = true
	}

	return nil
}

// InitGenesis initializes the attestation module's state from a genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState GenesisState) {
	// Set params
	k.SetParams(ctx, genState.Params)

	// Register default schemas
	for _, schema := range genState.Schemas {
		if schema.Creator == nil {
			// Genesis schemas created by module account
			schema.Creator = sdk.AccAddress{}
		}
		k.RegisterSchema(ctx, schema.Creator, schema.Schema, schema.Resolver, schema.Revocable)
	}

	// Import any genesis attestations
	for _, attestation := range genState.Attestations {
		k.ImportAttestation(ctx, attestation)
	}

	// Import encrypted attestations
	for _, encAttestation := range genState.EncryptedAttestations {
		k.ImportEncryptedAttestation(ctx, encAttestation)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *GenesisState {
	return &GenesisState{
		Params:                k.GetParams(ctx),
		Schemas:               k.GetAllSchemas(ctx),
		Attestations:          k.GetAllAttestations(ctx),
		EncryptedAttestations: k.GetAllEncryptedAttestations(ctx),
	}
}
