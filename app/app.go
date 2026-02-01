package app

import (
	"encoding/json"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/grpc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/reflect/protoreflect"
	protov2 "google.golang.org/protobuf/proto"
	evmapi "github.com/evmos/evmos/v20/api/ethermint/evm/v1"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/mempool"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	evmante "github.com/evmos/evmos/v20/app/ante/evm"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	// govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	// Crisis module disabled - empty stores cause IAVL v1.x "version does not exist" errors
	// "github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	// crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	// IBC imports
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	// CERT custom modules
	attestationmodule "github.com/chaincertify/certd/x/attestation"
	attestationkeeper "github.com/chaincertify/certd/x/attestation/keeper"
	attestationtypes "github.com/chaincertify/certd/x/attestation/types"

	// Evmos imports
	"github.com/evmos/evmos/v20/app/ante"
	"github.com/evmos/evmos/v20/x/evm"
	evmkeeper "github.com/evmos/evmos/v20/x/evm/keeper"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/evmos/evmos/v20/x/feemarket"
	feemarketkeeper "github.com/evmos/evmos/v20/x/feemarket/keeper"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
	etherminttypes "github.com/evmos/evmos/v20/types"
)

const (
	// AppName is the name of the CERT Blockchain application
	AppName = "certd"

	// AccountAddressPrefix is the Bech32 prefix for account addresses
	AccountAddressPrefix = "cert"

	// BondDenom is the staking/bonding denomination - CERT token
	BondDenom = "ucert"
)

var (
	// DefaultNodeHome is the default home directory for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager for the app
	// Must match modules registered in ModuleManager to avoid gRPC gateway registration issues
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		upgrade.AppModuleBasic{},
			BankModuleBasicGenesis{},
			StakingModuleBasicGenesis{},
		params.AppModuleBasic{},
		consensus.AppModuleBasic{},
		// Custom CERT modules
		attestationmodule.AppModuleBasic{},
		// Ethermint modules
		evm.AppModuleBasic{},
		feemarket.AppModuleBasic{},
		// Add overrides for unused modules
		SlashingModuleBasicGenesis{},
		GovModuleBasicGenesis{},
		MintModuleBasicGenesis{},
		// CrisisModuleBasicGenesis{},
	)

	// Module account permissions - only for modules in ModuleManager
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	// Add permissions for new modules
		minttypes.ModuleName: {authtypes.Minter},
		govtypes.ModuleName:  {authtypes.Burner},
		evmtypes.ModuleName:  {authtypes.Minter, authtypes.Burner},
		feemarkettypes.ModuleName: nil,
	}
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	DefaultNodeHome = filepath.Join(userHomeDir, ".certd")
}

// MakeEncodingConfig creates an EncodingConfig for the app with proper address codecs
func MakeEncodingConfig() (codec.Codec, codectypes.InterfaceRegistry, client.TxConfig) {
	// Create address codecs with the "cert" bech32 prefix
	addressCodec := address.NewBech32Codec(AccountAddressPrefix)
	validatorAddressCodec := address.NewBech32Codec(AccountAddressPrefix + "valoper")

	// Create interface registry with proper signing options using global proto registry
	interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: protoregistry.GlobalFiles,
		SigningOptions: signing.Options{
			AddressCodec:          addressCodec,
			ValidatorAddressCodec: validatorAddressCodec,
			CustomGetSigners: map[protoreflect.FullName]signing.GetSignersFunc{
				"ethermint.evm.v1.MsgEthereumTx": func(msg protov2.Message) ([][]byte, error) {
	// Use Evmos's GetSigners which recovers the sender address from the tx signature
					return evmapi.GetSigners(msg)
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	std.RegisterInterfaces(interfaceRegistry)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)
	etherminttypes.RegisterInterfaces(interfaceRegistry)
	
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)
	return appCodec, interfaceRegistry, txConfig
}

// GetTxConfig returns the default TxConfig for CLI commands
func GetTxConfig() client.TxConfig {
	_, _, txConfig := MakeEncodingConfig()
	return txConfig
}

// CertApp extends the Cosmos SDK BaseApp with CERT-specific functionality
type CertApp struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	// Keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// Cosmos SDK Keepers (only modules in ModuleManager)
	AccountKeeper   authkeeper.AccountKeeper
	BankKeeper      bankkeeper.Keeper
	StakingKeeper   *stakingkeeper.Keeper
	ParamsKeeper    paramskeeper.Keeper
	ConsensusKeeper consensuskeeper.Keeper

	// IBC Keeper (placeholder for future IBC integration)
	IBCKeeper *ibckeeper.Keeper

	// CERT Custom Module Keepers
	AttestationKeeper attestationkeeper.Keeper

	// New module keepers
	SlashingKeeper slashingkeeper.Keeper
	GovKeeper      govkeeper.Keeper
	MintKeeper     mintkeeper.Keeper
	CrisisKeeper   crisiskeeper.Keeper
	UpgradeKeeper  *upgradekeeper.Keeper

	// Ethermint Keepers
	EvmKeeper       *evmkeeper.Keeper
	FeeMarketKeeper feemarketkeeper.Keeper

	// Module Manager
	ModuleManager *module.Manager

	// Simulation Manager (for testing)
	sm *module.SimulationManager
}

