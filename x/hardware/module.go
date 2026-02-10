package hardware

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/chaincertify/certd/x/hardware/keeper"
	"github.com/chaincertify/certd/x/hardware/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

// GenesisState defines the hardware module's genesis state
type GenesisState struct {
	Devices []types.Device `json:"devices"`
}

// AppModuleBasic defines the basic application module
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the module's name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the module's types for legacy amino
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// TODO: Register amino types
}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// TODO: Register interface types
}

// DefaultGenesis returns default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	gs := GenesisState{Devices: []types.Device{}}
	bz, _ := json.Marshal(gs)
	return bz
}

// ValidateGenesis performs genesis state validation
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return gs.Validate()
}

// RegisterGRPCGatewayRoutes registers gRPC Gateway routes
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// TODO: Register gRPC gateway routes
}

// AppModule implements the sdk.AppModule interface
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule instance
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface
func (am AppModule) IsAppModule() {}

// InitGenesis initializes the module's state from genesis
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var gs GenesisState
	if err := json.Unmarshal(data, &gs); err != nil {
		panic(fmt.Sprintf("failed to unmarshal genesis: %v", err))
	}
	InitGenesis(ctx, am.keeper, gs)
}

// ExportGenesis exports the module's state to genesis
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	bz, _ := json.Marshal(gs)
	return bz
}

// ConsensusVersion returns the module's consensus version
func (am AppModule) ConsensusVersion() uint64 {
	return 1
}

// RegisterServices registers module services
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// TODO: Register msg server and query server
}

// Validate validates genesis state
func (gs GenesisState) Validate() error {
	return nil
}

// InitGenesis initializes state from genesis
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs GenesisState) {
	// TODO: Initialize devices from genesis
}

// ExportGenesis exports state to genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) GenesisState {
	return GenesisState{
		Devices: []types.Device{},
	}
}

// BeginBlock is called at the beginning of each block
func (am AppModule) BeginBlock(ctx context.Context) error {
	return nil
}

// EndBlock is called at the end of each block
func (am AppModule) EndBlock(ctx context.Context) error {
	return nil
}

