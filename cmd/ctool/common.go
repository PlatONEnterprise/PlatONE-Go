package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/packet"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"gopkg.in/urfave/cli.v1"
	"strconv"
)

// convert, convert user input from key to value
type convert struct {
	key1      string      // user input 1
	key2      string      // user input 2
	value1    interface{} // the convert value of user input 1
	value2    interface{} // the convert value of user input 2
	paramName string
}

// innerCall extract the common parts of the actions of fw and mig calls
func innerCall(c *cli.Context, funcName string, funcParams []string, txType int) interface{} {
	addr := c.Args().First()
	to := utl.ChainParamConvert(addr, "to").(common.Address)

	call := packet.InnerCallCommon(funcName, funcParams, txType)
	return messageCall(c, call, &to, "", call.TxType)
}

// contractCommon extract the common parts of the actions of contract execution
func contractCommon(c *cli.Context, funcParams []string, funcName, contract string) interface{} {
	abiPath := c.String(ContractAbiFilePathFlag.Name)
	vm := c.String("vm")
	value := c.String("value")

	utl.ParamValid(vm, "vm")
	value = utl.ChainParamConvert(value, "value").(string)

	// get the abi bytes of the contracts
	funcAbi := AbiParse(abiPath, contract)

	// judge whether the input string is contract address or contract name
	cns := packet.CnsParse(contract)
	to := utl.ChainParamConvert(cns.To, "to").(common.Address)

	call := packet.ContractCallCommon(funcName, funcParams, funcAbi, *cns, vm)
	return messageCall(c, call, &to, value, call.TxType)
}

// messageCall extract the common parts of the transaction based calls
// including eth_call, eth_sendTransaction, and eth_sendRawTransaction
func messageCall(c *cli.Context, call packet.MessageCall, to *common.Address, value string, txType int) interface{} {

	// get the global parameters
	address, keystore, gas, gasPrice, isSync, isDefault := getGlobalParam(c)
	from := common.HexToAddress(address)

	if call == nil {
		utils.Fatalf("")
	}

	// combine the data based on the types of the calls (contract call, inner call or deploy call)
	data, outputType, isWrite, err := call.CombineData()
	if err != nil {
		utils.Fatalf(utl.ErrPackDataFormat, err.Error())
	}

	// packet the transaction and select the transaction based calls
	tx := packet.NewTxParams(from, to, value, gas, gasPrice, data, txType)
	params, action := tx.SendMode(isWrite, keystore)

	// print the RPC JSON param to the terminal
	utl.PrintRequest(params)

	// send the RPC calls
	resp, err := utl.RpcCalls(action, params)
	if err != nil {
		utils.Fatalf(utl.ErrSendTransacionFormat, err.Error())
	}

	setDefault(address, keystore, isDefault)

	return packet.ParseTxResponse(resp, outputType, isWrite, isSync)
}

// CombineRule combines firewall rules
func CombineRule(addr, api string) string {
	return addr + ":" + api
}

// CombineFuncParams combines the function parameters
func CombineFuncParams(args ...string) []string {
	var strArray []string

	for _, value := range args {
		strArray = append(strArray, value)
	}

	return strArray
}

// getGlobalParam gets and converts the global parameters
func getGlobalParam(c *cli.Context) (string, string, string, string, bool, bool) {

	if c == nil {
		panic("the cli.context pointer is nill")
	}

	// set the url for http request
	setUrl(c)

	// get the global parameters from cli.context
	//TODO 分类规划???
	gas := c.String("gas")
	gasPrice := c.String("gasPrice")

	address := c.String("account")
	keystore := c.String("keystore")
	isLocal := c.Bool("local")
	isDefault := c.Bool("default")

	isSync := !c.Bool("sync")

	// check and covert the global parameters
	utl.OptionParamValid(address, "address")
	keystore = getKeystore(keystore, isLocal)
	if address == "" && keystore == "" {
		address = config.Account
		keystore = config.Keystore
	}

	gas = utl.ChainParamConvert(gas, "gas").(string)
	gasPrice = utl.ChainParamConvert(gasPrice, "gasPrice").(string)

	return address, keystore, gas, gasPrice, isSync, isDefault
}

