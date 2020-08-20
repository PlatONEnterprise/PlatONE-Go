package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/precompiled"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/platoneclient"

	"github.com/PlatONEnetwork/PlatONE-Go/core/types"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/packet"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"gopkg.in/urfave/cli.v1"
)

// temporary deprecated
/*
// innerCall extract the common parts of the actions of fw and mig calls
func innerCall(c *cli.Context, funcName string, funcParams []string, txType uint64) interface{} {
	addr := c.Args().First()
	to := chainParamConvert(addr, "to").(common.Address)

	call := packet.InnerCallCommon(funcName, funcParams, txType)
	return messageCall(c, call, &to, "")
}*/

// contractCall extract the common parts of the actions of contract execution
func contractCall(c *cli.Context, funcParams []string, funcName, contract string) interface{} {
	result := contractCallWrap(c, funcParams, funcName, contract)
	return result[0]
}

func contractCallWrap(c *cli.Context, funcParams []string, funcName, contract string) []interface{} {
	vm := c.String(ContractVmFlags.Name)
	paramValid(vm, "vm")

	// get the abi bytes of the contracts
	abiPath := c.String(ContractAbiFilePathFlag.Name)
	funcAbi := AbiParse(abiPath, contract)

	// judge whether the input string is contract address or contract name
	cns, err := CnsParse(contract)
	if err != nil {
		utils.Fatalf(err.Error())
	}

	to := chainParamConvert(cns.To, "to").(common.Address)

	dataGenerator := packet.NewContractDataGenWrap(funcName, funcParams, funcAbi, *cns, vm)

	return clientCommon(c, dataGenerator, &to)
}

// todo: rename genTxAndCall?
func clientCommon(c *cli.Context, dataGen packet.MsgDataGen, to *common.Address) []interface{} {

	// get the client global parameters
	account, isSync, isDefault, url := getClientConfig(c)
	pc, err := platoneclient.SetupClient(url)
	if err != nil {
		utils.Fatalf("set up client failed: %s\n", err.Error())
	}

	// form transaction
	tx := getTxParams(c)
	tx.From = account.address
	tx.To = to

	// do message call
	result, isTxHash, err := pc.MessageCall(dataGen, account.keyfile, tx)
	if err != nil {
		utils.Fatalf(err.Error())
	}

	// store default values to config file
	if isDefault && !reflect.ValueOf(result).IsZero() {
		runPath := utl.GetRunningTimePath()
		WriteConfig(runPath+defaultConfigFilePath, config)
	}

	// todo: move isSync from [pc.MessageCall] to here???
	if isSync && isTxHash {
		res, err := pc.GetReceiptByPolling(result[0].(string))
		if err != nil {
			return result
		}

		receiptBytes, _ := json.MarshalIndent(res, "", "\t")
		fmt.Println(string(receiptBytes))

		result[0] = dataGen.ReceiptParsing(res)
	}

	return result
}

// CombineRule combines firewall rules
func CombineRule(addr, api string) string {
	return addr + ":" + api
}

// CombineFuncParams combines the function parameters
func CombineFuncParams(args ...string) []string {
	strArray := append([]string{}, args...)
	return strArray
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
			paramValid(tmp, key)
			temp := ParamParse(tmp, key)
			m[key] = temp
		}
	}

	// required value
	for i, key := range arrayMust {
		m[key] = c.Args().Get(i)
		paramValid(m[key].(string), key)
	}

	if len(m) == 0 {
		utils.Fatalf(utl.ErrInputNullFormat, "json combination result")
	}

	bytes, _ = json.Marshal(m)
	/// utl.Logger.Printf("the combine json is %s\n", bytes)

	return string(bytes)
}

//===============================Abi Parsing========================================
// AbiParse gets the abi bytes by the input parameters provided
// The abi file can be obtained through following ways:
// 1. user provide the abi file path
// 2. abiBytes of precompiled contracts (see precompiled/bindata.go)
// (currently, the following features are not enabled)
// a. get the abi files from default abi file locations
// b. get the abi bytes on chain (wasm contract only).
func AbiParse(abiFilePath, str string) []byte {
	var err error
	var abiBytes []byte

	// todo: equalFold string?
	if p := precompile.List[str]; p != "" {
		precompiledAbi, _ := precompile.Asset(p)
		return precompiledAbi
	}

	if abiFilePath == "" {
		/// abiFilePath = getAbiFileFromLocal(str)
	}

	abiBytes, err = utl.ParseFileToBytes(abiFilePath)
	if err != nil {
		utils.Fatalf(utl.ErrParseFileFormat, "abi", err.Error())
	}

	return abiBytes
}

