package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/cmd"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/rest"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	app = utils.NewApp("", "PlatONE interaction command line interface")
)

func init() {

	// Initialize the CLI app
	app.Commands = []cli.Command{
		cmd.AccountCmd,  // see cmd_account.go
		cmd.ContractCmd, // see cmd_contract.go
		// AdminCmd,    // see cmd_admin.go, deprecated
		cmd.CnsCmd,       // see cmd_cns.go
		cmd.FwCmd,        // see cmd_firewall.go
		cmd.RoleCmd,      // see cmd_role.go
		cmd.NodeCmd,      // see cmd_node.go
		cmd.SysConfigCmd, // see cmd_sysconfig.go
		cmd.CaCmd,
		StartRest, // see rest
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.After = func(ctx *cli.Context) error {
		return nil
	}

}

var (
	// rest
	StartRest = cli.Command{
		Name:  "rest",
		Usage: "start a restful api server",
		Action: func(ctx *cli.Context) {
			port := ctx.String(cmd.RestPortFlags.Name)
			rest.StartServer(port)
			return
		},
		Flags:    []cli.Flag{cmd.RestPortFlags},
		Category: "rest",
	}
)

//go:generate go-bindata -pkg precompile -o precompiled/bindata.go ../../release/linux/conf/contracts/...
func main() {
	// Initialize the related file
	cmd.ConfigInit()
	/// abiInit()
	/// utl.LogInit()

	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
