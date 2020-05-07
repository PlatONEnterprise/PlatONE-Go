package packet

import (
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/test"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"strings"
	"testing"
	"time"
)

const (
	//TEST_ACCOUNT = "0x60ceca9c1290ee56b98d4e160ef0453f7c40d219"
	TEST_TX_HASH = "123"
	TEST_ADDRESS = "0x1e6c4c426dfc3ed2435f726fd9a0c2995df1afc7"
)

/*
func TestRpcCalls(t *testing.T){
	rpcCalls("eth_sendTransaction", "for test only")
}*/

func TestGetNonce(t *testing.T) {

	server := test.MockServer("rpc")
	utils.SetHttpUrl(server.URL)

	nonce := GetNonce(common.HexToAddress(test.TEST_ACCOUNT))
	t.Logf("the nonce is %v\n", nonce)
}

func TestGetAddressByName(t *testing.T) {

	server := test.MockServer("rpc")
	utils.SetHttpUrl(server.URL)

	address, _ := GetAddressByName("tofu")

	if strings.EqualFold(address, TEST_ADDRESS) {
		t.Logf("SUCCESS, the address is %s\n", address)
	} else {
		t.Logf("FAILED\n")
	}
}

func TestGetTransactionReceipt(t *testing.T) {

	server := test.MockServer("rpc")
	utils.SetHttpUrl(server.URL)

	ch := make(chan string, 1)

	go GetReceiptByPolling(TEST_TX_HASH, ch)

	select {
	case address := <-ch:
		t.Logf("result: %s\n", address)
	case <-time.After(time.Second * 10):
		t.Logf("get contract receipt timeout...more than 10 second.\n")
	}
}

func TestContractCallCommon(t *testing.T) {
	// ContractCallCommon("", nil, nil, Cns{}, "")
	ContractCallCommon("getName", nil, nil, Cns{}, "")
}

//TODO
/*
func TestCombineJson(t *testing.T) {

	ctx := cli.Context{

	}

	ctx.Args()
}*/
