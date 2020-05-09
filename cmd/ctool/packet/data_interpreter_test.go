package packet

import (
	"testing"
)

func TestStringConverter(t *testing.T) {
	var call = NewContractCallDemo(nil, "", 0)
	var interArray = []string{"wasm", "evm"}

	var testCase = []struct {
		value   string
		strType string
	}{
		{"false", "bool"},
		{"true", "bool"},
		{"", "bool"},
		{"1.87526", "float32"},
		{"1.87526", "float64"},
		{"1.875262675", "float32"},
		{"1.875262675", "float64"},
		{"-121", "int32"},
		{"-121", "uint32"},
		{"-121", "int64"},
		{"121", "int32"},
		{"121", "uint32"},
		{"121", "uint64"},
		{"test", "string"},
	}

	for _, interpreter := range interArray {
		call.SetInterpreter(interpreter)
		t.Logf("testing %s interpreter converter:\n", interpreter)
		t.Logf("--------------------Start---------------------\n")
		for _, data := range testCase {
			result, err := call.Interp.StringConverter(data.value, data.strType)
			if err != nil {
				t.Logf("(%s) %s convert error: %s\t", data.strType, data.value, err.Error())
			} else {
				t.Logf("(%s) %s convert result: %v\n", data.strType, data.value, result)
			}

		}
		t.Logf("--------------------End-----------------------\n")
	}

}
