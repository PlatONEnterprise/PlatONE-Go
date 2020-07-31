package main

import (
	"encoding/json"
	"fmt"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/precompiled"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/platoneclient"

	"github.com/PlatONEnetwork/PlatONE-Go/core/types"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/packet"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	// contract
	ContractCmd = cli.Command{
		Name:      "contract",
		Usage:     "Manage contracts",
		ArgsUsage: "",
		Category:  "contract",
		Subcommands: []cli.Command{
			ExecuteCmd,
			MethodCmd,
			MigrateCmd,
			DeployCmd,
			ReceiptCmd,
		},
		Description: `
Currently PlatONE supports two types of contracts:  wasm and evm 
Use --vm flag to select the correct interpreter when deploy and 
execute contracts on PlatONE`,
	}

	DeployCmd = cli.Command{
		Name:      "deploy",
		Usage:     "Deploy a wasm or evm contract",
		ArgsUsage: "<codeFile>",
		Action:    deploy,
		Flags:     contractDeployCmdFlags,
		Description: `
		platonecli contract deploy <codeFile>

Modify the genesis.json file 'interpreter' tag to support evm contract deployment
Evm: only support byzantium version 
Wasm: --abi flag is also needed when deploy wasm contract`,
	}

	ExecuteCmd = cli.Command{
		Name:      "execute",
		Usage:     "Execute a method of contract",
		ArgsUsage: "<contract> <function>",
		Action:    execute,
		Flags:     contractExecuteCmdFlags,
		Description: `
		platonecli contract execute <contract> <function>`,
	}

	MigrateCmd = cli.Command{
		Name:      "migrate",
		Usage:     "Migrate the storage trie of a contract to a new address",
		ArgsUsage: "<address> <to>",
		Action:    migrate,
		Flags:     globalCmdFlags,
		Description: `
		platonecli contract migrate <address> <to>`,
	}

	MethodCmd = cli.Command{
		Name:   "methods",
		Usage:  "List all the exported methods of a contract by its abi file or contract address",
		Action: contractMethods,
		Flags:  contractMethodsCmd,
		Description: `
		platonecli contract methods

To list the methods of a contract by contract address, the abi file of the contract
should be stored in the ./abi with the name of its address`,
	}

	ReceiptCmd = cli.Command{
		Name:      "receipt",
		Usage:     "Get the transaction receipt by transaction hash",
		ArgsUsage: "<tx hash>",
		Action:    contractReceipt,
		Description: `
		platonecli contract receipt <tx hash>

Get the full information of the transaction receipt by transaction hash`,
	}
)

func contractReceipt(c *cli.Context) {

	url := getUrl(c)
	client, err := platoneclient.SetupClient(url)
	if err != nil {
		utils.Fatalf("set up client failed: %s\n", err.Error())
	}

	txHash := c.Args().First()

	result, err := client.GetTransactionReceipt(txHash)
	if result == nil {
		fmt.Printf("the tx receipt is not generated, please try again later\n")
	}

	if err != nil {
		utils.Fatalf("get receipt failed: %s\n", err.Error())
	} else {
		resultBytes, _ := json.Marshal(result)
		strResult := utl.PrintJson(resultBytes)
		fmt.Printf("result:\n%s\n", strResult)
	}
}

// deploy a contract
func deploy(c *cli.Context) {
	var abiBytes []byte

	codePath := c.Args().First()                      // 必选参数
	abiPath := c.String(ContractAbiFilePathFlag.Name) // 可选参数
	vm := c.String(ContractVmFlags.Name)

	codeBytes := ParamParse(codePath, "code").([]byte)
	if abiPath != "" {
		abiBytes = ParamParse(abiPath, "abi").([]byte)
	}
	paramValid(vm, "vm")

	dataGenerator := packet.NewDeployDataGen(codeBytes, abiBytes, vm, types.CreateTxType)
	result := clientCommon(c, dataGenerator, nil)

	if utl.IsMatch(result.(string), "address") {
		/// storeAbiFile(result.(string), abiBytes)
		fmt.Printf("result: contract address is %s\n", result)
	} else {
		fmt.Printf("result: %s\n", result)
	}
}

// execute a method in the contract(evm or wasm).
func execute(c *cli.Context) {

	contract := c.Args().First()
	funcName := c.Args().Get(1)
	funcParams := c.StringSlice(ContractParamFlag.Name)
	isListMethods := c.Bool(ShowContractMethodsFlag.Name)

	paramValid(contract, "contract")

	if isListMethods {
		contractMethods(c)
		return
	}

	result := contractCall(c, funcParams, funcName, contract)
	fmt.Printf(" %v\n", result)
	//utl.PrintJson([]byte(result.(string))) //TODO
}

//TODO test
func migrate(c *cli.Context) {

	funcName := "migrateFrom" // 内置
	sourceAddr := c.Args().Get(0)
	targetAddr := c.Args().Get(1) // 必选参数

	paramValid(sourceAddr, "address")

	if targetAddr != "" {
		paramValid(targetAddr, "address")
		funcParams := CombineFuncParams(sourceAddr, targetAddr)
		result := contractCall(c, funcParams, funcName, precompile.ContractDataProcessorAddress)
		fmt.Printf("%s\n", result)
	} else {
		// future feature
		// txType := MIG_DP_TYPE
	}
}

func contractMethods(c *cli.Context) {
	var abiPath string

	abi := c.String(ContractAbiFilePathFlag.Name)
	// contract := c.String(ContractIDFlag.Name)

	switch {
	case abi != "":
		abiPath = abi
	// currently deprecated, used when file_abi.go is enabled
	/*
		case contract != "":
			paramValid(contract, "address")
			abiPath = getAbiFile(contract)*/
	default:
		utils.Fatalf("no argument provided\n")
	}

	result, err := listAbiFunctions(abiPath)
	if err != nil {
		utils.Fatalf("list contract methods error: %s\n", err.Error())
	}

	fmt.Printf(result)
}

func listAbiFunctions(abiPath string) (string, error) {

	if abiPath == "" {
		return "", fmt.Errorf("the abi file is not found\n")
	}

	abiBytes := ParamParse(abiPath, "abi").([]byte)
	//abiBytes := abiParse(abi, contract) //TODO

	return listAbiFunctionsByBytes(abiBytes)
}

func listAbiFunctionsByBytes(abiBytes []byte) (string, error) {
	abiFuncs, err := packet.ParseAbiFromJson(abiBytes)
	if err != nil {
		return "", err
	}

	result := packet.ListAbiFuncName(abiFuncs)
	return result, nil
}
