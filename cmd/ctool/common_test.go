package main

import (
	"reflect"
	"testing"
)

const (
	TEST_CONFIG_FILE_PATH = "./test/test_case/config.json"
	TEST_ACCOUNT          = "0x60ceca9c1290ee56b98d4e160ef0453f7c40d219"
)

//TODO ???
func TestWriteConfigFile(t *testing.T) {
	WriteConfigFile(TEST_CONFIG_FILE_PATH, "account", TEST_ACCOUNT)
	//writeConfigFile(TEST_CONFIG_FILE_PATH, "wrong_key", "0x000..00")
	config := ParseConfigJson(TEST_CONFIG_FILE_PATH)

	t.Logf("the config values are %+v", *config)
}

func TestParamParse(t *testing.T) {
	var r string

	testCase := []struct {
		param     string
		paramName string
		result    interface{}
	}{
		{TEST_ACCOUNT, "contract", true},
		{"Alice_02", "contract", false},
		//{"Alice.bob", "contract"},
		//{"na*&2", "contract"},
		//{"-1", "p2pPort"},
		{"123", "p2pPort", 123},
		//{"123456", "p2pPort"},
		{"123456", "", "123456"},
		{"invalid", "status", 2},
		{"approve", "operation", 2},
		{"observer", "type", 0},
	}

	for i, data := range testCase {
		result := ParamParse(data.param, data.paramName)
		if reflect.ValueOf(result) == data.result {
			r = "SUCCESS"
		} else {
			r = "FAILED"
			// t.Failed()
		}

		t.Logf("%s: case %d: Before: (%v) %s, After convert: (%v) %v\n", r, i, reflect.TypeOf(data.param), data.param, reflect.TypeOf(result), result)
	}
}

func TestMessageCall(t *testing.T) {
	// messageCall(nil, nil, nil, "", 0)
}
