package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"

	"github.com/stretchr/testify/assert"
)

const (
	testParseFile = "../test/test_case/wasm/contracta.wasm"
)

func TestParseFileToBytes(t *testing.T) {
	var testErr = errors.New("error is not nil")
	var nullFile = make([]byte, 0)

	testCase := []struct {
		path      string
		fileBytes []byte
		err       error
	}{
		{testParseFile, []byte("test"), nil},               // case 0: correct
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
			t.Logf("case %d: test file path %s: fileBytes: %v, the error is %v\n", i, data.path, fileBytes, err.Error())
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
		expParams  []string
	}{
		//{"", nil},
		{"set", []string{"123", "true"}, []string{"123", "true"}},
		{"set", []string{""}, []string{""}},
		{"set", nil, nil},
		{"set()", []string{"123", "true"}, []string{"123", "true"}},
	}

	for i, data := range funcArray {
		t.Logf("case %d: \n", i)
		name, params := FuncParse(data.funcName, data.funcParams)
		assert.Equal(t, "set", name, "name parse FAILED")
		assert.Equal(t, data.expParams, params, "params parse FAILED")
		t.Logf("Before: function name: %s, function params: %s\n", data.funcName, data.funcParams)
		t.Logf("After:  function name: %s, function params: %s\n", name, params)
	}
}

func TestGetFuncParam(t *testing.T) {
	testCases := []struct {
		function  string
		expName   string
		expParams []string
	}{
		{"set()", "set", nil},
		{"set(\"1\",'b' , 1.2, true)", "set", []string{"1", "b", "1.2", "true"}},
		{"set('[\"chainAdmin\",\"nodeAdmin\"]', [\"chainAdmin\",\"nodeAdmin\"])", "set", []string{"[\"chainAdmin\",\"nodeAdmin\"]", "[\"chainAdmin\",\"nodeAdmin\"]"}},
		{"set({\"key\":\"value\"})", "set", []string{"{\"key\":\"value\"}"}},
		{"set(\"{\"key\":\"{\"name\": \"alice\", \"score\": \"[12, 25.0, 35]\"}\"}\")", "set", []string{"{\"key\":\"{\"name\":\"alice\",\"score\":\"[12,25.0,35]\"}\"}"}},
		{"set(show(), 1000 ) ", "set", []string{"show()", "1000"}},
	}

	for i, data := range testCases {
		t.Logf("case %d: %s", i, data)
		name, params := GetFuncNameAndParams(data.function)
		assert.Equal(t, data.expName, name, "name parse FAILED")
		assert.Equal(t, data.expParams, params, "params parse FAILED")

		t.Logf("result: function name: %s, function params: %s\n", name, params)
	}
}

func TestGetFuncParams(t *testing.T) {
	testCase := "\"1\",'b' , 1.2, true"
	result := abi.GetFuncParams(testCase)

	t.Log(result)
}

func TestStructType(t *testing.T) {
	var testCase = make(map[string]interface{}, 0)
	var S struct{}
	var str = "{\"components\": [{\"internalType\": \"int32\",\"name\": \"x\",\"type\": \"int32\"},{\"internalType\": \"int32\",\"name\": \"y\",\"type\": \"int32\"}],\"internalType\": \"struct TupleTest.Point\",\"name\": \"num\",\"type\": \"tuple\"}"

	testCase["test"] = 2
	_ = json.Unmarshal([]byte(str), &S)

	t.Log(reflect.ValueOf(testCase).Kind())
	t.Log(S)
}

func TestPrintJson(t *testing.T) {
	str := "{\"account\":\"\",\"url\":\"http://127.0.0.1:6794\",\"keystore\":" +
		"\"../../release/linux/data/node-0/keystore/UTC--2020-07-27T03-08-50.310696196Z--8bc9cbeac3b9e89c47b3d0f21ba93b8a6e0aa818\"}"
	result := PrintJson([]byte(str))
	t.Logf("\n%s", result)
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
