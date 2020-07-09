package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
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

	app.After = func(ctx *cli.Context) error {
		return nil
	}

}

//go:generate go-bindata -pkg precompile -o precompiled/bindata.go ../../release/linux/conf/contracts/...
func main() {

	// Initialize the related file
	configInit()
	abiInit()

	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
