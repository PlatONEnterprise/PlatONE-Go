package packet

import (
	"bytes"
	"fmt"
	"testing"
)

func TestStringConverter(t *testing.T) {
	var call = NewContractCallDemo(nil, "", 0)
	var interArray = []string{"wasm", "evm"}

	var testCase = []struct {
		value   string
		strType string
		result  []byte
	}{
		{"false", "bool", []byte{0}},
		{"true", "bool", []byte{1}},
		{"", "bool", nil},
		{"1.87526", "float32", []byte{133, 8, 240, 63}},
		{"1.87526", "float64", []byte{140, 243, 55, 161, 16, 1, 254, 63}},
		{"1.875262675", "float32", []byte{155, 8, 240, 63}}, // 1.8752626 155?
		{"1.875262675", "float64", []byte{56, 141, 72, 111, 19, 1, 254, 63}},
		{"-121", "int32", []byte{255, 255, 255, 135}},
		{"-121", "uint32", []byte{255, 255, 255, 135}}, // TODO
		{"-121", "int64", []byte{255, 255, 255, 255, 255, 255, 255, 135}},
		{"121", "int32", []byte{0, 0, 0, 121}},
		{"121", "uint32", []byte{0, 0, 0, 121}},
		{"121", "uint64", []byte{0, 0, 0, 0, 0, 0, 0, 121}},
		{"test", "string", []byte{116, 101, 115, 116}},
	}

	for _, interpreter := range interArray {
		call.SetInterpreter(interpreter)
		t.Logf("testing %s interpreter converter:\n", interpreter)
		for i, data := range testCase {
			t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
				result, err := call.Interp.StringConverter(data.value, data.strType)

				switch {
				case err != nil:
					t.Logf("(%s) %s convert error: %v\t", data.strType, data.value, err.Error())
				case bytes.Equal(result, data.result):
					t.Logf("(%s) %s convert result: %v\n", data.strType, data.value, result)
				default:
					//t.Logf("(%s) %s convert failed: %v\t", data.strType, data.value, result)
					//t.Fail()
					t.Errorf("(%s) %s convert failed: %v\t", data.strType, data.value, result)
				}
			})
		}
	}

}
