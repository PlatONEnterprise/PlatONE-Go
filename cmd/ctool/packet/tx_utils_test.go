package packet

import (
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"testing"
)

const (
	TEST_ABI_FILE_PATH = "../test/test_case/wasm/contracta.cpp.abi.json"
)

func TestParseFuncFromAbi(t *testing.T) {

	abiBytes, err := utl.ParseFileToBytes(TEST_ABI_FILE_PATH)
	if err != nil {
		utils.Fatalf("%s\n", err.Error())
	}

	testCase := []struct {
		abiBytes []byte
		funcName string
	}{
		{nil, ""},
		{nil, "atransfer"},
		{abiBytes, "atransfer"},
		{abiBytes, ""},
	}

	for i, data := range testCase {
		t.Logf("case %d: \n", i)
		funcDesc, err := ParseFuncFromAbi(data.abiBytes, data.funcName)
		if err != nil {
			utils.Fatalf("%s\n", err.Error())
		} else {
			t.Logf("%s %s %s %s\n", funcDesc.Name, funcDesc.Inputs, funcDesc.Outputs, funcDesc.Constant)
		}
	}
}
