package main

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
	"os"
	"sort"
)

var (
	app = utils.NewApp("", "the wasm command line interface")
)

func init() {

	// Initialize the CLI app
	app.Commands = []cli.Command{
		// see cmd_account.go
		AccountCmd,
		// see cmd_contract.go
		ContractCmd,
		// see cmd_admin.go
		AdminCmd,
		// see cmd_cns.go
		CnsCmd,
		// see cmd_firewall.go
		FwCmd,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	//app.Flags = append(app.Flags, globalCmdFlags...)
	//app.Flags = append(app.Flags, contractFlags...)

	app.Version = "0.0.1 - Beta"

	//sort.Sort(cli.FlagsByName(app.Flags))

	app.After = func(ctx *cli.Context) error {
		return nil
	}

	//TODO 重新写
	//utl.LogFileSetup()
}

func main() {

	configInit()

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
