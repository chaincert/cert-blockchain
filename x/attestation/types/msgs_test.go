package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/chaincertify/certd/x/attestation/types"
)

// Helper function to create a valid test address
func createTestAddress(_ string) string {
	// Create a valid bech32 address for testing
	addr := sdk.AccAddress("testaddr1234567890ab")
	return addr.String()
}

func TestMsgRegisterSchema_ValidateBasic(t *testing.T) {
	// Initialize SDK config for cert prefix
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("cert", "certpub")

	validAddr := createTestAddress("cert")

	testCases := []struct {
		name      string
		msg       *types.MsgRegisterSchema
		expectErr bool
	}{
		{
			name: "valid message",
			msg: types.NewMsgRegisterSchema(
				validAddr,
				"string name, uint256 age",
				"",
				true,
			),
			expectErr: false,
		},
		{
			name: "empty creator",
			msg: types.NewMsgRegisterSchema(
				"",
				"string name, uint256 age",
				"",
				true,
			),
			expectErr: true,
		},
		{
			name: "empty schema",
			msg: types.NewMsgRegisterSchema(
				validAddr,
				"",
				"",
				true,
			),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgAttest_ValidateBasic(t *testing.T) {
	// Initialize SDK config for cert prefix
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("cert", "certpub")

	validAddr := createTestAddress("cert")
	validRecipient := sdk.AccAddress("recipient123456789a").String()

	testCases := []struct {
		name      string
		msg       *types.MsgAttest
		expectErr bool
	}{
		{
			name: "valid message",
			msg: types.NewMsgAttest(
				validAddr,
				"0x1234567890abcdef",
				validRecipient,
				0,
				true,
				"",
				[]byte(`{"test":"data"}`),
			),
			expectErr: false,
		},
		{
			name: "empty attester",
			msg: types.NewMsgAttest(
				"",
				"0x1234567890abcdef",
				validRecipient,
				0,
				true,
				"",
				[]byte(`{"test":"data"}`),
			),
			expectErr: true,
		},
		{
			name: "empty schema UID",
			msg: types.NewMsgAttest(
				validAddr,
				"",
				validRecipient,
				0,
				true,
				"",
				[]byte(`{"test":"data"}`),
			),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreateEncryptedAttestation_ValidateBasic(t *testing.T) {
	// Initialize SDK config for cert prefix
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("cert", "certpub")

	validAddr := createTestAddress("cert")
	validRecipient := sdk.AccAddress("recipient123456789a").String()

	testCases := []struct {
		name      string
		msg       *types.MsgCreateEncryptedAttestation
		expectErr bool
	}{
		{
			name: "valid message with single recipient",
			msg: types.NewMsgCreateEncryptedAttestation(
				validAddr,
				"0x1234567890abcdef",
				"QmTest123",
				"0xhash",
				[]string{validRecipient},
				map[string]string{validRecipient: "encryptedKey1"},
				true,
				0,
			),
			expectErr: false,
		},
		{
			name: "exceeds max recipients (50)",
			msg: func() *types.MsgCreateEncryptedAttestation {
				recipients := make([]string, 51)
				keys := make(map[string]string)
				for i := 0; i < 51; i++ {
					recipients[i] = validRecipient
					keys[recipients[i]] = "key"
				}
				return types.NewMsgCreateEncryptedAttestation(
					validAddr, "0x1234", "QmTest", "0xhash",
					recipients, keys, true, 0,
				)
			}(),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
