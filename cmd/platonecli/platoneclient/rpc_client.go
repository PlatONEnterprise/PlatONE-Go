package platoneclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/common/byteutil"

	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"

	"github.com/PlatONEnetwork/PlatONE-Go/rlp"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"

	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/packet"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/rpc"
)

type pClient struct {
	c *rpc.Client
}

func SetupClient(url string) *pClient {
	var client = new(pClient)
	var err error

	client.c, err = rpc.DialContext(context.Background(), "http://"+url)
	if err != nil {
		utils.Fatalf(err.Error())
	}

	return client
}

func (p *pClient) GetTransactionReceipt(txHash string) (*Receipt, error) {

	var response interface{}
	_ = p.c.Call(&response, "eth_getTransactionReceipt", txHash)

	// parse the rpc response
	receipt, err := ParseTxReceipt(response)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// messageCall extract the common parts of the transaction based calls
// including eth_call, eth_sendTransaction, and eth_sendRawTransaction
func (client *pClient) MessageCall(call packet.MessageCall, keyfile string, tx *packet.TxParams, isSync bool) interface{} {

	// combine the data based on the types of the calls (contract call, inner call or deploy call)
	data, outputType, isWrite, err := call.CombineData()
	if err != nil {
		utils.Fatalf(utl.ErrPackDataFormat, err.Error())
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
		utils.Fatalf(utl.ErrSendTransacionFormat, err.Error())
	}

	// parse transaction response
	respStr := fmt.Sprint(resp)

	switch {
	case !isWrite:
		return ParseNonConstantRespose(respStr, outputType)
	case isSync:
		return client.GetResponseByReceipt(respStr, call)
	default:
		return fmt.Sprintf("trasaction hash is %s\n", respStr)
	}
	/// return ParseTxResponse(resp, client, outputType, isWrite, isSync)
}

func (client *pClient) GetResponseByReceipt(respStr string, call packet.MessageCall) interface{} {
	ch := make(chan interface{}, 1)
	txHash := fmt.Sprintf("trasaction hash is %s\n", respStr)
	go client.GetReceiptByPolling(respStr, call, ch)

	select {
	case str := <-ch:
		// todo:
		// runesTrim := TrimSpecialChar([]rune(str))
		// str = string(runesTrim)
		/// utl.Logger.Printf("result: %s\n", str)
		// fmt.Printf(txHash)
		return str
	case <-time.After(time.Second * 10):
		temp := fmt.Sprintf("\nget contract receipt timeout...more than 10 second.\n")
		return temp + txHash
	}
}

// todo: end goroutine?
func (client *pClient) GetReceiptByPolling(txHash string, call packet.MessageCall, ch chan interface{}) {

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

		receiptBytes, _ := json.Marshal(receipt)
		utl.PrintJson(receiptBytes)

		switch {
		case len(receipt.Logs) != 0:

			// currently it only take the first log
			var resp []interface{}
			dataBytes, _ := hexutil.Decode(receipt.Logs[0].Data)
			topicTypes := findLogTopic(receipt.Logs[0].Topics[0], call.GetAbiBytes())
			err = rlp.DecodeBytes(dataBytes, &resp)
			if err != nil {
				fmt.Printf("the error is %v\n", err)
			}

			// future work: different calls(evm or wasm) may have diff. decoding
			result := parseReceiptLogData(resp, topicTypes)
			ch <- result

		case receipt.Status == txReceiptFailureCode:
			ch <- txReceiptFailureMsg

		case receipt.ContractAddress != "":
			ch <- receipt.ContractAddress

		case receipt.Status == txReceiptSuccessCode:
			ch <- txReceiptSuccessMsg

		}
	}
}

func findLogTopic(topic string, abiBytes []byte) []string {
	var types []string

	abiFunc, _ := packet.ParseAbiFromJson(abiBytes)

	for _, data := range abiFunc {
		if data.Type != "event" {
			continue
		}

		strings.EqualFold(logTopicEncode(data.Name), topic)
		for _, v := range data.Inputs {
			types = append(types, v.Type)
		}
		break
	}

	return types
}

func parseReceiptLogData(data []interface{}, types []string) string {
	var str string

	for i, v := range data {

		if len(v.([]uint8)) == 0 {
			continue
		}
		result := ConvertRlpBytesTo(v.([]uint8), types[i])
		str += fmt.Sprintf("%v", result)
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
	// return byteutil.BytesToUint64(b)
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
