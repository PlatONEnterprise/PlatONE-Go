package packet

import (
	"testing"
	"time"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
)

const (
	TEST_ABI_FILE_PATH = "../test/test_case/wasm/contracta.cpp.abi.json"
)

func TestGetNonceRand(t *testing.T) {
	r1 := getNonceRand()

	time.Sleep(1000000000) // one second
	r2 := getNonceRand()

	if r1 == r2 {
		t.Fail()
	}

	t.Logf("r1 is %v, r2 is %v\n", r1, r2)
}

func TestParseFuncFromAbi(t *testing.T) {

	abiBytes, err := utl.ParseFileToBytes(TEST_ABI_FILE_PATH)
	if err != nil {
		utils.Fatalf("%s\n", err.Error())
	}

	testCase := []struct {
		abiBytes []byte
		funcName string
	}{
		{abiBytes, "atransfer"},                 // case 1: correct
		{nil, ""},                               // case 2: null
		{abiBytes, ""},                          // case 3: null
		{abiBytes, " atran sfer "},              // case 4: function name invalid
		{[]byte{32, 13, 14, 23}, " atransfer "}, // case 5: abi bytes invalid
		{[]byte{}, " atransfer "},               // case 6: abi bytes invalid

	}

	for i, data := range testCase {
		funcDesc, err := ParseFuncFromAbi(data.abiBytes, data.funcName)

		switch {
		case err != nil:
			t.Logf("case %d: %s\n", i+1, err.Error())
		case funcDesc != nil && funcDesc.Name == data.funcName:
			t.Logf("case %d: %s %v %v %s\n", i+1, funcDesc.Name, funcDesc.Inputs, funcDesc.Outputs, funcDesc.Constant)
		default:
			t.Fail()
		}
	}
}
