package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/grpc"
	"google.golang.org/protobuf/reflect/protoregistry"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/tx/signing"

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
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
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
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// IBC imports
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	// CERT custom modules
	attestationmodule "github.com/chaincertify/certd/x/attestation"
	attestationkeeper "github.com/chaincertify/certd/x/attestation/keeper"
	attestationtypes "github.com/chaincertify/certd/x/attestation/types"
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
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		params.AppModuleBasic{},
		consensus.AppModuleBasic{},
		// Custom CERT modules
		attestationmodule.AppModuleBasic{},
	)

	// Module account permissions - only for modules in ModuleManager
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
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
		},
	})
	if err != nil {
		panic(err)
	}

	std.RegisterInterfaces(interfaceRegistry)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(appCodec, tx.DefaultSignModes)
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
	interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: protoregistry.GlobalFiles,
		SigningOptions: signing.Options{
			AddressCodec:          addressCodec,
			ValidatorAddressCodec: validatorAddressCodec,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create interface registry: %v", err))
	}

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(legacyAmino)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)

	appCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(appCodec, tx.DefaultSignModes)

	// Create base app with options (including chain ID)
	bApp := baseapp.NewBaseApp(AppName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion("v1.0.0")
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	// Define store keys - only include stores for modules in ModuleManager
	// Removed mint, distribution, slashing, gov as they're not in ModuleManager
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		paramstypes.StoreKey,
		consensustypes.StoreKey, // Required for consensus params storage
		attestationtypes.StoreKey,
	)
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
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

	// Create bech32 address codec for account keeper
	bech32Codec := address.NewBech32Codec(AccountAddressPrefix)

	// Initialize account keeper
	certApp.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
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

	// Mount stores
	certApp.MountKVStores(keys)
	certApp.MountTransientStores(tkeys)
	certApp.MountMemoryStores(memKeys)

	// Create module manager with all modules
	// Note: Order matters for genesis initialization
	certApp.ModuleManager = module.NewManager(
		genutil.NewAppModule(certApp.AccountKeeper, certApp.StakingKeeper, certApp, txConfig),
		auth.NewAppModule(appCodec, certApp.AccountKeeper, nil, nil),
		bank.NewAppModule(appCodec, certApp.BankKeeper, certApp.AccountKeeper, nil),
		staking.NewAppModule(appCodec, certApp.StakingKeeper, certApp.AccountKeeper, certApp.BankKeeper, nil),
		consensus.NewAppModule(appCodec, certApp.ConsensusKeeper),
		attestationmodule.NewAppModule(appCodec, certApp.AttestationKeeper, certApp.AccountKeeper, certApp.BankKeeper),
	)

	// Set order for genesis initialization
	certApp.ModuleManager.SetOrderInitGenesis(
		authtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		consensustypes.ModuleName,
		genutiltypes.ModuleName,
		attestationtypes.ModuleName,
	)

	// Set order for begin/end blockers
	certApp.ModuleManager.SetOrderBeginBlockers(
		stakingtypes.ModuleName,
		attestationtypes.ModuleName,
	)
	certApp.ModuleManager.SetOrderEndBlockers(
		stakingtypes.ModuleName,
		attestationtypes.ModuleName,
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

	if loadLatest {
		if err := certApp.LoadLatestVersion(); err != nil {
			panic(err)
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
	// Module services are already registered via RegisterServices in NewCertApp
	// This method is called by the server to allow apps to register additional gRPC services
	// The SDK automatically routes gRPC-gateway requests to the GRPCQueryRouter
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
