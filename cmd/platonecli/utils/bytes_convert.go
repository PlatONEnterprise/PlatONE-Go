package utils

import (
	"bytes"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
)

// BytesConverter converts the bytes to the specific data type
// it is the same as the BytesConverter in ctool/core/bytes_util.go
func BytesConverter(source []byte, t string) interface{} {
	switch t {
	case "int32":
		return common.CallResAsInt32(source)
	case "int64":
		return common.CallResAsInt64(source)
	case "int128":
		return common.CallResAsInt128(source)
	case "float32":
		return common.CallResAsFloat32(source)
	case "float64":
		return common.CallResAsFloat64(source)
	case "float128":
		return common.CallResAsFloat128(source)
	case "string", "int128_s", "uint128_s", "int256_s", "uint256_s":
		source = bytes.TrimRight(source, "\x00")

		if len(source) < 64 {
			return string(source[:])
		} else {
			return string(source[64:])
		}
	case "uint128":
		return common.CallResAsUint128(source)
	case "uint64":
		return common.CallResAsUint64(source)
	case "uint32":
		return common.CallResAsUint32(source)
	default:
		return source
	}
}
