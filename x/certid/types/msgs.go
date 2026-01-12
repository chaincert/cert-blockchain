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
	TypeMsgAwardBadge          = "award_badge"
	TypeMsgRevokeBadge         = "revoke_badge"
	TypeMsgUpdateTrustScore    = "update_trust_score"
	TypeMsgSetVerification     = "set_verification"
	TypeMsgAuthorizeOracle     = "authorize_oracle"
	TypeMsgRevokeOracle        = "revoke_oracle"
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

// MsgAwardBadge awards a soulbound badge to a user
type MsgAwardBadge struct {
	Authority   string `json:"authority" protobuf:"bytes,1,opt,name=authority,proto3"`
	User        string `json:"user" protobuf:"bytes,2,opt,name=user,proto3"`
	BadgeName   string `json:"badgeName" protobuf:"bytes,3,opt,name=badge_name,proto3"`
	Description string `json:"description,omitempty" protobuf:"bytes,4,opt,name=description,proto3"`
}

func (msg *MsgAwardBadge) Reset()         { *msg = MsgAwardBadge{} }
func (msg *MsgAwardBadge) String() string { return msg.Authority }
func (msg *MsgAwardBadge) ProtoMessage()  {}
func (msg *MsgAwardBadge) Route() string  { return RouterKey }
func (msg *MsgAwardBadge) Type() string   { return TypeMsgAwardBadge }

func (msg *MsgAwardBadge) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

func (msg *MsgAwardBadge) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.User); err != nil {
		return err
	}
	if msg.BadgeName == "" {
		return ErrInvalidBadgeName
	}
	return nil
}

// MsgRevokeBadge revokes a badge from a user
type MsgRevokeBadge struct {
	Authority string `json:"authority" protobuf:"bytes,1,opt,name=authority,proto3"`
	User      string `json:"user" protobuf:"bytes,2,opt,name=user,proto3"`
	BadgeName string `json:"badgeName" protobuf:"bytes,3,opt,name=badge_name,proto3"`
}

func (msg *MsgRevokeBadge) Reset()         { *msg = MsgRevokeBadge{} }
func (msg *MsgRevokeBadge) String() string { return msg.Authority }
func (msg *MsgRevokeBadge) ProtoMessage()  {}
func (msg *MsgRevokeBadge) Route() string  { return RouterKey }
func (msg *MsgRevokeBadge) Type() string   { return TypeMsgRevokeBadge }

func (msg *MsgRevokeBadge) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

func (msg *MsgRevokeBadge) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.User); err != nil {
		return err
	}
	if msg.BadgeName == "" {
		return ErrInvalidBadgeName
	}
	return nil
}

// MsgUpdateTrustScore updates a user's trust score
type MsgUpdateTrustScore struct {
	Authority string `json:"authority" protobuf:"bytes,1,opt,name=authority,proto3"`
	User      string `json:"user" protobuf:"bytes,2,opt,name=user,proto3"`
	Score     uint64 `json:"score" protobuf:"varint,3,opt,name=score,proto3"`
	Reason    string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason,proto3"`
}

func (msg *MsgUpdateTrustScore) Reset()         { *msg = MsgUpdateTrustScore{} }
func (msg *MsgUpdateTrustScore) String() string { return msg.Authority }
func (msg *MsgUpdateTrustScore) ProtoMessage()  {}
func (msg *MsgUpdateTrustScore) Route() string  { return RouterKey }
func (msg *MsgUpdateTrustScore) Type() string   { return TypeMsgUpdateTrustScore }

func (msg *MsgUpdateTrustScore) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

func (msg *MsgUpdateTrustScore) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.User); err != nil {
		return err
	}
	if msg.Score > 100 {
		return ErrInvalidTrustScore
	}
	return nil
}

// MsgSetVerificationStatus sets verification status for a profile
type MsgSetVerificationStatus struct {
	Authority  string `json:"authority" protobuf:"bytes,1,opt,name=authority,proto3"`
	User       string `json:"user" protobuf:"bytes,2,opt,name=user,proto3"`
	IsVerified bool   `json:"isVerified" protobuf:"varint,3,opt,name=is_verified,proto3"`
}

func (msg *MsgSetVerificationStatus) Reset()         { *msg = MsgSetVerificationStatus{} }
func (msg *MsgSetVerificationStatus) String() string { return msg.Authority }
func (msg *MsgSetVerificationStatus) ProtoMessage()  {}
func (msg *MsgSetVerificationStatus) Route() string  { return RouterKey }
func (msg *MsgSetVerificationStatus) Type() string   { return TypeMsgSetVerification }

func (msg *MsgSetVerificationStatus) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

func (msg *MsgSetVerificationStatus) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.User); err != nil {
		return err
	}
	return nil
}

