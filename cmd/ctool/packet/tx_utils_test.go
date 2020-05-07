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
		abiBytes   []byte
		funcName   string
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
		}else{
			t.Logf("%s %s %s %s\n", funcDesc.Name, funcDesc.Inputs, funcDesc.Outputs, funcDesc.Constant)
		}
	}
}

//没有测试待必要性？
/*
func TestParseAbiFromJson(t *testing.T) {

	//dir, _ := os.Getwd()
	//filePath := dir + TEST_ABI_FILE_PATH
	abiBytes, _ := parseFileToBytes(TEST_ABI_FILE_PATH)
	a, e := parseAbiFromJson(abiBytes)
	if e != nil {
		t.Fatalf("parse abi json error! \n， %s", e.Error())
	}
	t.Log(a)
	marshal, _ := json.Marshal(a)
	t.Log(string(marshal))
}*/
