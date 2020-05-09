package packet

import (
	"bytes"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
)

// InnerCallCommon
// extract the common part of all the inner calls
func InnerCallCommon(funcName string, funcParams []string, txType int) *InnerCall {
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
	call := NewContractCallDemo(data, cns.Name, cns.txType)
	call.SetInterpreter(vm) //TODO

	return call
}

// GetAddressByName wraps the RpcCalls used to get the contract address by cns name
// the parameters are packet into transaction before packet into rpc json data struct
func GetAddressByName(name string) (string, error) {

	// chain defined data type convert
	to := common.HexToAddress(CNS_PROXY_ADDRESS)
	from := common.HexToAddress("")

	// packet the contract all data
	rawData := NewData("getContractAddress", []string{name, "latest"}, nil)
	call := NewInnerCallDemo(rawData, EXECUTE_CONTRACT)
	data, _, _, _ := call.CombineData()

	tx := NewTxParams(from, &to, "", "", "", data, call.TxType)
	params := utl.CombineParams(tx, "latest")

	response, err := utl.RpcCalls("eth_call", params)
	if err != nil {
		return "", err
	}

	// parse the rpc response
	resultBytes, _ := hexutil.Decode(response.(string))
	bytesTrim := bytes.TrimRight(resultBytes, "\x00")
	result := utl.BytesConverter(bytesTrim, "string")

	return result.(string), nil
}
