package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"

	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient"

	"github.com/PlatONEnetwork/PlatONE-Go/core/types"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/packet"
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
		resultBytes, _ := json.MarshalIndent(result, "", "\t")
		fmt.Printf("result:\n%s\n", resultBytes)
	}
}

// deploy a contract
func deploy(c *cli.Context) {
	var abiBytes []byte
	var consArgs = make([]interface{}, 0)

	codePath := c.Args().First()                      // 必选参数
	abiPath := c.String(ContractAbiFilePathFlag.Name) // 可选参数
	vm := c.String(ContractVmFlags.Name)
	consParams := c.StringSlice(ContractParamFlag.Name)

	codeBytes := cmd_common.ParamParse(codePath, "code").([]byte)
	if abiPath != "" {
		abiBytes = cmd_common.ParamParse(abiPath, "abi").([]byte)
	}
	paramValid(vm, "vm")

	conAbi, _ := packet.ParseAbiFromJson(abiBytes)
	constructor := conAbi.GetConstructor()
	if constructor != nil {
		consArgs, _ = constructor.StringToArgs(consParams)
	}

	/// dataGenerator := packet.NewDeployDataGen(codeBytes, abiBytes, consArgs, vm, types.CreateTxType)
	dataGenerator := packet.NewDeployDataGen(conAbi, types.CreateTxType)
	// set the virtual machine interpreter
	dataGenerator.SetInterpreter(vm, abiBytes, codeBytes, consArgs, constructor)

	result := clientCommonV2(c, dataGenerator, nil)[0]

	fmt.Printf("result:\n%s\n", result.(string))
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

	funcName, funcParams = FuncParse(funcName, funcParams)
	result := contractCallWrap(c, funcParams, funcName, contract)
	for i, data := range result {
		if isTypeLenLong(reflect.ValueOf(data)) {
			fmt.Printf("result%d:\n%+v\n", i, data)
		} else {
			fmt.Printf("result%d:%+v\n", i, data)
		}
	}
}

func isTypeLenLong(v reflect.Value) bool {
	k := v.Kind()
	switch k {
	case reflect.Array, reflect.String, reflect.Slice, reflect.Map, reflect.Chan:
		return v.Len() > 20
	default:
		return false
	}
}

func migrate(c *cli.Context) {

	funcName := "migrate" // 内置
	sourceAddr := c.Args().Get(0)
	targetAddr := c.Args().Get(1) // 必选参数

	paramValid(sourceAddr, "address")

	if targetAddr != "" && sourceAddr != ""{
		paramValid(targetAddr, "address")
		funcParams := cmd_common.CombineFuncParams(sourceAddr, targetAddr)
		result := contractCall(c, funcParams, funcName, precompile.ContractDataProcessorAddress)
		fmt.Printf("%s\n", result)
	} else {
		// future feature
		// txType := MIG_DP_TYPE
		fmt.Println("invalid param")
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

	abiBytes := cmd_common.ParamParse(abiPath, "abi").([]byte)
	//abiBytes := abiParse(abi, contract) //TODO

	return listAbiFunctionsByBytes(abiBytes)
}

func listAbiFunctionsByBytes(abiBytes []byte) (string, error) {
	abiFuncs, err := packet.ParseAbiFromJson(abiBytes)
	if err != nil {
		return "", err
	}

	result := abiFuncs.ListAbiFuncName()
	return result, nil
}
