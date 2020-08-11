package abi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
)

type ArgumentMarshaling struct {
	Name         string
	Type         string
	InternalType string
	Components   []ArgumentMarshaling
	Indexed      bool
}

// Pack performs the operation Go format -> Hexdata
func (arguments Arguments) PackV2(args ...interface{}) ([]byte, error) {
	// Make sure arguments match up and pack them
	abiArgs := arguments
	if len(args) != len(abiArgs) {
		return nil, fmt.Errorf("argument count mismatch: got %d for %d", len(args), len(abiArgs))
	}
	// variable input is the output appended at the end of packed
	// output. This is used for strings and bytes types input.
	var variableInput []byte

	// input offset is the bytes offset for packed output
	inputOffset := 0
	for _, abiArg := range abiArgs {
		inputOffset += getTypeSize(abiArg.Type)
	}
	var ret []byte
	for i, a := range args {
		input := abiArgs[i]
		// pack the input
		packed, err := input.Type.packV2(reflect.ValueOf(a))
		if err != nil {
			return nil, err
		}
		// check for dynamic types
		if isDynamicType(input.Type) {
			// set the offset
			ret = append(ret, packNum(reflect.ValueOf(inputOffset))...)
			// calculate next offset
			inputOffset += len(packed)
			// append to variable input
			variableInput = append(variableInput, packed...)
		} else {
			// append the packed value to the input
			ret = append(ret, packed...)
		}
	}
	// append the variable input at the end of the packed input
	ret = append(ret, variableInput...)

	return ret, nil
}

// ToCamelCase converts an under-score string to a camel-case string
func ToCamelCase(input string) string {
	parts := strings.Split(input, "_")
	for i, s := range parts {
		if len(s) > 0 {
			parts[i] = strings.ToUpper(s[:1]) + s[1:]
		}
	}
	return strings.Join(parts, "")
}

//============================ Unpack Value ======================================

// Unpack performs the operation hexdata -> Go format
func (arguments Arguments) UnpackV2(v interface{}, data []byte) error {
	if len(data) == 0 {
		if len(arguments) != 0 {
			return fmt.Errorf("abi: attempting to unmarshall an empty string while arguments are expected")
		}
		return nil // Nothing to unmarshal, return
	}
	// make sure the passed value is arguments pointer
	if reflect.Ptr != reflect.ValueOf(v).Kind() {
		return fmt.Errorf("abi: Unpack(non-pointer %T)", v)
	}
	marshalledValues, err := arguments.UnpackValuesV2(data)
	if err != nil {
		return err
	}
	if len(marshalledValues) == 0 {
		return fmt.Errorf("abi: Unpack(no-values unmarshalled %T)", v)
	}
	if arguments.isTuple() {
		return arguments.unpackTupleV2(v, marshalledValues)
	}
	return arguments.unpackAtomicV2(v, marshalledValues[0])
}

/*
// UnpackIntoMap performs the operation hexdata -> mapping of argument name to argument value
func (arguments Arguments) UnpackIntoMap(v map[string]interface{}, data []byte) error {
	// Make sure map is not nil
	if v == nil {
		return fmt.Errorf("abi: cannot unpack into a nil map")
	}
	if len(data) == 0 {
		if len(arguments) != 0 {
			return fmt.Errorf("abi: attempting to unmarshall an empty string while arguments are expected")
		}
		return nil // Nothing to unmarshal, return
	}
	marshalledValues, err := arguments.UnpackValues(data)
	if err != nil {
		return err
	}
	for i, arg := range arguments.NonIndexed() {
		v[arg.Name] = marshalledValues[i]
	}
	return nil
}*/

// unpackAtomic unpacks ( hexdata -> go ) a single value
func (arguments Arguments) unpackAtomicV2(v interface{}, marshalledValues interface{}) error {
	dst := reflect.ValueOf(v).Elem()
	src := reflect.ValueOf(marshalledValues)

	if dst.Kind() == reflect.Struct && src.Kind() != reflect.Struct {
		return setV2(dst.Field(0), src)
	}
	return setV2(dst, src)
}

// unpackTuple unpacks ( hexdata -> go ) a batch of values.
func (arguments Arguments) unpackTupleV2(v interface{}, marshalledValues []interface{}) error {
	value := reflect.ValueOf(v).Elem()
	nonIndexedArgs := arguments.NonIndexed()

	switch value.Kind() {
	case reflect.Struct:
		argNames := make([]string, len(nonIndexedArgs))
		for i, arg := range nonIndexedArgs {
			argNames[i] = arg.Name
		}
		var err error
		abi2struct, err := mapArgNamesToStructFields(argNames, value)
		if err != nil {
			return err
		}
		for i, arg := range nonIndexedArgs {
			field := value.FieldByName(abi2struct[arg.Name])
			if !field.IsValid() {
				return fmt.Errorf("abi: field %s can't be found in the given value", arg.Name)
			}
			if err := setV2(field, reflect.ValueOf(marshalledValues[i])); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		if value.Len() < len(marshalledValues) {
			return fmt.Errorf("abi: insufficient number of arguments for unpack, want %d, got %d", len(arguments), value.Len())
		}
		for i := range nonIndexedArgs {
			if err := setV2(value.Index(i), reflect.ValueOf(marshalledValues[i])); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("abi:[2] cannot unmarshal tuple in to %v", value.Type())
	}
	return nil
}

// UnpackValues can be used to unpack ABI-encoded hexdata according to the ABI-specification,
// without supplying a struct to unpack into. Instead, this method returns a list containing the
// values. An atomic argument will be a list with one element.
func (arguments Arguments) UnpackValuesV2(data []byte) ([]interface{}, error) {
	nonIndexedArgs := arguments.NonIndexed()
	retval := make([]interface{}, 0, len(nonIndexedArgs))
	virtualArgs := 0
	for index, arg := range nonIndexedArgs {
		marshalledValue, err := toGoTypeV2((index+virtualArgs)*32, arg.Type, data)
		if arg.Type.T == ArrayTy && !isDynamicType(arg.Type) {
			// If we have a static array, like [3]uint256, these are coded as
			// just like uint256,uint256,uint256.
			// This means that we need to add two 'virtual' arguments when
			// we count the index from now on.
			//
			// Array values nested multiple levels deep are also encoded inline:
			// [2][3]uint256: uint256,uint256,uint256,uint256,uint256,uint256
			//
			// Calculate the full array size to get the correct offset for the next argument.
			// Decrement it by 1, as the normal index increment is still applied.
			virtualArgs += getTypeSize(arg.Type)/32 - 1
		} else if arg.Type.T == TupleTy && !isDynamicType(arg.Type) {
			// If we have a static tuple, like (uint256, bool, uint256), these are
			// coded as just like uint256,bool,uint256
			virtualArgs += getTypeSize(arg.Type)/32 - 1
		}
		if err != nil {
			return nil, err
		}
		retval = append(retval, marshalledValue)
	}
	return retval, nil
}

//==================== Custom Methods=====================================

func (arguments Arguments) ReturnBytesUnpack(data string) []interface{} {
	var list = make([]interface{}, len(arguments))
	var err error
	dataBytes, _ := hexutil.Decode(data)

	// todo: ?? isTuple()
	if arguments.isTuple() {
		err = arguments.UnpackV2(&list, dataBytes)
	} else {
		err = arguments.UnpackV2(&list[0], dataBytes)
	}

	if err != nil {
		// todo: error handle
		fmt.Printf("the error is %v\n", err)
	}

	return list
}
