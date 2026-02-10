package types

// Re-export scoring types from the hardware module for backward compatibility.
// The canonical scoring algorithm types live in x/hardware/types because
// the scoring logic was originally part of the hardware module.
//
// This package adds trustscore-specific message types.

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgCalculateHumanityScore triggers a humanity score calculation for an address
type MsgCalculateHumanityScore struct {
	// Creator is the address requesting the calculation
	Creator string `json:"creator"`

	// TargetAddress is the address to calculate the humanity score for
	TargetAddress string `json:"target_address"`
}

// Route implements sdk.Msg
func (msg MsgCalculateHumanityScore) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgCalculateHumanityScore) Type() string { return "calculate_humanity_score" }

// ValidateBasic implements sdk.Msg
func (msg MsgCalculateHumanityScore) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return ErrInvalidAddress.Wrap("invalid creator address")
	}

	_, err = sdk.AccAddressFromBech32(msg.TargetAddress)
	if err != nil {
		return ErrInvalidAddress.Wrap("invalid target address")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (msg MsgCalculateHumanityScore) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}
