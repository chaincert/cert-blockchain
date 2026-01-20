package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoregistry"

	"cosmossdk.io/log"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/chaincertify/certd/app"
)

// EncodingConfig specifies the concrete encoding types to use for CERT blockchain
type EncodingConfig struct {
	InterfaceRegistry codectypes.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

// MakeEncodingConfig creates the EncodingConfig for the CERT blockchain
// Note: app.SetConfig() must be called before this function to set up bech32 prefixes
func MakeEncodingConfig() EncodingConfig {
	amino := codec.NewLegacyAmino()

	// Create address codecs with the "cert" bech32 prefix
	addressCodec := address.NewBech32Codec(app.AccountAddressPrefix)
	validatorAddressCodec := address.NewBech32Codec(app.AccountAddressPrefix + "valoper")

	// Create interface registry with proper signing options
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

	// Register standard types
	std.RegisterLegacyAminoCodec(amino)
	std.RegisterInterfaces(interfaceRegistry)

	// Register module interfaces
	app.ModuleBasics.RegisterLegacyAminoCodec(amino)
	app.ModuleBasics.RegisterInterfaces(interfaceRegistry)

	cdc := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(cdc, tx.DefaultSignModes)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             cdc,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}

// NewRootCmd creates the root command for the CERT Blockchain daemon
func NewRootCmd() *cobra.Command {
	// Patch module basics so Cosmos SDK module tx/query commands can be safely
	// constructed via app.ModuleBasics.AddTxCommands/AddQueryCommands.
	patchAppModuleBasicsForCLI()

	encodingConfig := MakeEncodingConfig()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("CERT")

	// Store encoding config for use by app creation
	appEncodingConfig = encodingConfig

	// Create logger for app creation
	appLogger = log.NewNopLogger()

	rootCmd := &cobra.Command{
		Use:   app.AppName,
		Short: "CERT Blockchain - Native Privacy Layer-1 Protocol",
		Long: `CERT Blockchain is a specialized Layer-1 protocol engineered for native encrypted attestations.
Built on Cosmos SDK with full EVM compatibility via Ethermint.

Features:
- CometBFT consensus with ~2-second block times
- Full EVM compatibility (MetaMask, Hardhat, Solidity)
- Native encrypted attestation system
- CERT utility token with 1B fixed supply
- IBC interoperability`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customTMConfig := initTendermintConfig()

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customTMConfig)
		},
	}

	initRootCmd(rootCmd, app.ModuleBasics)

	return rootCmd
}

// initAppConfig sets up CERT-specific application configuration
func initAppConfig() (string, interface{}) {
	type CustomAppConfig struct {
		serverconfig.Config

		// CERT-specific configuration
		EVM struct {
			// JSON-RPC server configuration
			Enable     bool   `mapstructure:"enable"`
			Address    string `mapstructure:"address"`
			WsAddress  string `mapstructure:"ws-address"`
			API        string `mapstructure:"api"`
			GasCap     uint64 `mapstructure:"gas-cap"`
			EVMTimeout string `mapstructure:"evm-timeout"`
		} `mapstructure:"json-rpc"`
	}

	srvCfg := serverconfig.DefaultConfig()
	srvCfg.MinGasPrices = "0.0001ucert"

	customAppConfig := CustomAppConfig{
		Config: *srvCfg,
	}

	// Set CERT-specific defaults
	customAppConfig.EVM.Enable = true
	customAppConfig.EVM.Address = "0.0.0.0:8545"
	customAppConfig.EVM.WsAddress = "0.0.0.0:8546"
	customAppConfig.EVM.API = "eth,txpool,personal,net,debug,web3"
	customAppConfig.EVM.GasCap = 25000000
	customAppConfig.EVM.EVMTimeout = "5s"

	customAppTemplate := serverconfig.DefaultConfigTemplate + `
[json-rpc]
# Enable defines if the JSON-RPC server should be enabled.
enable = {{ .EVM.Enable }}

# Address defines the HTTP server to listen on
address = "{{ .EVM.Address }}"

# WsAddress defines the WebSocket server to listen on
ws-address = "{{ .EVM.WsAddress }}"

# API defines a list of JSON-RPC namespaces to be enabled
api = "{{ .EVM.API }}"

# GasCap sets a cap on gas that can be used in eth_call/estimateGas
gas-cap = {{ .EVM.GasCap }}

# EVMTimeout is the timeout for eth_call
evm-timeout = "{{ .EVM.EVMTimeout }}"
`

	return customAppTemplate, customAppConfig
}
