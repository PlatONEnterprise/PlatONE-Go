package utils

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/test"
	"testing"
)

func TestHttpPost(t *testing.T) {

	server := test.MockServer("http")
	SetHttpUrl(server.URL)

	param := JsonParam{
		Jsonrpc: "2.0",
		Params:  "test only",
		Id:      1,
	}

	for i := 1; i <= 3; i++ {
		param.Method = fmt.Sprintf("test%d", i)

		result, err := HttpPost(param)
		if err != nil {
			t.Logf("case %d: result is %s, error is %s", i, result, err.Error())
		} else {
			t.Logf("case %d: result is %s, http post success", i, result)
		}
	}

}
