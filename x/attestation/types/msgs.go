package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgRegisterSchema             = "register_schema"
	TypeMsgAttest                     = "attest"
	TypeMsgRevoke                     = "revoke"
	TypeMsgCreateEncryptedAttestation = "create_encrypted_attestation"
)

// MsgRegisterSchema registers a new attestation schema
type MsgRegisterSchema struct {
	Creator   string `json:"creator" protobuf:"bytes,1,opt,name=creator,proto3"`
	Schema    string `json:"schema" protobuf:"bytes,2,opt,name=schema,proto3"`
	Resolver  string `json:"resolver,omitempty" protobuf:"bytes,3,opt,name=resolver,proto3"`
	Revocable bool   `json:"revocable" protobuf:"varint,4,opt,name=revocable,proto3"`
}

// Proto interface implementations
func (msg *MsgRegisterSchema) Reset()         { *msg = MsgRegisterSchema{} }
func (msg *MsgRegisterSchema) String() string { return msg.Creator }
func (msg *MsgRegisterSchema) ProtoMessage()  {}

// XXX_MessageName returns the fully qualified protobuf message name for TypeURL registration
func (*MsgRegisterSchema) XXX_MessageName() string { return "cert.attestation.v1.MsgRegisterSchema" }

func NewMsgRegisterSchema(creator, schema, resolver string, revocable bool) *MsgRegisterSchema {
	return &MsgRegisterSchema{
		Creator:   creator,
		Schema:    schema,
		Resolver:  resolver,
		Revocable: revocable,
	}
}

func (msg MsgRegisterSchema) Route() string { return RouterKey }
func (msg MsgRegisterSchema) Type() string  { return TypeMsgRegisterSchema }

func (msg MsgRegisterSchema) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errors.New("invalid creator address")
	}
	if msg.Schema == "" {
		return errors.New("schema cannot be empty")
	}
	return nil
}

func (msg MsgRegisterSchema) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

// MsgAttest creates a new public attestation
type MsgAttest struct {
	Attester       string `json:"attester" protobuf:"bytes,1,opt,name=attester,proto3"`
	SchemaUID      string `json:"schema_uid" protobuf:"bytes,2,opt,name=schema_uid,proto3"`
	Recipient      string `json:"recipient,omitempty" protobuf:"bytes,3,opt,name=recipient,proto3"`
	ExpirationTime int64  `json:"expiration_time,omitempty" protobuf:"varint,4,opt,name=expiration_time,proto3"`
	Revocable      bool   `json:"revocable" protobuf:"varint,5,opt,name=revocable,proto3"`
	RefUID         string `json:"ref_uid,omitempty" protobuf:"bytes,6,opt,name=ref_uid,proto3"`
	Data           []byte `json:"data" protobuf:"bytes,7,opt,name=data,proto3"`
}

// Proto interface implementations
func (msg *MsgAttest) Reset()         { *msg = MsgAttest{} }
func (msg *MsgAttest) String() string { return msg.Attester }
func (msg *MsgAttest) ProtoMessage()  {}

// XXX_MessageName returns the fully qualified protobuf message name for TypeURL registration
func (*MsgAttest) XXX_MessageName() string { return "cert.attestation.v1.MsgAttest" }

func NewMsgAttest(attester, schemaUID, recipient string, expirationTime int64, revocable bool, refUID string, data []byte) *MsgAttest {
	return &MsgAttest{
		Attester:       attester,
		SchemaUID:      schemaUID,
		Recipient:      recipient,
		ExpirationTime: expirationTime,
		Revocable:      revocable,
		RefUID:         refUID,
		Data:           data,
	}
}

func (msg MsgAttest) Route() string { return RouterKey }
func (msg MsgAttest) Type() string  { return TypeMsgAttest }

func (msg MsgAttest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Attester)
	if err != nil {
		return errors.New("invalid attester address")
	}
	if msg.SchemaUID == "" {
		return errors.New("schema UID cannot be empty")
	}
	return nil
}

func (msg MsgAttest) GetSigners() []sdk.AccAddress {
	attester, _ := sdk.AccAddressFromBech32(msg.Attester)
	return []sdk.AccAddress{attester}
}

// MsgRevoke revokes an existing attestation
type MsgRevoke struct {
	Revoker string `json:"revoker" protobuf:"bytes,1,opt,name=revoker,proto3"`
	UID     string `json:"uid" protobuf:"bytes,2,opt,name=uid,proto3"`
}

// Proto interface implementations
func (msg *MsgRevoke) Reset()         { *msg = MsgRevoke{} }
func (msg *MsgRevoke) String() string { return msg.Revoker }
func (msg *MsgRevoke) ProtoMessage()  {}

// XXX_MessageName returns the fully qualified protobuf message name for TypeURL registration
func (*MsgRevoke) XXX_MessageName() string { return "cert.attestation.v1.MsgRevoke" }

