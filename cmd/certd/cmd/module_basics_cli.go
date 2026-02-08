package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	clientflags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/types/module"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"

	"github.com/chaincertify/certd/app"
	attestationmodule "github.com/chaincertify/certd/x/attestation"
	attestationcli "github.com/chaincertify/certd/x/attestation/client/cli"
	certidmodule "github.com/chaincertify/certd/x/certid"
	certidcli "github.com/chaincertify/certd/x/certid/client/cli"
)

// stakingModuleBasicCLI overrides the SDK staking module's CLI command builders.
//
// In cosmos-sdk v0.50.x, staking's [`staking.AppModuleBasic.GetTxCmd()`](https://github.com/cosmos/cosmos-sdk/blob/v0.50.5/x/staking/module.go)
// dereferences an internal codec that is not initialized when using a zero-value
// [`staking.AppModuleBasic`](https://github.com/cosmos/cosmos-sdk/blob/v0.50.5/x/staking/module.go#L37).
//
// We embed the original basic module to preserve genesis / interface registration
// behavior and only override CLI command construction to avoid nil dereferences.
type stakingModuleBasicCLI struct {
	staking.AppModuleBasic
}

func (stakingModuleBasicCLI) GetTxCmd() *cobra.Command {
	// Build address codecs using the chain's bech32 prefixes.
	acc := address.NewBech32Codec(app.AccountAddressPrefix)
	val := address.NewBech32Codec(app.AccountAddressPrefix + "valoper")

	return stakingcli.NewTxCmd(val, acc)
}

// Note: In SDK v0.50.x, staking queries are done via gRPC/REST, no CLI query commands

// bankModuleBasicCLI overrides the SDK bank module's CLI command builders.
// The default zero-value [`bank.AppModuleBasic`](https://github.com/cosmos/cosmos-sdk/blob/v0.50.5/x/bank/module.go#L42)
// does not have its address codec initialized, so we provide it here.
type bankModuleBasicCLI struct {
	bank.AppModuleBasic
}

func (bankModuleBasicCLI) GetTxCmd() *cobra.Command {
	acc := address.NewBech32Codec(app.AccountAddressPrefix)
	return bankcli.NewTxCmd(acc)
}

	// GetQueryCmd wires a minimal `query bank` command set, restoring the
	// familiar `certd query bank balances <address>` UX using the bank gRPC
	// query service.
	func (bankModuleBasicCLI) GetQueryCmd() *cobra.Command {
		cmd := &cobra.Command{
			Use:                        banktypes.ModuleName,
			Short:                      "Querying commands for the bank module",
			DisableFlagParsing:         true,
			SuggestionsMinimumDistance: 2,
			RunE:                       client.ValidateCmd,
		}

		cmd.AddCommand(newBankBalancesCmd())

		return cmd
	}

	const bankFlagDenom = "denom"

	// newBankBalancesCmd implements:
	//   certd query bank balances <address> [--denom <denom>]
	//
	// It queries the bank module via gRPC, so it stays consistent with REST
	// and other clients.
	func newBankBalancesCmd() *cobra.Command {
		cmd := &cobra.Command{
			Use:   "balances [address]",
			Short: "Query bank balances for an account",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				clientCtx, err := client.GetClientQueryContext(cmd)
				if err != nil {
					return err
				}

				denom, err := cmd.Flags().GetString(bankFlagDenom)
				if err != nil {
					return err
				}

				queryClient := banktypes.NewQueryClient(clientCtx)

				addr, err := sdk.AccAddressFromBech32(args[0])
				if err != nil {
					return err
				}

				pageReq, err := client.ReadPageRequest(cmd.Flags())
				if err != nil {
					return err
				}

				ctx := cmd.Context()

				if denom == "" {
					res, err := queryClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
						Address:    addr.String(),
						Pagination: pageReq,
					})
					if err != nil {
						return err
					}

					return clientCtx.PrintProto(res)
				}

				res, err := queryClient.Balance(ctx, &banktypes.QueryBalanceRequest{
					Address: addr.String(),
					Denom:   denom,
				})
				if err != nil {
					return err
				}

				return clientCtx.PrintProto(res.Balance)
			},
		}

		cmd.Flags().String(bankFlagDenom, "", "The specific balance denomination to query for")
		clientflags.AddQueryFlagsToCmd(cmd)
		clientflags.AddPaginationFlagsToCmd(cmd, "all balances")

		return cmd
	}

// govModuleBasicCLI overrides the SDK gov module's CLI command builders.
type govModuleBasicCLI struct {
	gov.AppModuleBasic
}

func (govModuleBasicCLI) GetTxCmd() *cobra.Command {
	// In SDK v0.50.x, NewTxCmd takes legacy proposal commands (can be nil/empty)
	return govcli.NewTxCmd(nil)
}

// Note: In SDK v0.50.x, gov queries are done via gRPC/REST, no CLI query commands

// patchAppModuleBasicsForCLI replaces any zero-value SDK module basics that
// would panic (or behave incorrectly) when their CLI commands are constructed.
//
// This keeps changes contained to the daemon CLI wiring layer, while still
// allowing us to call `app.ModuleBasics.AddTxCommands/AddQueryCommands`.
func patchAppModuleBasicsForCLI() {
	for i, b := range app.ModuleBasics {
		switch b.(type) {
		case staking.AppModuleBasic, app.StakingModuleBasicGenesis:
			app.ModuleBasics[i] = stakingModuleBasicCLI{}
		case bank.AppModuleBasic, app.BankModuleBasicGenesis:
			app.ModuleBasics[i] = bankModuleBasicCLI{}
		case gov.AppModuleBasic, app.GovModuleBasicGenesis:
			app.ModuleBasics[i] = govModuleBasicCLI{}
		case attestationmodule.AppModuleBasic:
			app.ModuleBasics[i] = attestationModuleBasicCLI{}
		case certidmodule.AppModuleBasic:
			app.ModuleBasics[i] = certidModuleBasicCLI{}
		}
	}

	// Defensive: ensure we didn't accidentally break the type.
	_ = module.BasicManager(app.ModuleBasics)
}

type attestationModuleBasicCLI struct {
	attestationmodule.AppModuleBasic
}

func (attestationModuleBasicCLI) GetTxCmd() *cobra.Command {
	return attestationcli.GetTxCmd()
}

func (attestationModuleBasicCLI) GetQueryCmd() *cobra.Command {
	return attestationcli.GetQueryCmd()
}

type certidModuleBasicCLI struct {
	certidmodule.AppModuleBasic
}

func (certidModuleBasicCLI) GetTxCmd() *cobra.Command {
	return certidcli.GetTxCmd()
}

func (certidModuleBasicCLI) GetQueryCmd() *cobra.Command {
	return certidcli.GetQueryCmd()
}
