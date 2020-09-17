package cmd

import (
	"fmt"
	"strings"

	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"

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

	paramValid(name, "name")
	paramValid(ver, "version")
	paramValid(address, "address")

	funcParams := cmd_common.CombineFuncParams(name, ver, address)
	result := contractCall(c, funcParams, "cnsRegister", precompile.CnsManagementAddress)
	fmt.Printf("%v\n", result)
}

func cnsRedirect(c *cli.Context) {

	name := c.Args().First()
	ver := c.Args().Get(1)

	paramValid(name, "name")
	paramValid(ver, "version")

	funcParams := cmd_common.CombineFuncParams(name, ver)
	result := contractCall(c, funcParams, "cnsRedirect", precompile.CnsManagementAddress)
	fmt.Printf("%s\n", result)
}

func cnsResolve(c *cli.Context) {

	name := c.Args().First()
	ver := c.String(CnsVersionFlags.Name)

	paramValid(name, "name")
	if !strings.EqualFold(ver, "latest") {
		paramValid(ver, "version")
	}

	funcParams := cmd_common.CombineFuncParams(name, ver)
	result := contractCall(c, funcParams, "getContractAddress", precompile.CnsManagementAddress)
	fmt.Printf("%s\n", result)

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

	validFlagsCount := c.NumFlags()

	if pageNum != "0" && pageSize != "0" {
		validFlagsCount -= 1
	}

	if validFlagsCount != 1 {
		utils.Fatalf("please select one and the only one search key")
	}

	switch {
	case all:
		funcParams := cmd_common.CombineFuncParams("0", "0")
		result = contractCall(c, funcParams, "getRegisteredContracts", precompile.CnsManagementAddress)

	case contract != "":
		isAddress := cmd_common.ParamParse(contract, "contract").(int32)
		if isAddress == cmd_common.CnsIsAddress {
			funcName = "getRegisteredContractsByAddress"
		} else {
			funcName = "getRegisteredContractsByName"
		}

		result = contractCall(c, []string{contract}, funcName, precompile.CnsManagementAddress)

	case user != "":
		paramValid(user, "address")

		funcName = "getRegisteredContractsByOrigin"
		funcParams := cmd_common.CombineFuncParams(user)
		result = contractCall(c, funcParams, funcName, precompile.CnsManagementAddress)

	default:
		paramValid(pageNum, "num")
		paramValid(pageSize, "num")

		funcParams := cmd_common.CombineFuncParams(pageNum, pageSize)
		result = contractCall(c, funcParams, "getRegisteredContracts", precompile.CnsManagementAddress)
	}

	strResult := PrintJson([]byte(result.(string)))
	fmt.Printf("result:\n%s\n", strResult)
}

func cnsState(c *cli.Context) {
	var funcName string
	contract := c.Args().First()

	isAddress := cmd_common.ParamParse(contract, "contract").(int32)
	if isAddress == cmd_common.CnsIsAddress {
		funcName = "ifRegisteredByAddress"
	} else {
		funcName = "ifRegisteredByName"
	}

	funcParams := cmd_common.CombineFuncParams(contract)
	result := contractCall(c, funcParams, funcName, precompile.CnsManagementAddress)

	if result.(int32) == 1 {
		fmt.Printf("result: the contract is registered in CNS\n")
	} else {
		fmt.Printf("result: the contract is not registered in CNS\n")
	}

}
