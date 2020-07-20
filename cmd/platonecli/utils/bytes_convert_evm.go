package utils

import (
	"bytes"
	"errors"
	"math/big"
	"reflect"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/math"
)

// EncodeBytesType encodes the byte type by the rule defined in solidity doc: contract ABI specification
// dynamic bytes: "bytes"
// fixed bytes: "bytes<M>", where 0 < M <= 32
func EncodeBytesType(source, t string) ([]byte, error) {
	var index = len("bytes")
	var sourceBytes = []byte(source)
	var isFixedBytes = IsInRange(t[index:], 32) || len(sourceBytes) <= 32

	switch {
	case len(t) == index: // dynamic bytes array
		k := len(sourceBytes)
		div := k / 32

		temp := math.PaddedBigBytes(new(big.Int).SetUint64(uint64(k)), 32)
		temp2 := common.RightPadBytes(sourceBytes[32*div:], 32)
		temp2 = append(sourceBytes[:32*div], temp2...)

		return append(temp, temp2...), nil
	case isFixedBytes: // fixed bytes array
		return common.RightPadBytes(sourceBytes, 32), nil
	default:
		return nil, errors.New("invalid bytes type")
	}
}

// EncodeBoolType encodes the bool type by the rule defined in solidity doc: contract ABI specification
func EncodeBoolType(source string) ([]byte, error) {
	switch source {
	case "true":
		return math.PaddedBigBytes(common.Big1, 32), nil
	case "false":
		return math.PaddedBigBytes(common.Big0, 32), nil
	default:
		return nil, errors.New("invalid bool syntax")
	}
}

// EncodeAddressType encodes the address type by the rule defined in solidity doc: contract ABI specification
func EncodeAddressType(source string) ([]byte, error) {
	dest := common.HexToAddress(source)

	reflectValue := reflect.ValueOf(dest)
	if reflectValue.Kind() == reflect.Array { // 这一步验证的必要性？
		reflectValue = mustArrayToByteSlice(reflectValue)
	}

	return common.LeftPadBytes(reflectValue.Bytes(), 32), nil
}

// EncodeInt encodes the int type by the rule defined in solidity doc: contract ABI specification
func EncodeInt(source string) ([]byte, error) {

	// 是否需要检查数据溢出？

	if !IsMatch(source, "num") {
		return nil, errors.New("invalid integer syntax")
	}

	n, ok := new(big.Int).SetString(source, 10)
	if !ok {
		return nil, errors.New("convert failed")
	}
	return math.PaddedBigBytes(math.U256(n), 32), nil
}

// EncodeOffset encodes the uint256 type by the rule defined in solidity doc: contract ABI specification
// it is used to calculate the offset of the arguments
func EncodeOffset(offset int) []byte {
	n := new(big.Int).SetInt64(int64((offset)))
	return math.PaddedBigBytes(math.U256(n), 32)
}

// TODO
func IsValidEvmIntType(t string) bool {
	return true
}

// RuneToBytesArray
func RuneToBytesArray(r []rune) []byte {
	var bytesArray []byte

	for _, value := range r {
		tempBytes := common.Int32ToBytes(value)
		tempBytes = bytes.TrimLeft(tempBytes, "\x00")
		bytesArray = append(bytesArray, tempBytes...)
	}

	return bytesArray
}

// mustArrayToBytesSlice creates a new byte slice with the exact same size as value
// and copies the bytes in value to the new slice.
func mustArrayToByteSlice(value reflect.Value) reflect.Value {
	slice := reflect.MakeSlice(reflect.TypeOf([]byte{}), value.Len(), value.Len())
	reflect.Copy(slice, value)
	return slice
}
