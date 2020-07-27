package main

import (
	"reflect"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common"

	"github.com/stretchr/testify/assert"
)

const (
	testConfigFilePath = "./test/test_case/config.json"
	testAccount        = "0x60ceca9c1290ee56b98d4e160ef0453f7c40d219"
)

//TODO ???
func TestWriteConfigFile(t *testing.T) {
	WriteConfigFile(testConfigFilePath, "account", testAccount)
	//writeConfigFile(TEST_CONFIG_FILE_PATH, "wrong_key", "0x000..00")
	config := ParseConfigJson(testConfigFilePath)

	t.Logf("the config values are %+v", *config)
}

func TestParamParse(t *testing.T) {
	var r string

	testCase := []struct {
		param     string
		paramName string
		result    interface{}
	}{
		{testAccount, "contract", true},
		{"Alice_02", "contract", false},
		//{"Alice.bob", "contract"},
		//{"na*&2", "contract"},
		//{"-1", "p2pPort"},
		{"123", "p2pPort", int64(123)},
		//{"123456", "p2pPort"},
		{"123456", "", "123456"},
		{"invalid", "status", 2},
		{"approve", "operation", "2"},
		{"observer", "type", 0},
	}

	for i, data := range testCase {
		result := ParamParse(data.param, data.paramName)
		assert.Equal(t, data.result, result, "FAILED")

		t.Logf("%s: case %d: Before: (%v) %s, After convert: (%v) %v\n", r, i, reflect.TypeOf(data.param), data.param, reflect.TypeOf(result), result)
	}
}

func TestChainParamConvert(t *testing.T) {
	testCase := []struct {
		param     string
		paramName string
		expResult interface{}
	}{
		{"0x002", "value", "0x2"},
		{"0020", "value", "0x14"},
		{"-20", "value", "-0x14"},
		{"0xFD", "value", "0xfd"}, //TODO 负数?
		{"0x020", "gas", "0x20"},
		{"002302", "gas", "0x8fe"},
		{testAccount, "to", common.HexToAddress(testAccount)},
	}

	for i, data := range testCase {
		result := chainParamConvert(data.param, data.paramName)
		assert.Equal(t, data.expResult, result, "FAILED")

		t.Logf("case %d: Before: (%v) %s, After convert: (%v) %v\n", i, reflect.TypeOf(data.param), data.param, reflect.TypeOf(result), result)
	}
}

func TestParamValid(t *testing.T) {
	testCase := []struct {
		param     string
		paramName string
	}{
		{"*", "fw"},
		{testAccount, "fw"},
		{testAccount, "contract"},
		{"Alice_02", "contract"},
		//{"Alice.bob", "contract"},
		{"accept", "action"},
		//{"xxx", "action"},
		{"127.0.0.1:6791", "url"},
		{"127.0.0.1", "externalIP"},
		{"[\"nodeAdmin \"]", "roles"},
		{"fd.deng@wxblockchain.com", "email"},
		{"13240283946", "mobile"},
		{"0.0.0.1", "version"},
		{"-123", "num"},
		{"+13", "num"},
		{"12459234", "num"},
		// {"+-123", "num"},
	}

	for i, data := range testCase {
		paramValid(data.param, data.paramName)
		t.Logf("case %d: the %s \"%s\" is valid\n", i, data.paramName, data.param)
	}

}
