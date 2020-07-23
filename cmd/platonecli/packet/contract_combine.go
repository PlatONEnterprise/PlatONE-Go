package packet

import (
	"fmt"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
)

// RawData, used to store function methods and abi
type RawData struct {
	funcName   string
	funcParams []string
	funcAbi    []byte
}

// NewData new a RawData object
func NewData(funcName string, funcParams []string, funcAbi []byte) *RawData {
	return &RawData{
		funcName:   funcName,
		funcParams: funcParams,
		funcAbi:    funcAbi,
	}
}

// ContractDataGen is used for combining the data of contract execution
type ContractDataGen struct {
	data   *RawData
	TxType uint64
	name   string
	Interp contractInter
}

// NewContractCallDemo new a ContractDataGen object
func ContractCallCommonTest(funcName string, funcParams []string, funcAbi []byte, cns Cns, vm string) *ContractDataGen {
	// parse the function parameters
	funcName, funcParams = utils.FuncParse(funcName, funcParams)

	// new an contract call, set the interpreter(wasm or evm contract)
	data := NewData(funcName, funcParams, funcAbi)
	call := NewContractDataGen(data, cns.Name, cns.txType)
	call.SetInterpreter(vm) //TODO

	return call
}

func NewContractDataGen(data *RawData, name string, txType uint64) *ContractDataGen {

	call := &ContractDataGen{
		data:   data,
		name:   name,
		TxType: txType,
	}

	return call
}

// SetInterpreter set the interpreter of ContractDataGen object
func (call *ContractDataGen) SetInterpreter(vm string) {
	switch vm {
	case "evm":
		call.Interp = &EvmContractInterpreter{}
	default: // the default interpreter is "wasm"
		call.Interp = &WasmContractInterpreter{
			cnsName: call.name,
			txType:  call.TxType,
		}
	}
}

// CombineData of Contractcall data struct is used for packeting the data of wasm or evm contracts execution
// Implement the MessageCallDemo interface
func (call *ContractDataGen) CombineData() (string, []string, bool, error) {

	// packet contract method and input parameters
	outputType, isWrite, funcBytes, err := call.combineFunc()
	if err != nil {
		return "", nil, false, err
	}

	// packet contract data
	data, err := call.combineContractData(funcBytes)
	return data, outputType, isWrite, err
}

// combineFunc of Contractcall data struct is used for combining the
func (call *ContractDataGen) combineFunc() ([]string, bool, [][]byte, error) {

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

func getOutputTypes(abiFunc *FuncDesc) []string {
	var outputTypes = make([]string, 0)

	/*
		if len(abiFunc.Outputs) != 0 {
			outputType = abiFunc.Outputs[0].Type
		}*/

	for _, output := range abiFunc.Outputs {
		outputTypes = append(outputTypes, output.Type)
	}

	return outputTypes
}

// combineContractData selects the interpreter for combining the contract call data
func (call *ContractDataGen) combineContractData(funcBytes [][]byte) (string, error) {
	return call.Interp.combineData(funcBytes)
}

func (call *ContractDataGen) GetAbiBytes() []byte {
	return call.data.funcAbi
}
