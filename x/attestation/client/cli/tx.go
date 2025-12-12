package cli

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/chaincertify/certd/x/attestation/types"
)

// GetTxCmd returns the transaction commands for the attestation module
func GetTxCmd() *cobra.Command {
	attestationTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Attestation transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	attestationTxCmd.AddCommand(
		CmdRegisterSchema(),
		CmdAttest(),
		CmdRevoke(),
		CmdCreateEncryptedAttestation(),
	)

	return attestationTxCmd
}

// CmdRegisterSchema returns the command for registering a new schema
func CmdRegisterSchema() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-schema [schema] --revocable [true/false]",
		Short: "Register a new attestation schema",
		Long: `Register a new attestation schema for use with the EAS protocol.
Example schema format: "string name, uint256 age, address wallet"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			schema := args[0]
			revocable, _ := cmd.Flags().GetBool("revocable")
			resolver, _ := cmd.Flags().GetString("resolver")

			msg := types.NewMsgRegisterSchema(
				clientCtx.GetFromAddress().String(),
				schema,
				resolver,
				revocable,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool("revocable", true, "Whether attestations using this schema can be revoked")
	cmd.Flags().String("resolver", "", "Optional resolver contract address")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdAttest returns the command for creating a public attestation
func CmdAttest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attest [schema-uid] [data-hex] --recipient [address]",
		Short: "Create a new public attestation",
		Long:  `Create a new public attestation using the specified schema.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			schemaUID := args[0]
			dataHex := args[1]

			data, err := hex.DecodeString(dataHex)
			if err != nil {
				return fmt.Errorf("invalid data hex: %w", err)
			}

			recipient, _ := cmd.Flags().GetString("recipient")
			expirationTime, _ := cmd.Flags().GetInt64("expiration")
			revocable, _ := cmd.Flags().GetBool("revocable")
			refUID, _ := cmd.Flags().GetString("ref-uid")

			msg := types.NewMsgAttest(
				clientCtx.GetFromAddress().String(),
				schemaUID,
				recipient,
				expirationTime,
				revocable,
				refUID,
				data,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("recipient", "", "Recipient address")
	cmd.Flags().Int64("expiration", 0, "Expiration timestamp (0 = never)")
	cmd.Flags().Bool("revocable", true, "Whether this attestation can be revoked")
	cmd.Flags().String("ref-uid", "", "Reference to another attestation UID")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdRevoke returns the command for revoking an attestation
func CmdRevoke() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [attestation-uid]",
		Short: "Revoke an existing attestation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRevoke(
				clientCtx.GetFromAddress().String(),
				args[0],
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCreateEncryptedAttestation returns the command for creating an encrypted attestation
func CmdCreateEncryptedAttestation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-encrypted [schema-uid] [ipfs-cid] [encrypted-hash] [recipients] [keys-file]",
		Short: "Create a new encrypted attestation",
		Long: `Create a new encrypted attestation with IPFS-stored encrypted data.
Recipients should be comma-separated addresses.
Keys file should be JSON with recipient->encrypted_key mapping.`,
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			schemaUID := args[0]
			ipfsCID := args[1]
			encryptedHash := args[2]
			recipients := strings.Split(args[3], ",")
			keysFile := args[4]

			// Read keys file - this is a placeholder for the actual implementation
			// In production, this would be handled by the SDK client library
			_ = keysFile

			msg := types.NewMsgCreateEncryptedAttestation(
				clientCtx.GetFromAddress().String(),
				schemaUID,
				ipfsCID,
				encryptedHash,
				recipients,
				nil, // Keys would be loaded from file
				true,
				0,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool("revocable", true, "Whether this attestation can be revoked")
	cmd.Flags().Int64("expiration", 0, "Expiration timestamp (0 = never)")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
