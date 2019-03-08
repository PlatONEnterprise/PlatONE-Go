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
	data = append(data, Int64ToBytes(2)) // tx type, 2 for normal
	data = append(data, []byte(funcName))
	
	for _, p := range params {
		switch p.(type) {
			// for intX
			case int:
				data = append(data, Int64ToBytes(int64(p.(int))))
			case int8:
				data = append(data, Int64ToBytes(int64(p.(int8))))
			case int16:
				data = append(data, Int64ToBytes(int64(p.(int16))))
			case int32:
				data = append(data, Int64ToBytes(int64(p.(int32))))
			case int64:
				data = append(data, Int64ToBytes(p.(int64)))

			// for uintX
			case uint:
				data = append(data, Uint64ToBytes(uint64(p.(uint))))
			case uint8:
				data = append(data, Uint64ToBytes(uint64(p.(uint8))))
			case uint16:
				data = append(data, Uint64ToBytes(uint64(p.(uint16))))
			case uint32:
				data = append(data, Uint64ToBytes(uint64(p.(uint32))))
			case uint64:
				data = append(data, Uint64ToBytes(p.(uint64)))

			// for bool
			case bool:
				if p.(bool) {
					data = append(data, Uint64ToBytes(uint64(1)))
				} else {
					data = append(data, Uint64ToBytes(uint64(0)))
				}

			// for stirng
			case string:
				data = append(data, []byte(p.(string)))

			// not support
			default:
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

	slen := int(bts[61])*256*256 + int(bts[62])*256 + int(bts[63])
	if slen > len(bts) - 64 {
		return ""
	}
	return string(bts[64:64+slen])
}

