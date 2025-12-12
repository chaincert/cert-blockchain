package cmd

import (
	"encoding/json"
	"io"

	"github.com/spf13/cobra"

	dbm "github.com/cosmos/cosmos-db"

	cmtcfg "github.com/cometbft/cometbft/config"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client/debug"
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

// initRootCmd initializes root command with all subcommands
func initRootCmd(rootCmd *cobra.Command, moduleBasics module.BasicManager) {
	rootCmd.AddCommand(
		genutilcli.InitCmd(moduleBasics, app.DefaultNodeHome),
		debug.Cmd(),
		pruning.Cmd(newApp, app.DefaultNodeHome),
		snapshot.Cmd(newApp),
	)

	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, appExport, addModuleInitFlags)

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
		RunE:                       nil,
	}

	cmd.AddCommand(
		rpc.QueryEventForTxCmd(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockResultsCmd(),
	)

	return cmd
}

// txCommand returns the transaction command group
func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       nil,
	}

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
