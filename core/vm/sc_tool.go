package vm

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common/byteutil"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/life/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"reflect"
)

var (
	errFuncNotFoundInExportFuncs = errors.New("the func not found in export function set")
	errParamsNumInvalid          = errors.New("the number of params is invalid")
)

func execSC(input []byte, fns SCExportFns) ([]byte, error) {
	_, fn, params, err := retrieveFnAndParams(input, fns)
	if nil != err {
		log.Error("failed to retrieve func name and params.", "error", err)
		return nil, err
	}

	//execute system contract method
	//all the export method of system contracts must return two results,
	//first result type is: primitive type, second result type: error
	result := reflect.ValueOf(fn).Call(params)
	if err, ok := result[1].Interface().(error); ok {
		log.Error("execute system contract failed.", "error", err)
		return nil, err
	}

	return primitiveToBytes(result[0]), nil
}

func primitiveToBytes(val reflect.Value) []byte {
	switch val.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return utils.Uint64ToBytes(val.Uint())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return utils.Int64ToBytes(val.Int())
	case reflect.String:
		return []byte(val.String())
		//case reflect.Bool:
		//case reflect.Float64, reflect.Float32:
	}

	panic("unsupported type")
}

func retrieveFnAndParams(input []byte, fns SCExportFns) (fnName string, fn SCExportFn, fnParams []reflect.Value, err error) {
	defer func() {
		if err := recover(); nil != err {
			fn, fnParams, err = nil, nil, fmt.Errorf("parse tx data failed:%s", err)
			log.Error("Failed to parse tx data", "error", err, "input", input)
		}
	}()

	var args [][]byte
	if err := rlp.DecodeBytes(input, &args); nil != err {
		log.Error("Failed to verify input of system contract,Decode rlp input failed", "error", err)
		return "", nil, nil, err
	}
	//txType := int(common.BytesToInt64(args[0]))
	fnName = string(args[1])

	var ok bool
	if fn, ok = fns[fnName]; !ok {
		return "", nil, nil, errFuncNotFoundInExportFuncs
	}

	fnType := reflect.TypeOf(fn)
	paramNum := fnType.NumIn()
	if paramNum != len(args)-2 {
		log.Warn("params number invalid. ", "expected:", paramNum, "got:", len(args)-2)
		return "", nil, nil, errParamsNumInvalid
	}

	fnParams = make([]reflect.Value, paramNum)
	for i := 0; i < paramNum; i++ {
		targetType := fnType.In(i).String()
		inputByte := args[i+2]
		fnParams[i] = byteutil.ConvertBytesTo(inputByte, targetType)
	}

	return fnName, fn, fnParams, nil
}

func CheckPublicKeyFormat(pub string) error {
	b, err := hex.DecodeString(pub)
	if err != nil {
		return err
	}

	_, err = crypto.UnmarshalPubkey(b)
	if err != nil {
		return err
	}

	return nil
}
