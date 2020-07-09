package main

import (
	"fmt"
	"strings"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	CnsCmd = cli.Command{
		Name:     "cns",
		Usage:    "Manage Contract Named Service",
		Category: "cns",
		Subcommands: []cli.Command{
			CnsResolveCmd,
			CnsRegisterCmd,
			CnsRedirectCmd,
			CnsQueryCmd,
			CnsStateCmd,
		},
	}

	CnsRegisterCmd = cli.Command{
		Name:      "register",
		Usage:     "Register a contract to the CNS",
		ArgsUsage: "<name> <version> <address>",
		Action:    cnsRegister,
		Flags:     globalCmdFlags,
		Description: `
		platonecli cns register <name> <version> <address>`,
	}

	CnsRedirectCmd = cli.Command{
		Name:      "redirect",
		Usage:     "redirect a contract name in the CNS to another contract address by specifying the version",
		ArgsUsage: "<name> <version>",
		Action:    cnsRedirect,
		Flags:     globalCmdFlags,
		Description: `
		platonecli cns redirect <name> <version>`,
	}

	CnsResolveCmd = cli.Command{
		Name:      "resolve",
		Usage:     "Shows the latest version (default) contract address binded with a name ",
		ArgsUsage: "<name>",
		Action:    cnsResolve,
		Flags:     cnsResolveCmdFlags,
		Description: `
		platonecli cns resolve <name>`,
	}

	CnsQueryCmd = cli.Command{
		Name:   "query",
		Usage:  "Query the CNS Info by the search key provided",
		Action: cnsQuery,
		Flags:  cnsQueryCmdFlags,
		Description: `
		platonecli cns query

List all the data object matching the search key. 
The --all flag has the highest priority than the other flags`,
	}

	CnsStateCmd = cli.Command{
		Name:      "state",
		Usage:     "Show the registration status of a contract name or contract address",
		ArgsUsage: "<contract>",
		Action:    cnsState,
		Flags:     globalCmdFlags,
		Description: `
		platonecli cns state <contract>`,
	}
)

func cnsRegister(c *cli.Context) {
	name := c.Args().First()
	ver := c.Args().Get(1)
	address := c.Args().Get(2)

	utl.ParamValid(name, "name")
	utl.ParamValid(ver, "version")
	utl.ParamValid(address, "address")

	funcParams := CombineFuncParams(name, ver, address)
	result := contractCommon(c, funcParams, "cnsRegister", cnsManagementAddress)
	fmt.Printf("result: %v\n", result)
}

func cnsRedirect(c *cli.Context) {

	name := c.Args().First()
	ver := c.Args().Get(1)

	utl.ParamValid(name, "name")
	utl.ParamValid(ver, "version")

	funcParams := CombineFuncParams(name, ver)
	result := contractCommon(c, funcParams, "cnsRedirect", cnsManagementAddress)
	fmt.Printf("result: %s\n", result)
}

func cnsResolve(c *cli.Context) {

	name := c.Args().First()
	ver := c.String(CnsVersionFlags.Name)

	utl.ParamValid(name, "name")
	if !strings.EqualFold(ver, "latest") {
		utl.ParamValid(ver, "version")
	}

	funcParams := CombineFuncParams(name, ver)
	result := contractCommon(c, funcParams, "getContractAddress", cnsManagementAddress)
	fmt.Printf("result: %s\n", result)

}

// todo: the code and the cmd flags need optimization
func cnsQuery(c *cli.Context) {
	var funcName string
	var result interface{}

	all := c.Bool(ShowAllFlags.Name)
	contract := c.String(ContractIDFlag.Name)
	user := c.String(AddressFlags.Name)
	pageNum := c.String(PageNumFlags.Name)
	pageSize := c.String(PageSizeFlags.Name)

	if user != "" && contract != "" {
		utils.Fatalf("please select one search key")
	}

	switch {
	case all:
		utl.ParamValid(pageNum, "num")
		utl.ParamValid(pageSize, "num")

		funcParams := CombineFuncParams(pageNum, pageSize)
		result = contractCommon(c, funcParams, "getRegisteredContracts", cnsManagementAddress)

	case contract != "":
		isAddress := ParamParse(contract, "contract").(bool)
		if isAddress {
			funcName = "getRegisteredContractsByAddress"
		} else {
			funcName = "getRegisteredContractsByName"
		}

		result = contractCommon(c, []string{contract}, funcName, cnsManagementAddress)

	case user != "":
		utl.ParamValid(user, "address")

		funcName = "getRegisteredContractsByOrigin"
		funcParams := CombineFuncParams(user)
		result = contractCommon(c, funcParams, funcName, cnsManagementAddress)

	default:
		result = "no search key provided!"
	}

	utl.PrintJson([]byte(result.(string)))
}

func cnsState(c *cli.Context) {
	var funcName string
	contract := c.Args().First()

	isAddress := ParamParse(contract, "contract").(bool)
	if isAddress {
		funcName = "ifRegisteredByAddress"
	} else {
		funcName = "ifRegisteredByName"
	}

	funcParams := CombineFuncParams(contract)
	result := contractCommon(c, funcParams, funcName, cnsManagementAddress)

	if result.(int32) == 1 {
		fmt.Printf("result: the contract is registered in CNS\n")
	} else {
		fmt.Printf("result: the contract is not registered in CNS\n")
	}

}
