// Copyright 2015 The go-ethereum Authors
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

package abi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

// Type enumerator
const (
	IntTy byte = iota
	UintTy
	BoolTy
	StringTy
	SliceTy
	ArrayTy
	// TupleTy
	AddressTy
	FixedBytesTy
	BytesTy
	HashTy
	FixedPointTy
	FunctionTy

	// newly Added [2020-08-06]
	TupleTy
)

// Type is the reflection of the supported argument type
type Type struct {
	Elem *Type

	Kind reflect.Kind
	Type reflect.Type
	Size int
	T    byte // Our own type checking

	stringKind string // holds the unparsed string for deriving signatures

	// newly Added [2020-08-06]
	// Tuple relative fields
	TupleRawName  string       // Raw struct name defined in source code, may be empty.
	TupleElems    []*Type      // Type information of all tuple fields
	TupleRawNames []string     // Raw field name of all tuple fields
	TupleType     reflect.Type // Underlying struct of the tuple
}

var (
	// typeRegex parses the abi sub types
	typeRegex = regexp.MustCompile("([a-zA-Z]+)(([0-9]+)(x([0-9]+))?)?")
)

// NewType creates a new reflection type of abi type given in t.
func NewType(t string) (typ Type, err error) {
	// check that array brackets are equal if they exist
	if strings.Count(t, "[") != strings.Count(t, "]") {
		return Type{}, fmt.Errorf("invalid arg type in abi")
	}

	typ.stringKind = t

	// if there are brackets, get ready to go into slice/array mode and
	// recursively create the type
	if strings.Count(t, "[") != 0 {
		i := strings.LastIndex(t, "[")
		// recursively embed the type
		embeddedType, err := NewType(t[:i])
		if err != nil {
			return Type{}, err
		}
		// grab the last cell and create a type from there
		sliced := t[i:]
		// grab the slice size with regexp
		re := regexp.MustCompile("[0-9]+")
		intz := re.FindAllString(sliced, -1)

		if len(intz) == 0 {
			// is a slice
			typ.T = SliceTy
			typ.Kind = reflect.Slice
			typ.Elem = &embeddedType
			typ.Type = reflect.SliceOf(embeddedType.Type)
		} else if len(intz) == 1 {
			// is a array
			typ.T = ArrayTy
			typ.Kind = reflect.Array
			typ.Elem = &embeddedType
			typ.Size, err = strconv.Atoi(intz[0])
			if err != nil {
				return Type{}, fmt.Errorf("abi: error parsing variable size: %v", err)
			}
			typ.Type = reflect.ArrayOf(typ.Size, embeddedType.Type)
		} else {
			return Type{}, fmt.Errorf("invalid formatting of array type")
		}
		return typ, err
	}
	// parse the type and size of the abi-type.
	matches := typeRegex.FindAllStringSubmatch(t, -1)
	if len(matches) == 0 {
		return Type{}, fmt.Errorf("invalid type '%v'", t)
	}
	parsedType := matches[0]

	// varSize is the size of the variable
	var varSize int
	if len(parsedType[3]) > 0 {
		var err error
		varSize, err = strconv.Atoi(parsedType[2])
		if err != nil {
			return Type{}, fmt.Errorf("abi: error parsing variable size: %v", err)
		}
	} else {
		if parsedType[0] == "uint" || parsedType[0] == "int" {
			// this should fail because it means that there's something wrong with
			// the abi type (the compiler should always format it to the size...always)
			return Type{}, fmt.Errorf("unsupported arg type: %s", t)
		}
	}
	// varType is the parsed abi type
	switch varType := parsedType[1]; varType {
	case "int":
		typ.Kind, typ.Type = reflectIntKindAndType(false, varSize)
		typ.Size = varSize
		typ.T = IntTy
	case "uint":
		typ.Kind, typ.Type = reflectIntKindAndType(true, varSize)
		typ.Size = varSize
		typ.T = UintTy
	case "bool":
		typ.Kind = reflect.Bool
		typ.T = BoolTy
		typ.Type = reflect.TypeOf(bool(false))
	case "address":
		typ.Kind = reflect.Array
		typ.Type = addressT
		typ.Size = 20
		typ.T = AddressTy
	case "string":
		typ.Kind = reflect.String
		typ.Type = reflect.TypeOf("")
		typ.T = StringTy
	case "bytes":
		if varSize == 0 {
			typ.T = BytesTy
			typ.Kind = reflect.Slice
			typ.Type = reflect.SliceOf(reflect.TypeOf(byte(0)))
		} else {
			typ.T = FixedBytesTy
			typ.Kind = reflect.Array
			typ.Size = varSize
			typ.Type = reflect.ArrayOf(varSize, reflect.TypeOf(byte(0)))
		}
	case "function":
		typ.Kind = reflect.Array
		typ.T = FunctionTy
		typ.Size = 24
		typ.Type = reflect.ArrayOf(24, reflect.TypeOf(byte(0)))
	default:
		return Type{}, fmt.Errorf("unsupported arg type: %s", t)
	}

	return
}

// String implements Stringer
func (t Type) String() (out string) {
	return t.stringKind
}

