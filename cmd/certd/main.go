package main

import (
	"os"

	"cosmossdk.io/log"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/chaincertify/certd/app"
	"github.com/chaincertify/certd/cmd/certd/cmd"
)

func main() {
	// Set the address prefixes and configuration
	app.SetConfig()

	rootCmd := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		log.NewLogger(os.Stderr).Error("failed to execute root command", "error", err.Error())
		os.Exit(1)
	}
}

