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

package vm

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/math"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// calculates the memory size required for a step
func calcMemSize(off, l *big.Int) *big.Int {
	if l.Sign() == 0 {
		return common.Big0
	}

	return new(big.Int).Add(off, l)
}

// getData returns a slice from the data based on the start and size and pads
// up to size with zero's. This function is overflow safe.
func getData(data []byte, start uint64, size uint64) []byte {
	length := uint64(len(data))
	if start > length {
		start = length
	}
	end := start + size
	if end > length {
		end = length
	}
	return common.RightPadBytes(data[start:end], int(size))
}

// getDataBig returns a slice from the data based on the start and size and pads
// up to size with zero's. This function is overflow safe.
func getDataBig(data []byte, start *big.Int, size *big.Int) []byte {
	dlen := big.NewInt(int64(len(data)))

	s := math.BigMin(start, dlen)
	e := math.BigMin(new(big.Int).Add(s, size), dlen)
	return common.RightPadBytes(data[s.Uint64():e.Uint64()], int(size.Uint64()))
}

// bigUint64 returns the integer casted to a uint64 and returns whether it
// overflowed in the process.
func bigUint64(v *big.Int) (uint64, bool) {
	return v.Uint64(), v.BitLen() > 64
}

// toWordSize returns the ceiled word size required for memory expansion.
func toWordSize(size uint64) uint64 {
	if size > math.MaxUint64-31 {
		return math.MaxUint64/32 + 1
	}

	return (size + 31) / 32
}

func allZero(b []byte) bool {
	for _, byte := range b {
		if byte != 0 {
			return false
		}
	}
	return true
}

const (
	InvokeContract       = 2
	ContractTypeWasm     = "wasm"
	ContractTypeSolidity = "sol"
)

type ContractType interface {
	GenerateInputData([]byte) ([]byte, error)
}

func GenerateInputData(ct ContractType, input []byte) ([]byte, error) {
	return ct.GenerateInputData(input)
}

type WasmInput struct {
	TxType     int      `json:"tx_type"`
	FuncName   string   `json:"func_name"`
	FuncParams []string `json:"func_params"`
}

// Generate the input data of the wasm contract
func (c *WasmInput) GenerateInputData(input []byte) ([]byte, error) {
	if err := json.Unmarshal(input, c); err != nil {
		common.ErrPrintln("Unmarshal wasm input error: ", err)
		return nil, err
	}

	if c.FuncName == "" {
		common.ErrPrintln("miss wasm func name")
		return nil, errors.New("miss wasm func name")
	}

	if c.TxType == 0 {
		c.TxType = InvokeContract
	}
	paramArr := [][]byte{
		common.Int64ToBytes(int64(c.TxType)),
		[]byte(c.FuncName),
	}

	for _, param := range c.FuncParams {
		paramType, paramValue, err := SpliceParam(param)
		if err != nil {
			common.ErrPrintln("SpliceParam wasm param err: ", err)
			return nil, err
		}
		p, err := StringConverter(paramValue, paramType)
		if err != nil {
			common.ErrPrintln("StringConverter wasm param err: ", err)
			return nil, err
		}
		paramArr = append(paramArr, p)
	}

	paramBytes, e := rlp.EncodeToBytes(paramArr)
	if e != nil {
		common.ErrPrintln("rlp.EncodeToBytes wasm param err: ", e)
		return nil, fmt.Errorf("rpl encode error,%s", e.Error())
	}
	return paramBytes, nil
}

func SpliceParam(param string) (paramType string, paramValue string, err error) {
	var (
		errMissParamValueType = errors.New("func param miss param value type")
		errParamFormat        = errors.New("func param format error")
	)

	if param == "" {
		err = errMissParamValueType
		return
	}
	reg := regexp.MustCompile(`(.*)\((.*)\)`)
	if p := reg.FindStringSubmatch(param); len(p) == 3 {
		paramType = p[1]
		paramValue = p[2]
		return
	}
	err = errParamFormat
	return
}

func StringConverter(source string, t string) ([]byte, error) {
	switch t {
	case "int32", "uint32", "uint", "int":
		dest, err := strconv.Atoi(source)
		return common.Int32ToBytes(int32(dest)), err
	case "int64", "uint64":
		dest, err := strconv.ParseInt(source, 10, 64)
		return common.Int64ToBytes(dest), err
	case "float32":
		dest, err := strconv.ParseFloat(source, 32)
		return common.Float32ToBytes(float32(dest)), err
	case "float64":
		dest, err := strconv.ParseFloat(source, 64)
		return common.Float64ToBytes(dest), err
	case "bool":
		if "true" == source || "false" == source {
			return common.BoolToBytes("true" == source), nil
		} else {
			return []byte{}, errors.New("invalid boolean param")
		}
	default:
		return []byte(source), nil
	}
}

type SolInput struct {
	FuncName   string   `json:"func_name"`
	FuncParams []string `json:"func_params"`
}