// NewCertApp creates and initializes a new CERT blockchain application
func NewCertApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *CertApp {
	// Create encoding config with proper address codecs for Cosmos SDK v0.50.x
	legacyAmino := codec.NewLegacyAmino()

	// Create address codecs with the "cert" bech32 prefix
	addressCodec := address.NewBech32Codec(AccountAddressPrefix)
	validatorAddressCodec := address.NewBech32Codec(AccountAddressPrefix + "valoper")

	// Create interface registry with proper signing options (required for address conversion)
	// IMPORTANT: Must include CustomGetSigners for MsgEthereumTx to handle EVM transactions
	interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: protoregistry.GlobalFiles,
		SigningOptions: signing.Options{
			AddressCodec:          addressCodec,
			ValidatorAddressCodec: validatorAddressCodec,
			CustomGetSigners: map[protoreflect.FullName]signing.GetSignersFunc{
				(&evmapi.MsgEthereumTx{}).ProtoReflect().Descriptor().FullName(): func(msg protov2.Message) ([][]byte, error) {
					// Robust handler for MsgEthereumTx
					// 1. Try generic evmapi.GetSigners which recovers from signature
					// 2. Fallback to "from" field if available (legacy/debug)
					// 3. Log everything to stderr to debug issues

					fmt.Fprintf(os.Stderr, "DEBUG: CustomGetSigners called for MsgEthereumTx\n")

					// Defensive panic recovery


	// Defensive panic recovery
	defer func() {
						if r := recover(); r != nil {
							fmt.Fprintf(os.Stderr, "ERROR: CustomGetSigners panicked: %v\n", r)
						}
					}()

					signers, err := evmapi.GetSigners(msg)
					if err == nil && len(signers) > 0 {
						// Check for empty bytes
						valid := true
						for _, s := range signers {
							if len(s) == 0 {
								valid = false
								break
							}
						}
						if valid {
							fmt.Fprintf(os.Stderr, "DEBUG: evmapi.GetSigners returned %d valid signers\n", len(signers))
							return signers, nil
						}
					}

					if err != nil {
						fmt.Fprintf(os.Stderr, "DEBUG: evmapi.GetSigners returned error: %v\n", err)
					} else {
						fmt.Fprintf(os.Stderr, "DEBUG: evmapi.GetSigners returned 0 signers\n")
					}

					// Fallback: Try decoding 'from' field directly using reflection
					fmt.Fprintf(os.Stderr, "DEBUG: Attempting fallback to 'from' field\n")
					md := msg.ProtoReflect().Descriptor()
					fromField := md.Fields().ByName("from")
					if fromField != nil {
						fromVal := msg.ProtoReflect().Get(fromField)
						fromStr := fromVal.String()
						fmt.Fprintf(os.Stderr, "DEBUG: Found 'from' field: %s\n", fromStr)
						
						if fromStr != "" {
							// Try hex decoding
							fromBytes, err := hex.DecodeString(fromStr)
							if err == nil && len(fromBytes) > 0 {
								fmt.Fprintf(os.Stderr, "DEBUG: Successfully decoded 'from' field via hex\n")
								return [][]byte{fromBytes}, nil
							}
							// Try 0x hex decoding
							if len(fromStr) > 2 && fromStr[:2] == "0x" {
								fromBytes, err = hex.DecodeString(fromStr[2:])
								if err == nil && len(fromBytes) > 0 {
									fmt.Fprintf(os.Stderr, "DEBUG: Successfully decoded 'from' field via 0x hex\n")
									return [][]byte{fromBytes}, nil
								}
							}
							// If generic string (e.g. valid bech32? unlikely for EVM msg)
							fmt.Fprintf(os.Stderr, "DEBUG: Failed to hex decode 'from' field\n")
							// Assuming it's already bytes stringified? 
							return [][]byte{[]byte(fromStr)}, nil
						}
					}
					
					fmt.Fprintf(os.Stderr, "ERROR: Failed to extract any signers from MsgEthereumTx\n")
					return nil, nil // Return nil, nil which triggers "tx must have at least one signer"
				},
			},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create interface registry: %v", err))
	}

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(legacyAmino)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)
	etherminttypes.RegisterInterfaces(interfaceRegistry)
	evmtypes.RegisterInterfaces(interfaceRegistry) // Register ExtensionOptionsEthereumTx

	appCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)

	// Create base app with options (including chain ID)
	bApp := baseapp.NewBaseApp(AppName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	
	// Configure custom Mempool with EVM support
	// Use PriorityMempool which allows setting a custom SignerExtractor.
	// SenderNonceMempool was ignoring MsgEthereumTx signers.
	signerExtractor := NewCertSignerExtractor()
	mempoolConfig := mempool.PriorityNonceMempoolConfig[int64]{
		TxPriority:      mempool.NewDefaultTxPriority(),
		SignerExtractor: signerExtractor,
		MaxTx:           5000,
	}
	nonceMempool := mempool.NewPriorityMempool(mempoolConfig)
	bApp.SetMempool(nonceMempool)

	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion("v1.0.0")
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	// Define store keys - only include stores for modules in ModuleManager
	// Note: crisis module is disabled, so we don't include its store key
	// Empty stores cause IAVL v1.x "version does not exist" errors
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		paramstypes.StoreKey,
		consensustypes.StoreKey, // Required for consensus params storage
		attestationtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		minttypes.StoreKey,
		evmtypes.StoreKey,
		feemarkettypes.StoreKey,
		upgradetypes.StoreKey,
		// crisistypes.StoreKey, // Disabled - crisis module not in use
	)
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey, evmtypes.TransientKey)
	memKeys := storetypes.NewMemoryStoreKeys()

	// Create the CertApp instance
	certApp := &CertApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	// Initialize params keeper
	certApp.ParamsKeeper = paramskeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	// Set param subspaces for modules (only for modules in ModuleManager)
	certApp.ParamsKeeper.Subspace(authtypes.ModuleName)
	certApp.ParamsKeeper.Subspace(banktypes.ModuleName)
	certApp.ParamsKeeper.Subspace(stakingtypes.ModuleName)
	certApp.ParamsKeeper.Subspace(slashingtypes.ModuleName)
	certApp.ParamsKeeper.Subspace(govtypes.ModuleName)
	certApp.ParamsKeeper.Subspace(minttypes.ModuleName)
	certApp.ParamsKeeper.Subspace(evmtypes.ModuleName)
	certApp.ParamsKeeper.Subspace(feemarkettypes.ModuleName)
	// certApp.ParamsKeeper.Subspace(crisistypes.ModuleName)

	// Create bech32 address codec for account keeper
	bech32Codec := address.NewBech32Codec(AccountAddressPrefix)

	// Initialize account keeper
	// IMPORTANT: Use etherminttypes.ProtoAccount instead of authtypes.ProtoBaseAccount
	// This creates EthAccount instances that support SetCodeHash for EVM contracts
	certApp.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		etherminttypes.ProtoAccount,
		maccPerms,
		bech32Codec,
		AccountAddressPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Initialize bank keeper
	certApp.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		certApp.AccountKeeper,
		BlockedAddresses(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)

	// Initialize staking keeper with proper address codecs
	// ValidatorAddressCodec uses "certvaloper" prefix for validator addresses
	// Note: validatorAddressCodec was defined earlier in this function for the InterfaceRegistry
	certApp.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		certApp.AccountKeeper,
		certApp.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		validatorAddressCodec,
		certApp.AccountKeeper.AddressCodec(),
	)

	// Initialize consensus keeper (required for consensus params storage)
	certApp.ConsensusKeeper = consensuskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensustypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		runtime.EventService{},
	)

	// Set consensus params keeper on BaseApp (required for InitChain)
	bApp.SetParamStore(certApp.ConsensusKeeper.ParamsStore)

	// Initialize attestation keeper (CERT custom module)
	certApp.AttestationKeeper = attestationkeeper.NewKeeper(
		appCodec,
		keys[attestationtypes.StoreKey],
		nil, // memKey - not used
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

		// Initialize slashing keeper
		certApp.SlashingKeeper = slashingkeeper.NewKeeper(
			appCodec,
			legacyAmino,
			runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
			certApp.StakingKeeper,
			authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		)

		// Wire slashing hooks into staking so validator signing info is created
		// for all validators (including the genesis validator). Without this,
		// x/slashing's BeginBlocker will fail with "no validator signing info
		// found" errors as soon as it tries to handle signatures.
		certApp.StakingKeeper.SetHooks(
			stakingtypes.NewMultiStakingHooks(certApp.SlashingKeeper.Hooks()),
		)
	
	// Initialize gov keeper
	certApp.GovKeeper = *govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[govtypes.StoreKey]),
		certApp.AccountKeeper,
		certApp.BankKeeper,
		certApp.StakingKeeper,
		nil, // distribution keeper not needed for basic gov
		nil, // message router not needed for basic gov
		govtypes.DefaultConfig(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

		// Initialize mint keeper
		certApp.MintKeeper = mintkeeper.NewKeeper(
			appCodec,
			runtime.NewKVStoreService(keys[minttypes.StoreKey]),
			certApp.StakingKeeper,
			certApp.AccountKeeper,
			certApp.BankKeeper,
			authtypes.FeeCollectorName,                               // fee collector module account ("fee_collector")
			authtypes.NewModuleAddress(govtypes.ModuleName).String(), // governance authority
		)

	// Initialize Upgrade Keeper
	certApp.UpgradeKeeper = upgradekeeper.NewKeeper(
		nil, // skip-upgrade-heights
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		"", // upgrade-info-path
		certApp.BaseApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(), // authority
	)

	// Initialize FeeMarket keeper
	certApp.FeeMarketKeeper = feemarketkeeper.NewKeeper(
		appCodec,
		authtypes.NewModuleAddress(govtypes.ModuleName),
		keys[feemarkettypes.StoreKey], // Evmos v20 uses raw StoreKey
		tkeys[evmtypes.TransientKey],
		certApp.GetSubspace(feemarkettypes.ModuleName),
	)

	// Initialize EVM keeper
	certApp.EvmKeeper = evmkeeper.NewKeeper(
		appCodec,
		keys[evmtypes.StoreKey],       // Evmos v20 uses raw StoreKey
		tkeys[evmtypes.TransientKey],
		authtypes.NewModuleAddress(govtypes.ModuleName),
		certApp.AccountKeeper,
		certApp.BankKeeper,
		certApp.StakingKeeper,
		certApp.FeeMarketKeeper,
		nil, // erc20Keeper (nil if module not used)
		"",  // tracer
		certApp.GetSubspace(evmtypes.ModuleName),
	)

	// Initialize crisis keeper
	// certApp.CrisisKeeper = *crisiskeeper.NewKeeper(
	// 	appCodec,
	// 	runtime.NewKVStoreService(keys[crisistypes.StoreKey]),
	// 	1000000000, // invariant check period
	// 	certApp.BankKeeper,
	// 	authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	// 	authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	// 	address.NewBech32Codec(AccountAddressPrefix),
	// )

	// Mount stores
	certApp.MountKVStores(keys)
	certApp.MountTransientStores(tkeys)
	certApp.MountMemoryStores(memKeys)

	// Set the default store loader (required for proper store initialization)
	certApp.SetStoreLoader(baseapp.DefaultStoreLoader)

	// Create module manager with all modules
	// Note: Order matters for genesis initialization
	// IMPORTANT: All modules with store keys MUST be in the module manager
	// to ensure their stores are properly initialized during genesis.
	// Empty stores cause IAVL v1.x "version does not exist" errors.
	certApp.ModuleManager = module.NewManager(
		genutil.NewAppModule(certApp.AccountKeeper, certApp.StakingKeeper, certApp, txConfig),
		auth.NewAppModule(appCodec, certApp.AccountKeeper, nil, nil),
		bank.NewAppModule(appCodec, certApp.BankKeeper, certApp.AccountKeeper, nil),
		staking.NewAppModule(appCodec, certApp.StakingKeeper, certApp.AccountKeeper, certApp.BankKeeper, nil),
		params.NewAppModule(certApp.ParamsKeeper), // Required to initialize params store
		consensus.NewAppModule(appCodec, certApp.ConsensusKeeper),
		attestationmodule.NewAppModule(appCodec, certApp.AttestationKeeper, certApp.AccountKeeper, certApp.BankKeeper),
		slashing.NewAppModule(appCodec, certApp.SlashingKeeper, certApp.AccountKeeper, certApp.BankKeeper, certApp.StakingKeeper, nil, nil),
		gov.NewAppModule(appCodec, &certApp.GovKeeper, certApp.AccountKeeper, certApp.BankKeeper, nil),
		mint.NewAppModule(appCodec, certApp.MintKeeper, certApp.AccountKeeper, nil, nil),
		evm.NewAppModule(certApp.EvmKeeper, certApp.AccountKeeper, certApp.GetSubspace(evmtypes.ModuleName)),
		feemarket.NewAppModule(certApp.FeeMarketKeeper, certApp.GetSubspace(feemarkettypes.ModuleName)),
		upgrade.NewAppModule(certApp.UpgradeKeeper, addressCodec),
		// crisis.NewAppModule(appCodec, &certApp.CrisisKeeper, false),
	)

	// Set order for genesis initialization
	// IMPORTANT: params must be initialized early as other modules depend on it
	certApp.ModuleManager.SetOrderInitGenesis(
		authtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		// crisistypes.ModuleName,
		paramstypes.ModuleName, // Initialize params store to avoid IAVL version issues
		consensustypes.ModuleName,
		genutiltypes.ModuleName,
		attestationtypes.ModuleName,
		evmtypes.ModuleName,
		feemarkettypes.ModuleName,
		upgradetypes.ModuleName,
	)

	// Set order for begin/end blockers
	certApp.ModuleManager.SetOrderBeginBlockers(
		minttypes.ModuleName,
		slashingtypes.ModuleName,
		stakingtypes.ModuleName,
		attestationtypes.ModuleName,
		feemarkettypes.ModuleName,
		evmtypes.ModuleName,
	)
	certApp.ModuleManager.SetOrderEndBlockers(
			govtypes.ModuleName,
			stakingtypes.ModuleName,
			attestationtypes.ModuleName,
			evmtypes.ModuleName,
			feemarkettypes.ModuleName,
			upgradetypes.ModuleName,
	)

	// Register services (message handlers and query handlers)
	// This is REQUIRED in Cosmos SDK v0.50.x for message routing to work
	configurator := module.NewConfigurator(appCodec, bApp.MsgServiceRouter(), bApp.GRPCQueryRouter())
	if err := certApp.ModuleManager.RegisterServices(configurator); err != nil {
		panic(fmt.Sprintf("failed to register module services: %v", err))
	}

	// Set init chainer, begin/end blockers
	certApp.SetInitChainer(certApp.InitChainer)
	certApp.SetPreBlocker(certApp.PreBlocker)
	certApp.SetBeginBlocker(certApp.BeginBlocker)
	certApp.SetEndBlocker(certApp.EndBlocker)

	// Set AnteHandler using Ethermint's logic (handles both Cosmos and EVM txs)
	anteOptions := ante.HandlerOptions{
		AccountKeeper:          certApp.AccountKeeper,
		BankKeeper:             certApp.BankKeeper,
		IBCKeeper:              certApp.IBCKeeper,
		FeeMarketKeeper:        certApp.FeeMarketKeeper,
		EvmKeeper:              certApp.EvmKeeper,
		DistributionKeeper:     MockDistributionKeeper{}, // TODO: Use actual keeper if available, but mock satisfies interface
		StakingKeeper:          certApp.StakingKeeper,
		SignModeHandler:        txConfig.SignModeHandler(),
		SigGasConsumer:         authante.DefaultSigVerificationGasConsumer,
		MaxTxGasWanted:         MaxGasPerBlock,
		ExtensionOptionChecker: etherminttypes.HasDynamicFeeExtensionOption,
		TxFeeChecker:           evmante.NewDynamicFeeChecker(certApp.EvmKeeper),
	}

	anteHandler, err := NewAnteHandler(anteOptions)
	if err != nil {
		panic(fmt.Sprintf("failed to create ante handler: %v", err))
	}

	// Wrap with Debug logic to inspect Msgs
	debugAnteHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		for i, msg := range tx.GetMsgs() {
			if pMsg, ok := msg.(protov2.Message); ok {
				fmt.Fprintf(os.Stderr, "DEBUG: Ante Msg[%d] FullName: %s\n", i, pMsg.ProtoReflect().Descriptor().FullName())
			} else {
				fmt.Fprintf(os.Stderr, "DEBUG: Ante Msg[%d] is not protov2.Message: %T\n", i, msg)
			}
		}
		return anteHandler(ctx, tx, simulate)
	}

	certApp.SetAnteHandler(debugAnteHandler)

	if loadLatest {
		if err := certApp.LoadLatestVersion(); err != nil {
			panic(err)
		}

		// Defensive fix for existing chains that were started before slashing
		// hooks were properly wired into staking. In that scenario, the
		// slashing module never created ValidatorSigningInfo records for
		// validators (including the genesis validator), which causes
		// FinalizeBlock to fail with:
		//   "no validator signing info found".
		//
		// To make the node forward-compatible without requiring operators to
		// wipe chain state, we ensure that every current validator has a
		// signing info entry. This migration is idempotent and runs on every
		// startup after the latest version is loaded.
		if err := certApp.ensureValidatorSigningInfos(); err != nil {
			panic(fmt.Sprintf("failed to ensure validator signing infos: %v", err))
		}
	}

	return certApp
}

