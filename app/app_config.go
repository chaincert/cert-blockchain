package app

import (
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// CERT Blockchain Network Parameters from Whitepaper Section 4 & 12
const (
	// ChainID for CERT Blockchain (testnet initially, mainnet TBA per whitepaper)
	ChainIDTestnet = "cert-testnet-1"
	ChainIDMainnet = "cert-mainnet-1"

	// Block time target: ~2 seconds (Whitepaper 4.1)
	BlockTimeTarget = 2 * time.Second

	// Max Validators: 80 (Whitepaper 4.1)
	MaxValidators = 80

	// Unbonding Period: 21 days (Whitepaper 4.1)
	UnbondingTime = 21 * 24 * time.Hour

	// Slashing Parameters (Whitepaper 4.1)
	DowntimeSlashFraction   = "0.0001" // 0.01%
	DoubleSignSlashFraction = "0.05"   // 5%

	// Max Gas per Block: 30,000,000 (Whitepaper 12)
	MaxGasPerBlock = 30_000_000

	// CERT Token Parameters (Whitepaper 5.1)
	TotalSupply   = 1_000_000_000_000_000 // 1 Billion CERT in ucert (1M ucert = 1 CERT)
	TokenSymbol   = "CERT"
	TokenDenom    = "ucert"
	TokenDecimals = 6

	// Encrypted Attestation Parameters (Whitepaper 12)
	MaxEncryptedFileSize        = 100 * 1024 * 1024 // 100 MB
	MaxRecipientsPerAttestation = 50

	// Minimum validator stake: 10,000 CERT (Whitepaper 10)
	MinValidatorStake = 10_000_000_000 // 10,000 CERT in ucert
)

// SetConfig sets the Bech32 address prefixes and coin type for CERT
func SetConfig() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(AccountAddressPrefix, AccountAddressPrefix+"pub")
	config.SetBech32PrefixForValidator(AccountAddressPrefix+"valoper", AccountAddressPrefix+"valoperpub")
	config.SetBech32PrefixForConsensusNode(AccountAddressPrefix+"valcons", AccountAddressPrefix+"valconspub")
	config.SetCoinType(60) // Ethereum coin type for EVM compatibility
	config.Seal()
}

// GetDefaultStakingParams returns the default staking parameters per whitepaper specs
func GetDefaultStakingParams() interface{} {
	return struct {
		UnbondingTime     time.Duration
		MaxValidators     uint32
		MaxEntries        uint32
		HistoricalEntries uint32
		BondDenom         string
		MinCommissionRate string
	}{
		UnbondingTime:     UnbondingTime,
		MaxValidators:     MaxValidators,
		MaxEntries:        7,
		HistoricalEntries: 10000,
		BondDenom:         TokenDenom,
		MinCommissionRate: "0.05", // 5% minimum commission
	}
}

// GetDefaultSlashingParams returns slashing parameters per whitepaper section 4.1
func GetDefaultSlashingParams() interface{} {
	return struct {
		SignedBlocksWindow      int64
		MinSignedPerWindow      string
		DowntimeJailDuration    time.Duration
		SlashFractionDoubleSign string
		SlashFractionDowntime   string
	}{
		SignedBlocksWindow:      10000,
		MinSignedPerWindow:      "0.05", // Must sign 5% of blocks
		DowntimeJailDuration:    10 * time.Minute,
		SlashFractionDoubleSign: DoubleSignSlashFraction,
		SlashFractionDowntime:   DowntimeSlashFraction,
	}
}

// GetDefaultGovParams returns governance parameters
func GetDefaultGovParams() govv1.Params {
	return govv1.Params{
		MinDeposit:                 sdk.NewCoins(sdk.NewCoin(TokenDenom, math.NewInt(10_000_000_000))), // 10,000 CERT
		MaxDepositPeriod:           nil,                                                                // Will be set with proper duration
		VotingPeriod:               nil,                                                                // Will be set with proper duration
		Quorum:                     "0.334",                                                            // 33.4% quorum
		Threshold:                  "0.5",                                                              // 50% threshold
		VetoThreshold:              "0.334",                                                            // 33.4% veto threshold
		MinInitialDepositRatio:     "0.0",
		ProposalCancelRatio:        "0.5",
		ProposalCancelDest:         "",
		ExpeditedVotingPeriod:      nil,
		ExpeditedThreshold:         "0.667",
		ExpeditedMinDeposit:        sdk.NewCoins(sdk.NewCoin(TokenDenom, math.NewInt(50_000_000_000))), // 50,000 CERT
		BurnVoteQuorum:             false,
		BurnProposalDepositPrevote: false,
		BurnVoteVeto:               true,
	}
}

// GetMaccPerms returns the module account permissions
func GetMaccPerms() map[string][]string {
	return maccPerms
}
