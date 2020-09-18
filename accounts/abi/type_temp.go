package abi

import (
	"errors"
	"math/big"
	"reflect"
	"strconv"

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
		// source = bytes.TrimRight(source, "\x00")
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

// related to the cpp.abi.json file
func StringToArg(source string, t string) (interface{}, error) {
	switch t {
	case "int32", "uint32", "uint", "int":
		return strconv.Atoi(source)
	case "int64", "uint64":
		return strconv.ParseInt(source, 10, 64)
	case "float32":
		return strconv.ParseFloat(source, 32)
	case "float64":
		return strconv.ParseFloat(source, 64)
	case "bool":
		switch source {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, errors.New("invalid boolean param")
		}
	default:
		return source, nil
	}
}

// only used for cpp contract
func WasmArgToBytes(arg interface{}) []byte {
	v := reflect.ValueOf(arg)

	switch v.Kind() {
	case reflect.Int:
		return common.Int32ToBytes(int32(arg.(int)))
	case reflect.Int32:
		return common.Int32ToBytes(arg.(int32))
	case reflect.Int64:
		return common.Int64ToBytes(arg.(int64))
	case reflect.Uint:
		return common.Uint32ToBytes(uint32(arg.(uint)))
	case reflect.Uint32:
		return common.Uint32ToBytes(arg.(uint32))
	case reflect.Uint64:
		return common.Uint64ToBytes(arg.(uint64))
	case reflect.Float32:
		return common.Float32ToBytes(arg.(float32))
	case reflect.Float64:
		return common.Float64ToBytes(arg.(float64))
	case reflect.Bool:
		return common.BoolToBytes(arg.(bool))
	case reflect.String:
		return []byte(arg.(string))
	default:
		panic("unsupported type")
	}
}

func StringConverterV2(source string, t string) ([]byte, error) {
	res, err := StringToArg(source, t)
	return WasmArgToBytes(res), err
}

// packNum packs the given number (using the reflect value) and will cast it to appropriate number representation
func packWasmNum(value reflect.Value) []byte {
	switch kind := value.Kind(); kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return U256(new(big.Int).SetUint64(value.Uint()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return U256(big.NewInt(value.Int()))
	case reflect.Ptr:
		return U256(value.Interface().(*big.Int))
	default:
		panic("abi: fatal error")
	}

}

/*
func (t Type) WasmPack(v interface{}) ([]byte, error) {
	reflectValue := reflect.ValueOf(v)

	switch t.T {
	case IntTy, UintTy:
		return packWasmNum(reflectValue)
	case StringTy:
		return packBytesSlice([]byte(reflectValue.String()), reflectValue.Len())
	case BoolTy:
		if reflectValue.Bool() {
			return math.PaddedBigBytes(common.Big1, 32)
		}
		return math.PaddedBigBytes(common.Big0, 32)
	case BytesTy:
		if reflectValue.Kind() == reflect.Array {
			reflectValue = mustArrayToByteSlice(reflectValue)
		}
		return packBytesSlice(reflectValue.Bytes(), reflectValue.Len())
	default:
		panic("abi: fatal error")
	}
}

func (t Type) WasmUnpack(output []byte) (interface{}, error) {
	switch t.T {
	case SliceTy:
		return forEachUnpackV2(t, output[begin:], 0, length)
	case ArrayTy:
		if isDynamicType(*t.Elem) {
			offset := int64(binary.BigEndian.Uint64(returnOutput[len(returnOutput)-8:]))
			return forEachUnpackV2(t, output[offset:], 0, t.Size)
		}
		return forEachUnpackV2(t, output[index:], 0, t.Size)
	case StringTy: // variable arrays are written at the end of the return bytes
		return string(output[begin : begin+length]), nil
	case IntTy, UintTy:
		return ReadInteger(t, returnOutput), nil
	case BoolTy:
		return readBool(returnOutput)
	case BytesTy:
		return output[begin : begin+length], nil
	default:
		return nil, fmt.Errorf("abi: unknown type %v", t.T)
	}
}*/
