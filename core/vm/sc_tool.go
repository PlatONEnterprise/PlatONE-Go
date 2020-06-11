package vm

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common/byteutil"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"reflect"
)

func execSC(input []byte, fns SCExportFns) ([]byte, error) {
	_, fn, params, err := retrieveFnAndParams(input, fns)
	if nil != err {
		log.Error("failed to retrieve func name and params.", "error", err)
		return nil, err
	}

	//execute system contract method
	//all the export method of system contracts must return two results,
	//first result type is: []byte, second result type: error
	result := reflect.ValueOf(fn).Call(params)
	if err, ok := result[1].Interface().(error); ok {
		log.Error("execute system contract failed.", "error", err)
		return nil, err
	}

	return result[0].Bytes(), nil
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
		return "", nil, nil, SC_ERR_FUNC_NOT_FOUND
	}

	fnType := reflect.TypeOf(fn)
	paramNum := fnType.NumIn()
	if paramNum != len(args)-2 {
		log.Info("params number invalid. ", "expected:", paramNum, "got:", len(args)-2)
		return "", nil, nil, SC_ERR_PARAMS_NUMBER_INVALID
	}

	fnParams = make([]reflect.Value, paramNum)
	for i := 0; i < paramNum; i++ {
		targetType := fnType.In(i).String()
		inputByte := args[i+2]
		fnParams[i] = byteutil.ConvertBytesTo(inputByte, targetType)
	}

	return fnName, fn, fnParams, nil
}

func IsEmpty(input []byte) bool {
	if len(input) == 0 {
		return true
	} else {
		return false
	}
}
