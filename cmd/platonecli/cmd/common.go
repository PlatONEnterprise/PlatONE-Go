package cmd

import (
	"encoding/json"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"

	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/packet"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/utils"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

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
	funcAbi := cmd_common.AbiParse(abiPath, contract)
	// abi bytes parsing
	contractAbi, _ := packet.ParseAbiFromJson(funcAbi)
	// find the method in abi obj.
	methodAbi, _ := contractAbi.GetFuncFromAbi(funcName)
	// convert user input string to args in Golang
	funcArgs, _ := methodAbi.StringToArgs(funcParams)

	// judge whether the input string is contract address or contract name
	cns, to, err := cmd_common.CnsParse(contract)
	if err != nil {
		utl.Fatalf(err.Error())
	}

	/// dataGenerator := packet.NewContractDataGenWrap(funcName, funcParams, funcAbi, *cns, vm)
	data := packet.NewData(funcArgs, methodAbi)
	dataGenerator := packet.NewContractDataGen(data, contractAbi, cns.TxType)
	dataGenerator.SetInterpreter(vm, cns.Name, cns.TxType)

	return clientCommonV2(c, dataGenerator, &to)
}

// =====================================================================

// FuncParse wraps the GetFuncNameAndParams
// it separates the function method name and the parameters
func FuncParse(funcName string, funcParams []string) (string, []string) {
	var funcParamsNew []string

	if funcName == "" {
		return "", nil
	}

	funcName, funcParamsNew = GetFuncNameAndParams(funcName)
	if len(funcParamsNew) != 0 && len(funcParams) != 0 {
		utl.Fatalf(utils.ErrParamInValidSyntax, "function parameters")
	}
	funcParams = append(funcParams, funcParamsNew...)

	/// Logger.Printf("after function parse, the function is %s, %s", funcName, funcParams)
	return funcName, funcParams
}

// GetFuncNameAndParams parse the function params from the input string
func GetFuncNameAndParams(funcAndParams string) (string, []string) {
	// eliminate space
	f := TrimSpace(funcAndParams)

	hasBracket := strings.Contains(f, "(") && strings.Contains(f, ")")
	if !hasBracket {
		return f, nil
	}

	funcName := f[0:strings.Index(f, "(")]

	paramString := f[strings.Index(f, "(")+1 : strings.LastIndex(f, ")")]
	params := abi.GetFuncParams(paramString)

	return funcName, params
}

// TrimSpace trims all the space in the string
func TrimSpace(str string) string {
	strNoSpace := strings.Split(str, " ")
	return strings.Join(strNoSpace, "")
}

// ==================================================================

// PrintJson reformats the json printing style, easier for users to read
func PrintJson(marshalJson []byte) string {

	var addBytes = []byte{'\n'}
	var newJson = make([]byte, 0)

	for _, v := range marshalJson {
		switch v {
		case '}':
			addBytes = addBytes[:len(addBytes)-1]
			newJson = append(newJson, addBytes...)
			newJson = append(newJson, v)
		case '{':
			addBytes = append(addBytes, byte('\t'))
			newJson = append(newJson, v)
			newJson = append(newJson, addBytes...)
		case ',':
			newJson = append(newJson, v)
			newJson = append(newJson, addBytes...)
		default:
			newJson = append(newJson, v)
		}
	}

	return string(newJson)
}

// ========================= Combine Json =================================
// todo: deprecated
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
			temp := cmd_common.ParamParse(tmp, key)
			m[key] = temp
		}
	}

	// required value
	for i, key := range arrayMust {
		m[key] = c.Args().Get(i)
		paramValid(m[key].(string), key)
	}

	if len(m) == 0 {
		utl.Fatalf(utils.ErrInputNullFormat, "json combination result")
	}

	bytes, _ = json.Marshal(m)
	/// utl.Logger.Printf("the combine json is %s\n", bytes)

	return string(bytes)
}

// OptionParamValid wraps ParamValid, it allows the input to be null
func optionParamValid(param, paramName string) {
	if param != "" {
		paramValid(param, paramName)
	}
}

// ParamValid check if the input is valid
func paramValid(param, paramName string) {

	valid := cmd_common.ParamValidWrap(param, paramName)

	if !valid {
		utl.Fatalf(utils.ErrParamInValidSyntax, paramName)
	}
}
