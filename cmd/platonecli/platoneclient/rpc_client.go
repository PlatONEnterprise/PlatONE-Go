package platoneclient

import (
	"context"
	"errors"
	"fmt"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"

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
func (client *pClient) MessageCall(dataGen packet.MsgDataGen, keyfile *utils.Keyfile, tx *packet.TxParams) ([]interface{}, bool, error) {
	var result = make([]interface{}, 1)

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
	/// utils.PrintRequest(params)

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
		result = dataGen.ParseNonConstantResponse(respStr, outputType)
	} else {
		result[0] = respStr
	}

	return result, isWrite, nil
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

func (p *pClient) GetRevertMsg(msg *packet.TxParams, blockNum uint64) ([]byte, error) {

	var hex = new(hexutil.Bytes)
	err := p.c.Call(hex, "eth_call", msg, hexutil.EncodeUint64(blockNum))
	if err != nil {
		return nil, err
	}

	return *hex, nil
}
