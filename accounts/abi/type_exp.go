package abi

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
)

// NewType creates a new reflection type of abi type given in t.
func NewTypeV2(t string, internalType string, components []ArgumentMarshaling) (typ Type, err error) {
	// check that array brackets are equal if they exist
	if strings.Count(t, "[") != strings.Count(t, "]") {
		return Type{}, fmt.Errorf("invalid arg type in abi")
	}
	typ.stringKind = t

	// if there are brackets, get ready to go into slice/array mode and
	// recursively create the type
	if strings.Count(t, "[") != 0 {
		// Note internalType can be empty here.
		subInternal := internalType
		if i := strings.LastIndex(internalType, "["); i != -1 {
			subInternal = subInternal[:i]
		}
		// recursively embed the type
		i := strings.LastIndex(t, "[")
		embeddedType, err := NewTypeV2(t[:i], subInternal, components)
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
			typ.Elem = &embeddedType
			typ.stringKind = embeddedType.stringKind + sliced
		} else if len(intz) == 1 {
			// is an array
			typ.T = ArrayTy
			typ.Elem = &embeddedType
			typ.Size, err = strconv.Atoi(intz[0])
			if err != nil {
				return Type{}, fmt.Errorf("abi: error parsing variable size: %v", err)
			}
			typ.stringKind = embeddedType.stringKind + sliced
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
		typ.Size = varSize
		typ.T = IntTy
	case "uint":
		typ.Size = varSize
		typ.T = UintTy
	case "bool":
		typ.T = BoolTy
	case "address":
		typ.Size = 20
		typ.T = AddressTy
	case "string":
		typ.T = StringTy
	case "bytes":
		if varSize == 0 {
			typ.T = BytesTy
		} else {
			typ.T = FixedBytesTy
			typ.Size = varSize
		}
	case "tuple":
		var (
			fields     []reflect.StructField
			elems      []*Type
			names      []string
			expression string // canonical parameter expression
		)
		expression += "("
		overloadedNames := make(map[string]string)
		for idx, c := range components {
			cType, err := NewTypeV2(c.Type, c.InternalType, c.Components)
			if err != nil {
				return Type{}, err
			}
			fieldName, err := overloadedArgName(c.Name, overloadedNames)
			if err != nil {
				return Type{}, err
			}
			overloadedNames[fieldName] = fieldName
			fields = append(fields, reflect.StructField{
				Name: fieldName, // reflect.StructOf will panic for any exported field.
				Type: cType.GetType(),
				Tag:  reflect.StructTag("json:\"" + c.Name + "\""),
			})
			elems = append(elems, &cType)
			names = append(names, c.Name)
			expression += cType.stringKind
			if idx != len(components)-1 {
				expression += ","
			}
		}
		expression += ")"

		typ.TupleType = reflect.StructOf(fields)
		typ.TupleElems = elems
		typ.TupleRawNames = names
		typ.T = TupleTy
		typ.stringKind = expression

		const structPrefix = "struct "
		// After solidity 0.5.10, a new field of abi "internalType"
		// is introduced. From that we can obtain the struct name
		// user defined in the source code.
		if internalType != "" && strings.HasPrefix(internalType, structPrefix) {
			// Foo.Bar type definition is not allowed in golang,
			// convert the format to FooBar
			typ.TupleRawName = strings.Replace(internalType[len(structPrefix):], ".", "", -1)
		}

	case "function":
		typ.T = FunctionTy
		typ.Size = 24
	default:
		return Type{}, fmt.Errorf("unsupported arg type: %s", t)
	}

	return
}

// GetType returns the reflection type of the ABI type.
func (t Type) GetType() reflect.Type {
	switch t.T {
	case IntTy:
		return reflectIntType(false, t.Size)
	case UintTy:
		return reflectIntType(true, t.Size)
	case BoolTy:
		return reflect.TypeOf(false)
	case StringTy:
		return reflect.TypeOf("")
	case SliceTy:
		return reflect.SliceOf(t.Elem.GetType())
	case ArrayTy:
		return reflect.ArrayOf(t.Size, t.Elem.GetType())
	case TupleTy:
		return t.TupleType
	case AddressTy:
		return reflect.TypeOf(common.Address{})
	case FixedBytesTy:
		return reflect.ArrayOf(t.Size, reflect.TypeOf(byte(0)))
	case BytesTy:
		return reflect.SliceOf(reflect.TypeOf(byte(0)))
	case HashTy:
		// hashtype currently not used
		return reflect.ArrayOf(32, reflect.TypeOf(byte(0)))
	case FixedPointTy:
		// fixedpoint type currently not used
		return reflect.ArrayOf(32, reflect.TypeOf(byte(0)))
	case FunctionTy:
		return reflect.ArrayOf(24, reflect.TypeOf(byte(0)))
	default:
		panic("Invalid type")
	}
}

