package platoneclient

import (
	"context"
	"fmt"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"

	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/packet"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/rpc"
)

type pClient struct {
	c *rpc.Client
}

func SetupClient(url string) *pClient {
	var client = new(pClient)
	var err error

	// client.c, _ = ethclient.Dial(url)
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
	utl.PrintRequest(params)

	// send the RPC calls
	/*
		resp, err := utl.RpcCalls(action, params)
		if err != nil {
			utils.Fatalf(utl.ErrSendTransacionFormat, err.Error())
		}*/
	var resp interface{}
	err = client.c.Call(&resp, action, params...)
	if err != nil {
		utils.Fatalf(utl.ErrSendTransacionFormat, err.Error())
	}

	return ParseTxResponse(resp, client, outputType, isWrite, isSync)
}

func (client *pClient) GetResponseByReceipt(respStr string) interface{} {
	ch := make(chan string, 1)
	go client.GetReceiptByPolling(respStr, ch)

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

func (client *pClient) GetReceiptByPolling(txHash string, ch chan string) {

	for {
		receipt, err := client.GetTransactionReceipt(txHash)

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
