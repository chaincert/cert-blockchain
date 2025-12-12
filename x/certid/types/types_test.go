package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chaincertify/certd/x/certid/types"
)

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()

	require.Equal(t, uint32(32), params.MaxUsernameLength)
	require.Equal(t, uint32(100), params.MaxDisplayNameLength)
	require.Equal(t, uint32(500), params.MaxBioLength)
	require.Equal(t, uint32(50), params.MaxCredentials)
	require.True(t, params.RegistrationFee.IsZero())
}

func TestValidateUsername(t *testing.T) {
	testCases := []struct {
		name      string
		username  string
		expectErr bool
	}{
		{
			name:      "valid lowercase",
			username:  "alice",
			expectErr: false,
		},
		{
			name:      "valid with numbers",
			username:  "alice123",
			expectErr: false,
		},
		{
			name:      "valid with underscore",
			username:  "alice_smith",
			expectErr: false,
		},
		{
			name:      "empty",
			username:  "",
			expectErr: true,
		},
		{
			name:      "too short (< 3 chars)",
			username:  "ab",
			expectErr: true,
		},
		{
			name:      "too long (> 32 chars)",
			username:  "thisusernameiswaytoolongtobevalid",
			expectErr: true,
		},
		{
			name:      "invalid characters",
			username:  "alice@smith",
			expectErr: true,
		},
		{
			name:      "starts with number",
			username:  "123alice",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateUsername(tc.username)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSupportedPlatforms(t *testing.T) {
	supported := types.SupportedSocialPlatforms()

	require.Contains(t, supported, "twitter")
	require.Contains(t, supported, "github")
	require.Contains(t, supported, "linkedin")
	require.Contains(t, supported, "discord")
}

func TestModuleConstants(t *testing.T) {
	require.Equal(t, "certid", types.ModuleName)
	require.Equal(t, "certid", types.StoreKey)
	require.Equal(t, "certid", types.RouterKey)
}

func TestCredentialTypes(t *testing.T) {
	validTypes := types.ValidCredentialTypes()

	require.Contains(t, validTypes, "education")
	require.Contains(t, validTypes, "employment")
	require.Contains(t, validTypes, "certification")
	require.Contains(t, validTypes, "identity")
	require.Contains(t, validTypes, "membership")
}