func (t Type) pack(v reflect.Value) ([]byte, error) {
	// dereference pointer first if it's a pointer
	v = indirect(v)

	if err := typeCheck(t, v); err != nil {
		return nil, err
	}

	if t.T == SliceTy || t.T == ArrayTy {
		var packed []byte

		for i := 0; i < v.Len(); i++ {
			val, err := t.Elem.pack(v.Index(i))
			if err != nil {
				return nil, err
			}
			packed = append(packed, val...)
		}
		if t.T == SliceTy {
			return packBytesSlice(packed, v.Len()), nil
		} else if t.T == ArrayTy {
			return packed, nil
		}
	}
	return packElement(t, v), nil
}

// requireLengthPrefix returns whether the type requires any sort of length
// prefixing.
func (t Type) requiresLengthPrefix() bool {
	return t.T == StringTy || t.T == BytesTy || t.T == SliceTy
}

const (
	ContractTypeWasm     = "wasm"
	ContractTypeSolidity = "sol"
)

type ContractType interface {
	GenerateInputData() ([]byte, error)
	NewContractTypeFromJson([]byte) error
}

func GenerateInputData(ct ContractType, input []byte) ([]byte, error) {
	if err := ct.NewContractTypeFromJson(input); err != nil {
		return input, err
	}
	return ct.GenerateInputData()
}

type WasmInput struct {
	TxType     int      `json:"-"`
	FuncName   string   `json:"func_name"`
	FuncParams []string `json:"func_params"`
}

func (c *WasmInput) NewContractTypeFromJson(input []byte) error {
	var err error
	if err = json.Unmarshal(input, c); err != nil {
		common.ErrPrintln("GenerateInputData sol json unmarshal error: ", err)
		return err
	}
	return err
}

// Generate the input data of the wasm contract
func (c *WasmInput) GenerateInputData() ([]byte, error) {

	if c.FuncName == "" {
		common.ErrPrintln("miss wasm func name")
		return nil, errors.New("miss wasm func name")
	}
	c.TxType = common.TxTypeCallSollCompatibleWasm

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

func (s *SolInput) NewContractTypeFromJson(input []byte) error {
	var err error
	if err = json.Unmarshal(input, s); err != nil {
		common.ErrPrintln("GenerateInputData sol json unmarshal error: ", err)
		return err
	}
	return err
}

// Generate input data for solidity contract
func (s *SolInput) GenerateInputData() ([]byte, error) {
	var arguments Arguments
	var args []interface{}
	var paramTypes []string

	for _, param := range s.FuncParams {
		paramType, paramValue, err := SpliceParam(param)
		if err != nil {
			common.ErrPrintln("sol SpliceParam error: ", err)
			return nil, err
		}
		// parsing arg type
		var argument Argument
		if argument.Type, err = NewType(paramType); err != nil {
			common.ErrPrintln("sol NewType error: ", err)
			return nil, err
		}
		arguments = append(arguments, argument)

		// parsing arg value
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
	// sig func
	inputBytes := crypto.Keccak256([]byte(Sig(s.FuncName, paramTypes)))[:4]
	// append params byte stream
	inputBytes = append(inputBytes, paramsBytes...)
	return inputBytes, nil
}

func SetInputLength(input []byte) (res []byte) {
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
	//case strings.HasPrefix(t, "bytes"):
	//	if len(v) < 3 {
	//		return nil, fmt.Errorf("input format error: %s", v)
	//	}
	//
	//	v = v[1 : len(v)-1]
	//	vs := strings.Split(v, ",")
	//	var res []byte
	//	for _, value := range vs {
	//		intV, err := strconv.Atoi(value)
	//		if err != nil || intV > 255 {
	//			return nil, fmt.Errorf("bytes input strconv a to i error || value > 255 : %s", value)
	//		}
	//		res = append(res, byte(intV))
	//	}
	//	return res, nil
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
			if parts[1] != "" && strings.Contains(v, "-") {
				return nil, fmt.Errorf("value does not match type: Unsigned type passes negative number")
			}
			value, ok := big.NewInt(0).SetString(v, 10)
			bit, _ := strconv.Atoi(parts[2])
			if !ok || !common.IsSafeNumber(v, bit, parts[1] != "") {
				return nil, fmt.Errorf("paring big int string error")
			}
			return value, nil
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

func ParseWasmCallSolInput(input []byte) ([]byte, error) {
	// Only used in compatibility mode
	ptr := new(interface{})
	err := rlp.Decode(bytes.NewReader(input), &ptr)
	if err != nil {
		return input, err
	}
	rlpList := reflect.ValueOf(ptr).Elem().Interface()
	if _, ok := rlpList.([]interface{}); !ok {
		return input, fmt.Errorf("error input rlp data type")
	}
	iRlpList := rlpList.([]interface{})
	if len(iRlpList) < 2 {
		return input, fmt.Errorf("error input rlp data count")
	}
	var (
		funcName string
		params   []string
	)
	v, ok := iRlpList[1].([]byte)
	if !ok {
		return input, fmt.Errorf("error input rlp data funcname")
	}
	funcName = string(v)

	for _, v := range iRlpList[2:] {
		vv, ok := v.([]byte)
		if !ok {
			return input, fmt.Errorf("error input rlp data funcparams")
		}
		params = append(params, string(vv))
	}
	solInput := SolInput{
		FuncName:   funcName,
		FuncParams: params,
	}
	newInput, err := solInput.GenerateInputData()
	if err != nil {
		return input, err
	}
	return newInput, nil
}
