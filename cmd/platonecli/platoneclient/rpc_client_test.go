package platoneclient

import (
	"reflect"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

const (
	arg0        = "this is a test"
	arg1 uint64 = 17
	arg2 bool   = true
)

var types = []string{"string", "uint64", "bool"}

func rlpEncode(params ...interface{}) []byte {

	bin, err := rlp.EncodeToBytes(params)
	if nil != err {
		panic("error")
	}
	return bin
}

func rlpDecode(bin []byte) interface{} {
	var result interface{}
	_ = rlp.DecodeBytes(bin, &result)
	return result
}

func TestParseReceiptLogData(t *testing.T) {

	bin := rlpEncode(arg0, arg1, arg2)
	result := rlpDecode(bin)
	strResult := parseReceiptLogData(result.([]interface{}), types)
	t.Logf("the result is %v type: %v\n", result, reflect.TypeOf(result))
	t.Logf("the strResult is %v\n", strResult)
}