// BlockedAddresses returns all the app's blocked module account addresses
func BlockedAddresses() map[string]bool {
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}
	return blockedAddrs
}

// CertSignerExtractor implements mempool.SignerExtractionAdapter to handle both Cosmos and EVM transactions
type CertSignerExtractor struct{
	defaultExtractor mempool.SignerExtractionAdapter
}

func NewCertSignerExtractor() CertSignerExtractor {
	return CertSignerExtractor{
		defaultExtractor: mempool.NewDefaultSignerExtractionAdapter(),
	}
}

func (cse CertSignerExtractor) GetSigners(tx sdk.Tx) ([]mempool.SignerData, error) {
	// DEBUG LOGGING
	fmt.Fprintf(os.Stderr, "DEBUG: CertSignerExtractor.GetSigners called with %d msgs\n", len(tx.GetMsgs()))

	for i, msg := range tx.GetMsgs() {
		fmt.Fprintf(os.Stderr, "DEBUG: Msg[%d] Type: %T\n", i, msg)
		if ethMsg, ok := msg.(*evmtypes.MsgEthereumTx); ok {
			fmt.Fprintf(os.Stderr, "DEBUG: Matched evmtypes.MsgEthereumTx\n")
			
			// Get sender from the message (recovers from signature)
			from := ethMsg.GetFrom()
			fmt.Fprintf(os.Stderr, "DEBUG: MsgEthereumTx From: %s\n", from.String())
			if from.Empty() {
				fmt.Fprintf(os.Stderr, "DEBUG: MsgEthereumTx From is empty, returning nil\n")
				return nil, nil
			}
			
			// Get nonce from transaction data
			txData, err := evmtypes.UnpackTxData(ethMsg.Data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "DEBUG: MsgEthereumTx UnpackTxData failed: %v\n", err)
				return nil, err
			}
			nonce := txData.GetNonce()
			fmt.Fprintf(os.Stderr, "DEBUG: MsgEthereumTx nonce: %d\n", nonce)

			// Return SignerData
			return []mempool.SignerData{{
				Signer:   from,
				Sequence: nonce,
			}}, nil
		}
	}
	
	// Fallback to default for standard Cosmos Transactions
	// Note: Default extractor handles getting Sequence from AuthInfo
	fmt.Fprintf(os.Stderr, "DEBUG: Delegating to DefaultSignerExtractionAdapter\n")
	return cse.defaultExtractor.GetSigners(tx)
}

