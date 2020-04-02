package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/core"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	app = utils.NewApp("", "the wasm command line interface")
)

func init() {

	// Initialize the CLI app
	app.Commands = []cli.Command{
		/*
		core.DeployCmd,
		core.InvokeCmd,
		core.CnsInvokeCmd,
		core.CodeGenCmd,
		core.SendTransactionCmd,
		core.SendRawTransactionCmd,
		core.GetTxReceiptCmd,
		core.StabilityCmd,
		core.StabPrepareCmd,
		core.FwInvokeCmd,
		core.MigrateCmd,*/

		core.AccountCmd,
		core.ContractCmd,
		core.AdminCmd,
		core.CnsCmd,
		core.FwCmd,

	}

	app.Flags = []cli.Flag{
		core.AccountCmdFlags,
		core.GasCmdFlags,
		core.GasPriceCmdFlags,
		core.KeystoreCmdFlags,
		core.LocalCmdFlags,
		core.SyncCmdFlags,
	}

	sort.Sort(cli.CommandsByName(app.Commands))
	sort.Sort(cli.FlagsByName(app.Flags))

	app.After = func(ctx *cli.Context) error {
		return nil
	}
}

func main() {

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
