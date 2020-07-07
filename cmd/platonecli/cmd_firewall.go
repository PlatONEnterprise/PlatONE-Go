package main

import (
	"fmt"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

const (
	DEFAULT_FIREWALL_FILE_PATH = "./config/fireWall.json"
)

var (
	// fire wall
	FwCmd = cli.Command{
		Name:     "fw",
		Usage:    "Manage contract fire wall",
		Category: "fw",
		Subcommands: []cli.Command{
			FwStatusCmd,
			FwStartCmd,
			FwStopCmd,
			FwExportCmd,
			FwImportCmd,
			FwNewCmd,
			FwDeleteCmd,
			FwResetCmd,
			FwClearCmd,
		},
	}

	FwStartCmd = cli.Command{
		Name:      "start",
		Usage:     "Start the fire wall of an specific contract",
		ArgsUsage: "<address>",
		Action:    fwStart,
		Flags:     globalCmdFlags,
		Description: `
		platonecli fw start <address>`,
	}

	FwStopCmd = cli.Command{
		Name:      "stop",
		Usage:     "Stop the fire wall of an specific contract",
		ArgsUsage: "<address>",
		Action:    fwStop,
		Flags:     globalCmdFlags,
		Description: `
		platonecli fw stop <address>`,
	}

	FwStatusCmd = cli.Command{
		Name:      "query",
		Usage:     "Show the fire wall Info of a contract",
		ArgsUsage: "<address>",
		Action:    fwStatus,
		Flags:     globalCmdFlags,
		Description: `
		platonecli fw query <address>`,
	}

	FwExportCmd = cli.Command{
		Name:      "export",
		Usage:     "Export all the fire wall rules of a contract",
		ArgsUsage: "<address>",
		Action:    fwExport,
		Flags:     fwImportCmdFlags,
		Description: `
		platonecli fw export <address>`,
	}

	FwImportCmd = cli.Command{
		Name:      "import",
		Usage:     "Import fire wall rules to a contract",
		ArgsUsage: "<address>",
		Action:    fwImport,
		Flags:     fwImportCmdFlags,
		Description: `
		platonecli fw import <address>`,
	}

	FwNewCmd = cli.Command{
		Name:      "new",
		Usage:     "New a fire wall rule",
		ArgsUsage: "<address> <action> <account> <api>",
		Action:    fwNew,
		Flags:     globalCmdFlags,
		Description: `
		platonecli fw new <address> <action> <account> <api>

Example: ./platonecli fw new 0xcce493dcb135a19928627a7d5a0df0b1477fbce7 \
accept 0x16c8a21295E68f039B8406d13eE0dc6c3a481C76 function1

The action of the fire wall rules can be either accept or reject.
The * is stand for all account addresses or APIs`,
	}

	FwDeleteCmd = cli.Command{
		Name:      "delete",
		Usage:     "Delete a fire wall rule",
		ArgsUsage: "<address> <action> <account> <api>",
		Action:    fwDelete,
		Flags:     globalCmdFlags,
		Description: `
		platonecli fw delete <address> <action> <account> <api>`,
	}

	FwResetCmd = cli.Command{
		Name:      "reset",
		Usage:     "Reset all the fire wall rules of an action",
		ArgsUsage: "<address> <action> <account> <api>",
		Action:    fwReset,
		Flags:     globalCmdFlags,
		Description: `
		platonecli fw reset <address> <action> <account> <api>`,
	}

	FwClearCmd = cli.Command{
		Name:      "clear",
		Usage:     "Clear all the fire wall rules of an action",
		ArgsUsage: "<address>",
		Action:    fwClear,
		Flags:     fwClearCmdFlags,
		Description: `
		platonecli fw clear <address>`,
	}
)

func fwStart(c *cli.Context) {
	// todo which method is better? method 2 lack of abi
	/// funcName := "__sys_FwOpen"
	/// result := innerCall(c, funcName, []string{}, packet.FW_TX_TYPE)

	addr := c.Args().First()
	result := contractCommon(c, []string{addr}, "__sys_FwOpen", firewallManagementAddress)
	fmt.Printf("result: %s\n", result)
}

func fwStop(c *cli.Context) {
	funcName := "__sys_FwClose"
	result := innerCall(c, funcName, []string{}, fwTxType)
	fmt.Printf("result: %s\n", result)
}

func fwStatus(c *cli.Context) {
	funcName := "__sys_FwStatus"
	result := innerCall(c, funcName, []string{}, fwTxType)
	utl.PrintJson([]byte(result.(string)))
}

func fwExport(c *cli.Context) {
	funcName := "__sys_FwStatus"
	filePath := c.String(FilePathFlags.Name)
	result := innerCall(c, funcName, []string{}, fwTxType)
	utl.WriteFile([]byte(result.(string)), filePath)
}

func fwImport(c *cli.Context) {
	funcName := "__sys_FwImport"
	filePath := c.String(FilePathFlags.Name)

	fileBytes, err := utl.ParseFileToBytes(filePath)
	if err != nil {
		utils.Fatalf(utl.ErrParseFileFormat, "fire wall", err.Error())
	}

	funcParams := []string{string(fileBytes)}
	result := innerCall(c, funcName, funcParams, fwTxType)
	fmt.Printf("result: %s\n", result)
}

func fwCommon(c *cli.Context, funcName string) {

	action := c.Args().Get(1)
	addr := c.Args().Get(2)
	api := c.Args().Get(3)

	utl.ParamValid(action, "action")
	utl.ParamValid(addr, "fw")
	utl.ParamValid(api, "name")

	rules := CombineRule(addr, api) //TODO batch rules
	// stringslice --rule addr1:func1 --rule addr2:func2
	// string --rule addr1:func1|addr2:func2|...
	// string --addr addr1 --api func1

	funcParams := CombineFuncParams(action, rules)
	result := innerCall(c, funcName, funcParams, fwTxType)
	fmt.Printf("result: %s\n", result)
}

func fwNew(c *cli.Context) {
	funcName := "__sys_FwAdd"
	fwCommon(c, funcName)
}

func fwDelete(c *cli.Context) {
	funcName := "__sys_FwDel"
	fwCommon(c, funcName)
}

func fwClear(c *cli.Context) {
	funcName := "__sys_FwClear"
	action := c.String(FwActionFlags.Name)
	all := c.Bool(FwClearAllFlags.Name)

	switch {
	case all: // clear all fire wall rules
		result1 := innerCall(c, funcName, []string{"accept"}, fwTxType)
		result2 := innerCall(c, funcName, []string{"reject"}, fwTxType)
		if result1 == result2 && result1 == "Operation Succeeded" {
			fmt.Printf("result: Operation Succeeded\n")
		} else {
			fmt.Printf("result: something wrong happens, only part of the rules are cleared\n")

		}
	case action != "": // clear the fire wall rules of a specific action
		utl.ParamValid(action, "action")

		funcParams := []string{action}
		result := innerCall(c, funcName, funcParams, fwTxType)
		fmt.Printf("result: %s\n", result)
	default:
		fmt.Printf("no action is specified.\n")
	}
}

func fwReset(c *cli.Context) {
	funcName := "__sys_FwSet"
	fwCommon(c, funcName)
	return
}
