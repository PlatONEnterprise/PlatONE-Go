package packet

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/common"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"

	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
)

func (call *ContractCallTest) GetAbiBytes() []byte {
	return call.data.funcAbi
}

// CombineData of Contractcall data struct is used for packeting the data of wasm or evm contracts execution
// Implement the MessageCallDemo interface
func (call *ContractCallTest) CombineData() (string, []string, bool, error) {

	// packet contract method and input parameters
	outputType, isWrite, funcBytes, err := call.combineFunc()
	if err != nil {
		return "", nil, false, err
	}

	// packet contract data
	data, err := call.combineContractData(funcBytes)
	return data, outputType, isWrite, err
}

// combineContractData selects the interpreter for combining the contract call data
func (call *ContractCallTest) combineContractData(funcBytes [][]byte) (string, error) {
	return call.Interp.combineData(funcBytes)
}

// combineFunc of Contractcall data struct is used for combining the
func (call *ContractCallTest) combineFunc() ([]string, bool, [][]byte, error) {

	// Judging whether this method exists or not by abi file
	abiFunc, err := ParseFuncFromAbi(call.data.funcAbi, call.data.funcName) //修改
	if err != nil {
		return nil, false, nil, err
	}

	// Judging whether the number of inputs matches
	if len(abiFunc.Inputs) != len(call.data.funcParams) {
		return nil, false, nil, fmt.Errorf("param check error, required %d inputs, recieved %d.\n", len(abiFunc.Inputs), len(call.data.funcParams))
	}

	// encode the function and get the function constant
	funcByte, err := call.Interp.encodeFunction(abiFunc, call.data.funcParams, call.data.funcName)
	if err != nil {
		return nil, false, nil, err
	}

	// get the function constant
	isWrite := call.Interp.setIsWrite(abiFunc)

	// Get the function output type for further use
	outputType := getOutputTypes(abiFunc)

	return outputType, isWrite, funcByte, nil
}

// encodeFunction converts the function params to bytes and combine them by specific encoding rules
func (i *EvmContractInterpreter) encodeFunction(abiFunc *FuncDesc, funcParams []string, funcName string) ([][]byte, error) {
	var arguments abi.Arguments
	var funcByte = make([][]byte, 1)
	var paramTypes = make([]string, 0)
	var args = make([]interface{}, 0)
	var argument abi.Argument
	var err error

	// converts the function params to bytes
	for i, v := range funcParams {
		input := abiFunc.Inputs[i]
		if argument.Type, err = abi.NewType(input.Type); err != nil {
			return nil, err
		}
		arguments = append(arguments, argument)

		arg, err := abi.SolInputTypeConversion(input.Type, v)
		if err != nil {
			return nil, err
		}

		args = append(args, arg)
		paramTypes = append(paramTypes, input.Type)
	}

	i.typeName = paramTypes
	paramsBytes, err := arguments.Pack(args...)
	if err != nil {
		/// common.ErrPrintln("pack args error: ", err)
		return nil, err
	}

	// encode the contract method
	funcByte[0] = i.encodeFuncName(funcName)
	funcByte = append(funcByte, paramsBytes)

	/// utl.Logger.Printf("the function byte is %v, the write operation is %v\n", funcByte, isWrite)
	return funcByte, nil
}

func (i *WasmContractInterpreter) encodeFunction(abiFunc *FuncDesc, funcParams []string, funcName string) ([][]byte, error) {

	var funcByte = make([][]byte, 1)

	// converts the function params to bytes
	for i, v := range funcParams {
		input := abiFunc.Inputs[i]
		p, err := abi.StringConverter(v, input.Type)
		if err != nil {
			return nil, err
		}

		funcByte = append(funcByte, p)
	}

	// encode the contract method
	funcByte[0] = i.encodeFuncName(funcName)

	/// utl.Logger.Printf("the function byte is %v, the write operation is %v\n", funcByte, isWrite)
	return funcByte, nil
}

// encodeFuncName encodes the contract method in the way defined by the wasm virtual mechine
// Implement the Interpreter interface
func (i *WasmContractInterpreter) encodeFuncName(funcName string) []byte {
	/// utl.Logger.Printf("combine functoin in wasm")
	return []byte(funcName)
}

// encodeFuncName encodes the contract method in the way defined by the evm virtual mechine
// Implement the Interpreter interface
func (i *EvmContractInterpreter) encodeFuncName(funcName string) []byte {

	funcNameStr := fmt.Sprintf("%v(%v)", funcName, strings.Join(i.typeName, ","))
	funcNameHash := crypto.Keccak256([]byte(funcNameStr))[:4]
	funcByte := funcNameHash

	return funcByte
}

// combineData packet the data in the way defined by the evm virtual mechine
// Implement the Interpreter interface
func (i EvmContractInterpreter) combineData(funcBytes [][]byte) (string, error) {
	/// utl.Logger.Printf("combine data in evm")
	return hexutil.Encode(bytes.Join(funcBytes, []byte(""))), nil
}

// combineData packet the data in the way defined by the wasm virtual mechine
// Implement the Interpreter interface
func (i WasmContractInterpreter) combineData(funcBytes [][]byte) (string, error) {
	dataParams := make([][]byte, 0)
	dataParams = append(dataParams, common.Int64ToBytes(int64(i.txType)))

	if i.cnsName != "" {
		dataParams = append(dataParams, []byte(i.cnsName))
	}

	// apend function params (contract method and parameters) to data
	dataParams = append(dataParams, funcBytes...)
	/// utl.Logger.Printf("combine data in wasm, dataParam is %v", dataParams)
	return rlpEncode(dataParams)
}

//------------------------------------------------------------------------------------

// CombineData of DeployCall data struct is used for packeting the data of wasm or evm contracts deployment
// Implement the MessageCallDemo interface
func (call DeployCallTest) CombineData() (string, []string, bool, error) {
	if call.Interpreter == nil {
		return "", nil, false, errors.New("interpreter is not provided")
	}

	data, err := call.Interpreter.combineData()
	return data, nil, true, err
}

// combineDeployData packet the data in the way defined by the evm virtual mechine
// Implement the Interpreter interface
func (i *EvmDeployInterpreter) combineData() (string, error) {
	return "0x" + string(i.codeBytes), nil
}

// combineDeployData packet the data in the way defined by the wasm virtual mechine
// Implement the Interpreter interface
func (i *WasmDeployInterpreter) combineData() (string, error) {
	/// utl.Logger.Printf("int wasm combineDeployData()")

	dataParams := make([][]byte, 0)
	dataParams = append(dataParams, common.Int64ToBytes(int64(i.txType)))
	dataParams = append(dataParams, i.codeBytes)
	dataParams = append(dataParams, i.abiBytes)

	return rlpEncode(dataParams)
}