// MsgRemoveCredential removes a credential from a CertID profile
type MsgRemoveCredential struct {
	Creator        string `json:"creator" protobuf:"bytes,1,opt,name=creator,proto3"`
	AttestationUID string `json:"attestationUid" protobuf:"bytes,2,opt,name=attestation_uid,proto3"`
}

func (msg *MsgRemoveCredential) Reset()         { *msg = MsgRemoveCredential{} }
func (msg *MsgRemoveCredential) String() string { return msg.Creator }
func (msg *MsgRemoveCredential) ProtoMessage()  {}
func (msg *MsgRemoveCredential) Route() string  { return RouterKey }
func (msg *MsgRemoveCredential) Type() string   { return TypeMsgRemoveCredential }

func (msg *MsgRemoveCredential) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

func (msg *MsgRemoveCredential) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return err
	}
	if msg.AttestationUID == "" {
		return ErrInvalidAttestationUID
	}
	return nil
}

// MsgRequestVerification requests verification for a profile
type MsgRequestVerification struct {
	Creator          string `json:"creator" protobuf:"bytes,1,opt,name=creator,proto3"`
	VerificationType string `json:"verificationType" protobuf:"bytes,2,opt,name=verification_type,proto3"`
	VerificationData string `json:"verificationData" protobuf:"bytes,3,opt,name=verification_data,proto3"`
}

func (msg *MsgRequestVerification) Reset()         { *msg = MsgRequestVerification{} }
func (msg *MsgRequestVerification) String() string { return msg.Creator }
func (msg *MsgRequestVerification) ProtoMessage()  {}
func (msg *MsgRequestVerification) Route() string  { return RouterKey }
func (msg *MsgRequestVerification) Type() string   { return TypeMsgRequestVerification }

func (msg *MsgRequestVerification) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

func (msg *MsgRequestVerification) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return err
	}
	if msg.VerificationType == "" {
		return ErrInvalidSocialVerification.Wrap("verification type required")
	}
	return nil
}

// MsgAuthorizeOracle authorizes an oracle to perform trust score updates
type MsgAuthorizeOracle struct {
	Authority string `json:"authority" protobuf:"bytes,1,opt,name=authority,proto3"`
	Oracle    string `json:"oracle" protobuf:"bytes,2,opt,name=oracle,proto3"`
}

func (msg *MsgAuthorizeOracle) Reset()         { *msg = MsgAuthorizeOracle{} }
func (msg *MsgAuthorizeOracle) String() string { return msg.Authority }
func (msg *MsgAuthorizeOracle) ProtoMessage()  {}
func (msg *MsgAuthorizeOracle) Route() string  { return RouterKey }
func (msg *MsgAuthorizeOracle) Type() string   { return TypeMsgAuthorizeOracle }

func (msg *MsgAuthorizeOracle) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

func (msg *MsgAuthorizeOracle) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.Oracle); err != nil {
		return err
	}
	return nil
}

// MsgRevokeOracle revokes an oracle's authorization
type MsgRevokeOracle struct {
	Authority string `json:"authority" protobuf:"bytes,1,opt,name=authority,proto3"`
	Oracle    string `json:"oracle" protobuf:"bytes,2,opt,name=oracle,proto3"`
}

func (msg *MsgRevokeOracle) Reset()         { *msg = MsgRevokeOracle{} }
func (msg *MsgRevokeOracle) String() string { return msg.Authority }
func (msg *MsgRevokeOracle) ProtoMessage()  {}
func (msg *MsgRevokeOracle) Route() string  { return RouterKey }
func (msg *MsgRevokeOracle) Type() string   { return TypeMsgRevokeOracle }

func (msg *MsgRevokeOracle) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

func (msg *MsgRevokeOracle) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.Oracle); err != nil {
		return err
	}
	return nil
}

// MsgRegisterHandle registers a unique handle for a profile
type MsgRegisterHandle struct {
	Creator string `json:"creator" protobuf:"bytes,1,opt,name=creator,proto3"`
	Handle  string `json:"handle" protobuf:"bytes,2,opt,name=handle,proto3"`
}

func (msg *MsgRegisterHandle) Reset()         { *msg = MsgRegisterHandle{} }
func (msg *MsgRegisterHandle) String() string { return msg.Creator }
func (msg *MsgRegisterHandle) ProtoMessage()  {}
func (msg *MsgRegisterHandle) Route() string  { return RouterKey }
func (msg *MsgRegisterHandle) Type() string   { return "register_handle" }

func (msg *MsgRegisterHandle) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

func (msg *MsgRegisterHandle) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return err
	}
	if msg.Handle == "" {
		return ErrInvalidUsername.Wrap("handle cannot be empty")
	}
	return nil
}
