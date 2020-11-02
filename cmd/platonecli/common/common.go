package common

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/core/types"

	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/utils"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
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

	abiBytes, err = utils.ParseFileToBytes(abiFilePath)
	if err != nil {
		utl.Fatalf(utils.ErrParseFileFormat, "abi", err.Error())
	}

	return abiBytes
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
		utl.Fatalf("")
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

func (conv *Convert) Parse(param interface{}) string {

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

// ========================== Param Convert ========================================

// ParamParse convert the user inputs to the value needed
func ParamParse(param, paramName string) interface{} {
	var err error
	var i interface{}

	switch paramName {
	case "contract", "user":
		i = IsNameOrAddress(param)
		if i == CnsIsUndefined {
			err = fmt.Errorf(utils.ErrParamInValidSyntax, "name or contract address")
		}
	case "delayNum", "p2pPort", "rpcPort":
		if utils.IsInRange(param, 65535) {
			i, err = strconv.ParseInt(param, 10, 0)
		} else {
			err = errors.New("value out of range")
		}
	case "operation", "status", "type":
		i, err = ConvertSelect(param, paramName)
	case "code", "abi":
		i, err = utils.ParseFileToBytes(param)
	default:
		i, err = param, nil
	}

	if err != nil {
		utl.Fatalf(utils.ErrParamParseFormat, paramName, err.Error())
	}

	return i
}

// ChainParamConvert convert the string to chain defined type
func ChainParamConvert(param, paramName string) interface{} {
	var err error
	var i interface{}

	switch paramName {
	case "value", "gasPrice":
		i, err = utils.IntValueConvert(param)
	case "gas":
		i, err = utils.UintValueConvert(param)
	case "address", "to", "from":
		i, err = utils.AddressConvert(param)
	default:
		i, err = param, nil //TODO
	}

	if err != nil {
		utl.Fatalf(utils.ErrParamParseFormat, paramName, err.Error())
	}

	return i
}

//================================CNS=================================

type Cns struct {
	/// To     common.Address
	Name   string // the cns name of contract
	TxType uint64 // the transaction type of the contract execution (EXECUTE_CONTRACT or CNS_TX_TYPE)
}

func NewCns(name string, txType uint64) *Cns {
	return &Cns{
		/// To:     common.HexToAddress(to),
		Name:   name,
		TxType: txType,
	}
}

const (
	CnsIsName int32 = iota
	CnsIsAddress
	CnsIsUndefined
)

// CnsParse judge whether the input string is contract address or contract name
// and return the corresponding infos
func CnsParse(contract string) (*Cns, common.Address, error) {
	isAddress := IsNameOrAddress(contract)

	switch isAddress {
	case CnsIsAddress:
		return NewCns("", types.NormalTxType), common.HexToAddress(contract), nil
	case CnsIsName:
		return NewCns(contract, types.CnsTxType), common.HexToAddress(precompile.CnsInvokeAddress), nil
	default:
		return nil, common.Address{}, fmt.Errorf(utils.ErrParamInValidSyntax, "contract address")
	}
}

// IsNameOrAddress Judge whether the input string is an address or a name
func IsNameOrAddress(str string) int32 {
	var valid int32

	switch {
	case utils.IsMatch(str, "address"):
		valid = CnsIsAddress
	case utils.IsMatch(str, "name") &&
		!strings.HasPrefix(strings.ToLower(str), "0x"):
		valid = CnsIsName
	default:
		valid = CnsIsUndefined
	}

	return valid
}

//===============================User Input Verification=======================================

func ParamValidWrap(param, paramName string) bool {
	var valid = true

	switch paramName {
	case "fw":
		if param != "*" {
			valid = utils.IsMatch(param, "address")
		}
	case "to":
		valid = param == "" || utils.IsMatch(param, "address")
	case "contract":
		valid = utils.IsMatch(param, "address") || utils.IsMatch(param, "name")
	case "action":
		valid = strings.EqualFold(param, "accept") || strings.EqualFold(param, "reject")
	case "vm":
		valid = param == "" || strings.EqualFold(param, "evm") || strings.EqualFold(param, "wasm")
	case "ipAddress":
		valid = utils.IsUrl(param)
	case "externalIP", "internalIP":
		valid = utils.IsUrl(param + ":0")
	//case "version":
	//	valid = utl.IsVersion(param)
	case "role":
		valid = utils.IsRoleMatch(param)
	case "roles":
		valid = utils.IsValidRoles(param)
	case "email", "mobile", "version", "num":
		valid = utils.IsMatch(param, paramName)

	// newly added for restful server
	// todo; fix the toLower problem
	case "orgin", "address":
		valid = utils.IsMatch(param, "address")
	case "contractname", "name":
		valid = utils.IsMatch(param, "name")
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