//===============================User Input Convert=======================================

// convert, convert user input from key to value
type convert struct {
	key1      string      // user input 1
	key2      string      // user input 2
	value1    interface{} // the convert value of user input 1
	value2    interface{} // the convert value of user input 2
	paramName string
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

	return conv.convert(param)
}

func (conv *convert) convert(param string) (interface{}, error) {
	key1NotEqual := !strings.EqualFold(param, conv.key1)
	key2NotEqual := !strings.EqualFold(param, conv.key2)

	if key1NotEqual && key2NotEqual {
		return nil, fmt.Errorf("the %s should be either \"%s\" or \"%s\"", conv.paramName, conv.key1, conv.key2)
	}

	if key2NotEqual {
		return conv.value1, nil
	} else {
		return conv.value2, nil
	}
}

func (conv *convert) parse(param interface{}) string {

	value1NotEqual := param != conv.value1
	value2NotEqual := param != conv.value2

	if value1NotEqual && value2NotEqual {
		panic("not match")
	}

	if value2NotEqual {
		return conv.key1
	} else {
		return conv.key2
	}
}

// 2020.7.6 modified, moved from tx_utils.go
// CnsParse judge whether the input string is contract address or contract name
// and return the corresponding infos
func CnsParse(contract string) (*packet.Cns, error) {
	isAddress := utl.IsNameOrAddress(contract)

	switch isAddress {
	case utl.CnsIsAddress:
		return packet.NewCns(contract, "", types.NormalTxType), nil
	case utl.CnsIsName:
		return packet.NewCns(precompile.CnsInvokeAddress, contract, types.CnsTxType), nil
	case utl.CnsIsUndefined:
		return nil, fmt.Errorf(utl.ErrParamInValidSyntax, "contract address")
	default:
		panic("common.go CnsParse: unexpected error")
	}
}

// ParamParse convert the user inputs to the value needed
func ParamParse(param, paramName string) interface{} {
	var err error
	var i interface{}

	switch paramName {
	case "contract", "user":
		i = utl.IsNameOrAddress(param)
		if i == utl.CnsIsUndefined {
			err = fmt.Errorf(utl.ErrParamInValidSyntax, "contract address")
		}
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

// ChainParamConvert convert the string to chain defined type
func chainParamConvert(param, paramName string) interface{} {
	var err error
	var i interface{}

	switch paramName {
	case "value", "gasPrice":
		i, err = utl.IntValueConvert(param)
	case "gas":
		i, err = utl.UintValueConvert(param)
	case "address", "to", "from":
		i, err = utl.AddressConvert(param)
	default:
		i, err = param, nil //TODO
	}

	if err != nil {
		utils.Fatalf(utl.ErrParamParseFormat, paramName, err.Error())
	}

	return i
}

//===============================User Input Verification=======================================

// OptionParamValid wraps ParamValid, it allows the input to be null
func optionParamValid(param, paramName string) {
	if param != "" {
		paramValid(param, paramName)
	}
}

// ParamValid check if the input is valid
func paramValid(param, paramName string) {
	var valid = true

	switch paramName {
	case "fw":
		if param != "*" {
			valid = utl.IsMatch(param, "address")
		}
	case "to":
		valid = param == "" || utl.IsMatch(param, "address")
	case "contract":
		valid = utl.IsMatch(param, "address") || utl.IsMatch(param, "name")
	case "action":
		valid = strings.EqualFold(param, "accept") || strings.EqualFold(param, "reject")
	case "vm":
		valid = param == "" || strings.EqualFold(param, "evm") || strings.EqualFold(param, "wasm")
	case "ipAddress":
		valid = utl.IsUrl(param)
	case "externalIP", "internalIP":
		valid = utl.IsUrl(param + ":0")
	//case "version":
	//	valid = utl.IsVersion(param)
	case "roles":
		valid = utl.IsValidRoles(param)
	case "email", "mobile", "version", "name", "address", "num":
		valid = utl.IsMatch(param, paramName)
	default:
		/// Logger.Printf("param valid function used but not validate the <%s> param\n", paramName)
	}

	if !valid {
		utils.Fatalf(utl.ErrParamInValidSyntax, paramName)
	}
}
