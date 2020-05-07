package packet

import (
	"bytes"
	"errors"
	"fmt"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"strings"
)

// CombineData of Contractcall data struct is used for packeting the data of wasm or evm contracts execution
// Implement the MessageCallDemo interface
func (call *ContractCall) CombineData() (string, string, bool, error) {

	// only transfer value, no data provided
	if call.data == nil {
		return "", "", true, nil
	}

	// packet contract method and input parameters
	outputType, isWrite, funcBytes, err := call.combineFunc()
	if err != nil {
		return "", "", false, err
	}

	// packet contract data
	data, err := call.combineContractData(funcBytes)
	return data, outputType, isWrite, err
}

// combineContractData selects the interpreter for combining the contract call data
func (call *ContractCall) combineContractData(funcBytes [][]byte) (string, error) {
	if call.Interp == nil {
		return "", errors.New("interpreter is not provided")
	}

	return call.Interp.combineData(funcBytes)
}

// combineData packet the data in the way defined by the evm virtual mechine
// Implement the Interpreter interface
func (i EvmInterpreter) combineData(funcBytes [][]byte) (string, error) {
	utl.Logger.Printf("combine data in evm")
	return hexutil.Encode(bytes.Join(funcBytes, []byte(""))), nil
}

// combineData packet the data in the way defined by the wasm virtual mechine
// Implement the Interpreter interface
func (i WasmInterpreter) combineData(funcBytes [][]byte) (string, error) {
	dataParams := make([][]byte, 0)
	dataParams = append(dataParams, utl.Int64ToBytes(int64(i.txType)))

	if i.cnsName != "" {
		dataParams = append(dataParams, []byte(i.cnsName))
	}

	// apend function params (contract method and parameters) to data
	dataParams = append(dataParams, funcBytes...)
	utl.Logger.Printf("combine data in wasm, dataParam is %v", dataParams)
	return rlpEncode(dataParams)
}

// combineFunc of Contractcall data struct is used for combining the
func (call *ContractCall) combineFunc() (string, bool, [][]byte, error) {
	var outputType string

	if call.data == nil {
		return "", false, nil, errors.New("no data provided")
	}

	// Judging whether this method exists or not by abi file
	abiFunc, err := ParseFuncFromAbi(call.data.funcAbi, call.data.funcName) //修改
	if err != nil {
		return "", false, nil, err
	}

	// Judging whether the number of inputs matches
	if len(abiFunc.Inputs) != len(call.data.funcParams) {
		return "", false, nil, fmt.Errorf(utl.ErrParamNumCheckFormat, len(abiFunc.Inputs), len(call.data.funcParams))
	}

	// encode the function and get the function constant
	funcByte, isWrite := call.encodeFunction(abiFunc)

	// Get the function output type for further use
	if len(abiFunc.Outputs) != 0 {
		outputType = abiFunc.Outputs[0].Type
	}

	return outputType, isWrite, funcByte, nil
}

// encodeFunction converts the function params to bytes and combine them by specific encoding rules
func (call *ContractCall) encodeFunction(abiFunc *FuncDesc) ([][]byte, bool) {

	var funcByte = make([][]byte, 1)

	// TODO
	if call.Interp == nil {
		utils.Fatalf("interpreter is not provided")
	}

	// converts the function params to bytes
	for i, v := range call.data.funcParams {
		input := abiFunc.Inputs[i]
		p, err := call.Interp.StringConverter(v, input.Type)
		if err != nil {
			utils.Fatalf(utl.ErrParamTypeFormat, v, i)
		}

		funcByte = append(funcByte, p)
	}

	// encode the contract method
	funcByte[0] = call.Interp.encodeFuncName(call.data.funcName)

	// get the function constant
	isWrite := call.Interp.setIsWrite(abiFunc)

	utl.Logger.Printf("the function byte is %v, the write operation is %v\n", funcByte, isWrite)
	return funcByte, isWrite
}

// encodeFuncName encodes the contract method in the way defined by the wasm virtual mechine
// Implement the Interpreter interface
func (i *WasmInterpreter) encodeFuncName(funcName string) []byte {
	utl.Logger.Printf("combine functoin in wasm")
	return []byte(funcName)
}

// encodeFuncName encodes the contract method in the way defined by the evm virtual mechine
// Implement the Interpreter interface
func (i *EvmInterpreter) encodeFuncName(funcName string) []byte {
	funcNameStr := fmt.Sprintf("%v(%v)", funcName, strings.Join(i.str, ","))
	utl.Logger.Printf("the function is %s\n", funcNameStr)
	funcNameHash := crypto.Keccak256([]byte(funcNameStr))[:4]
	funcByte := funcNameHash
	utl.Logger.Printf("the function byte is %v\n", funcByte)

	return funcByte
}

// CombineData of InnerCall data struct is used for packeting the data of the inner calls including fw, mig, etc.
// Implement the MessageCallDemo interface
func (call *InnerCall) CombineData() (string, string, bool, error) {

	outputType, isWrite, funcBytes, err := call.combineFunc()
	data, err := call.combineInnerData(funcBytes)

	return data, outputType, isWrite, err
}

func (call *InnerCall) combineFunc() (string, bool, [][]byte, error) {
	var outputType string
	var isWrite = true

	if call.data == nil {
		return "", false, nil, errors.New("no data provided")
	}

	// combine the function method and parameters
	funcByte := [][]byte{
		[]byte(call.data.funcName),
	}

	for _, input := range call.data.funcParams {
		funcByte = append(funcByte, []byte(input))
	}

	// get the inner call method constant and output type
	if call.data.funcName == "__sys_FwStatus" {
		isWrite = false
		outputType = "string"
	}

	return outputType, isWrite, funcByte, nil
}

func (call InnerCall) combineInnerData(funcBytes [][]byte) (string, error) {
	dataParams := make([][]byte, 0)
	dataParams = append(dataParams, utl.Int64ToBytes(int64(call.TxType)))
	dataParams = append(dataParams, funcBytes...)

	return rlpEncode(dataParams)
}

// CombineData of DeployCall data struct is used for packeting the data of wasm or evm contracts deployment
// Implement the MessageCallDemo interface
func (call DeployCall) CombineData() (string, string, bool, error) {
	if call.Interpreter == nil {
		return "", "", false, errors.New("interpreter is not provided")
	}

	data, err := call.Interpreter.combineDeployData() //TODO seperate?
	return data, "", true, err
}

// combineDeployData packet the data in the way defined by the evm virtual mechine
// Implement the Interpreter interface
func (i *EvmInterpreter) combineDeployData() (string, error) {
	return "0x" + string(i.codeBytes), nil
}

// combineDeployData packet the data in the way defined by the wasm virtual mechine
// Implement the Interpreter interface
func (i *WasmInterpreter) combineDeployData() (string, error) {
	utl.Logger.Printf("int wasm combineDeployData()")

	dataParams := make([][]byte, 0)
	dataParams = append(dataParams, utl.Int64ToBytes(int64(i.txType)))
	dataParams = append(dataParams, i.codeBytes)
	dataParams = append(dataParams, i.abiBytes)

	return rlpEncode(dataParams)
}

// rlpEncode encode the input value by RLP and convert the output bytes to hex string
func rlpEncode(val interface{}) (string, error) {

	dataRlp, err := rlp.EncodeToBytes(val)
	if err != nil {
		return "", fmt.Errorf(utl.ErrRlpEncodeFormat, err.Error())
	}

	return hexutil.Encode(dataRlp), nil

}
