package main

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/packet"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	// cns
	CnsCmd = cli.Command{
		Name:     "cns",
		Usage:    "Manage Contract Named Service",
		Category: "cns",
		Subcommands: []cli.Command{
			CnsResolveCmd,
			CnsRegisterCmd,
			CnsUnregisterCmd,
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
		ctool cns register <name> <version> <address>`,
	}

	CnsUnregisterCmd = cli.Command{
		Name:      "unregister",
		Usage:     "Unregister a contract name in the CNS by specifying the version",
		ArgsUsage: "<name>",
		Action:    cnsUnregister,
		Flags:     cnsResolveCmdFlags,
		Description: `
		ctool cns unregister <name>`,
	}

	CnsResolveCmd = cli.Command{
		Name:      "resolve",
		Usage:     "Shows the latest version (default) contract address binded with a name ",
		ArgsUsage: "<name>",
		Action:    cnsResolve,
		Flags:     cnsResolveCmdFlags,
		Description: `
		ctool cns resolve <name>`,
	}

	//TODO 仔细梳理
	CnsQueryCmd = cli.Command{
		Name:   "query",
		Usage:  "Query the CNS Info by the search key provided",
		Action: cnsQuery,
		Flags:  cnsQueryCmdFlags,
		Description: `
		ctool cns query

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
		ctool cns state <contract>`,
	}
)

func cnsRegister(c *cli.Context) {
	name := c.Args().First()
	ver := c.Args().Get(1)
	address := c.Args().Get(2)

	//paramNumCheck(3,len(c.Args()))
	utl.ParamValid(name, "name")
	utl.ParamValid(ver, "version")
	utl.ParamValid(address, "address")

	funcParams := CombineFuncParams(name, ver, address)

	// __sys_CnsManager不能进行名字访问 未注册
	result := contractCommon(c, funcParams, "cnsRegister", packet.CNS_PROXY_ADDRESS)
	fmt.Printf("result: %v\n", result)
}

func cnsUnregister(c *cli.Context) {

	name := c.Args().First()
	ver := c.String(CnsVersionFlags.Name)

	//paramNumCheck(2,len(c.Args()))
	utl.ParamValid(name, "name")
	if ver != "latest" {
		utl.ParamValid(ver, "version")
	}

	funcParams := CombineFuncParams(name, ver)

	// __sys_CnsManager不能进行名字访问 未注册
	result := contractCommon(c, funcParams, "cnsUnregister", packet.CNS_PROXY_ADDRESS)
	fmt.Printf("result: %s\n", result)
}

func cnsResolve(c *cli.Context) {

	name := c.Args().First()
	ver := c.String(CnsVersionFlags.Name)

	utl.ParamValid(name, "name")
	if ver != "latest" {
		utl.ParamValid(ver, "version")
	}

	funcParams := CombineFuncParams(name, ver)

	// __sys_CnsManager不能进行名字访问 未注册
	result := contractCommon(c, funcParams, "getContractAddress", packet.CNS_PROXY_ADDRESS)
	fmt.Printf("result: %s\n", result)

}

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
		result = contractCommon(c, funcParams, "getRegisteredContracts", packet.CNS_PROXY_ADDRESS)
	case contract != "":
		isAddress := ParamParse(contract, "contract").(bool)
		if isAddress {
			funcName = "getContractInfoByAddress"
		} else {
			funcName = "getHistoryContractsByName"
		}

		result = contractCommon(c, []string{contract}, funcName, packet.CNS_PROXY_ADDRESS)
	case user != "":
		utl.ParamValid(user, "address")
		utl.ParamValid(pageNum, "num")
		utl.ParamValid(pageSize, "num")

		funcName = "getRegisteredContractsByAddress"
		funcParams := CombineFuncParams(user, pageNum, pageSize)

		result = contractCommon(c, funcParams, funcName, packet.CNS_PROXY_ADDRESS)
	default:
		result = "no search key provided!"
	}

	// fmt.Printf("result: %s\n", result)
	utl.PrintJson([]byte(result.(string)))
}

func cnsState(c *cli.Context) {
	var funcName string
	contract := c.Args().First()

	// __sys_CnsManager不能进行名字访问 未注册
	isAddress := ParamParse(contract, "contract").(bool)
	if isAddress {
		funcName = "ifRegisteredByAddress"
	} else {
		funcName = "ifRegisteredByName"
	}

	funcParams := CombineFuncParams(contract)
	result := contractCommon(c, funcParams, funcName, packet.CNS_PROXY_ADDRESS)

	if result.(int32) == 1 {
		fmt.Printf("result: the contract is registered in CNS\n")
	} else {
		fmt.Printf("result: the contract is not registered in CNS\n")
	}

}
