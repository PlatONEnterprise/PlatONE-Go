package platoneclient

import (
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/test"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
)

/*
func TestRpcCalls(t *testing.T){
	rpcCalls("eth_sendTransaction", "for test only")
}*/

func TestGetNonce(t *testing.T) {

	server := test.MockServer("rpc")
	SetHttpUrl(server.URL)

	nonce, _ := GetNonce(common.HexToAddress(test.TestAccount))
	t.Logf("the nonce is %v\n", nonce)
}
