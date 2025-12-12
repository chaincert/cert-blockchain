package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/chaincertify/certd/x/attestation/types"
)

// GetQueryCmd returns the query commands for the attestation module
func GetQueryCmd() *cobra.Command {
	attestationQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the attestation module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	attestationQueryCmd.AddCommand(
		CmdQuerySchema(),
		CmdQueryAttestation(),
		CmdQueryAttestationsByAttester(),
		CmdQueryAttestationsByRecipient(),
		CmdQueryEncryptedAttestation(),
		CmdQueryStats(),
	)

	return attestationQueryCmd
}

// CmdQuerySchema queries a schema by UID
func CmdQuerySchema() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema [uid]",
		Short: "Query a schema by UID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Schema(cmd.Context(), &types.QuerySchemaRequest{
				Uid: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryAttestation queries an attestation by UID
func CmdQueryAttestation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attestation [uid]",
		Short: "Query an attestation by UID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Attestation(cmd.Context(), &types.QueryAttestationRequest{
				Uid: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryAttestationsByAttester queries all attestations by an attester
func CmdQueryAttestationsByAttester() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "by-attester [address]",
		Short: "Query all attestations created by an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.AttestationsByAttester(cmd.Context(), &types.QueryAttestationsByAttesterRequest{
				Attester: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryAttestationsByRecipient queries all attestations for a recipient
func CmdQueryAttestationsByRecipient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "by-recipient [address]",
		Short: "Query all attestations for a recipient",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.AttestationsByRecipient(cmd.Context(), &types.QueryAttestationsByRecipientRequest{
				Recipient: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryEncryptedAttestation queries an encrypted attestation (with access control)
func CmdQueryEncryptedAttestation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encrypted [uid]",
		Short: "Query an encrypted attestation (requires authorization)",
		Long:  `Query an encrypted attestation. Only the attester or authorized recipients can access the encrypted key.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.EncryptedAttestation(cmd.Context(), &types.QueryEncryptedAttestationRequest{
				Uid:       args[0],
				Requester: clientCtx.GetFromAddress().String(),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryStats queries attestation statistics
func CmdQueryStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Query attestation statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Stats(cmd.Context(), &types.QueryStatsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
