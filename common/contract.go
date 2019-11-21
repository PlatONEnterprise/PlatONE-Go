package common

import (
	"bytes"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"reflect"
)

var currentInterpreterType string

func SetCurrentInterpreterType(cit string) {
	currentInterpreterType = cit
}

func GetCurrentInterpreterType() string {
	return currentInterpreterType
}

func IsWasmContractCode(code []byte) (ok bool, tx_type int64, abi, bytecode []byte) {
	var err error
	tx_type, abi, bytecode, err = ParseWasmCodeRlpData(code)
	if err != nil {
		return
	}
	if bytes.Equal(bytecode[:8], []byte{0, 97, 115, 109, 1, 0, 0, 0}) {
		ok = true
		return
	}
	return
}

func ParseWasmCodeRlpData(rlpData []byte) (int64, []byte, []byte, error) {
	ptr := new(interface{})
	err := rlp.Decode(bytes.NewReader(rlpData), &ptr)
	if err != nil {
		return -1, nil, nil, err
	}
	rlpList := reflect.ValueOf(ptr).Elem().Interface()

	if _, ok := rlpList.([]interface{}); !ok {
		return -1, nil, nil, fmt.Errorf("invalid rlp format.")
	}

	iRlpList := rlpList.([]interface{})
	if len(iRlpList) <= 2 {
		return -1, nil, nil, fmt.Errorf("invalid input. ele must greater than 2")
	}
	var (
		txType int64
		code   []byte
		abi    []byte
	)
	if v, ok := iRlpList[0].([]byte); ok {
		txType = BytesToInt64(v)
	}
	if v, ok := iRlpList[1].([]byte); ok {
		code = v
	}
	if v, ok := iRlpList[2].([]byte); ok {
		abi = v
	}
	return txType, abi, code, nil
}
