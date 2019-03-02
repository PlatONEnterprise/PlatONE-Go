package common

import (
	"github.com/PlatONnetwork/PlatON-Go/rlp"
)

var (
	innerCall func(Address, []byte) ([]byte) = nil
)

func SetInnerCallFunc(f func(Address, []byte) ([]byte)) {
	innerCall = f
}

func InnerCall(conAddr Address, funcName string, params []interface{}) ([]byte) {
	if innerCall == nil {
		return nil
	} else {
		return innerCall(conAddr, GenCallData(funcName, params))
	}
}

func GenCallData(funcName string, params []interface{}) ([]byte) {
	data := [][]byte{}
	data = append(data, Int64ToBytes(2))
	data = append(data, []byte(funcName))
	
	for _, p := range params {
		var t int
		var i int64
		var u uint64
		var s string
		switch p.(type) {
			case int:
				i = int64(p.(int))
				t = 1
			case int8:
				i = int64(p.(int8))
				t = 1
			case int16:
				i = int64(p.(int16))
				t = 1
			case int32:
				i = int64(p.(int32))
				t = 1
			case int64:
				i = int64(p.(int64))
				t = 1

			case uint:
				u = uint64(p.(uint))
				t = 2
			case uint8:
				u = uint64(p.(uint8))
				t = 2
			case uint16:
				u = uint64(p.(uint16))
				t = 2
			case uint32:
				u = uint64(p.(uint32))
				t = 2
			case uint64:
				u = uint64(p.(uint64))
				t = 2

			case bool:
				if p.(bool) {
					u = 1;
				} else {
					u = 0;
				}
				t = 3

			case string:
				s = string(p.(string))
				t = 4

			default:
				t = 0
		}

		if t == 1 {
			data = append(data, Int64ToBytes(i))
		} else if t == 2 {
			data = append(data, Uint64ToBytes(u))
		} else if t == 3 {
			data = append(data, Uint64ToBytes(u))
		} else if t == 4 {
			data = append(data, []byte(s))
		} else {
			return nil
		}
	}

	res, err := rlp.EncodeToBytes(data)
	if err == nil {
		return res
	} else {
		return nil
	}
}

func CallResAsUint64(bts []byte) (uint64) {
	if len(bts) < 32 {
		return 0
	}

	var n uint64 = 0
	for _, b := range bts[:32] {
		n = n * 256 + uint64(b)
	}
	return n
}

func CallResAsInt64(bts []byte) (int64) {
	if len(bts) < 32 {
		return 0
	}

	var n int64 = 0
	for _, b := range bts[:32] {
		n = n * 256 + int64(b)
	}
	return n
}

func CallResAsBool(bts []byte) (bool) {
	if len(bts) < 32 {
		return false
	}

	if bts[31] == 1 {
		return true
	} else {
		return false
	}
}

func CallResAsString(bts []byte) (string) {
	if len(bts) < 64 {
		return ""
	}

	slen := int(bts[63])
	if slen > len(bts) - 64 {
		return ""
	}
	return string(bts[64:64+slen])
}