// Generate input data for solidity contract
func (s *SolInput) GenerateInputData(input []byte) ([]byte, error) {
	if err := json.Unmarshal(input, s); err != nil {
		common.ErrPrintln("GenerateInputData sol json unmarshal error: ", err)
		return nil, err
	}
	var arguments abi.Arguments
	var args []interface{}
	var paramTypes []string

	for _, param := range s.FuncParams {
		paramType, paramValue, err := SpliceParam(param)
		if err != nil {
			common.ErrPrintln("sol SpliceParam error: ", err)
			return nil, err
		}
		var argument abi.Argument
		if argument.Type, err = abi.NewType(paramType); err != nil {
			common.ErrPrintln("sol NewType error: ", err)
			return nil, err
		}
		arguments = append(arguments, argument)

		arg, err := SolInputTypeConversion(paramType, paramValue)
		if err != nil {
			common.ErrPrintln("sol SolInputTypeConversion error: ", err)
			return nil, err
		}
		args = append(args, arg)
		paramTypes = append(paramTypes, paramType)
	}
	paramsBytes, err := arguments.Pack(args...)
	if err != nil {
		common.ErrPrintln("pack args error: ", err)
		return nil, err
	}
	inputBytes := crypto.Keccak256([]byte(Sig(s.FuncName, paramTypes)))[:4]
	inputBytes = append(inputBytes, paramsBytes...)
	return s.SetInputLength(inputBytes), nil
}

func (s *SolInput) SetInputLength(input []byte) (res []byte) {
	length := len(input)
	if length == 0 {
		return
	}
	lengthStr := strconv.Itoa(length)
	res = append(res, byte(len(lengthStr)))
	for _, value := range strings.Split(lengthStr, "") {
		v, _ := strconv.Atoi(value)
		res = append(res, byte(v))
	}
	res = append(res, input...)
	return
}

func Sig(funcName string, types []string) string {
	ts := make([]string, len(types))
	for i, t := range types {
		ts[i] = t
	}
	return fmt.Sprintf("%v(%v)", funcName, strings.Join(ts, ","))
}

func SolInputTypeConversion(t string, v string) (interface{}, error) {
	switch {
	case strings.HasPrefix(t, "address"):
		return common.HexToAddress(v), nil
	case strings.HasPrefix(t, "bytes"):
		if len(v) < 3 {
			return nil, fmt.Errorf("input format error: %s", v)
		}

		v = v[1 : len(v)-1]
		vs := strings.Split(v, ",")
		var res []byte
		for _, value := range vs {
			intV, err := strconv.Atoi(value)
			if err != nil || intV > 255 {
				return nil, fmt.Errorf("bytes input strconv a to i error || value > 255 : %s", value)
			}
			res = append(res, byte(intV))
		}
		return res, nil
		//todo: 反生成数组形式
		//parts := regexp.MustCompile(`bytes([0-9]*)`).FindStringSubmatch(t)
		//if parts[1] != "" {
		//	length, err := strconv.Atoi(parts[1])
		//	if err != nil {
		//		return nil, err
		//	}
		//
		//	if len(v) < 3 {
		//		return nil, fmt.Errorf("input format error: %s", v)
		//	}
		//
		//	v = v[1 : len(v)-1]
		//	vs := strings.Split(v, ",")
		//	if len(vs) != length {
		//		return nil, fmt.Errorf("input format error: %s", v)
		//	}
		//}

	case strings.HasPrefix(t, "int") || strings.HasPrefix(t, "uint"):
		parts := regexp.MustCompile(`(u)?int([0-9]*)`).FindStringSubmatch(t)
		switch parts[2] {
		case "8":
			return SolInputStringTOInt(v, 8, parts[1] == "")
		case "16":
			return SolInputStringTOInt(v, 16, parts[1] == "")
		case "32":
			return SolInputStringTOInt(v, 32, parts[1] == "")
		case "64":
			return SolInputStringTOInt(v, 64, parts[1] == "")
		case "128", "256":
			if parts[1] == "" {
				value, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return nil, err
				}
				return big.NewInt(0).SetInt64(value), nil
			}
			value, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return nil, err
			}
			return big.NewInt(0).SetUint64(value), nil
		}
		return nil, errors.New("parse input type int has err bitsize")
	case strings.HasPrefix(t, "bool"):
		if v == "false" {
			return false, nil
		} else if v == "true" {
			return true, nil
		} else {
			return false, errors.New("parse bool type error")
		}
	case strings.HasPrefix(t, "string"):
		return v, nil
	default:
		return nil, errors.New("sol input type error")
	}
}

func SolInputStringTOInt(v string, bitSize int, hasNotPrefixU bool) (interface{}, error) {
	if hasNotPrefixU {
		res, err := strconv.ParseInt(v, 10, bitSize)
		if err != nil {
			return nil, err
		}
		switch strconv.Itoa(bitSize) {
		case "8":
			return int8(res), nil
		case "16":
			return int16(res), nil
		case "32":
			return int32(res), nil
		case "64":
			return int64(res), nil
		default:
			return nil, fmt.Errorf("SolInputStringTOInt parsing int error. res: %d", res)
		}
	}
	res, err := strconv.ParseUint(v, 10, bitSize)
	if err != nil {
		return nil, err
	}
	switch strconv.Itoa(bitSize) {
	case "8":
		return uint8(res), nil
	case "16":
		return uint16(res), nil
	case "32":
		return uint32(res), nil
	case "64":
		return uint64(res), nil
	default:
		return nil, fmt.Errorf("SolInputStringTOInt parsing uint error. res: %d", res)
	}
}
