package cmd_common

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/packet"
	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/precompiled"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
)

// CombineRule combines firewall rules
func CombineRule(addr, api string) string {
	return addr + ":" + api
}

// CombineFuncParams combines the function parameters
func CombineFuncParams(args ...string) []string {
	strArray := append([]string{}, args...)
	return strArray
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

//===============================User Input Convert=======================================

// convert, convert user input from key to value
type Convert struct {
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
func NewConvert(key1, key2 string, value1, value2 interface{}, paramName string) *Convert {
	return &Convert{
		key1:      key1,
		key2:      key2,
		value1:    value1,
		value2:    value2,
		paramName: paramName,
	}
}

func ConvertSelect(param, paramName string) (interface{}, error) {
	var conv *Convert

	switch paramName {
	case "operation": // registration operation
		conv = NewConvert("approve", "reject", "2", "3", paramName)
	case "status": // node status
		conv = NewConvert("valid", "invalid", 1, 2, paramName)
	case "type": // node type
		conv = NewConvert("consensus", "observer", 1, 0, paramName)
	default:
		utils.Fatalf("")
	}

	return conv.Convert(param)
}

func (conv *Convert) Convert(param string) (interface{}, error) {
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

func (conv *Convert) parse(param interface{}) string {

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
		i, err = ConvertSelect(param, paramName)
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
func ChainParamConvert(param, paramName string) interface{} {
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
func OptionParamValid(param, paramName string) {
	if param != "" {
		ParamValid(param, paramName)
	}
}

func ParamValidWrap(param, paramName string) bool {
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
	case "role":
		valid = utl.IsRoleMatch(param)
	case "roles":
		valid = utl.IsValidRoles(param)
	case "email", "mobile", "version", "num":
		valid = utl.IsMatch(param, paramName)

	// newly added for restful server
	// todo; fix the toLower problem
	case "orgin", "address":
		valid = utl.IsMatch(param, "address")
	case "contractname", "name":
		valid = utl.IsMatch(param, "name")
	case "sysparam":
		valid = strings.EqualFold(param, "0") || strings.EqualFold(param, "1")
	case "blockgaslimit":
		num, err := strconv.ParseUint(param, 10, 0)
		if err != nil {
			return false
		}
		valid = vm.BlockGasLimitMinValue <= num && vm.BlockGasLimitMaxValue >= num
	case "txgaslimit":
		num, err := strconv.ParseUint(param, 10, 0)
		if err != nil {
			return false
		}
		valid = vm.TxGasLimitMinValue <= num && vm.TxGasLimitMaxValue >= num
	default:
		/// Logger.Printf("param valid function used but not validate the <%s> param\n", paramName)
	}

	return valid
}

// ParamValid check if the input is valid
func ParamValid(param, paramName string) {

	valid := ParamValidWrap(param, paramName)

	if !valid {
		utils.Fatalf(utl.ErrParamInValidSyntax, paramName)
	}
}

// ========================= others =============================
func KeyfileParsing(keyfilePath string) (*utl.Keyfile, error) {

	var keyfile = new(utl.Keyfile)
	// Load the keyfile.
	if keyfilePath != "" {
		keyJson, err := utl.ParseFileToBytes(keyfilePath)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(keyJson, keyfile)
		if err != nil {
			return nil, fmt.Errorf(utl.ErrUnmarshalBytesFormat, keyJson, err.Error())
		}

		keyfile.Json = keyJson
	}

	return keyfile, nil
}
