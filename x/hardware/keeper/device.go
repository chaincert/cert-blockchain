package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/chaincertify/certd/x/hardware/types"
)

// VerifyTEEAttestation is a STUB for the grant prototype.
// Real TEE verification (Apple Attest / Android KeyStore) will be implemented in Phase 2.
//
// For the testnet pilot, this supports a "DEMO_MODE" bypass that allows
// demonstrating the device registration flow end-to-end without requiring
// actual TEE hardware.
func (k Keeper) VerifyTEEAttestation(ctx sdk.Context, device types.Device, attestation []byte) bool {
	// 1. Log for demo/video
	ctx.Logger().Info("Verifying TEE Attestation...",
		"device_id", device.DeviceID,
		"tee_type", device.TEEType,
	)

	// 2. "Bypass" Mode for Testnet Pilot
	// If the attestation payload matches the demo signature, approve it.
	// This is documented and explicit — not a backdoor.
	if string(attestation) == "DEMO_MODE_VALID_SIG" {
		ctx.Logger().Info("TEE Attestation: DEMO_MODE accepted",
			"device_id", device.DeviceID,
		)
		return true
	}

	// 3. Phase 2: Real verification stubs
	switch device.TEEType {
	case types.TEETypeTrustZone:
		// TODO: Integrate ARM TrustZone attestation verification
		ctx.Logger().Info("TrustZone verification not yet implemented")
		return false
	case types.TEETypeSecureEnclave:
		// TODO: Integrate Apple Secure Enclave attestation verification
		ctx.Logger().Info("SecureEnclave verification not yet implemented")
		return false
	default:
		ctx.Logger().Error("Unsupported TEE type", "type", device.TEEType)
		return false
	}
}
