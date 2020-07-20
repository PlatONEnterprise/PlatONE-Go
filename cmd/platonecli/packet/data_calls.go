package packet

import (
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
)

// InnerCallCommon
// extract the common part of all the inner calls
func InnerCallCommon(funcName string, funcParams []string, txType uint64) *InnerCall {
	// parse the function parameters
	funcName, funcParams = utl.FuncParse(funcName, funcParams)

	// new an inner call
	data := NewData(funcName, funcParams, nil)
	call := NewInnerCallDemo(data, txType)

	return call
}

// ContractCallCommon
// extract the common part of all the contract calls
func ContractCallCommon(funcName string, funcParams []string, funcAbi []byte, cns Cns, vm string) *ContractCall {
	// parse the function parameters
	funcName, funcParams = utl.FuncParse(funcName, funcParams)

	// new an contract call, set the interpreter(wasm or evm contract)
	data := NewData(funcName, funcParams, funcAbi)
	call := NewContractCall(data, cns.Name, cns.txType)
	call.SetInterpreter(vm) //TODO

	return call
}

// ContractCallCommon
// extract the common part of all the contract calls
func ContractCallCommonTest(funcName string, funcParams []string, funcAbi []byte, cns Cns, vm string) *ContractCallTest {
	// parse the function parameters
	funcName, funcParams = utl.FuncParse(funcName, funcParams)

	// new an contract call, set the interpreter(wasm or evm contract)
	data := NewData(funcName, funcParams, funcAbi)
	call := NewContractCallTest(data, cns.Name, cns.txType)
	call.SetInterpreter(vm) //TODO

	return call
}
