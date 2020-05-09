package utils

import (
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/test"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"testing"
)

/*
func TestRpcCalls(t *testing.T){
	rpcCalls("eth_sendTransaction", "for test only")
}*/

func TestGetNonce(t *testing.T) {

	server := test.MockServer("rpc")
	SetHttpUrl(server.URL)

	nonce := GetNonce(common.HexToAddress(test.TEST_ACCOUNT))
	t.Logf("the nonce is %v\n", nonce)
}
