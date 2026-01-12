package app


import (
	"crypto/sha256"
	"encoding/json"
	"time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisis "github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	mint "github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	// authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	// genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	// paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	// consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	// attestationtypes "github.com/chaincertify/certd/x/attestation/types"
)

// GenesisState represents the entire genesis state for the CERT blockchain
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the CERT blockchain
// Per Whitepaper Sections 4, 5, and 12
func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	// Start from the SDK defaults (with our custom AppModuleBasic overrides)
	genesis := ModuleBasics.DefaultGenesis(cdc)

	// Override with custom genesis states for modules with custom helpers
	genesis[stakingtypes.ModuleName] = cdc.MustMarshalJSON(GetStakingGenesisState(cdc))
	genesis[slashingtypes.ModuleName] = cdc.MustMarshalJSON(GetSlashingGenesisState(cdc))
	genesis[govtypes.ModuleName] = cdc.MustMarshalJSON(GetGovGenesisState(cdc))
	genesis[minttypes.ModuleName] = cdc.MustMarshalJSON(GetMintGenesisState(cdc))
	// genesis[crisistypes.ModuleName] = cdc.MustMarshalJSON(GetCrisisGenesisState(cdc))

	// Add genesis accounts with Tokenomics v2.1 distribution
	genesis[banktypes.ModuleName] = cdc.MustMarshalJSON(GetBankGenesisStateWithAccounts(cdc))

	return genesis
}

// StakingModuleBasicGenesis overrides the SDK staking module's default
// genesis state so it matches the CERT whitepaper parameters.
//
// It still implements module.AppModuleBasic so it integrates cleanly with
// module.BasicManager and genutil's InitCmd.
type StakingModuleBasicGenesis struct {
	staking.AppModuleBasic
}

// Ensure StakingModuleBasicGenesis satisfies the AppModuleBasic interface.
var _ module.AppModuleBasic = StakingModuleBasicGenesis{}

// DefaultGenesis returns the staking genesis state with CERT-specific params.
func (StakingModuleBasicGenesis) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(GetStakingGenesisState(cdc))
}

// BankModuleBasicGenesis overrides the SDK bank module's default genesis state
// to include CERT token metadata per the whitepaper.
type BankModuleBasicGenesis struct {
	bank.AppModuleBasic
}

// Ensure BankModuleBasicGenesis satisfies the AppModuleBasic interface.
var _ module.AppModuleBasic = BankModuleBasicGenesis{}

// DefaultGenesis returns the bank genesis state with CERT token metadata.
func (BankModuleBasicGenesis) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(GetBankGenesisState(cdc))
}

// SlashingModuleBasicGenesis overrides the SDK slashing module's default
// genesis state to set CERT-specific slashing parameters.
type SlashingModuleBasicGenesis struct {
	slashing.AppModuleBasic
}

// Ensure SlashingModuleBasicGenesis satisfies the AppModuleBasic interface.
var _ module.AppModuleBasic = SlashingModuleBasicGenesis{}

// DefaultGenesis returns the slashing genesis state with CERT-specific params.
func (SlashingModuleBasicGenesis) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(GetSlashingGenesisState(cdc))
}

// GovModuleBasicGenesis overrides the SDK gov module's default
// genesis state to set CERT-specific governance parameters.
type GovModuleBasicGenesis struct {
	gov.AppModuleBasic
}

// Ensure GovModuleBasicGenesis satisfies the AppModuleBasic interface.
var _ module.AppModuleBasic = GovModuleBasicGenesis{}

// DefaultGenesis returns the gov genesis state with CERT-specific params.
func (GovModuleBasicGenesis) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(GetGovGenesisState(cdc))
}

// MintModuleBasicGenesis overrides the SDK mint module's default
// genesis state to set CERT-specific mint parameters (zero inflation).
type MintModuleBasicGenesis struct {
	mint.AppModuleBasic
}

// Ensure MintModuleBasicGenesis satisfies the AppModuleBasic interface.
var _ module.AppModuleBasic = MintModuleBasicGenesis{}

// DefaultGenesis returns the mint genesis state with CERT-specific params.
func (MintModuleBasicGenesis) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(GetMintGenesisState(cdc))
}

