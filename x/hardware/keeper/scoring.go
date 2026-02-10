package keeper

import (
	"github.com/chaincertify/certd/x/hardware/types"
)

// CalculateDeviceTrustScore computes the Device Trust Score using the deterministic algorithm
// Per Whitepaper v3.0: "Deterministic, Hard to Game, Transparent"
//
// Weighting:
//   - TEE Attestation: 40% (Pass/Fail - Critical)
//   - Uptime Reliability: 25% (Rolling 7-day average)
//   - Data Congruence: 20% (Statistical deviation from neighbors)
//   - Firmware Integrity: 15% (Latest signed firmware)
//
// Slasher Logic:
//   - If TEE fails -> Score = 0, device banned
//   - If Data Congruence <50% for 3 days -> Flagged for audit
func CalculateDeviceTrustScore(factors types.DeviceTrustFactors) types.DeviceTrustResult {
	config := types.DefaultTrustScoreConfig()
	result := types.DeviceTrustResult{}

	// ==================================================
	// CRITICAL FAIL STATE: TEE Attestation
	// If the TEE signature is invalid, score is 0 and device is banned
	// This prevents software emulators from participating
	// ==================================================
	if !factors.TEEAttestationValid {
		result.Score = 0
		result.TEEPassed = false
		result.Banned = true
		return result
	}
	result.TEEPassed = true
	result.Score = config.TEEAttestationWeight // +40

	// ==================================================
	// UPTIME RELIABILITY (25 points max)
	// Formula: (Hours Online / 24) * 25, rolling 7-day average
	// DePIN networks need stable nodes; flaky devices get lower scores
	// ==================================================
	uptime := clampFloat(factors.Uptime, 0.0, 1.0)
	result.UptimePoints = uint64(uptime * float64(config.UptimeWeight))
	result.Score += result.UptimePoints

	// ==================================================
	// DATA CONGRUENCE (20 points max)
	// Statistical deviation from neighbor nodes
	// If a sensor says 100°F and neighbors say 70°F, score drops
	// Prevents "lying" nodes that send fake data to farm rewards
	// ==================================================
	congruence := clampFloat(factors.DataCongruence, 0.0, 1.0)
	result.CongruencePoints = uint64(congruence * float64(config.DataCongruenceWeight))
	result.Score += result.CongruencePoints

	// Check for audit flag: <50% congruence for 3+ consecutive days
	if factors.DataCongruence < 0.5 && factors.ConsecutiveLowCongruenceDays >= int(config.DataCongruenceAuditDays) {
		result.FlaggedForAudit = true
	}

	// ==================================================
	// FIRMWARE INTEGRITY (15 points max)
	// +15 for latest signed firmware
	// -5 for every version behind
	// Ensures security patches are applied, prevents hacked firmware
	// ==================================================
	if factors.FirmwareVersion == types.LatestFirmwareVersion {
		result.FirmwarePoints = config.FirmwareIntegrityWeight // +15
	} else {
		versionsBehind := types.LatestFirmwareVersion - factors.FirmwareVersion
		if versionsBehind < 0 {
			versionsBehind = 0 // Future version? Give full points
			result.FirmwarePoints = config.FirmwareIntegrityWeight
		} else {
			penalty := uint64(versionsBehind * 5)
			if penalty >= config.FirmwareIntegrityWeight {
				result.FirmwarePoints = 0
			} else {
				result.FirmwarePoints = config.FirmwareIntegrityWeight - penalty
			}
		}
	}
	result.Score += result.FirmwarePoints

	// Cap at 100
	if result.Score > 100 {
		result.Score = 100
	}

	return result
}

// CalculateHumanityScore computes the Humanity Score for Sybil resistance
// Per Whitepaper v3.0: "Prove the account owner is a single human, not a bot farm"
//
// Weighting:
//   - Hardware Anchor: 40% (Linked to unique high-trust device)
//   - Social Staking: 30% (Verified aged social accounts)
//   - On-Chain History: 20% (Account age + transaction history)
//   - Network Fees Paid: 10% (Burned $CERT/ETH fees)
//
// Sybil Logic:
//   - 1 Device, 1 Human: If device linked to multiple accounts, split points
//   - Threshold for "Verified Human" = 60+
func CalculateHumanityScore(factors types.HumanityFactors) types.HumanityResult {
	config := types.DefaultTrustScoreConfig()
	result := types.HumanityResult{}

	// ==================================================
	// HARDWARE ANCHOR (40 points max)
	// +40 if linked to unique high-trust device (Device Score > 80)
	// "It is expensive to buy 1,000 physical devices"
	// ==================================================
	if factors.LinkedDeviceScore >= config.HighTrustDeviceThreshold {
		basePoints := config.HardwareAnchorWeight // 40

		// SYBIL DEFENSE: If device shared across multiple accounts, split the points
		// 1 Device, 1 Human principle
		if factors.LinkedDeviceSharedAccounts > 1 {
			result.SybilMultiplier = 1.0 / float64(factors.LinkedDeviceSharedAccounts)
			result.HardwarePoints = uint64(float64(basePoints) * result.SybilMultiplier)
		} else {
			result.SybilMultiplier = 1.0
			result.HardwarePoints = basePoints
		}
	} else {
		result.SybilMultiplier = 1.0
		result.HardwarePoints = 0
	}
	result.Score = result.HardwarePoints

	// ==================================================
	// SOCIAL STAKING (30 points max)
	// +10 per verified aged account (X >6mo, GitHub w/ commits, Discord, LinkedIn)
	// "Hard to fake aged social accounts at scale"
	// Max 3 accounts = 30 points
	// ==================================================
	socialAccounts := factors.VerifiedSocialAccounts
	if socialAccounts > 3 {
		socialAccounts = 3 // Cap at 3
	}
	result.SocialPoints = uint64(socialAccounts * 10)
	result.Score += result.SocialPoints

	// ==================================================
	// ON-CHAIN HISTORY (20 points max)
	// +10 for account age > 6 months
	// +10 for > 5 transactions on Eth/Arbitrum
	// "Bots usually use fresh wallets. Real humans have history."
	// ==================================================
	if factors.AccountAgeMonths >= 6 {
		result.OnChainPoints += 10
	}
	if factors.TransactionCount >= 5 {
		result.OnChainPoints += 10
	}
	result.Score += result.OnChainPoints

	// ==================================================
	// NETWORK FEES PAID (10 points max)
	// +10 if user has burned > $10 in fees over lifetime
	// "Cost of Forgery. Spammers hate spending real money."
	// ==================================================
	if factors.TotalFeesBurnedUSD >= 10.0 {
		result.FeePoints = config.NetworkFeesPaidWeight
	} else {
		// Proportional scoring for partial fees
		result.FeePoints = uint64((factors.TotalFeesBurnedUSD / 10.0) * float64(config.NetworkFeesPaidWeight))
	}
	result.Score += result.FeePoints

	// Cap at 100
	if result.Score > 100 {
		result.Score = 100
	}

	// Determine verification status
	result.IsVerifiedHuman = result.Score >= config.VerifiedHumanityThreshold

	return result
}

// clampFloat clamps a float64 value between min and max
func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