// setUrl sets the url for http request
func setUrl(c *cli.Context) {
	if c == nil {
		panic("the cli.context pointer is nill")
	}

	url := c.String("url")

	runPath := utl.GetRunningTimePath()

	switch {
	case url != "":
		utl.ParamValid(url, "url")
		config.Url = url
		WriteConfigFile(runPath+DEFAULT_CONFIG_FILE_PATH, "url", config.Url)
	case config.Url == "":
		utils.Fatalf("Please set url first.\n")
	default:
		//utils.Fatalf(utl.PanicUnexpSituation, "SetUrl")
	}

	utl.SetHttpUrl(config.Url)
}

// getKeystore gets the path of keystore file based on the keystore and isLocal flags
func getKeystore(keystore string, isLocal bool) string {
	if isLocal && keystore == "" {
		keystore, _ = utl.GetFileInDirt(DEFAULT_KEYSTORE_DIRT)
	}
	return keystore
}

// setDefault write values of account and keystore to config file if default flag provided
func setDefault(account, keystore string, isDefault bool) {
	if isDefault {

		runPath := utl.GetRunningTimePath()

		WriteConfigFile(runPath+DEFAULT_CONFIG_FILE_PATH, "account", account)
		WriteConfigFile(runPath+DEFAULT_CONFIG_FILE_PATH, "keystore", keystore)
	}
}

// Some of the contract function inputs are in complex json format,
// To simplify the user input, the user only need to input the values of the json keys,
// and the function will packet multiple user inputs into json format
func combineJson(c *cli.Context, arrayMust []string, bytes []byte) string {
	m := make(map[string]interface{}, 0)
	mTemp := make(map[string]interface{}, 0)

	_ = json.Unmarshal(bytes, &mTemp)

	for key := range mTemp {
		// default value
		if mTemp[key] != "" {
			m[key] = mTemp[key]
		}
		// user input
		tmp := c.String(key)
		if tmp != "" {
			utl.ParamValid(tmp, key)
			temp := ParamParse(tmp, key)
			m[key] = temp
		}
	}

	// required value
	for i, key := range arrayMust {
		m[key] = c.Args().Get(i)
		utl.ParamValid(m[key].(string), key)
	}

	if len(m) == 0 {
		utils.Fatalf(utl.ErrInputNullFormat, "json combination result")
	}

	bytes, _ = json.Marshal(m)
	utl.Logger.Printf("the combine json is %s\n", bytes)

	return string(bytes)
}

// ParamParse convert the user inputs to the value needed
func ParamParse(param, paramName string) interface{} {
	var err error
	var i interface{}

	switch paramName {
	case "contract", "user":
		i, err = utl.IsNameOrAddress(param)
	case "delayNum", "p2pPort", "rpcPort":
		if utl.IsInRange(param, 65535) {
			i, err = strconv.ParseInt(param, 10, 0)
		} else {
			err = errors.New("value out of range")
		}
	case "operation", "status", "type":
		i, err = convertSelect(param, paramName)
	case "code", "abi":
		i, err = utl.ParseFileToBytes(param)
	default:
		i, err = param, nil
	}

	if err != nil {
		utils.Fatalf(utl.ErrParamParseFormat, paramName, err.Error())
	}

	return i
}

// Some of the contract function inputs are numbers,
// these numbers are hard for users to remember the meanings behind them,
// Thus, to simplify the user input, we convert the meaningful strings to number automatically
// For example, if user input: "valid", the converter will convert the string to 1
func newConvert(key1, key2 string, value1, value2 interface{}, paramName string) *convert {
	return &convert{
		key1:      key1,
		key2:      key2,
		value1:    value1,
		value2:    value2,
		paramName: paramName,
	}
}

func convertSelect(param, paramName string) (interface{}, error) {
	var conv *convert

	switch paramName {
	case "operation": // registration operation
		conv = newConvert("approve", "reject", "2", "3", paramName)
	case "status": // node status
		conv = newConvert("valid", "invalid", 1, 2, paramName)
	case "type": // node type
		conv = newConvert("consensus", "observer", 1, 0, paramName)
	default:
		utils.Fatalf("")
	}

	return conv.typeConvert(param)
}

func (conv *convert) typeConvert(param string) (interface{}, error) {
	if param != conv.key1 && param != conv.key2 {
		return nil, fmt.Errorf("the %s should be either \"%s\" or \"%s\"", conv.paramName, conv.key1, conv.key2)
	}

	if param == conv.key1 {
		return conv.value1, nil
	} else {
		return conv.value2, nil
	}
}
