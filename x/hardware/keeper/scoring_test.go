package keeper_test

import (
	"testing"

	"github.com/chaincertify/certd/x/hardware/keeper"
	"github.com/chaincertify/certd/x/hardware/types"
)

func TestCalculateDeviceTrustScore_ValidAttestation(t *testing.T) {
	factors := types.DeviceTrustFactors{
		TEEAttestationValid: true,
		Uptime:              1.0, // 100% uptime
		DataCongruence:      1.0, // Perfect congruence
		FirmwareVersion:     types.LatestFirmwareVersion,
	}

	result := keeper.CalculateDeviceTrustScore(factors)

	if result.Score != 100 {
		t.Errorf("Expected score 100, got %d", result.Score)
	}
	if !result.TEEPassed {
		t.Error("Expected TEEPassed to be true")
	}
	if result.Banned {
		t.Error("Expected device not to be banned")
	}
}

func TestCalculateDeviceTrustScore_FailedTEE(t *testing.T) {
	factors := types.DeviceTrustFactors{
		TEEAttestationValid: false, // CRITICAL FAIL
		Uptime:              1.0,
		DataCongruence:      1.0,
		FirmwareVersion:     types.LatestFirmwareVersion,
	}

	result := keeper.CalculateDeviceTrustScore(factors)

	if result.Score != 0 {
		t.Errorf("Expected score 0 for failed TEE, got %d", result.Score)
	}
	if result.TEEPassed {
		t.Error("Expected TEEPassed to be false")
	}
	if !result.Banned {
		t.Error("Expected device to be banned")
	}
}

func TestCalculateDeviceTrustScore_LowCongruenceAudit(t *testing.T) {
	factors := types.DeviceTrustFactors{
		TEEAttestationValid:          true,
		Uptime:                       0.8,
		DataCongruence:               0.4, // Below 50%
		FirmwareVersion:              types.LatestFirmwareVersion,
		ConsecutiveLowCongruenceDays: 3, // 3 days threshold
	}

	result := keeper.CalculateDeviceTrustScore(factors)

	if !result.FlaggedForAudit {
		t.Error("Expected device to be flagged for audit")
	}
}

func TestCalculateDeviceTrustScore_OutdatedFirmware(t *testing.T) {
	factors := types.DeviceTrustFactors{
		TEEAttestationValid: true,
		Uptime:              1.0,
		DataCongruence:      1.0,
		FirmwareVersion:     types.LatestFirmwareVersion - 2, // 2 versions behind
	}

	result := keeper.CalculateDeviceTrustScore(factors)

	// Should get 15 - (2*5) = 5 firmware points
	if result.FirmwarePoints != 5 {
		t.Errorf("Expected firmware points 5, got %d", result.FirmwarePoints)
	}
}

func TestCalculateHumanityScore_VerifiedHuman(t *testing.T) {
	factors := types.HumanityFactors{
		LinkedDeviceScore:          85,  // Above 80 threshold
		LinkedDeviceSharedAccounts: 1,   // Not shared
		VerifiedSocialAccounts:     3,   // Max social points
		AccountAgeMonths:           12,  // Over 6 months
		TransactionCount:           10,  // Over 5 txs
		TotalFeesBurnedUSD:         15.0, // Over $10
	}

	result := keeper.CalculateHumanityScore(factors)

	if result.Score != 100 {
		t.Errorf("Expected score 100, got %d", result.Score)
	}
	if !result.IsVerifiedHuman {
		t.Error("Expected IsVerifiedHuman to be true")
	}
}

func TestCalculateHumanityScore_SybilDefense(t *testing.T) {
	factors := types.HumanityFactors{
		LinkedDeviceScore:          85,
		LinkedDeviceSharedAccounts: 5, // 5-way split!
		VerifiedSocialAccounts:     0,
		AccountAgeMonths:           0,
		TransactionCount:           0,
		TotalFeesBurnedUSD:         0,
	}

	result := keeper.CalculateHumanityScore(factors)

	// Hardware points should be 40 / 5 = 8
	if result.HardwarePoints != 8 {
		t.Errorf("Expected hardware points 8 (split 5 ways), got %d", result.HardwarePoints)
	}
	if result.SybilMultiplier != 0.2 {
		t.Errorf("Expected sybil multiplier 0.2, got %f", result.SybilMultiplier)
	}
}

func TestCalculateHumanityScore_LowTrustDevice(t *testing.T) {
	factors := types.HumanityFactors{
		LinkedDeviceScore:          50, // Below 80 threshold
		LinkedDeviceSharedAccounts: 1,
		VerifiedSocialAccounts:     2,
		AccountAgeMonths:           8,
		TransactionCount:           6,
		TotalFeesBurnedUSD:         12.0,
	}

	result := keeper.CalculateHumanityScore(factors)

	// No hardware points because device score < 80
	if result.HardwarePoints != 0 {
		t.Errorf("Expected hardware points 0 for low-trust device, got %d", result.HardwarePoints)
	}
	// Should still get social (20) + onchain (20) + fees (10) = 50
	if result.Score != 50 {
		t.Errorf("Expected score 50, got %d", result.Score)
	}
	if result.IsVerifiedHuman {
		t.Error("Expected IsVerifiedHuman to be false (below 60 threshold)")
	}
}
