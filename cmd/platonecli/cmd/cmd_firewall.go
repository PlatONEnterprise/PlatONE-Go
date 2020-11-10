package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"

	"github.com/PlatONEnetwork/PlatONE-Go/core/state"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

const (
	defaultFwFilePath = "./fireWall.json"
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
	funcName := "__sys_FwOpen"
	addr := c.Args().First()
	funcParams := []string{addr}

	result := contractCall(c, funcParams, funcName, precompile.FirewallManagementAddress)
	fmt.Printf("result: %s\n", result)
}

func fwStop(c *cli.Context) {
	funcName := "__sys_FwClose"
	addr := c.Args().First()
	funcParams := []string{addr}

	result := contractCall(c, funcParams, funcName, precompile.FirewallManagementAddress)
	fmt.Printf("result: %s\n", result)
}

func fwStatus(c *cli.Context) {
	funcName := "__sys_FwStatus"
	addr := c.Args().First()
	funcParams := []string{addr}

	result := contractCall(c, funcParams, funcName, precompile.FirewallManagementAddress)
	strResult := PrintJson([]byte(result.(string)))
	fmt.Printf("result:\n%s\n", strResult)
}

type ExportFwStatus struct {
	// ContractAddr common.Address
	// Active       bool
	AcceptedList []state.FwElem
	RejectedList []state.FwElem
}

func fwExport(c *cli.Context) {
	funcName := "__sys_FwExport"
	filePath := c.String(FilePathFlags.Name)
	addr := c.Args().First()

	funcParams := []string{addr}
	result := contractCall(c, funcParams, funcName, precompile.FirewallManagementAddress)

	// trim the firewall "ContractAddr" and "Active" tag
	var r = new(ExportFwStatus)
	_ = json.Unmarshal([]byte(result.(string)), r)
	rBytes, err := json.Marshal(r)
	if err != nil{
		log.Fatal(err)
	}
	err = utl.WriteFile(rBytes, filePath)
	if err != nil{
		log.Fatal(err)
	}
	log.Println("Success")
}

func fwImport(c *cli.Context) {
	funcName := "__sys_FwImport"
	filePath := c.String(FilePathFlags.Name)
	addr := c.Args().First()

	fileBytes, err := utl.ParseFileToBytes(filePath)
	fmt.Println(fileBytes)
	if err != nil {
		utils.Fatalf(utl.ErrParseFileFormat, "fire wall", err.Error())
	}

	funcParams := []string{addr, string(fileBytes)}
	result := contractCall(c, funcParams, funcName, precompile.FirewallManagementAddress)
	fmt.Printf("result: %s\n", result)
}

func fwCommon(c *cli.Context, funcName string) {

	addr := c.Args().First()
	action := c.Args().Get(1)
	targetAddr := c.Args().Get(2)
	api := c.Args().Get(3)

	paramValid(action, "action")
	paramValid(targetAddr, "fw")
	paramValid(api, "name")

	rules := cmd_common.CombineRule(targetAddr, api) //TODO batch rules
	// stringslice --rule addr1:func1 --rule addr2:func2
	// string --rule addr1:func1|addr2:func2|...
	// string --addr addr1 --api func1

	funcParams := cmd_common.CombineFuncParams(addr, action, rules)
	result := contractCall(c, funcParams, funcName, precompile.FirewallManagementAddress)
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

func fwClearCommon(c *cli.Context, addr, action string) {
	funcName := "__sys_FwClear"
	paramValid(action, "action")

	funcParams := []string{addr, action}
	result := contractCall(c, funcParams, funcName, precompile.FirewallManagementAddress)
	fmt.Printf("result: clear '%s' rule lists %s\n", action, result)
}

func fwClear(c *cli.Context) {
	// funcName := "__sys_FwClear"
	addr := c.Args().First()
	action := c.String(FwActionFlags.Name)
	all := c.Bool(FwClearAllFlags.Name)

	switch {
	case all: // clear all fire wall rules
		fwClearCommon(c, addr, "reject")
		fwClearCommon(c, addr, "accept")

	case action != "": // clear the fire wall rules of a specific action
		fwClearCommon(c, addr, action)
	default:
		fmt.Printf("no action is specified.\n")
	}
}

func fwReset(c *cli.Context) {
	funcName := "__sys_FwSet"
	fwCommon(c, funcName)
	return
}
