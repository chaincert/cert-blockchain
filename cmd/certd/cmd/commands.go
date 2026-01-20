package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	dbm "github.com/cosmos/cosmos-db"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	cmttypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	"github.com/chaincertify/certd/app"

	ethermintserver "github.com/evmos/evmos/v20/server"
)

// Package-level variables for app creation
var (
	appEncodingConfig EncodingConfig
	appLogger         log.Logger
)

// newApp creates a new CERT blockchain application instance
func newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	// Get chain ID from server context for baseapp options
	baseAppOptions := server.DefaultBaseappOptions(appOpts)

	return app.NewCertApp(
		logger,
		db,
		traceStore,
		true, // loadLatest
		appOpts,
		baseAppOptions...,
	)
}

// appExport exports the application state for genesis
func appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	// Get chain ID from server context for baseapp options
	baseAppOptions := server.DefaultBaseappOptions(appOpts)

	certApp := app.NewCertApp(
		logger,
		db,
		traceStore,
		height == -1, // loadLatest only if height is -1
		appOpts,
		baseAppOptions...,
	)

	if height != -1 {
		if err := certApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	return certApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

// addModuleInitFlags adds module-specific initialization flags
func addModuleInitFlags(startCmd *cobra.Command) {
	// Add any module-specific init flags here
	// Currently no additional flags needed
}

// GenesisState represents the genesis state map
type GenesisState map[string]json.RawMessage

// initTendermintConfig sets up CometBFT/Tendermint configuration per whitepaper specs
func initTendermintConfig() *cmtcfg.Config {
	cfg := cmtcfg.DefaultConfig()

	// Consensus parameters from Whitepaper Section 4.1
	// Block time target: ~2 seconds
	cfg.Consensus.TimeoutCommit = 2000 * 1_000_000 // 2 seconds in nanoseconds
	cfg.Consensus.TimeoutPropose = 3000 * 1_000_000
	cfg.Consensus.TimeoutProposeDelta = 500 * 1_000_000
	cfg.Consensus.TimeoutPrevote = 1000 * 1_000_000
	cfg.Consensus.TimeoutPrevoteDelta = 500 * 1_000_000
	cfg.Consensus.TimeoutPrecommit = 1000 * 1_000_000
	cfg.Consensus.TimeoutPrecommitDelta = 500 * 1_000_000

	// P2P configuration for network performance
	cfg.P2P.MaxNumInboundPeers = 40
	cfg.P2P.MaxNumOutboundPeers = 10
	cfg.P2P.SendRate = 5120000 // 5 MB/s
	cfg.P2P.RecvRate = 5120000 // 5 MB/s
	cfg.P2P.FlushThrottleTimeout = 100 * 1_000_000

	// Mempool configuration
	cfg.Mempool.Size = 5000
	cfg.Mempool.MaxTxsBytes = 1073741824 // 1 GB
	cfg.Mempool.CacheSize = 10000

	return cfg
}

// certInitCmd creates a custom init command that uses app.NewDefaultGenesisState
func certInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long: `Initialize the files needed for a validator and node to run.

'init' creates a new key named 'validator' (with the bip44 path: 44'/118'/0'/0/0) or
the one specified in --keyring-key-name by --keyring-backend <backend>,
existing keys are renamed by appending '_backup' to the name. The mnemonic
is a 24-word string that allows you to recover and restore your account.

It is better to store the mnemonic as a file rather than plain text.
To encrypt the mnemonic in a file you can run the following command:

	echo "your mnemonic here" | certd keys add <yourKey> --dry-run --output json

Read more about how to use the seed in the following docs:
https://docs.cosmos.network/main/run-node/keyring.html

Note: For nameservice tutorial, when creating the chain for the first time,
the validator key will act as the chain orchestrator (it will create and sign
the GENESIS transactions). When creating a child chain (i.e., a chain created
from a previous state), any key can be used as orchestrator.

Note: The first time you run this command, make sure add the --chain-id flag.
It is recommended that you use as chain-id a unique name with alphanumeric
and - characters that will uniquely identify your blockchain.

Example: cert-testnet-1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			// Set the chain ID from the command line flag
			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainID == "" {
				chainID = app.ChainIDTestnet
			}

			config.SetRoot(clientCtx.HomeDir)

				// Generate genesis file using app.NewDefaultGenesisState
				cmd.Println("DEBUG: Using codec from clientCtx")
				genesisState := app.NewDefaultGenesisState(clientCtx.Codec)
				cmd.Println("DEBUG: Generated genesis state with", len(genesisState), "modules")

				// Marshal app state for inclusion in a full CometBFT genesis document
				appState, err := json.Marshal(genesisState)
				if err != nil {
					return err
				}

				genDoc := &cmttypes.GenesisDoc{
					ChainID:       chainID,
					GenesisTime:   time.Now().UTC(),
					InitialHeight: 1,
					AppState:      appState,
				}

				// Write full genesis document to disk
				genesisFile := config.GenesisFile()
				if err := genDoc.SaveAs(genesisFile); err != nil {
					return err
				}

			// Create private validator file if it doesn't exist
			privValKeyFile := config.PrivValidatorKeyFile()
			privValStateFile := config.PrivValidatorStateFile()

			if !fileExists(privValKeyFile) {
				if err := os.MkdirAll(filepath.Dir(privValKeyFile), 0700); err != nil {
					return err
				}

				pv := privval.GenFilePV(privValKeyFile, privValStateFile)
				pv.Save()
			}

			// Create node key if it doesn't exist
			nodeKeyFile := config.NodeKeyFile()
			if !fileExists(nodeKeyFile) {
				if err := os.MkdirAll(filepath.Dir(nodeKeyFile), 0700); err != nil {
					return err
				}

				if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
					return err
				}
			}

			cmd.Println("Successfully initialized validator node")
			cmd.Println("Genesis file created with CERT-specific parameters")
			cmd.Println("Use 'certd start' to start the node")

			return nil
		},
	}

	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(flags.FlagKeyringDir, "", "The client Keyring directory; if omitted, the default 'home' directory will be used")
	cmd.Flags().String(flags.FlagHome, app.DefaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagOutput, "text", "Output format (text|json)")

	return cmd
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// initRootCmd initializes root command with all subcommands
func initRootCmd(rootCmd *cobra.Command, moduleBasics module.BasicManager) {
	rootCmd.AddCommand(
		certInitCmd(), // Use custom init command instead of genutilcli.InitCmd
		debug.Cmd(),
		pruning.Cmd(newApp, app.DefaultNodeHome),
		snapshot.Cmd(newApp),
	)

	ethermintserver.AddCommands(
		rootCmd,
		ethermintserver.NewDefaultStartOptions(newApp, app.DefaultNodeHome),
		appExport,
		addModuleInitFlags,
	)

	// Add genesis commands (add-genesis-account, gentx, collect-gentxs, validate-genesis)
	rootCmd.AddCommand(genesisCommand(moduleBasics))

	// Add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		queryCommand(),
		txCommand(),
		keys.Commands(),
	)

	// Add CERT-specific commands
	rootCmd.AddCommand(
		CertStatusCmd(),
		ValidatorCmd(),
	)
}

// queryCommand returns the query command group
func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Add base query commands
	// Note: In Cosmos SDK v0.50.x, module queries (bank, auth) are accessed via gRPC/REST
	cmd.AddCommand(
		rpc.QueryEventForTxCmd(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockResultsCmd(),
	)

	// Add module query commands (e.g. `attestation`, `bank`, ...)
	app.ModuleBasics.AddQueryCommands(cmd)

	// Add standard query flags so `certd query --node ... --output json ...` is accepted.
	// Note: module leaf commands may define their own flags as well.
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// txCommand returns the transaction command group
func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Add base tx commands
	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetSimulateCmd(),
	)

	// Add module tx commands (e.g. `attestation`, `bank`, ...)
	app.ModuleBasics.AddTxCommands(cmd)

	// Add standard tx flags so `certd tx ... --chain-id ... --node ... --output json ...` works.
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CertStatusCmd returns the CERT blockchain status command
func CertStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cert-status",
		Short: "Query CERT blockchain status including encrypted attestation stats",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement CERT-specific status including attestation counts
			cmd.Println("CERT Blockchain Status")
			cmd.Println("======================")
			cmd.Println("Chain ID: cert-testnet-1")
			cmd.Println("Max Validators: 80")
			cmd.Println("Block Time: ~2 seconds")
			cmd.Println("Token: CERT (ucert)")
			return nil
		},
	}
}

// ValidatorCmd returns validator management commands
func ValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator",
		Short: "Validator management commands",
	}
	// Subcommands will be added for validator operations
	return cmd
}

// genesisCommand returns the genesis command group with all genesis subcommands
// Uses the SDK's built-in Commands function which includes:
// - add-genesis-account
// - gentx
// - collect-gentxs
// - validate-genesis
// - migrate
func genesisCommand(moduleBasics module.BasicManager) *cobra.Command {
	txConfig := app.GetTxConfig()
	return genutilcli.Commands(txConfig, moduleBasics, app.DefaultNodeHome)
}