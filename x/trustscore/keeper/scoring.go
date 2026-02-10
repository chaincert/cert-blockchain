package keeper

import (
	hardwaretypes "github.com/chaincertify/certd/x/hardware/types"
)

// CalculateHumanityScore is the trustscore module's local wrapper around the
// deterministic scoring algorithm. The algorithm is identical to the one in
// x/hardware/keeper/scoring.go — extracted here so the trustscore module can
// call it without a cross-keeper dependency.
//
// Per Whitepaper v3.0: "Prove the account owner is a single human, not a bot farm"
//
// Weighting:
//   - Hardware Anchor: 40% (Linked to unique high-trust device)
//   - Social Staking: 30% (Verified aged social accounts)
//   - On-Chain History: 20% (Account age + transaction history)
//   - Network Fees Paid: 10% (Burned $CERT/ETH fees)
func CalculateHumanityScore(factors hardwaretypes.HumanityFactors) hardwaretypes.HumanityResult {
	config := hardwaretypes.DefaultTrustScoreConfig()
	result := hardwaretypes.HumanityResult{}

	// HARDWARE ANCHOR (40 points max)
	if factors.LinkedDeviceScore >= config.HighTrustDeviceThreshold {
		basePoints := config.HardwareAnchorWeight
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

	// SOCIAL STAKING (30 points max)
	socialAccounts := factors.VerifiedSocialAccounts
	if socialAccounts > 3 {
		socialAccounts = 3
	}
	result.SocialPoints = uint64(socialAccounts * 10)
	result.Score += result.SocialPoints

	// ON-CHAIN HISTORY (20 points max)
	if factors.AccountAgeMonths >= 6 {
		result.OnChainPoints += 10
	}
	if factors.TransactionCount >= 5 {
		result.OnChainPoints += 10
	}
	result.Score += result.OnChainPoints

	// NETWORK FEES PAID (10 points max)
	if factors.TotalFeesBurnedUSD >= 10.0 {
		result.FeePoints = config.NetworkFeesPaidWeight
	} else {
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
