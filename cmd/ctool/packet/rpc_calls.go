package packet

import (
	"bytes"
	"fmt"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
)

// Receipt, eth_getTransactionReceipt return data struct
type Receipt struct {
	BlockHash         string `json:"blockHash"`          // hash of the block
	BlockNumber       string `json:"blockNumber"`        // height of the block
	ContractAddress   string `json:"contractAddress"`    // contract address of the contract deployment. otherwise null
	CumulativeGasUsed string `json:"cumulativeGas_used"` //
	From              string `json:"from"`               // the account address used to send the transaction
	GasUsed           string `json:"gasUsed"`            // gas used by executing the transaction
	Root              string `json:"root"`
	To                string `json:"to"`               // the address the transaction is sent to
	TransactionHash   string `json:"transactionHash"`  // the hash of the transaction
	TransactionIndex  string `json:"transactionIndex"` // the index of the transaction
	Logs              []struct {
		Address string   `json:"address"`
		Topics  []string `json:"topics"`
		Data    string   `json:"data"`
	} `json:"logs"`
	Status string `json:"status"` // the execution status of the transaction, "0x1" for success
}

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

// RpcCalls
func RpcCalls(action string, params []interface{}) (interface{}, error) {

	// new a rpc json data struct
	rpcJson := utl.NewRpcJson(action, params)

	// post http request
	response, err := utl.HttpPost(rpcJson)
	if err != nil {
		return nil, fmt.Errorf(utl.ErrHttpSendFormat, err.Error())
	}

	// parse the result from Rpc response
	result, err := utl.ParseRpcResponse(response)
	if err != nil {
		return nil, fmt.Errorf(utl.ErrRpcExecuationFormat, action, err.Error())
	}

	//Logger.Printf("response json: %v\n", result)
	return result, nil
}

// CombineParams combines multiple rpc json parameters into an array
func CombineParams(args ...interface{}) []interface{} {
	params := make([]interface{}, 0)
	params = append(params, args...)
	return params
}

// GetTransactionReceipt wraps the RpcCalls used to get the transaction receipt
func GetTransactionReceipt(txHash string) (*Receipt, error) {

	params := CombineParams(txHash)

	response, err := RpcCalls("eth_getTransactionReceipt", params)
	if err != nil {
		return nil, err
	}

	// parse the rpc response
	receipt := ParseTxReceipt(response)
	return receipt, nil
}

// GetNonce wraps the RpcCalls used to get the nonce based on the latest block number
func GetNonce(addr common.Address) uint64 {

	params := CombineParams(addr, "latest")

	response, err := RpcCalls("eth_getTransactionCount", params)
	if err != nil {
		utl.LogErr.Println(err.Error())
		utils.Fatalf(utl.ErrTODO, utl.DEFAULT_LOG_DIRT)
	}

	// parse the rpc response
	nonce, _ := hexutil.DecodeBig(response.(string))
	return nonce.Uint64()
}

// getCodeByAddress wraps the RpcCalls used to get the code of an contract by contract address
func getCodeByAddress(addr string) (string, error) {

	params := CombineParams(addr, "latest")

	response, err := RpcCalls("eth_getCode", params)
	if err != nil {
		return "", err
	}

	// parse the rpc response and return
	return response.(string), nil
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
	params := CombineParams(tx, "latest")

	response, err := RpcCalls("eth_call", params)
	if err != nil {
		return "", err
	}

	// parse the rpc response
	resultBytes, _ := hexutil.Decode(response.(string))
	bytesTrim := bytes.TrimRight(resultBytes, "\x00")
	result := utl.BytesConverter(bytesTrim, "string")

	return result.(string), nil
}
