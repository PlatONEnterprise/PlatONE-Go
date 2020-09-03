package packet

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
)

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

func GenUnpackArgs(data []abi.ArgumentMarshaling) (arguments abi.Arguments) {
	var argument abi.Argument

	for _, v := range data {
		argument.Type, _ = abi.NewTypeV2(v.Type, v.InternalType, v.Components)
		argument.Name = v.Name
		argument.Indexed = v.Indexed

		arguments = append(arguments, argument)
	}

	return
}

// todo: similar to function selector???
// todo: optimization
func evmLogTopicEncode(data FuncDesc) string {
	var strArray = make([]string, 0)

	for _, event := range data.Inputs {
		strArray = append(strArray, event.Type)
	}

	topic := data.Name + "(" + strings.Join(strArray, ",") + ")"
	return common.BytesToHash(crypto.Keccak256([]byte(topic))).String()
}
