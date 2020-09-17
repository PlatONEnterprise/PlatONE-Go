package packet

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"
	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

type eventParsingFunc func(eLog *Log, abiBytes []byte) string

func getSysEventAbis(SysEventList []string) (abiBytesArr [][]byte) {
	for _, data := range SysEventList {
		p := precompile.List[data]
		abiBytes, _ := precompile.Asset(p)
		abiBytesArr = append(abiBytesArr, abiBytes)
	}

	return
}

func EventParsing(logs RecptLogs, abiBytesArr [][]byte, fn eventParsingFunc) []string {
	var res []string

	for _, logData := range logs {
		for _, data := range abiBytesArr {
			result := fn(logData, data)
			if result != "" {
				res = append(res, result)
				break
			}
		}
	}

	return res
}

// ============================ WASM ==================================

func WasmEventParsingPerLog(eLog *Log, abiBytes []byte) string {
	var rlpList []interface{}

	eventName, topicTypes := findWasmLogTopic(eLog.Topics[0], abiBytes)

	if len(topicTypes) == 0 {
		return ""
	}

	dataBytes, _ := hexutil.Decode(eLog.Data)
	err := rlp.DecodeBytes(dataBytes, &rlpList)
	if err != nil {
		// todo: error handle
		fmt.Printf("the error is %v\n", err)
	}

	result := fmt.Sprintf("Event %s: ", eventName)
	result += parseReceiptLogData(rlpList, topicTypes)

	return result
}

func findWasmLogTopic(topic string, abiBytes []byte) (string, []string) {
	abiFunc, err := ParseAbiFromJson(abiBytes)
	if err != nil {
		return "", nil
	}

	for _, data := range abiFunc {
		if data.Type != "event" {
			continue
		}

		if strings.EqualFold(wasmLogTopicEncode(data.Name), topic) {
			topicTypes := make([]string, 0)
			name := data.Name
			for _, v := range data.Inputs {
				topicTypes = append(topicTypes, v.Type)
			}
			return name, topicTypes
		}
	}

	return "", nil
}

// ================================== EVM ========================
func EvmEventParsingPerLog(eLog *Log, abiBytes []byte) string {
	eventName, arguments := findEvmLogTopic(eLog.Topics[0], abiBytes)
	if arguments == nil {
		return ""
	}

	result := fmt.Sprintf("Event %s: ", eventName)
	rlpList := arguments.ReturnBytesUnpack(eLog.Data)

	for _, data := range rlpList {
		if data != nil && !reflect.ValueOf(data).IsZero() {
			result += fmt.Sprintf("%v ", data)
		}
	}

	return result
}

func findEvmLogTopic(topic string, abiBytes []byte) (string, abi.Arguments) {
	abiFunc, err := ParseAbiFromJson(abiBytes)
	if err != nil {
		return "", nil
	}

	for _, data := range abiFunc {
		if data.Type != "event" {
			continue
		}

		if strings.EqualFold(evmLogTopicEncode(data), topic) {
			name := data.Name
			arguments := GenUnpackArgs(data.Inputs)
			return name, arguments
		}
	}

	return "", nil
}
