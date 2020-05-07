package packet

import (
	"bytes"
	"errors"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"math/big"
	"strconv"
)

// MessageCallDemo, the interface for different types of data package methods
type MessageCall interface {
	CombineData() (string, string, bool, error)
}

// ContractCallDemo, used for combining the data of contract execution
type ContractCall struct {
	data   *RawData
	TxType int
	name   string
	Interp Interpreter
}

// InnerCallDemo, used for combining the data of inner call methods (fw, mig, etc.)
type InnerCall struct {
	data   *RawData
	TxType int
}

// RawData, used to store function methods and abi
type RawData struct {
	funcName   string
	funcParams []string
	funcAbi    []byte
}

// DeployCall, used for combining the data of contract deployment
type DeployCall struct {
	codeBytes   []byte
	abiBytes    []byte
	TxType      int
	Interpreter Interpreter
}

// EvmInterpreter, packet data in the way defined by the evm virtual machine
type EvmInterpreter struct {
	str       []string
	codeBytes []byte // code bytes for evm contract deployment
}

// WasmInterpreter, packet data in the way defined by the evm virtual machine
type WasmInterpreter struct {
	cnsName   string // contract name for contract execution by contract name
	codeBytes []byte // code bytes for wasm contract deployment
	abiBytes  []byte // abi bytes for wasm contract deployment
	txType    int    // transaction type for contract deployment and execution
}

// Interpreter, the interface for different types of virtual machine(wasm or evm)
type Interpreter interface {
	encodeFuncName(string) []byte
	setIsWrite(*FuncDesc) bool
	combineData([][]byte) (string, error)
	combineDeployData() (string, error)
	StringConverter(string, string) ([]byte, error)
}

// NewData new a RawData object
func NewData(funcName string, funcParams []string, funcAbi []byte) *RawData {
	return &RawData{
		funcName:   funcName,
		funcParams: funcParams,
		funcAbi:    funcAbi,
	}
}

// NewContractCallDemo new a ContractCallDemo object
func NewContractCallDemo(data *RawData, name string, txType int) *ContractCall {

	call := &ContractCall{
		data:   data,
		name:   name,
		TxType: txType,
	}

	return call
}

// NewInnerCallDemo new a InnerCallDemo object
func NewInnerCallDemo(data *RawData, txType int) *InnerCall {

	call := &InnerCall{
		data:   data,
		TxType: txType,
	}

	return call
}

// NewDeployCall new a DeployCall object
func NewDeployCall(codeBytes, abiBytes []byte, vm string, txType int) *DeployCall {

	call := &DeployCall{
		codeBytes: codeBytes,
		abiBytes:  abiBytes,
		TxType:    txType,
	}

	// set the virtual machine interpreter
	call.SetInterpreter(vm)

	return call
}

// SetInterpreter set the interpreter of ContractCallDemo object
func (call *ContractCall) SetInterpreter(vm string) {
	switch vm {
	case "wasm":
		call.Interp = &WasmInterpreter{cnsName: call.name, txType: call.TxType}
	case "evm":
		call.Interp = &EvmInterpreter{}
	default:
		call.Interp = &WasmInterpreter{cnsName: call.name, txType: call.TxType}
	}
}

// SetInterpreter set the interpreter of DeployCall object
func (call *DeployCall) SetInterpreter(vm string) {
	switch vm {
	case "evm":
		if IsWasmContract(call.codeBytes) {
			utils.Fatalf("the input  is not evm byte code")
		}
		call.Interpreter = &EvmInterpreter{codeBytes: call.codeBytes}
	default:
		if !IsWasmContract(call.codeBytes) {
			utils.Fatalf("the input  is not wasm byte code")
		}
		call.Interpreter = &WasmInterpreter{
			codeBytes: call.codeBytes,
			abiBytes:  call.abiBytes,
		}
	}
}

// IsWasmContract judge whether the bytes satisfy the code format of wasm virtual machine
func IsWasmContract(codeBytes []byte) bool {
	if bytes.Equal(codeBytes[:8], []byte{0, 97, 115, 109, 1, 0, 0, 0}) {
		return true
	}
	return false
}

// setIsWrite judge the constant of the contract method based on evm
// Implement the Interpreter interface
func (i EvmInterpreter) setIsWrite(abiFunc *FuncDesc) bool {
	return abiFunc.StateMutability != "pure" && abiFunc.StateMutability != "view"
}

// setIsWrite judge the constant of the contract method based on wasm
// Implement the Interpreter interface
func (i WasmInterpreter) setIsWrite(abiFunc *FuncDesc) bool {
	return abiFunc.Constant != "true"
}

// set append all the function parameters and type into an array in EvmInterpreter object
func (i *EvmInterpreter) set(s string) {
	i.str = append(i.str, s)
}

// StringConverter encodes different types of function parameters into bytes in the way defined by the evm virtual machine
// Implement the Interpreter interface
func (i *EvmInterpreter) StringConverter(source string, t string) ([]byte, error) {
	i.set(t)

	switch t {
	case "uint32", "uint16", "uint8", "uint":
		dest, err := strconv.Atoi(source)
		return utl.U256(new(big.Int).SetUint64(uint64(dest))), err
	case "int", "int8", "int16", "int32":
		dest, err := strconv.Atoi(source)
		return utl.U256(big.NewInt(int64(dest))), err
	case "int64", "uint64":
		dest, err := strconv.ParseInt(source, 10, 64)
		return utl.Int64ToBytes(dest), err
	case "float32":
		dest, err := strconv.ParseFloat(source, 32)
		return utl.Float32ToBytes(float32(dest)), err
	case "float64":
		dest, err := strconv.ParseFloat(source, 64)
		return utl.Float64ToBytes(dest), err
	case "bool":
		if "true" == source || "false" == source {
			return utl.BoolToBytes("true" == source), nil
		} else {
			return []byte{}, errors.New("invalid boolean param")
		}
	default:
		return []byte(source), nil
	}
}

// StringConverter encodes different types of function parameters into bytes in the way defined by the wasm virtual machine
// Implement the Interpreter interface
func (i WasmInterpreter) StringConverter(source string, t string) ([]byte, error) {
	switch t {
	case "int32", "uint32", "uint", "int":
		dest, err := strconv.Atoi(source)
		return utl.Int32ToBytes(int32(dest)), err
	case "int64", "uint64":
		dest, err := strconv.ParseInt(source, 10, 64)
		return utl.Int64ToBytes(dest), err
	case "float32":
		dest, err := strconv.ParseFloat(source, 32)
		return utl.Float32ToBytes(float32(dest)), err
	case "float64":
		dest, err := strconv.ParseFloat(source, 64)
		return utl.Float64ToBytes(dest), err
	case "bool":
		if "true" == source || "false" == source {
			return utl.BoolToBytes("true" == source), nil
		} else {
			return []byte{}, errors.New("invalid boolean param")
		}
	default:
		return []byte(source), nil
	}
}