// CrisisModuleBasicGenesis overrides the SDK crisis module's default
// genesis state to set CERT-specific constant fee.
type CrisisModuleBasicGenesis struct {
	crisis.AppModuleBasic
}

// Ensure CrisisModuleBasicGenesis satisfies the AppModuleBasic interface.
var _ module.AppModuleBasic = CrisisModuleBasicGenesis{}

// DefaultGenesis returns the crisis genesis state with CERT-specific constant fee.
func (CrisisModuleBasicGenesis) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(GetCrisisGenesisState(cdc))
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

	// Expedited proposals must have a higher minimum deposit than normal proposals
	expeditedMinDeposit := sdk.NewCoins(sdk.NewCoin(TokenDenom, math.NewInt(20_000_000_000))) // 20,000 CERT
	govGenState.Params.ExpeditedMinDeposit = expeditedMinDeposit

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

// GetBankGenesisStateWithAccounts returns bank genesis state with Tokenomics v2.1 distribution
func GetBankGenesisStateWithAccounts(cdc codec.JSONCodec) *banktypes.GenesisState {
	bankGenState := GetBankGenesisState(cdc)
		
	// Tokenomics v2.1 Distribution
	// Total Supply: 1,000,000,000 CERT = 1,000,000,000,000,000 ucert
	const totalSupply = 1_000_000_000_000_000
	
	// Distribution percentages
	const treasuryPercent = 32  // 320,000,000 CERT
	const stakingPercent = 30   // 300,000,000 CERT
	const teamPrivatePercent = 30 // 300,000,000 CERT (15% + 15%)
	const advisorsPercent = 5   // 50,000,000 CERT
	const airdropPercent = 3    // 30,000,000 CERT
	
	// Calculate amounts in ucert
	treasuryAmount := totalSupply * treasuryPercent / 100
	stakingAmount := totalSupply * stakingPercent / 100
	teamPrivateAmount := totalSupply * teamPrivatePercent / 100
	advisorsAmount := totalSupply * advisorsPercent / 100
	airdropAmount := totalSupply * airdropPercent / 100
	
		// Helper to deterministically derive valid CERT bech32 addresses from labels
		mustAddressForLabel := func(label string) string {
			b := sha256.Sum256([]byte(label))
			addr := sdk.AccAddress(b[:20])
			return addr.String()
		}
		
		// Create genesis accounts using deterministic, valid CERT addresses
		genesisAccounts := []banktypes.Balance{
			{
				Address: mustAddressForLabel("treasury"), // Treasury address
				Coins:   sdk.NewCoins(sdk.NewCoin(TokenDenom, math.NewInt(int64(treasuryAmount)))),
			},
			{
				Address: mustAddressForLabel("staking-reserve"), // Staking pool address
				Coins:   sdk.NewCoins(sdk.NewCoin(TokenDenom, math.NewInt(int64(stakingAmount)))),
			},
			{
				Address: mustAddressForLabel("team-private"), // Team/Private address
				Coins:   sdk.NewCoins(sdk.NewCoin(TokenDenom, math.NewInt(int64(teamPrivateAmount)))),
			},
			{
				Address: mustAddressForLabel("advisors"), // Advisors address
				Coins:   sdk.NewCoins(sdk.NewCoin(TokenDenom, math.NewInt(int64(advisorsAmount)))),
			},
			{
				Address: mustAddressForLabel("airdrop"), // Airdrop address
				Coins:   sdk.NewCoins(sdk.NewCoin(TokenDenom, math.NewInt(int64(airdropAmount)))),
			},
		}
		
	bankGenState.Balances = genesisAccounts
	bankGenState.Supply = sdk.NewCoins(sdk.NewCoin(TokenDenom, math.NewInt(totalSupply)))
	
	return bankGenState
}

// GetCrisisGenesisState returns crisis genesis state
func GetCrisisGenesisState(cdc codec.JSONCodec) *crisistypes.GenesisState {
	crisisGenState := crisistypes.DefaultGenesisState()
	crisisGenState.ConstantFee = sdk.NewCoin(TokenDenom, math.NewInt(1_000_000_000)) // 1,000 CERT
	return crisisGenState
}