func NewMsgRevoke(revoker, uid string) *MsgRevoke {
	return &MsgRevoke{
		Revoker: revoker,
		UID:     uid,
	}
}

func (msg MsgRevoke) Route() string { return RouterKey }
func (msg MsgRevoke) Type() string  { return TypeMsgRevoke }

func (msg MsgRevoke) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Revoker)
	if err != nil {
		return errors.New("invalid revoker address")
	}
	if msg.UID == "" {
		return errors.New("attestation UID cannot be empty")
	}
	return nil
}

func (msg MsgRevoke) GetSigners() []sdk.AccAddress {
	revoker, _ := sdk.AccAddressFromBech32(msg.Revoker)
	return []sdk.AccAddress{revoker}
}

// MsgCreateEncryptedAttestation creates a new encrypted attestation
// Per Whitepaper Section 3.2 - Step 4: On-Chain Anchoring
type MsgCreateEncryptedAttestation struct {
	Attester               string            `json:"attester" protobuf:"bytes,1,opt,name=attester,proto3"`
	SchemaUID              string            `json:"schema_uid" protobuf:"bytes,2,opt,name=schema_uid,proto3"`
	IPFSCID                string            `json:"ipfs_cid" protobuf:"bytes,3,opt,name=ipfs_cid,proto3"`
	EncryptedDataHash      string            `json:"encrypted_data_hash" protobuf:"bytes,4,opt,name=encrypted_data_hash,proto3"`
	Recipients             []string          `json:"recipients" protobuf:"bytes,5,rep,name=recipients,proto3"`
	EncryptedSymmetricKeys map[string]string `json:"encrypted_symmetric_keys" protobuf:"bytes,6,rep,name=encrypted_symmetric_keys,proto3"`
	Revocable              bool              `json:"revocable" protobuf:"varint,7,opt,name=revocable,proto3"`
	ExpirationTime         int64             `json:"expiration_time,omitempty" protobuf:"varint,8,opt,name=expiration_time,proto3"`
}

// Proto interface implementations
func (msg *MsgCreateEncryptedAttestation) Reset()         { *msg = MsgCreateEncryptedAttestation{} }
func (msg *MsgCreateEncryptedAttestation) String() string { return msg.Attester }
func (msg *MsgCreateEncryptedAttestation) ProtoMessage()  {}

// XXX_MessageName returns the fully qualified protobuf message name for TypeURL registration
func (*MsgCreateEncryptedAttestation) XXX_MessageName() string {
	return "cert.attestation.v1.MsgCreateEncryptedAttestation"
}

func NewMsgCreateEncryptedAttestation(
	attester, schemaUID, ipfsCID, encryptedDataHash string,
	recipients []string,
	encryptedSymmetricKeys map[string]string,
	revocable bool,
	expirationTime int64,
) *MsgCreateEncryptedAttestation {
	return &MsgCreateEncryptedAttestation{
		Attester:               attester,
		SchemaUID:              schemaUID,
		IPFSCID:                ipfsCID,
		EncryptedDataHash:      encryptedDataHash,
		Recipients:             recipients,
		EncryptedSymmetricKeys: encryptedSymmetricKeys,
		Revocable:              revocable,
		ExpirationTime:         expirationTime,
	}
}

func (msg MsgCreateEncryptedAttestation) Route() string { return RouterKey }
func (msg MsgCreateEncryptedAttestation) Type() string  { return TypeMsgCreateEncryptedAttestation }

func (msg MsgCreateEncryptedAttestation) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Attester)
	if err != nil {
		return errors.New("invalid attester address")
	}
	if msg.SchemaUID == "" {
		return errors.New("schema UID cannot be empty")
	}
	if msg.IPFSCID == "" {
		return errors.New("IPFS CID cannot be empty")
	}
	if msg.EncryptedDataHash == "" {
		return errors.New("encrypted data hash cannot be empty")
	}
	if len(msg.Recipients) == 0 {
		return errors.New("at least one recipient is required")
	}
	// Whitepaper Section 12: Max 50 recipients per attestation
	if len(msg.Recipients) > 50 {
		return errors.New("maximum 50 recipients allowed per attestation")
	}
	// Verify each recipient has an encrypted key
	for _, recipient := range msg.Recipients {
		_, err := sdk.AccAddressFromBech32(recipient)
		if err != nil {
			return errors.New("invalid recipient address: " + recipient)
		}
		if _, ok := msg.EncryptedSymmetricKeys[recipient]; !ok {
			return errors.New("missing encrypted key for recipient: " + recipient)
		}
	}
	return nil
}

func (msg MsgCreateEncryptedAttestation) GetSigners() []sdk.AccAddress {
	attester, _ := sdk.AccAddressFromBech32(msg.Attester)
	return []sdk.AccAddress{attester}
}
