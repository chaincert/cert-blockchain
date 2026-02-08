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
// MsgCreateProfile is defined in tx.pb.go

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
// MsgUpdateProfile is defined in tx.pb.go

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
// MsgAddCredential is defined in tx.pb.go

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
// MsgVerifySocial is defined in tx.pb.go

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
// MsgAwardBadge is defined in tx.pb.go
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
// MsgRevokeBadge is defined in tx.pb.go
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
// MsgUpdateTrustScore is defined in tx.pb.go
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
// MsgSetVerificationStatus is defined in tx.pb.go
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
// MsgRemoveCredential is defined in tx.pb.go
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
// MsgRequestVerification is defined in tx.pb.go
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
// MsgAuthorizeOracle is defined in tx.pb.go
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
// MsgRevokeOracle is defined in tx.pb.go
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
// MsgRegisterHandle is defined in tx.pb.go
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
