package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message types for the hardware module
const (
	TypeMsgRegisterDevice    = "register_device"
	TypeMsgSubmitAttestation = "submit_attestation"
	TypeMsgLinkDeviceToCertID = "link_device_to_certid"
	TypeMsgSuspendDevice     = "suspend_device"
	TypeMsgReactivateDevice  = "reactivate_device"
)

// MsgRegisterDevice registers a new hardware device
type MsgRegisterDevice struct {
	// Creator is the address registering the device
	Creator string `json:"creator"`

	// Manufacturer of the device
	Manufacturer string `json:"manufacturer"`

	// Model of the device (optional)
	Model string `json:"model,omitempty"`

	// TEEType specifies the TEE type (ARM_TRUSTZONE, APPLE_SECURE_ENCLAVE)
	TEEType TEEType `json:"tee_type"`

	// PublicKey is the hardware-bound public key from TEE
	PublicKey []byte `json:"public_key"`

	// InitialAttestation is the attestation proof for registration
	InitialAttestation []byte `json:"initial_attestation"`
}

// Route implements sdk.Msg
func (msg MsgRegisterDevice) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgRegisterDevice) Type() string { return TypeMsgRegisterDevice }

// ValidateBasic implements sdk.Msg
func (msg MsgRegisterDevice) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return ErrInvalidAddress.Wrap("invalid creator address")
	}

	if msg.Manufacturer == "" {
		return ErrInvalidDevice.Wrap("manufacturer cannot be empty")
	}

	if !ValidateTEEType(msg.TEEType) {
		return ErrUnsupportedTEE.Wrapf("unsupported TEE type: %s", msg.TEEType)
	}

	if len(msg.PublicKey) == 0 {
		return ErrInvalidDevice.Wrap("public key cannot be empty")
	}

	if len(msg.InitialAttestation) == 0 {
		return ErrInvalidAttestation.Wrap("initial attestation required")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (msg MsgRegisterDevice) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

// MsgSubmitAttestation submits a TEE attestation for verification
type MsgSubmitAttestation struct {
	// Submitter is the device owner submitting attestation
	Submitter string `json:"submitter"`

	// DeviceID identifies the attesting device
	DeviceID string `json:"device_id"`

	// AttestationData is the raw TEE attestation blob
	AttestationData []byte `json:"attestation_data"`

	// Nonce is the challenge nonce (if challenge-response)
	Nonce []byte `json:"nonce,omitempty"`

	// AttestationType indicates attestation context
	AttestationType AttestationType `json:"attestation_type"`
}

// Route implements sdk.Msg
func (msg MsgSubmitAttestation) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgSubmitAttestation) Type() string { return TypeMsgSubmitAttestation }

// ValidateBasic implements sdk.Msg
func (msg MsgSubmitAttestation) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil {
		return ErrInvalidAddress.Wrap("invalid submitter address")
	}

	if msg.DeviceID == "" {
		return ErrInvalidDevice.Wrap("device ID cannot be empty")
	}

	if len(msg.AttestationData) == 0 {
		return ErrInvalidAttestation.Wrap("attestation data cannot be empty")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (msg MsgSubmitAttestation) GetSigners() []sdk.AccAddress {
	submitter, _ := sdk.AccAddressFromBech32(msg.Submitter)
	return []sdk.AccAddress{submitter}
}

// MsgLinkDeviceToCertID links a verified device to a CertID profile
type MsgLinkDeviceToCertID struct {
	// Owner is the CertID profile owner
	Owner string `json:"owner"`

	// DeviceID is the device to link
	DeviceID string `json:"device_id"`

	// CertIDAddress is the CertID profile address
	CertIDAddress string `json:"certid_address"`
}

// Route implements sdk.Msg
func (msg MsgLinkDeviceToCertID) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgLinkDeviceToCertID) Type() string { return TypeMsgLinkDeviceToCertID }

// ValidateBasic implements sdk.Msg
func (msg MsgLinkDeviceToCertID) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return ErrInvalidAddress.Wrap("invalid owner address")
	}

	if msg.DeviceID == "" {
		return ErrInvalidDevice.Wrap("device ID cannot be empty")
	}

	_, err = sdk.AccAddressFromBech32(msg.CertIDAddress)
	if err != nil {
		return ErrInvalidAddress.Wrap("invalid CertID address")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (msg MsgLinkDeviceToCertID) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{owner}
}

// MsgSuspendDevice suspends a device for suspicious activity
type MsgSuspendDevice struct {
	// Authority is the authorized oracle/admin
	Authority string `json:"authority"`

	// DeviceID is the device to suspend
	DeviceID string `json:"device_id"`

	// Reason for suspension
	Reason string `json:"reason"`
}

// Route implements sdk.Msg
func (msg MsgSuspendDevice) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgSuspendDevice) Type() string { return TypeMsgSuspendDevice }

// ValidateBasic implements sdk.Msg
func (msg MsgSuspendDevice) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}

	if msg.DeviceID == "" {
		return ErrInvalidDevice.Wrap("device ID cannot be empty")
	}

	if msg.Reason == "" {
		return ErrInvalidDevice.Wrap("suspension reason required")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (msg MsgSuspendDevice) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}