func overloadedArgName(rawName string, names map[string]string) (string, error) {
	fieldName := ToCamelCase(rawName)
	if fieldName == "" {
		return "", errors.New("abi: purely anonymous or underscored field is not supported")
	}
	// Handle overloaded fieldNames
	_, ok := names[fieldName]
	for idx := 0; ok; idx++ {
		fieldName = fmt.Sprintf("%s%d", ToCamelCase(rawName), idx)
		_, ok = names[fieldName]
	}
	return fieldName, nil
}

func (t Type) packV2(v reflect.Value) ([]byte, error) {
	// dereference pointer first if it's a pointer
	v = indirectV2(v)
	if err := typeCheckV2(t, v); err != nil {
		return nil, err
	}

	switch t.T {
	case SliceTy, ArrayTy:
		var ret []byte

		if t.requiresLengthPrefix() {
			// append length
			ret = append(ret, packNum(reflect.ValueOf(v.Len()))...)
		}

		// calculate offset if any
		offset := 0
		offsetReq := isDynamicType(*t.Elem)
		if offsetReq {
			offset = getTypeSize(*t.Elem) * v.Len()
		}
		var tail []byte
		for i := 0; i < v.Len(); i++ {
			val, err := t.Elem.packV2(v.Index(i))
			if err != nil {
				return nil, err
			}
			if !offsetReq {
				ret = append(ret, val...)
				continue
			}
			ret = append(ret, packNum(reflect.ValueOf(offset))...)
			offset += len(val)
			tail = append(tail, val...)
		}
		return append(ret, tail...), nil
	case TupleTy:
		// (T1,...,Tk) for k >= 0 and any types T1, …, Tk
		// enc(X) = head(X(1)) ... head(X(k)) tail(X(1)) ... tail(X(k))
		// where X = (X(1), ..., X(k)) and head and tail are defined for Ti being a static
		// type as
		//     head(X(i)) = enc(X(i)) and tail(X(i)) = "" (the empty string)
		// and as
		//     head(X(i)) = enc(len(head(X(1)) ... head(X(k)) tail(X(1)) ... tail(X(i-1))))
		//     tail(X(i)) = enc(X(i))
		// otherwise, i.e. if Ti is a dynamic type.
		fieldmap, err := mapArgNamesToStructFields(t.TupleRawNames, v)
		if err != nil {
			return nil, err
		}
		// Calculate prefix occupied size.
		offset := 0
		for _, elem := range t.TupleElems {
			offset += getTypeSize(*elem)
		}
		var ret, tail []byte
		for i, elem := range t.TupleElems {
			field := v.FieldByName(fieldmap[t.TupleRawNames[i]])
			if !field.IsValid() {
				return nil, fmt.Errorf("field %s for tuple not found in the given struct", t.TupleRawNames[i])
			}
			val, err := elem.packV2(field)
			if err != nil {
				return nil, err
			}
			if isDynamicType(*elem) {
				ret = append(ret, packNum(reflect.ValueOf(offset))...)
				tail = append(tail, val...)
				offset += len(val)
			} else {
				ret = append(ret, val...)
			}
		}
		return append(ret, tail...), nil

	default:
		return packElement(t, v), nil
	}
}

// isDynamicType returns true if the type is dynamic.
// The following types are called “dynamic”:
// * bytes
// * string
// * T[] for any T
// * T[k] for any dynamic T and any k >= 0
// * (T1,...,Tk) if Ti is dynamic for some 1 <= i <= k
func isDynamicType(t Type) bool {
	if t.T == TupleTy {
		for _, elem := range t.TupleElems {
			if isDynamicType(*elem) {
				return true
			}
		}
		return false
	}
	return t.T == StringTy || t.T == BytesTy || t.T == SliceTy || (t.T == ArrayTy && isDynamicType(*t.Elem))
}

// getTypeSize returns the size that this type needs to occupy.
// We distinguish static and dynamic types. Static types are encoded in-place
// and dynamic types are encoded at a separately allocated location after the
// current block.
// So for a static variable, the size returned represents the size that the
// variable actually occupies.
// For a dynamic variable, the returned size is fixed 32 bytes, which is used
// to store the location reference for actual value storage.
func getTypeSize(t Type) int {
	if t.T == ArrayTy && !isDynamicType(*t.Elem) {
		// Recursively calculate type size if it is a nested array
		if t.Elem.T == ArrayTy {
			return t.Size * getTypeSize(*t.Elem)
		} else if t.Elem.T == TupleTy {
			return t.Size * getTypeSize(*t.Elem)
		}
		return t.Size * 32
	} else if t.T == TupleTy && !isDynamicType(t) {
		total := 0
		for _, elem := range t.TupleElems {
			total += getTypeSize(*elem)
		}
		return total
	}
	return 32
}

