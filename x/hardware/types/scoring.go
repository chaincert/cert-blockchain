package types

// TrustScoreConfig defines the weighting parameters for trust score calculation
// Per Whitepaper v3.0: Deterministic, Hard to Game, Transparent
type TrustScoreConfig struct {
	// Device Trust Score Weights (total = 100)
	TEEAttestationWeight    uint64 // 40% - Pass/Fail critical
	UptimeWeight            uint64 // 25% - Rolling 7-day average
	DataCongruenceWeight    uint64 // 20% - Statistical deviation from neighbors
	FirmwareIntegrityWeight uint64 // 15% - Latest signed firmware bonus

	// Humanity Score Weights (total = 100)
	HardwareAnchorWeight  uint64 // 40% - Linked to high-trust device
	SocialStakingWeight   uint64 // 30% - Verified social accounts
	OnChainHistoryWeight  uint64 // 20% - Account age + tx history
	NetworkFeesPaidWeight uint64 // 10% - Burned $CERT/ETH fees

	// Thresholds
	VerifiedHumanityThreshold uint64 // 60 - Minimum for "Verified Human"
	HighTrustDeviceThreshold  uint64 // 80 - Device qualifies for Hardware Anchor
	DataCongruenceAuditDays   uint64 // 3 - Days of <50% before audit flag
}

// DefaultTrustScoreConfig returns the standard scoring weights
func DefaultTrustScoreConfig() TrustScoreConfig {
	return TrustScoreConfig{
		// Device Trust Score (out of 100)
		TEEAttestationWeight:    40,
		UptimeWeight:            25,
		DataCongruenceWeight:    20,
		FirmwareIntegrityWeight: 15,

		// Humanity Score (out of 100)
		HardwareAnchorWeight:  40,
		SocialStakingWeight:   30,
		OnChainHistoryWeight:  20,
		NetworkFeesPaidWeight: 10,

		// Thresholds
		VerifiedHumanityThreshold: 60,
		HighTrustDeviceThreshold:  80,
		DataCongruenceAuditDays:   3,
	}
}

// LatestFirmwareVersion is the current expected firmware version
// This should be updated via governance when new firmware is released
const LatestFirmwareVersion = 1

// DeviceTrustFactors contains the inputs for device trust calculation
type DeviceTrustFactors struct {
	// TEEAttestationValid: true if TEE signature verified (critical fail if false)
	TEEAttestationValid bool

	// Uptime: 0.0 to 1.0 (percentage of online hours / 24 over 7-day rolling avg)
	Uptime float64

	// DataCongruence: 0.0 to 1.0 (statistical match with neighbor nodes)
	DataCongruence float64

	// FirmwareVersion: current firmware version number
	FirmwareVersion int

	// ConsecutiveLowCongruenceDays: days with <50% data congruence
	ConsecutiveLowCongruenceDays int
}

// HumanityFactors contains the inputs for humanity score calculation
type HumanityFactors struct {
	// LinkedDeviceScore: trust score of linked device (0 if no device)
	LinkedDeviceScore uint64

	// LinkedDeviceSharedAccounts: number of accounts sharing this device
	LinkedDeviceSharedAccounts uint64

	// VerifiedSocialAccounts: count of verified aged social accounts
	VerifiedSocialAccounts int

	// AccountAgeMonths: age of crypto account in months
	AccountAgeMonths int

	// TransactionCount: number of on-chain transactions
	TransactionCount int

	// TotalFeesBurnedUSD: total fees burned in USD equivalent
	TotalFeesBurnedUSD float64
}

// DeviceTrustResult contains the calculated device trust score and status
type DeviceTrustResult struct {
	Score           uint64
	TEEPassed       bool
	UptimePoints    uint64
	CongruencePoints uint64
	FirmwarePoints  uint64
	FlaggedForAudit bool
	Banned          bool
}

// HumanityResult contains the calculated humanity score and verification status
type HumanityResult struct {
	Score             uint64
	HardwarePoints    uint64
	SocialPoints      uint64
	OnChainPoints     uint64
	FeePoints         uint64
	IsVerifiedHuman   bool
	SybilMultiplier   float64 // 1.0 = no split, 0.2 = 5-way split
}
