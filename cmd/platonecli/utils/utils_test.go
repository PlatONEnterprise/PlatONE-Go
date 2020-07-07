package utils

import (
	"bytes"
	"errors"
	"testing"
)

const (
	TEST_PARSE_FILE_PATH = "../test/test_case/wasm/contracta.wasm"
)

func TestParseFileToBytes(t *testing.T) {
	var testErr = errors.New("error is not nil")
	var nullFile = make([]byte, 0)

	testCase := []struct {
		path      string
		fileBytes []byte
		err       error
	}{
		{TEST_PARSE_FILE_PATH, []byte("test"), nil},        // case 0: correct
		{"", nil, testErr},                                 // case 1: path is null
		{"../test/test_case", nil, testErr},                // case 2: file directory
		{".//", nil, testErr},                              // case 3: ?
		{"../test/test_case/config.txt", nil, testErr},     // case 4: file not exist
		{"../test/test_case/test_null.txt", nullFile, nil}, // case 5: content of file is null
		{".@\"", nil, testErr},                             // case 6: invalid input
	}

	for i, data := range testCase {
		fileBytes, err := ParseFileToBytes(data.path)

		errCorrect := (err != nil) == (data.err != nil)
		fileBytesCorrect := bytes.EqualFold(fileBytes, data.fileBytes) || len(data.fileBytes) > 0

		switch {
		case err != nil && errCorrect:
			t.Logf("case %d: test file path %s: the error is %v\n", i, data.path, err.Error())
		case fileBytesCorrect:
			t.Logf("case %d: test file path %s: no error\n", i, data.path)
		default:
			t.Fail()
		}
	}
}

//TODO 能否进行合并 funcparse 与 get function params
func TestFuncParse(t *testing.T) {
	funcArray := []struct {
		funcName   string
		funcParams []string
	}{
		//{"", nil},
		{"set", []string{"123", "true"}},
		{"set", []string{""}},
		{"set", nil},
		{"set()", []string{"123", "true"}},
	}

	for i, data := range funcArray {
		t.Logf("case %d: \n", i)
		name, params := FuncParse(data.funcName, data.funcParams)
		t.Logf("Before: function name: %s, function params: %s\n", data.funcName, data.funcParams)
		t.Logf("After:  function name: %s, function params: %s\n", name, params)
	}
}

func TestGetFuncParam(t *testing.T) {
	f := []string{
		//"set",
		"set()",
		"set(\"1\",'b' , 1.2, true)",
		"set('[\"chainAdmin\",\"nodeAdmin\"]', [\"chainAdmin\",\"nodeAdmin\"])",
		"set({\"key\":\"value\"})",
		"set(\"{\"key\":\"{\"name\": \"alice\", \"score\": \"[12, 25.0, 35]\"}\"}\")",
		"set(show(), 1000 ) ",
	}
	for i, data := range f {
		t.Logf("case %d: %s", i, data)
		name, params := GetFuncNameAndParams(data)
		t.Logf("result: function name: %s, function params: %s\n", name, params)
	}

}

//TODO 重新设计测试
/*
func TestAbiParse(t *testing.T){
	testCase := []struct{
		abiPath string
		contract string
	}{
		//{"", ""},
		//{"", "__sys_UserManager"},
		//{"", CNS_PROXY_ADDRESS},
		{TEST_ABI_FILE_PATH, ""},
		{TEST_ABI_FILE_PATH, CNS_PROXY_ADDRESS},
		//{"", contract_name}, //TODO get abi on chain
		//{"", contract_address},
	}

	for i, data := range testCase{
		t.Logf("case %d: \n",i)
		_ = abiParse(data.abiPath, data.contract)

	}
}*/
