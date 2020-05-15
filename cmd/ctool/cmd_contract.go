package main

import (
	"encoding/json"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/packet"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
	"strings"
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
		ctool contract deploy <codeFile>

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
		ctool contract execute <contract> <function>`,
	}

	MigrateCmd = cli.Command{
		Name:      "migrate",
		Usage:     "Migrate the storage trie of a contract to a new address",
		ArgsUsage: "<address> <to>",
		Action:    migrate,
		Flags:     globalCmdFlags,
		Description: `
		ctool contract migrate <address> <to>`,
	}

	MethodCmd = cli.Command{
		Name:   "methods",
		Usage:  "List all the exported methods of a contract by its abi file or contract address",
		Action: contractMethods,
		Flags:  contractMethodsCmd,
		Description: `
		ctool contract methods

To list the methods of a contract by contract address, the abi file of the contract
should be stored in the ./abi with the name of its address`,
	}

	ReceiptCmd = cli.Command{
		Name:      "receipt",
		Usage:     "Get the transaction receipt by transaction hash",
		ArgsUsage: "<tx hash>",
		Action:    contractReceipt,
		Description: `
		ctool contract receipt <tx hash>

Get the full information of the transaction receipt by transaction hash`,
	}
)

func contractReceipt(c *cli.Context) {

	//TODO 是否还能优化？
	setUrl(c)

	txHash := c.Args().First()
	//paramValid(txHash, "txHash") //TODO

	result, err := utl.GetTransactionReceipt(txHash)
	if err != nil {
		utils.Fatalf("get receipt failed: %s\n", err.Error())
	} else {
		resultBytes, _ := json.Marshal(result)
		utl.PrintJson(resultBytes)
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
	utl.ParamValid(vm, "vm")

	call := packet.NewDeployCall(codeBytes, abiBytes, vm, packet.DEPLOY_CONTRACT)

	result := messageCall(c, call, nil, "", call.TxType)
	fmt.Printf("result: contract address is %s\n", result)

	if utl.IsMatch(result.(string), "address") {
		storeAbiFile(result.(string), abiBytes)
	}
}

// execute a method in the contract(evm or wasm).
func execute(c *cli.Context) {

	contract := c.Args().First()
	funcName := c.Args().Get(1)
	funcParams := c.StringSlice("param")
	isListMethods := c.Bool("methods")

	utl.ParamValid(contract, "contract")

	if isListMethods {
		abiPath := getAbiFile(contract)
		_ = listAbiFunctions(abiPath)
		return
	}

	//TODO bug fix
	/*
		if len(c.Args()) != 2 {
			utils.Fatalf("param check error, required %d inputs, recieved %d\n",2, len(c.Args()))
		}*/

	result := contractCommon(c, funcParams, funcName, contract)
	fmt.Printf("result: %v\n", result)
	//utl.PrintJson([]byte(result.(string))) //TODO
}

//TODO test
func migrate(c *cli.Context) {

	funcName := "migrateFrom"     // 内置
	sourceAddr := c.Args().Get(1) // 必选参数

	if sourceAddr != "" {
		utl.ParamValid(sourceAddr, "address")
		funcParams := CombineFuncParams(sourceAddr)
		result := innerCall(c, funcName, funcParams, packet.MIG_TX_TYPE)
		fmt.Printf("result: %s\n", result)
	} else {
		// future feature
		// txType := MIG_DP_TYPE
	}
}

func contractMethods(c *cli.Context) {
	var abiPath string

	abi := c.String(ContractAbiFilePathFlag.Name)
	contract := c.String("contract")

	utl.ParamValid(contract, "address")

	switch {
	case abi != "":
		abiPath = abi
	case contract != "":
		abiPath = getAbiFile(contract)
	default:
		utils.Fatalf("no argument provided\n")
	}

	err := listAbiFunctions(abiPath)
	if err != nil {
		utils.Fatalf("list contract methods error: %s\n", err.Error())
	}
}

func listAbiFunctions(abiPath string) error {
	var strInput []string
	var strOutput []string

	if abiPath == "" {
		return fmt.Errorf("the abi file is not found\n")
	}

	abiBytes := ParamParse(abiPath, "abi").([]byte)
	//abiBytes := abiParse(abi, contract) //TODO

	abiFuncs, err := packet.ParseAbiFromJson(abiBytes)
	if err != nil {
		return err
	}

	fmt.Printf("-------------------contract methods list------------------------\n")

	for i, function := range abiFuncs {
		strInput = []string{}
		strOutput = []string{}
		for _, param := range function.Inputs {
			strInput = append(strInput, param.Name+" "+param.Type)
		}
		for _, param := range function.Outputs {
			strOutput = append(strOutput, param.Name+" "+param.Type)
		}
		fmt.Printf("Method %d:", i+1)
		fmt.Printf("%s(%s)%s\n", function.Name, strings.Join(strInput, ","), strings.Join(strOutput, ","))
	}

	return nil
}
