package platoneclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/rlp"

	"github.com/PlatONEnetwork/PlatONE-Go/common/byteutil"

	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"

	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/packet"
	"github.com/PlatONEnetwork/PlatONE-Go/rpc"
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

type pClient struct {
	c *rpc.Client
}

func SetupClient(url string) (*pClient, error) {
	var client = new(pClient)
	var err error

	client.c, err = rpc.DialContext(context.Background(), url)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (p *pClient) GetTransactionReceipt(txHash string) (*Receipt, error) {

	var response interface{}
	_ = p.c.Call(&response, "eth_getTransactionReceipt", txHash)
	if response == nil {
		return nil, nil
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

	temp, _ := json.Marshal(response)
	err := json.Unmarshal(temp, receipt)
	if err != nil {
		// LogErr.Printf(ErrUnmarshalBytesFormat, "transaction receipt", err.Error())
		errStr := fmt.Sprintf(utl.ErrUnmarshalBytesFormat, "transaction receipt", err.Error())
		return nil, errors.New(errStr)
	}

	return receipt, nil
}

// messageCall extract the common parts of the transaction based calls
// including eth_call, eth_sendTransaction, and eth_sendRawTransaction
func (client *pClient) MessageCall(dataGen packet.MsgDataGen, keyfile string, tx *packet.TxParams) (interface{}, bool, error) {

	// combine the data based on the types of the calls (contract call, inner call or deploy call)
	data, outputType, isWrite, err := dataGen.CombineData()
	if err != nil {
		errStr := fmt.Sprintf(utl.ErrPackDataFormat, err.Error())
		return nil, false, errors.New(errStr)
	}

	// packet the transaction and select the transaction based calls
	tx.Data = data
	params, action := tx.SendMode(isWrite, keyfile)

	// print the RPC JSON param to the terminal
	/// utl.PrintRequest(params)

	// send the RPC calls
	var resp interface{}
	err = client.c.Call(&resp, action, params...)
	if err != nil {
		errStr := fmt.Sprintf(utl.ErrSendTransacionFormat, err.Error())
		return nil, false, errors.New(errStr)
	}

	// parse transaction response
	respStr := fmt.Sprint(resp)
	if !isWrite {
		return ParseNonConstantResponse(respStr, outputType), false, nil
	}

	return respStr, true, nil
}

func (client *pClient) MessageCallOld(dataGenerator packet.MsgDataGen, keyfile string, tx *packet.TxParams, isSync bool) (interface{}, error) {

	// combine the data based on the types of the calls (contract call, inner call or deploy call)
	data, outputType, isWrite, err := dataGenerator.CombineData()
	if err != nil {
		errStr := fmt.Sprintf(utl.ErrPackDataFormat, err.Error())
		return nil, errors.New(errStr)
	}

	// packet the transaction and select the transaction based calls
	tx.Data = data
	params, action := tx.SendMode(isWrite, keyfile)

	// print the RPC JSON param to the terminal
	/// utl.PrintRequest(params)

	// send the RPC calls
	var resp interface{}
	err = client.c.Call(&resp, action, params...)
	if err != nil {
		errStr := fmt.Sprintf(utl.ErrSendTransacionFormat, err.Error())
		return nil, errors.New(errStr)
	}

	// parse transaction response
	respStr := fmt.Sprint(resp)

	switch {
	case !isWrite:
		return ParseNonConstantResponse(respStr, outputType), nil
	case isSync:
		result, err := client.GetReceiptByPolling(respStr)
		if err != nil {
			return respStr, nil
		}

		receiptBytes, _ := json.Marshal(result)
		return string(receiptBytes), nil
	default:
		/// return fmt.Sprintf("trasaction hash: %s\n", respStr), nil
		return respStr, nil
	}
}

// ParseNonConstantRespose wraps the utl.BytesConverter,
// it converts the hex string response based the output type provided
func ParseNonConstantResponse(respStr string, outputType []string) interface{} {
	if len(outputType) != 0 {
		b, _ := hexutil.Decode(respStr)
		// utl.Logger.Printf("result: %v\n", utl.BytesConverter(bytesTrim, outputType))
		return utl.BytesConverter(b, outputType[0])
	} else {
		return fmt.Sprintf("message call has no return value\n")
	}
}

func (client *pClient) GetReceiptByPolling(txHash string) (*Receipt, error) {
	ch := make(chan interface{}, 1)
	go client.getReceiptByPolling(txHash, ch)

	select {
	case receipt := <-ch:
		return receipt.(*Receipt), nil

	case <-time.After(time.Second * 10):
		// temp := fmt.Sprintf("\nget contract receipt timeout...more than %d second.\n", 10)
		// return temp + txHash

		errStr := fmt.Sprintf("get contract receipt timeout...more than %d second.", 10)
		return nil, errors.New(errStr)
	}
}

// todo: end goroutine?
func (client *pClient) getReceiptByPolling(txHash string, ch chan interface{}) {

	for {
		receipt, err := client.GetTransactionReceipt(txHash)

		// limit the times of the polling
		if err != nil {
			fmt.Println(err.Error())
			fmt.Printf("try again 5s later...")
			time.Sleep(5 * sleepTime)
			fmt.Printf("try again...\n")
			continue
		}

		if receipt == nil {
			time.Sleep(1 * sleepTime)
			continue
		}

		ch <- receipt
	}
}

func ReceiptParsing(receipt *Receipt, abiBytes []byte) string {
	var result string

	switch {
	case len(receipt.Logs) != 0:
		for i, elog := range receipt.Logs {
			var rlpList []interface{}

			eventName, topicTypes := findLogTopic(elog.Topics[0], abiBytes)
			if len(topicTypes) == 0 {
				continue
			}

			dataBytes, _ := hexutil.Decode(elog.Data)
			err := rlp.DecodeBytes(dataBytes, &rlpList)
			if err != nil {
				fmt.Printf("the error is %v\n", err)
			}
			result = fmt.Sprintf("\nEvent[%d]: %s ", i, eventName)
			result += parseReceiptLogData(rlpList, topicTypes)
		}

	case receipt.Status == txReceiptFailureCode:
		result = txReceiptFailureMsg

	case receipt.ContractAddress != "":
		result = receipt.ContractAddress

	case receipt.Status == txReceiptSuccessCode:
		result = txReceiptSuccessMsg
	}

	return result
}

func findLogTopic(topic string, abiBytes []byte) (string, []string) {
	var types []string
	var name string

	abiFunc, _ := packet.ParseAbiFromJson(abiBytes)

	for _, data := range abiFunc {
		if data.Type != "event" {
			continue
		}

		if strings.EqualFold(logTopicEncode(data.Name), topic) {
			name = data.Name
			for _, v := range data.Inputs {
				types = append(types, v.Type)
			}
			break
		}
	}

	return name, types
}

func parseReceiptLogData(data []interface{}, types []string) string {
	var str string

	for i, v := range data {
		result := ConvertRlpBytesTo(v.([]uint8), types[i])
		str += fmt.Sprintf("%v ", result)
	}

	return str
}

func logTopicEncode(name string) string {
	return common.BytesToHash(crypto.Keccak256([]byte(name))).String()
}

func ConvertRlpBytesTo(input []byte, targetType string) interface{} {
	v, ok := Bytes2X_CMD[targetType]
	if !ok {
		panic("unsupported type")
	}

	return reflect.ValueOf(v).Call([]reflect.Value{reflect.ValueOf(input)})[0].Interface()
}

var Bytes2X_CMD = map[string]interface{}{
	"string": byteutil.BytesToString,

	// "uint8":  RlpBytesToUint,
	"uint16": RlpBytesToUint16,
	"uint32": RlpBytesToUint32,
	"uint64": RlpBytesToUint64,

	// "uint8":  RlpBytesToUint,
	"int16": RlpBytesToUint16,
	"int32": RlpBytesToUint32,
	"int64": RlpBytesToUint64,

	"bool": RlpBytesToBool,
}

func RlpBytesToUint16(b []byte) uint16 {
	b = common.LeftPadBytes(b, 32)
	result := common.CallResAsUint32(b)
	return uint16(result)
}

func RlpBytesToUint32(b []byte) uint32 {
	b = common.LeftPadBytes(b, 32)
	return common.CallResAsUint32(b)
}

func RlpBytesToUint64(b []byte) uint64 {
	b = common.LeftPadBytes(b, 32)
	return common.CallResAsUint64(b)
}

func RlpBytesToBool(b []byte) bool {
	if bytes.Compare(b, []byte{1}) == 0 {
		return true
	}
	return false
}

/*
func RlpBytesToUintV2(b []byte) interface{} {
	var val interface{}

	for _, v := range b {
		val = val << 8
		val |= uint(v)
	}

	return val
}*/
