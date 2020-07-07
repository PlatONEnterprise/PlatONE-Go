package utils

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/packet"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
)

const (
	EVM_TRUE        = "0000000000000000000000000000000000000000000000000000000000000001"
	EVM_FALSE       = "0000000000000000000000000000000000000000000000000000000000000000"
	EVM_69          = "0000000000000000000000000000000000000000000000000000000000000045"
	EVM_MINUS69     = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFBB"
	EVM_121         = "0000000000000000000000000000000000000000000000000000000000000079"
	EVM_MINUS121    = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF87"
	EVM_STRING_TEST = "7465737400000000000000000000000000000000000000000000000000000000"
	EVM_STRING_DAVE = "00000000000000000000000000000000000000000000000000000000000000046461766500000000000000000000000000000000000000000000000000000000"
)

func TestStringConverterDemo(t *testing.T) {
	var testCase = []struct {
		value   string
		strType string
		result  string
	}{
		{"false", "bool", EVM_FALSE},
		{"true", "bool", EVM_TRUE},
		{"", "bool", ""},
		{"-69", "int32", EVM_MINUS69},
		{"69", "uint32", EVM_69},
		{"-121", "int", EVM_MINUS121},
		{"121", "uint", EVM_121},
		{"test", "bytes4", EVM_STRING_TEST},
		{"dave", "bytes", EVM_STRING_DAVE},
		{"dave", "string", EVM_STRING_DAVE},
	}

	var call = packet.NewContractCallDemo(nil, "", 0)
	call.SetInterpreter("evm")

	for i, data := range testCase {
		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			result, err := call.Interp.StringConverter(data.value, data.strType)
			expectedBytes := common.Hex2BytesFixed(data.result, len(data.result)/2)

			switch {
			case err != nil:
				t.Logf("(%s) %s convert error: %v\t", data.strType, data.value, err.Error())
			case bytes.Equal(result, expectedBytes):
				t.Logf("(%s) %s convert result: %v\n", data.strType, data.value, result)
			default:
				t.Errorf("(%s) %s convert failed: %v expected: %v\t", data.strType, data.value, result, expectedBytes)
			}
		})
	}
}
