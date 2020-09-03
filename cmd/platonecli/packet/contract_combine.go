package packet

import (
	"fmt"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"

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

// NewContractDataGen new a ContractDataGen object
func NewContractDataGenWrap(funcName string, funcParams []string, funcAbi []byte, cns Cns, vm string) *ContractDataGen {
	// parse the function parameters
	funcName, funcParams = utils.FuncParse(funcName, funcParams)

	// new an contract call, set the interpreter(wasm or evm contract)
	data := NewData(funcName, funcParams, funcAbi)
	dataGen := NewContractDataGen(data, cns.Name, cns.txType)
	dataGen.SetInterpreter(vm) //TODO

	return dataGen
}

func NewContractDataGen(data *RawData, name string, txType uint64) *ContractDataGen {

	dataGen := &ContractDataGen{
		data:   data,
		name:   name,
		TxType: txType,
	}

	return dataGen
}

// SetInterpreter set the interpreter of ContractDataGen object
func (dataGen *ContractDataGen) SetInterpreter(vm string) {
	switch vm {
	case "evm":
		dataGen.Interp = &EvmContractInterpreter{}
	default: // the default interpreter is "wasm"
		dataGen.Interp = &WasmContractInterpreter{
			cnsName: dataGen.name,
			txType:  dataGen.TxType,
		}
	}
}

func (dataGen *ContractDataGen) ReceiptParsing(receipt *Receipt) *ReceiptParsingReturn {
	return dataGen.Interp.ReceiptParsing(receipt, dataGen.data.funcAbi)
}

// CombineData of Contractcall data struct is used for packeting the data of wasm or evm contracts execution
// Implement the MessageCallDemo interface
func (dataGen *ContractDataGen) CombineData() (string, []abi.ArgumentMarshaling, bool, error) {

	// packet contract method and input parameters
	outputType, isWrite, funcBytes, err := dataGen.combineFunc()
	if err != nil {
		return "", nil, false, err
	}

	// packet contract data
	data, err := dataGen.combineContractData(funcBytes)
	return data, outputType, isWrite, err
}

// combineFunc of Contractcall data struct is used for combining the
func (dataGen *ContractDataGen) combineFunc() ([]abi.ArgumentMarshaling, bool, [][]byte, error) {

	// Judging whether this method exists or not by abi file
	abiFunc, err := ParseFuncFromAbi(dataGen.data.funcAbi, dataGen.data.funcName) //修改
	if err != nil {
		return nil, false, nil, err
	}

	// Judging whether the number of inputs matches
	if len(abiFunc.Inputs) != len(dataGen.data.funcParams) {
		return nil, false, nil, fmt.Errorf("param check error, required %d inputs, recieved %d.\n", len(abiFunc.Inputs), len(dataGen.data.funcParams))
	}

	// encode the function and get the function constant
	funcByte, err := dataGen.Interp.encodeFunction(abiFunc, dataGen.data.funcParams, dataGen.data.funcName)
	if err != nil {
		return nil, false, nil, err
	}

	// get the function constant
	isWrite := dataGen.Interp.setIsWrite(abiFunc)

	// Get the function output type for further use
	/// outputType := getOutputTypes(abiFunc)
	outputType := abiFunc.Outputs

	return outputType, isWrite, funcByte, nil
}

func getOutputTypes(abiFunc *FuncDesc) []string {
	var outputTypes = make([]string, 0)

	for _, output := range abiFunc.Outputs {
		outputTypes = append(outputTypes, GenFuncSig(output))
	}

	return outputTypes
}

// combineContractData selects the interpreter for combining the contract call data
func (dataGen *ContractDataGen) combineContractData(funcBytes [][]byte) (string, error) {
	return dataGen.Interp.combineData(funcBytes)
}

func (dataGen *ContractDataGen) ParseNonConstantResponse(respStr string, outputType []abi.ArgumentMarshaling) []interface{} {
	return dataGen.Interp.ParseNonConstantResponse(respStr, outputType)
}
