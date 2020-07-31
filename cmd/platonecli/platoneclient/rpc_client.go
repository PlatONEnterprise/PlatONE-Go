package platoneclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"

	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/packet"
	"github.com/PlatONEnetwork/PlatONE-Go/rpc"
)

const (
	sleepTime = 1000000000 // 1 seconds
)

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

func (p *pClient) GetTransactionReceipt(txHash string) (*packet.Receipt, error) {

	var response interface{}
	_ = p.c.Call(&response, "eth_getTransactionReceipt", txHash)
	if response == nil {
		return nil, nil
	}

	// parse the rpc response
	receipt, err := packet.ParseTxReceipt(response)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// messageCall extract the common parts of the transaction based calls
// including eth_call, eth_sendTransaction, and eth_sendRawTransaction
func (client *pClient) MessageCall(dataGen packet.MsgDataGen, keyfile string, tx *packet.TxParams) (interface{}, bool, error) {

	// combine the data based on the types of the calls (contract call, inner call or deploy call)
	data, outputType, isWrite, err := dataGen.CombineData()
	if err != nil {
		errStr := fmt.Sprintf(utils.ErrPackDataFormat, err.Error())
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
		errStr := fmt.Sprintf(utils.ErrSendTransacionFormat, err.Error())
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
		errStr := fmt.Sprintf(utils.ErrPackDataFormat, err.Error())
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
		errStr := fmt.Sprintf(utils.ErrSendTransacionFormat, err.Error())
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
		return utils.BytesConverter(b, outputType[0])
	} else {
		return fmt.Sprintf("message call has no return value\n")
	}
}

func (client *pClient) GetReceiptByPolling(txHash string) (*packet.Receipt, error) {
	ch := make(chan interface{}, 1)
	go client.getReceiptByPolling(txHash, ch)

	select {
	case receipt := <-ch:
		return receipt.(*packet.Receipt), nil

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
