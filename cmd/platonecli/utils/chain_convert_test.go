package utils

import (
	"reflect"
	"testing"
)

func TestChainParamConvert(t *testing.T) {
	testCase := []struct {
		param     string
		paramName string
	}{
		{"0x002", "value"},
		{"0020", "value"},
		{"-20", "value"},
		{"0xFD", "value"}, //TODO 负数?
		{"0x020", "gas"},
		{"002302", "gas"},
		{TEST_ACCOUNT, "to"},
	}

	for i, data := range testCase {
		result := ChainParamConvert(data.param, data.paramName)
		t.Logf("case %d: Before: (%v) %s, After convert: (%v) %v\n", i, reflect.TypeOf(data.param), data.param, reflect.TypeOf(result), result)
	}
}