// LoadHeight loads a particular height from the store
func (app *CertApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ExportAppStateAndValidators exports the state of the application for a genesis file
func (app *CertApp) ExportAppStateAndValidators(
	forZeroHeight bool,
	jailAllowedAddrs []string,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	// Similar to simapp, export the state
	ctx := app.NewContextLegacy(true, sdk.Context{}.BlockHeader())

	// Export genesis state
	genState, err := app.ModuleManager.ExportGenesisForModules(ctx, app.appCodec, modulesToExport)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	appState, err := codec.MarshalJSONIndent(app.legacyAmino, genState)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	// Get validators
	validators, err := staking.WriteValidators(ctx, app.StakingKeeper)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	return servertypes.ExportedApp{
		AppState:        appState,
		Validators:      validators,
		Height:          app.LastBlockHeight(),
		ConsensusParams: app.GetConsensusParams(ctx),
	}, nil
}

// ensureValidatorSigningInfos makes sure every current validator has a
// corresponding x/slashing ValidatorSigningInfo record.
//
// This is primarily a safety net for chains that were started before
// slashing hooks were wired into staking, which would otherwise panic with
// "no validator signing info found" during FinalizeBlock.
//
// The migration is idempotent and safe to run on every startup.
func (app *CertApp) ensureValidatorSigningInfos() error {
	sdkCtx := app.NewContextLegacy(true, sdk.Context{}.BlockHeader())
	ctx := sdk.WrapSDKContext(sdkCtx)

	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return err
	}

	for _, v := range validators {
		consAddr, err := v.GetConsAddr()
		if err != nil {
			return err
		}

		if app.SlashingKeeper.HasValidatorSigningInfo(ctx, consAddr) {
			continue
		}

		info := slashingtypes.NewValidatorSigningInfo(
			consAddr,
			sdkCtx.BlockHeight(), // start height: current block height
			0,                    // index offset
			time.Unix(0, 0),      // jailed until (zero time)
			false,                // tombstoned
			0,                    // missed blocks counter
		)

		if err := app.SlashingKeeper.SetValidatorSigningInfo(ctx, consAddr, info); err != nil {
			return err
		}
	}

	return nil
}

