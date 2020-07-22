package platoneclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
)

const (
	txReceiptSuccessCode = "0x1"
	txReceiptFailureCode = "0x0"

	txReceiptSuccessMsg = "Operation Succeeded"
	txReceiptFailureMsg = "Operation Failed"

	sleepTime = 1000000000 // 1 seconds
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
		return nil, fmt.Errorf(utl.ErrHttpSendFormat, err.Error())
	}

	// parse the result from Rpc response
	result, err := ParseRpcResponse(response)
	if err != nil {
		return nil, fmt.Errorf(utl.ErrRpcExecuationFormat, action, err.Error())
	}

	//Logger.Printf("response json: %v\n", result)
	return result, nil
}

// GetTransactionReceipt wraps the RpcCalls used to get the transaction receipt
func GetTransactionReceipt(txHash string) (*Receipt, error) {

	params := utl.CombineParams(txHash)

	response, err := RpcCalls("eth_getTransactionReceipt", params)
	if err != nil {
		return nil, err
	}

	// parse the rpc response
	receipt, err := ParseTxReceipt(response)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// ParseSysContractResult parsed the rpc response to Receipt object
func ParseTxReceipt(response interface{}) (*Receipt, error) {
	var receipt = &Receipt{}

	if response == nil {
		return nil, nil
	}

	temp, _ := json.Marshal(response)
	err := json.Unmarshal(temp, receipt)
	if err != nil {
		// LogErr.Printf(ErrUnmarshalBytesFormat, "transaction receipt", err.Error())
		// utils.Fatalf(ErrTODO, DEFAULT_LOG_DIRT)
		errStr := fmt.Sprintf(utl.ErrUnmarshalBytesFormat, "transaction receipt", err.Error())
		return nil, errors.New(errStr)
	}

	return receipt, nil
}

// GetNonce wraps the RpcCalls used to get the nonce based on the latest block number
func GetNonce(addr common.Address) (uint64, error) {

	params := utl.CombineParams(addr, "latest")

	response, err := RpcCalls("eth_getTransactionCount", params)
	if err != nil {
		// LogErr.Println(err.Error())
		// utils.Fatalf(ErrTODO, DEFAULT_LOG_DIRT)
		return 0, err
	}

	// parse the rpc response
	nonce, _ := hexutil.DecodeBig(response.(string))
	return nonce.Uint64(), nil
}

// getCodeByAddress wraps the RpcCalls used to get the code of an contract by contract address
func GetCodeByAddress(addr string) (string, error) {

	params := utl.CombineParams(addr, "latest")

	response, err := RpcCalls("eth_getCode", params)
	if err != nil {
		return "", err
	}

	// parse the rpc response and return
	return response.(string), nil
}

// ParseTxResponse parse result based on the function constant and output type
// if the isSync is ture, the function will get the receipt of the transaction in further
func ParseTxResponse(resp interface{}, outputType []string, isWrite, isSync bool) interface{} {

	var respStr string

	//TODO
	temp, _ := json.Marshal(resp)
	_ = json.Unmarshal(temp, &respStr)

	switch {
	case !isWrite:
		return ParseNonConstantRespose(respStr, outputType)
	case isSync:
		return GetResponseByReceipt(respStr)
	default:
		return fmt.Sprintf("trasaction hash is %s\n", respStr)
	}
}

// ParseNonConstantRespose wraps the utl.BytesConverter,
// it converts the hex string response based the output type provided
func ParseNonConstantRespose(respStr string, outputType []string) interface{} {
	if len(outputType) != 0 {
		b, _ := hexutil.Decode(respStr)
		// bytesTrim := bytes.TrimRight(b, "\x00") // TODO
		// utl.Logger.Printf("result: %v\n", utl.BytesConverter(bytesTrim, outputType))
		return utl.BytesConverter(b, outputType[0])
	} else {
		return fmt.Sprintf("message call has no return value\n")
	}
}

// GetReceiptByPolling creates a channel to get the transaction receipt by polling
// The timeout is setted to 10 seconds
func GetResponseByReceipt(respStr string) interface{} {
	ch := make(chan string, 1)
	go GetReceiptByPolling(respStr, ch)

	select {
	case str := <-ch:
		runesTrim := TrimSpecialChar([]rune(str))
		str = string(runesTrim)
		/// utl.Logger.Printf("result: %s\n", str)
		return str
	case <-time.After(time.Second * 10):
		temp1 := fmt.Sprintf("\nget contract receipt timeout...more than 10 second.\n")
		temp2 := fmt.Sprintf("trasaction hash is %s\n", respStr)
		return temp1 + temp2
	}
}

func TrimSpecialChar(trimRunes []rune) []rune {

	var newBytes = make([]rune, 0)

	for _, v := range trimRunes {
		if !isSpecialChar(v) {
			newBytes = append(newBytes, v)
		}
	}

	return newBytes
}

func isSpecialChar(r rune) bool {

	if r >= 32 && r <= 126 { // ascii char
		return false
	} else if r >= 19968 && r <= 40869 { // unicode \u4e00-\u9fa5
		return false
	} else {
		return true
	}
}

// GetReceiptByPolling gets transaction receipt by polling. After getting the receipt, it
// parses the receipt and get the infos (contract address, transaction status, logs, etc.)
// The sleep time is designed to limit the times of the polling
func GetReceiptByPolling(txHash string, ch chan string) {

	// todo: fix the bug

	for {
		receipt, err := GetTransactionReceipt(txHash)

		// limit the times of the polling
		switch {
		case err != nil:
			fmt.Println(err.Error())
			fmt.Printf("try again 5s later...")
			time.Sleep(5 * sleepTime)
			fmt.Printf("try again...\n")
			continue
		case receipt == nil:
			time.Sleep(2 * sleepTime)
			continue
		}

		switch {
		case len(receipt.Logs) != 0:
			tmp, _ := hexutil.Decode(receipt.Logs[0].Data) // currently it only take the first topic
			ch <- string(tmp)
			break

		case receipt.Status == txReceiptFailureCode:
			ch <- txReceiptFailureMsg
			break

		case receipt.ContractAddress != "":
			ch <- receipt.ContractAddress
			break

		case receipt.Status == txReceiptSuccessCode:
			ch <- txReceiptSuccessMsg
			break

		}
	}
}
