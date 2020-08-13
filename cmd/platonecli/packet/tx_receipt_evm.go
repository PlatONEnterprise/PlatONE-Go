package packet

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
)

// EventParsing parsing all the events recorded in receipt log.
// The event should be written in the abiBytes provided.
// Otherwise, the event will not be parsed
func EventParsingV2(logs RecptLogs, abiBytes []byte) (result string) {

	for i, eLog := range logs {
		eventName, arguments := findLogTopicV2(eLog.Topics[0], abiBytes)
		if arguments == nil {
			continue
		}

		result += fmt.Sprintf("\nEvent[%d]: %s ", i, eventName)
		rlpList := arguments.ReturnBytesUnpack(eLog.Data)

		for _, data := range rlpList {
			if data != nil && !reflect.ValueOf(data).IsZero() {
				result += fmt.Sprintf("%v ", data)
			}
		}

		result += "\n"
	}

	return
}

func findLogTopicV2(topic string, abiBytes []byte) (string, abi.Arguments) {
	abiFunc, err := ParseAbiFromJson(abiBytes)
	if err != nil {
		return "", nil
	}

	for _, data := range abiFunc {
		if data.Type != "event" {
			continue
		}

		if strings.EqualFold(logTopicEncodeV2(data), topic) {
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
		argument.Type, _ = abi.NewTypeV2(v.Type, "", v.Components)
		argument.Name = v.Name
		argument.Indexed = v.Indexed

		arguments = append(arguments, argument)
	}

	return
}

// todo: similar to function selector???
// todo: optimization
func logTopicEncodeV2(data FuncDesc) string {
	var strArray = make([]string, 0)

	for _, event := range data.Inputs {
		strArray = append(strArray, event.Type)
	}

	topic := data.Name + "(" + strings.Join(strArray, ",") + ")"
	return common.BytesToHash(crypto.Keccak256([]byte(topic))).String()
}
