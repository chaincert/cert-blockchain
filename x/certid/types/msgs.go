package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgCreateProfile       = "create_profile"
	TypeMsgUpdateProfile       = "update_profile"
	TypeMsgAddCredential       = "add_credential"
	TypeMsgRemoveCredential    = "remove_credential"
	TypeMsgVerifySocial        = "verify_social"
	TypeMsgRequestVerification = "request_verification"
)

// MsgCreateProfile creates a new CertID profile
type MsgCreateProfile struct {
	Creator     string            `json:"creator" protobuf:"bytes,1,opt,name=creator,proto3"`
	Name        string            `json:"name,omitempty" protobuf:"bytes,2,opt,name=name,proto3"`
	Bio         string            `json:"bio,omitempty" protobuf:"bytes,3,opt,name=bio,proto3"`
	AvatarCID   string            `json:"avatarCid,omitempty" protobuf:"bytes,4,opt,name=avatar_cid,proto3"`
	PublicKey   string            `json:"publicKey,omitempty" protobuf:"bytes,5,opt,name=public_key,proto3"`
	SocialLinks map[string]string `json:"socialLinks,omitempty" protobuf:"bytes,6,rep,name=social_links,proto3"`
}

// Proto interface implementations
func (msg *MsgCreateProfile) Reset()         { *msg = MsgCreateProfile{} }
func (msg *MsgCreateProfile) String() string { return msg.Creator }
func (msg *MsgCreateProfile) ProtoMessage()  {}

func NewMsgCreateProfile(creator, name, bio, avatarCID, publicKey string) *MsgCreateProfile {
	return &MsgCreateProfile{
		Creator:     creator,
		Name:        name,
		Bio:         bio,
		AvatarCID:   avatarCID,
		PublicKey:   publicKey,
		SocialLinks: make(map[string]string),
	}
}

func (msg *MsgCreateProfile) Route() string { return RouterKey }
func (msg *MsgCreateProfile) Type() string  { return TypeMsgCreateProfile }

func (msg *MsgCreateProfile) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateProfile) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return err
	}
	if len(msg.Name) > 100 {
		return ErrInvalidName
	}
	if len(msg.Bio) > 500 {
		return ErrInvalidBio
	}
	return nil
}

// MsgUpdateProfile updates an existing CertID profile
type MsgUpdateProfile struct {
	Creator     string            `json:"creator" protobuf:"bytes,1,opt,name=creator,proto3"`
	Name        string            `json:"name,omitempty" protobuf:"bytes,2,opt,name=name,proto3"`
	Bio         string            `json:"bio,omitempty" protobuf:"bytes,3,opt,name=bio,proto3"`
	AvatarCID   string            `json:"avatarCid,omitempty" protobuf:"bytes,4,opt,name=avatar_cid,proto3"`
	PublicKey   string            `json:"publicKey,omitempty" protobuf:"bytes,5,opt,name=public_key,proto3"`
	SocialLinks map[string]string `json:"socialLinks,omitempty" protobuf:"bytes,6,rep,name=social_links,proto3"`
}

// Proto interface implementations
func (msg *MsgUpdateProfile) Reset()         { *msg = MsgUpdateProfile{} }
func (msg *MsgUpdateProfile) String() string { return msg.Creator }
func (msg *MsgUpdateProfile) ProtoMessage()  {}

func NewMsgUpdateProfile(creator string) *MsgUpdateProfile {
	return &MsgUpdateProfile{
		Creator:     creator,
		SocialLinks: make(map[string]string),
	}
}

func (msg *MsgUpdateProfile) Route() string { return RouterKey }
func (msg *MsgUpdateProfile) Type() string  { return TypeMsgUpdateProfile }

func (msg *MsgUpdateProfile) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateProfile) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return err
	}
	return nil
}

// MsgAddCredential adds a credential to a CertID profile
type MsgAddCredential struct {
	Creator        string `json:"creator" protobuf:"bytes,1,opt,name=creator,proto3"`
	AttestationUID string `json:"attestationUid" protobuf:"bytes,2,opt,name=attestation_uid,proto3"`
	CredentialType string `json:"credentialType" protobuf:"bytes,3,opt,name=credential_type,proto3"`
	Title          string `json:"title" protobuf:"bytes,4,opt,name=title,proto3"`
}

// Proto interface implementations
func (msg *MsgAddCredential) Reset()         { *msg = MsgAddCredential{} }
func (msg *MsgAddCredential) String() string { return msg.Creator }
func (msg *MsgAddCredential) ProtoMessage()  {}

func (msg *MsgAddCredential) Route() string { return RouterKey }
func (msg *MsgAddCredential) Type() string  { return TypeMsgAddCredential }

func (msg *MsgAddCredential) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

func (msg *MsgAddCredential) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return err
	}
	if msg.AttestationUID == "" {
		return ErrInvalidAttestationUID
	}
	return nil
}

// MsgVerifySocial verifies a social media account
type MsgVerifySocial struct {
	Creator  string `json:"creator" protobuf:"bytes,1,opt,name=creator,proto3"`
	Platform string `json:"platform" protobuf:"bytes,2,opt,name=platform,proto3"`
	Handle   string `json:"handle" protobuf:"bytes,3,opt,name=handle,proto3"`
	Proof    string `json:"proof" protobuf:"bytes,4,opt,name=proof,proto3"`
}

// Proto interface implementations
func (msg *MsgVerifySocial) Reset()         { *msg = MsgVerifySocial{} }
func (msg *MsgVerifySocial) String() string { return msg.Creator }
func (msg *MsgVerifySocial) ProtoMessage()  {}

func (msg *MsgVerifySocial) Route() string { return RouterKey }
func (msg *MsgVerifySocial) Type() string  { return TypeMsgVerifySocial }

func (msg *MsgVerifySocial) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

func (msg *MsgVerifySocial) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return err
	}
	if msg.Platform == "" || msg.Handle == "" {
		return ErrInvalidSocialVerification
	}
	return nil
}
