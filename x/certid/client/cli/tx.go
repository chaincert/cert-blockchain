package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/chaincertify/certd/x/certid/types"
)

// GetTxCmd returns the transaction commands for the certid module
func GetTxCmd() *cobra.Command {
	certidTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "CertID transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	certidTxCmd.AddCommand(
		CmdCreateProfile(),
		CmdUpdateProfile(),
		CmdRegisterHandle(),
	)

	return certidTxCmd
}

func CmdCreateProfile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-profile --name [name] --bio [bio]",
		Short: "Create a new CertID profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			name, _ := cmd.Flags().GetString("name")
			bio, _ := cmd.Flags().GetString("bio")

			msg := types.NewMsgCreateProfile(
				clientCtx.GetFromAddress().String(),
				name,
				bio,
				"", // AvatarCID
				"", // PublicKey
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("name", "", "Display name")
	cmd.Flags().String("bio", "", "Short biography")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdUpdateProfile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-profile --name [name] --bio [bio]",
		Short: "Update an existing CertID profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateProfile(clientCtx.GetFromAddress().String())
			
			name, _ := cmd.Flags().GetString("name")
			if name != "" {
				msg.Name = name
			}
			bio, _ := cmd.Flags().GetString("bio")
			if bio != "" {
				msg.Bio = bio
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("name", "", "Display name")
	cmd.Flags().String("bio", "", "Short biography")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdRegisterHandle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-handle [handle]",
		Short: "Register a unique handle for your profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgRegisterHandle{
				Creator: clientCtx.GetFromAddress().String(),
				Handle:  args[0],
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