// InitChainer application update at chain initialization
func (app *CertApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState map[string]json.RawMessage
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		return nil, err
	}
	return app.ModuleManager.InitGenesis(ctx, app.appCodec, genesisState)
}

// PreBlocker runs before BeginBlock
func (app *CertApp) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *CertApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *CertApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

// RegisterAPIRoutes registers all application module routes with the provided API server
func (app *CertApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	// Register new tx routes from grpc-gateway
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new CometBFT queries routes from grpc-gateway
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register swagger API from root so that other applications can override easily.
	// This enables the /swagger/swagger.json endpoint used by tooling and docs UIs.
	// If swagger assets are missing in the container, we log the error but do NOT
	// crash the node, so the REST/gRPC gateway can still serve all other routes.
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		fmt.Printf("swagger registration failed: %v\n", err)
	}
}

// RegisterTxService implements the Application.RegisterTxService method
func (app *CertApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method
func (app *CertApp) RegisterTendermintService(clientCtx client.Context) {
	// Use CometABCIWrapper for proper query handling
	cmtApp := server.NewCometABCIWrapper(app)
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		cmtApp.Query,
	)
}

// RegisterNodeService registers the node gRPC service on the provided query router
func (app *CertApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// RegisterGRPCServer registers gRPC services directly with the gRPC server
// This method is called by the Cosmos SDK server when starting the gRPC service
func (app *CertApp) RegisterGRPCServer(grpcServer grpc.Server) {
	// IMPORTANT:
	// We MUST call the embedded BaseApp.RegisterGRPCServer here.
	//
	// BaseApp.RegisterGRPCServer is what copies all gRPC query services registered on
	// the BaseApp's GRPCQueryRouter (via module.Manager.RegisterServices) onto the
	// actual network gRPC server (9090). If we don't call it, the gRPC server will
	// start without module query services and clients will see:
	//   Unimplemented desc = unknown service <pkg>.Query
	app.BaseApp.RegisterGRPCServer(grpcServer)

	// Register attestation module gRPC services directly.
	// The attestation module uses manually-defined Go types (not protoc-generated),
	// so Cosmos SDK's Configurator cannot auto-register them via proto reflection.
	// We bypass the Configurator by registering directly on the gRPC server here.
	attestationtypes.RegisterQueryServer(grpcServer, attestationkeeper.NewQueryServerImpl(app.AttestationKeeper))
	attestationtypes.RegisterMsgServer(grpcServer, attestationkeeper.NewMsgServerImpl(app.AttestationKeeper))
}

// TxConfig returns the app's TxConfig
func (app *CertApp) TxConfig() client.TxConfig {
	return app.txConfig
}

// AppCodec returns the app's codec
func (app *CertApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns the app's interface registry
func (app *CertApp) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetSubspace returns a param subspace for a given module name.
func (app *CertApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, ok := app.ParamsKeeper.GetSubspace(moduleName)
	if !ok {
		panic(fmt.Sprintf("subspace not found: %s", moduleName))
	}
	return subspace
}

