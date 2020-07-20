package utils

import (
	"strconv"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
)

const (
	testLongString = "11111111 11111111 11111111 11111111 11111111 11111111 11111111 11111111 "
)

func ByteConvertSwitch(value, expectValue interface{}, strType string) (interface{}, bool) {
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

	return result, skip || result == expectValue
}

func TestBytesConvert(t *testing.T) {

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
		value     interface{}
		strType   string
		expResult interface{}
	}{

		{f, "float32", f},
		{f2, "float32", f2},
		{d, "float64", d},
		{i, "int32", i},
		{i2, "int32", i2},
		{i3, "int64", i3},
		{u, "uint64", u},
		{"123456", "string", "123456"},
		{testLongString, "string", testLongString[64:]},
		{[]byte("wxblockchain"), "[]byte", []byte{119, 120, 98, 108, 111, 99, 107, 99, 104, 97, 105, 110}},
	}

	for j, data := range testCase {

		result, isCorrect := ByteConvertSwitch(data.value, data.expResult, data.strType)
		if isCorrect {
			str = "SUCCESS"
		} else {
			str = "FAILED"
			t.Fail()
		}
		t.Logf("case %d: %s (%s) %v convert result: %v", j, str, data.strType, data.value, result)
		//t.Logf("(%s) %v convert result: %v", data.strType, data.value, result)

	}

	// expected input is "1.875262675" (float64), the abi file is "float32"
	dest, err := strconv.ParseFloat("1.875262675", 32)
	if err != nil {
		t.Log(err)
	}
	t.Logf("the result is %v\n", dest)

	hash := Float32ToBytes(float32(dest))
	result := BytesConverter(hash, "float32")
	t.Logf("(%s) %v convert result: %v\n", "float32", 1.8752626, result)

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
