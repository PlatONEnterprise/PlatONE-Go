package utils

import (
	"encoding/json"
	"fmt"
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

// RpcCalls
func RpcCalls(action string, params []interface{}) (interface{}, error) {

	// new a rpc json data struct
	rpcJson := NewRpcJson(action, params)

	// post http request
	response, err := HttpPost(rpcJson)
	if err != nil {
		return nil, fmt.Errorf(ErrHttpSendFormat, err.Error())
	}

	// parse the result from Rpc response
	result, err := ParseRpcResponse(response)
	if err != nil {
		return nil, fmt.Errorf(ErrRpcExecuationFormat, action, err.Error())
	}

	//Logger.Printf("response json: %v\n", result)
	return result, nil
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

// ParseSysContractResult parsed the rpc response to Receipt object
func ParseTxReceipt(response interface{}) *Receipt {
	var receipt = &Receipt{}

	if response == nil {
		return nil
	}

	temp, _ := json.Marshal(response)
	err := json.Unmarshal(temp, receipt)
	if err != nil {
		LogErr.Printf(ErrUnmarshalBytesFormat, "transaction receipt", err.Error())
		utils.Fatalf(ErrTODO, DEFAULT_LOG_DIRT)
	}

	return receipt
}

// GetNonce wraps the RpcCalls used to get the nonce based on the latest block number
func GetNonce(addr common.Address) uint64 {

	params := CombineParams(addr, "latest")

	response, err := RpcCalls("eth_getTransactionCount", params)
	if err != nil {
		LogErr.Println(err.Error())
		utils.Fatalf(ErrTODO, DEFAULT_LOG_DIRT)
	}

	// parse the rpc response
	nonce, _ := hexutil.DecodeBig(response.(string))
	return nonce.Uint64()
}

// getCodeByAddress wraps the RpcCalls used to get the code of an contract by contract address
func GetCodeByAddress(addr string) (string, error) {

	params := CombineParams(addr, "latest")

	response, err := RpcCalls("eth_getCode", params)
	if err != nil {
		return "", err
	}

	// parse the rpc response and return
	return response.(string), nil
}

// CombineParams combines multiple rpc json parameters into an array
func CombineParams(args ...interface{}) []interface{} {
	params := make([]interface{}, 0)
	params = append(params, args...)
	return params
}
