// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package common contains various helper functions.
package common

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"math"
	"math/big"
	"strings"
)

// ToHex returns the hex representation of b, prefixed with '0x'.
// For empty slices, the return value is "0x0".
//
// Deprecated: use hexutil.Encode instead.
func ToHex(b []byte) string {
	hex := Bytes2Hex(b)
	if len(hex) == 0 {
		hex = "0"
	}
	return "0x" + hex
}

// FromHex returns the bytes represented by the hexadecimal string s.
// s may be prefixed with "0x".
func FromHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

// CopyBytes returns an exact copy of the provided bytes.
func CopyBytes(b []byte) (copiedBytes []byte) {
	if b == nil {
		return nil
	}
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)

	return
}

// hasHexPrefix validates str begins with '0x' or '0X'.
func hasHexPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// isHexCharacter returns bool of c being a valid hexadecimal.
func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

// isHex validates whether each byte is valid hexadecimal string.
func isHex(str string) bool {
	if len(str)%2 != 0 {
		return false
	}
	for _, c := range []byte(str) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}

// Bytes2Hex returns the hexadecimal encoding of d.
func Bytes2Hex(d []byte) string {
	return hex.EncodeToString(d)
}

// Hex2Bytes returns the bytes represented by the hexadecimal string str.
func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}

// Hex2BytesFixed returns bytes of a specified fixed length flen.
func Hex2BytesFixed(str string, flen int) []byte {
	h, _ := hex.DecodeString(str)
	if len(h) == flen {
		return h
	}
	if len(h) > flen {
		return h[len(h)-flen:]
	}
	hh := make([]byte, flen)
	copy(hh[flen-len(h):flen], h)
	return hh
}

// RightPadBytes zero-pads slice to the right up to length l.
func RightPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded, slice)

	return padded
}

// LeftPadBytes zero-pads slice to the left up to length l.
func LeftPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}

func BytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

func IntToBytes(n int) []byte {
	tmp := int(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func Int8ToBytes(n int8) []byte {
	tmp := int8(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func Int16ToBytes(n int16) []byte {
	tmp := int16(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func Int32ToBytes(n int32) []byte {
	tmp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func Int64ToBytes(n int64) []byte {
	tmp := int64(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func UintToBytes(n uint) []byte {
	tmp := uint(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func Uint8ToBytes(n uint8) []byte {
	tmp := uint8(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func Uint16ToBytes(n uint16) []byte {
	tmp := uint16(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func Uint32ToBytes(n uint32) []byte {
	tmp := uint32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func Uint64ToBytes(n uint64) []byte {
	tmp := uint64(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func BytesToInt32(b []byte) int32 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int32(tmp)
}

func BytesToInt64(b []byte) int64 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int64
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int64(tmp)
}

func Float32ToBytes(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, bits)
	return bytes
}

func BytesToFloat32(bytes []byte) float32 {
	bits := binary.BigEndian.Uint32(bytes)
	return math.Float32frombits(bits)
}

func Float64ToBytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, bits)
	return bytes
}

func BytesToFloat64(bytes []byte) float64 {
	bits := binary.BigEndian.Uint64(bytes)
	return math.Float64frombits(bits)
}

func PaddingLeft(src []byte, bytes int) []byte {
	if len(src) >= bytes {
		return src
	}
	dst := make([]byte, bytes)
	copy(dst, src)
	return reverse(dst)
}

func reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func BoolToBytes(b bool) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, b)
	return buf.Bytes()
}

func ToBytes(source interface{}) ([]byte, error) {
	switch dest := source.(type) {
	case string:
		return []byte(dest), nil
	case int32:
		return Int32ToBytes(dest), nil
	case uint32:
		return Int32ToBytes(int32(dest)), nil
	case uint:
		return Int32ToBytes(int32(dest)), nil
	case int:
		return Int32ToBytes(int32(dest)), nil
	case uint64:
		return Int64ToBytes(int64(dest)), nil
	case int64:
		return Int64ToBytes(dest), nil
	case float32:
		return Float32ToBytes(dest), nil
	case float64:
		return Float64ToBytes(dest), nil
	case bool:
		return BoolToBytes(dest), nil
	case *big.Int:
		return dest.Bytes(), nil
	}

	return nil, errors.New(fmt.Sprintf("ToBytes function not support %v", source))
}

func WasmCallResultCompatibleSolString(res []byte) []byte {
	// The first 32 bits are offsets, and the next 32 bits are string lengths
	rl := len(res)
	// Sol contract string type return value, longer than 64, and the first byte value is 0
	if rl >= 64 && res[0] == byte(0) && strings.EqualFold(GetCurrentInterpreterType(), "all") {
		// Get string length
		lengthBytes := bytes.TrimLeft(res[32:64], "\x00")
		length := 0
		for i := 0; i < len(lengthBytes); i++ {
			length += int(lengthBytes[i]) * int(math.Pow(256, float64(len(lengthBytes)-1-i)))
		}
		if rl < 64+length {
			return res
		}
		// Intercept string
		res = res[64 : 64+length]
		res = append(res, byte(0))
	}
	return res
}

func WasmCallResultCompatibleSolInt64(res []byte) []byte {
	if !strings.EqualFold(GetCurrentInterpreterType(), "all") || len(res) != 32 {
		return res
	}
	// The first 24 bytes are all 255 for negative numbers, all 0 for positive numbers
	if !bytes.Equal(bytes.Repeat([]byte{255}, 24), res[:24]) && !bytes.Equal(res[:24], make([]byte, 24)) {
		return nil
	}
	return res[24:]
}

func IsSafeNumber(number string, bit int, isUnsigned bool) (res bool) {
	if bit%8 != 0 {
		return
	}
	count := bit / 8
	var max, min *big.Int
	if isUnsigned {
		max = big.NewInt(0).SetBytes(bytes.Repeat([]byte{255}, count))
		min = big.NewInt(0)
	} else {
		max = big.NewInt(0).SetBytes(bytes.Repeat([]byte{255}, count))
		max = max.Div(big.NewInt(0).Sub(max, big.NewInt(1)), big.NewInt(2))
		min = big.NewInt(0).Neg(big.NewInt(0).Add(max, big.NewInt(1)))
	}
	fmt.Println(max, min)
	src, ok := big.NewInt(0).SetString(number, 10)
	if !ok {
		return
	}
	return src.Cmp(min) >= 0 && src.Cmp(max) <= 0
}

func IsBytesEmpty(input []byte) bool {
	if len(input) == 0 {
		return true
	} else {
		return false
	}
}

func GenerateWasmData(txType int64, funcName string, params []interface{}) ([]byte, error){
	paramArr := [][]byte{
		Int64ToBytes(txType),
		[]byte(funcName),
	}

	for _, v := range params {
		p, e := ToBytes(v)
		if e != nil {
			err := fmt.Errorf("convert %v to string failed", v)
			return nil,  err
		}
		paramArr = append(paramArr, p)
	}

	paramBytes, err := rlp.EncodeToBytes(paramArr)
	if err != nil{
		return nil, err
	}

	return paramBytes, nil
}

