package attestation

import (
	"github.com/spf13/cobra"

	"github.com/chaincertify/certd/x/attestation/client/cli"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

