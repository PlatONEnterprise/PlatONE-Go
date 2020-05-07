package utils

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"testing"
)

const (
	TEST_LONG_STRING = "11111111 11111111 11111111 11111111 11111111 11111111 11111111 11111111 "
)

func ByteConvertSwitch(value interface{}, strType string) (interface{}, bool) {
	var result interface{}
	var skip bool

	switch value.(type) {
	case int32:
		hash := common.BytesToHash(Int32ToBytes(value.(int32)))
		result = BytesConverter(hash.Bytes(), strType)
	case int64:
		hash := common.BytesToHash(Int64ToBytes(value.(int64)))
		result = BytesConverter(hash.Bytes(), strType)
	case uint64:
		hash := common.BytesToHash(common.Uint64ToBytes(value.(uint64)))
		result = BytesConverter(hash.Bytes(), strType)
	case float32:
		hash := Float32ToBytes(value.(float32))
		result = BytesConverter(hash, strType)
	case float64:
		hash := Float64ToBytes(value.(float64))
		result = BytesConverter(hash, strType)
	case string:
		hash := []byte(value.(string))
		result = BytesConverter(hash, strType)
	case []byte:
		hash := []byte(value.([]byte))
		result = BytesConverter(hash, strType)
		skip = true
	}

	return result, skip || result == value
}

func TestByteConvert(t *testing.T) {

	var i int32 = -121
	var i2 int32 = 121
	var i3 int64 = -121
	var u uint64 = 121
	//var u2 uint64 = -121
	var f float32 = 1.87526
	var f2 float32 = 1.875262675
	var d float64 = 1.875262675
	var str string

	var testCase = []struct {
		value   interface{}
		strType string
	}{

		{f, "float32"},
		{f2, "float32"},
		{d, "float64"},
		{i, "int32"},
		{i2, "int32"},
		{i3, "int64"},
		{u, "uint64"},
		{"123456", "string"},
		{TEST_LONG_STRING, "string"},
		{[]byte("wxblockchain"), "[]byte"},
	}

	for j, data := range testCase {

		result, isCorrect := ByteConvertSwitch(data.value, data.strType)
		if isCorrect {
			str = "SUCCESS"
		} else {
			str = "FAILED"
		}
		t.Logf("case %d: %s (%s) %v convert result: %v", j, str, data.strType, data.value, result)
		//t.Logf("(%s) %v convert result: %v", data.strType, data.value, result)

	}

	hash := Float32ToBytes(1.875262675)
	result := BytesConverter(hash, "float32")
	t.Logf("(%s) %v convert result: %v\n", "float32", 1.875262675, result)

}

/*
func TestByteConvert2(t *testing.T) {

	hash := common.BytesToHash(Int32ToBytes(121))
	result := BytesConverter(hash.Bytes(), "int32")
	fmt.Printf("%s %v convert result: %v\n","int32",121,result)

	hash = common.BytesToHash(Int32ToBytes(-121))
	result = BytesConverter(hash.Bytes(), "int32")
	fmt.Printf("%s %v convert result: %v\n","int32",-121,result)

	hash = common.BytesToHash(common.Uint64ToBytes(121))
	result = BytesConverter(hash.Bytes(), "uint64")
	fmt.Printf("%s %v convert result: %v\n","uint64",121,result)


	hash = common.BytesToHash(common.Uint64ToBytes(-121))
	result = BytesConverter(hash.Bytes(), "uint64")
	fmt.Printf("%s %v convert result: %v\n","uint64",-121,result)

	hash2 := Float32ToBytes(1.57829)
	result = BytesConverter(hash2, "float32")
	fmt.Printf("%s %v convert result: %v\n","float32",1.57829,result)

	hash2 = Float32ToBytes(1.578291234)
	result = BytesConverter(hash2, "float32")
	fmt.Printf("%s %v convert result: %v\n","float32",1.578291234,result)

	hash2 = Float64ToBytes(1.578291234)
	result = BytesConverter(hash2, "float64")
	fmt.Printf("%s %v convert result: %v\n","float64",1.578291234,result)

	hash2 = []byte("wxblockchain")
	result = BytesConverter(hash2, "")
	fmt.Printf("%s %v convert result: %v\n","default",hash2,result)

}*/
