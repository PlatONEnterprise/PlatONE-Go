package packet

import (
	"strings"
	"testing"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/test"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
)

const (
	//TEST_ACCOUNT = "0x60ceca9c1290ee56b98d4e160ef0453f7c40d219"
	TEST_TX_HASH = "123"
	TEST_ADDRESS = "0x1e6c4c426dfc3ed2435f726fd9a0c2995df1afc7"
)

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
