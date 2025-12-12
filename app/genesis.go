package app

import (
	"encoding/json"
	"time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GenesisState represents the entire genesis state for the CERT blockchain
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the CERT blockchain
// Per Whitepaper Sections 4, 5, and 12
func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	return ModuleBasics.DefaultGenesis(cdc)
}

// GetStakingGenesisState returns staking genesis state per Whitepaper Section 4.1
func GetStakingGenesisState(cdc codec.JSONCodec) *stakingtypes.GenesisState {
	stakingGenState := stakingtypes.DefaultGenesisState()

	// Per Whitepaper Section 4.1
	stakingGenState.Params.UnbondingTime = UnbondingTime                       // 21 days
	stakingGenState.Params.MaxValidators = MaxValidators                       // 80 validators
	stakingGenState.Params.BondDenom = TokenDenom                              // ucert
	stakingGenState.Params.MinCommissionRate = math.LegacyNewDecWithPrec(5, 2) // 5%

	return stakingGenState
}

// GetSlashingGenesisState returns slashing genesis state per Whitepaper Section 4.1
func GetSlashingGenesisState(cdc codec.JSONCodec) *slashingtypes.GenesisState {
	slashingGenState := slashingtypes.DefaultGenesisState()

	// Per Whitepaper Section 4.1
	slashingGenState.Params.SignedBlocksWindow = 10000
	slashingGenState.Params.MinSignedPerWindow = math.LegacyNewDecWithPrec(5, 2) // 5%
	slashingGenState.Params.DowntimeJailDuration = 10 * time.Minute
	slashingGenState.Params.SlashFractionDoubleSign = math.LegacyNewDecWithPrec(5, 2) // 5% - Whitepaper
	slashingGenState.Params.SlashFractionDowntime = math.LegacyNewDecWithPrec(1, 4)   // 0.01% - Whitepaper

	return slashingGenState
}

// GetGovGenesisState returns governance genesis state
func GetGovGenesisState(cdc codec.JSONCodec) *govv1.GenesisState {
	govGenState := govv1.DefaultGenesisState()

	// Set governance parameters
	minDeposit := sdk.NewCoins(sdk.NewCoin(TokenDenom, math.NewInt(10_000_000_000))) // 10,000 CERT
	govGenState.Params.MinDeposit = minDeposit

	votingPeriod := 7 * 24 * time.Hour // 7 days
	govGenState.Params.VotingPeriod = &votingPeriod

	maxDepositPeriod := 14 * 24 * time.Hour // 14 days
	govGenState.Params.MaxDepositPeriod = &maxDepositPeriod

	govGenState.Params.Quorum = "0.334"  // 33.4%
	govGenState.Params.Threshold = "0.5" // 50%
	govGenState.Params.VetoThreshold = "0.334"

	return govGenState
}

// GetMintGenesisState returns mint genesis state
// Per Whitepaper Section 5.3 - Non-inflationary, zero protocol-level inflation
func GetMintGenesisState(cdc codec.JSONCodec) *minttypes.GenesisState {
	mintGenState := minttypes.DefaultGenesisState()

	// Set inflation to 0 per Whitepaper Section 5.3
	mintGenState.Params.InflationRateChange = math.LegacyZeroDec()
	mintGenState.Params.InflationMax = math.LegacyZeroDec()
	mintGenState.Params.InflationMin = math.LegacyZeroDec()
	mintGenState.Params.GoalBonded = math.LegacyNewDecWithPrec(67, 2) // 67% bonded
	mintGenState.Params.MintDenom = TokenDenom
	mintGenState.Minter.Inflation = math.LegacyZeroDec()
	mintGenState.Minter.AnnualProvisions = math.LegacyZeroDec()

	return mintGenState
}

// GetBankGenesisState returns bank genesis state with CERT token metadata
func GetBankGenesisState(cdc codec.JSONCodec) *banktypes.GenesisState {
	bankGenState := banktypes.DefaultGenesisState()

	// Set CERT token metadata per Whitepaper Section 5.1
	bankGenState.DenomMetadata = []banktypes.Metadata{
		{
			Description: "The native staking and utility token of CERT Blockchain",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    TokenDenom,
					Exponent: 0,
					Aliases:  []string{"microcert"},
				},
				{
					Denom:    "mcert",
					Exponent: 3,
					Aliases:  []string{"millicert"},
				},
				{
					Denom:    "cert",
					Exponent: 6,
					Aliases:  []string{TokenSymbol},
				},
			},
			Base:    TokenDenom,
			Display: "cert",
			Name:    "CERT",
			Symbol:  TokenSymbol,
			URI:     "",
			URIHash: "",
		},
	}

	return bankGenState
}

// GetCrisisGenesisState returns crisis genesis state
func GetCrisisGenesisState(cdc codec.JSONCodec) *crisistypes.GenesisState {
	crisisGenState := crisistypes.DefaultGenesisState()
	crisisGenState.ConstantFee = sdk.NewCoin(TokenDenom, math.NewInt(1_000_000_000)) // 1,000 CERT
	return crisisGenState
}
