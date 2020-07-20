package packet

import (
	"errors"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
)

type deployInter interface {
	combineData() (string, error)
}

type contractInter interface {
	encodeFuncName(string) []byte
	encodeFunction(*FuncDesc, []string, string) ([][]byte, error)
	combineData([][]byte) (string, error)
	setIsWrite(*FuncDesc) bool
}

// ContractCallDemo, used for combining the data of contract execution
type ContractCallTest struct {
	data   *RawData
	TxType uint64
	name   string
	Interp contractInter
}

// DeployCall, used for combining the data of contract deployment
type DeployCallTest struct {
	codeBytes   []byte
	abiBytes    []byte
	TxType      uint64
	Interpreter deployInter
}

// EvmInterpreter, packet data in the way defined by the evm virtual machine
type EvmDeployInterpreter struct {
	codeBytes []byte // code bytes for evm contract deployment
}

type EvmContractInterpreter struct {
	typeName []string // contract parameter types
}

// WasmInterpreter, packet data in the way defined by the evm virtual machine
type WasmDeployInterpreter struct {
	codeBytes []byte // code bytes for wasm contract deployment
	abiBytes  []byte // abi bytes for wasm contract deployment
	txType    uint64 // transaction type for contract deployment and execution
}

type WasmContractInterpreter struct {
	txType    uint64 // transaction type for contract deployment and execution
	cnsName   string // contract name for contract execution by contract name
}

// NewContractCallDemo new a ContractCallDemo object
func NewContractCallTest(data *RawData, name string, txType uint64) *ContractCallTest {

	call := &ContractCallTest{
		data:   data,
		name:   name,
		TxType: txType,
	}

	return call
}

// NewDeployCall new a DeployCall object
func NewDeployCallTest(codeBytes, abiBytes []byte, vm string, txType uint64) *DeployCallTest {

	call := &DeployCallTest{
		codeBytes: codeBytes,
		abiBytes:  abiBytes,
		TxType:    txType,
	}

	// set the virtual machine interpreter
	call.SetInterpreter(vm)

	return call
}

// SetInterpreter set the interpreter of ContractCallDemo object
func (call *ContractCallTest) SetInterpreter(vm string) {
	switch vm {
	case "evm":
		call.Interp = &EvmContractInterpreter{}
	default:	// the default interpreter is "wasm"
		call.Interp = &WasmContractInterpreter{
			cnsName:	call.name,
			txType:		call.TxType,
		}
	}
}

// SetInterpreter set the interpreter of DeployCall object
func (call *DeployCallTest) SetInterpreter(vm string) error {
	switch vm {
	case "evm":
		if IsWasmContract(call.codeBytes) {
			// utils.Fatalf("the input  is not evm byte code")
			return errors.New("the input  is not evm byte code")
		}
		call.Interpreter = &EvmDeployInterpreter{codeBytes: call.codeBytes}
	default:
		if !IsWasmContract(call.codeBytes) {
			// utils.Fatalf("the input  is not wasm byte code")
			return errors.New("the input  is not wasm byte code")
		}
		call.Interpreter = &WasmDeployInterpreter{
			codeBytes: call.codeBytes,
			abiBytes:  call.abiBytes,
		}
	}

	return nil
}

// setIsWrite judge the constant of the contract method based on evm
// Implement the Interpreter interface
func (i EvmContractInterpreter) setIsWrite(abiFunc *FuncDesc) bool {
	return abiFunc.StateMutability != "pure" && abiFunc.StateMutability != "view"
}

// setIsWrite judge the constant of the contract method based on wasm
// Implement the Interpreter interface
func (i WasmContractInterpreter) setIsWrite(abiFunc *FuncDesc) bool {
	return abiFunc.Constant != "true"
}

// combineDeployData packet the data in the way defined by the evm virtual mechine
// Implement the Interpreter interface
func (i *EvmDeployInterpreter) combineDeployData() (string, error) {
	return "0x" + string(i.codeBytes), nil
}

// combineDeployData packet the data in the way defined by the wasm virtual mechine
// Implement the Interpreter interface
func (i *WasmDeployInterpreter) combineDeployData() (string, error) {
	/// utl.Logger.Printf("int wasm combineDeployData()")

	dataParams := make([][]byte, 0)
	dataParams = append(dataParams, utl.Int64ToBytes(int64(i.txType)))
	dataParams = append(dataParams, i.codeBytes)
	dataParams = append(dataParams, i.abiBytes)

	return rlpEncode(dataParams)
}