// newly added
// integerParsing parses the string input to integer/big.Int type
func integerParsing(value string, size int, unsigned bool) (interface{}, error) {
	if size <= 64 {
		return SolInputStringTOInt(value, size, !unsigned)
	} else {
		var isValid bool
		var intValue = new(big.Int)

		if strings.HasPrefix(strings.ToLower(value), "0x") {
			intValue, isValid = big.NewInt(0).SetString(value[2:], 16)
		} else {
			intValue, isValid = big.NewInt(0).SetString(value, 10)
			isValid = isValid && common.IsSafeNumber(value, size, unsigned)
		}

		if !isValid {
			return nil, fmt.Errorf("paring big int string error")
		}
		return intValue, nil
	}
}

// newly added
// StringConvert converts the input string to the actual type defined in golang
func (t Type) StringConvert(value string) (interface{}, error) {
	// eliminate space
	value = TrimSpace(value)

	switch t.T {
	case IntTy:
		return integerParsing(value, t.Size, false)
	case UintTy:
		return integerParsing(value, t.Size, true)
	case BoolTy:
		if value == "false" {
			return false, nil
		} else if value == "true" {
			return true, nil
		} else {
			return false, errors.New("parse bool type error")
		}
	case StringTy:
		return value, nil
	case SliceTy:
		paramArray := GetFuncParamWrap(value)
		size := len(paramArray)
		v := reflect.MakeSlice(t.GetType(), size, size*2)

		for i, vTup := range paramArray {
			paramType := t.Elem
			argTup, err := paramType.StringConvert(vTup)
			if err != nil {
				return nil, err
			}

			v.Index(i).Set(reflect.ValueOf(argTup))
		}
		return v.Interface(), nil
	case ArrayTy:
		paramArray := GetFuncParamWrap(value)
		v := reflect.New(t.GetType())
		vSet := v.Elem()

		if t.Size != len(paramArray) {
			return nil, errors.New(fmt.Sprintf("array has %d args, received %d", t.Size, len(paramArray)))
		}

		for i, vTup := range paramArray {
			paramType := t.Elem
			argTup, err := paramType.StringConvert(vTup)
			if err != nil {
				return nil, err
			}

			vSet.Index(i).Set(reflect.ValueOf(argTup))
		}
		return vSet.Interface(), nil
	case TupleTy:
		tupleArray := GetFuncParamWrap(value)
		v := reflect.New(t.GetType())
		vSet := v.Elem()

		for i, vTup := range tupleArray {
			paramType := t.TupleElems[i]
			argTup, err := paramType.StringConvert(vTup)
			if err != nil {
				return nil, err
			}

			vSet.Field(i).Set(reflect.ValueOf(argTup))
		}
		return vSet.Interface(), nil
	case AddressTy:
		return common.HexToAddress(value), nil
	case FixedBytesTy:
		vBytes, err := hexutil.Decode(value)
		if err != nil {
			return nil, err
		}

		p := unsafe.Pointer(&vBytes[0])
		v := reflect.NewAt(t.GetType(), p)

		return v.Elem().Interface(), nil
		/// return setBytes(t, vBytes)
	case BytesTy:
		// todo: untested
		vBytes, err := hexutil.Decode(value)
		if err != nil {
			return nil, err
		}

		return vBytes, nil
	default:
		// see more details in GetType
		// hashtype currently not used
		// FixedPointTy currently not used
		panic("todo")
	}
}

// deprecated, realized by reflect.NewAt()
func setBytes(t Type, b []byte) (interface{}, error) {
	v := reflect.New(t.GetType())
	vSet := v.Elem()

	if t.Size < len(b) {
		return nil, errors.New(fmt.Sprintf("invalid input type: expected %s(len:%d), actual: %v", t.stringKind, t.Size, b))
	}

	switch t.Size {
	case 1:
		var a [1]byte
		copy(a[:], b)
		vSet.Set(reflect.ValueOf(a))
	case 3:
		var a [3]byte
		copy(a[:], b)
		vSet.Set(reflect.ValueOf(a))
	case 32:
		var a [32]byte
		copy(a[:], b)
		vSet.Set(reflect.ValueOf(a))
	default:
		panic("unsupported byte array")
	}

	return vSet.Interface(), nil
}
