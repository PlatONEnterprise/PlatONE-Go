package byteutil

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"reflect"
)

func ConvertBytesTo(input []byte, targetType string) reflect.Value {
	return reflect.ValueOf(Bytes2X_CMD[targetType]).Call([]reflect.Value{reflect.ValueOf(input)})[0]
}

var Bytes2X_CMD = map[string]interface{}{
	"string": BytesToString,

	"uint16": BytesToUint16,
	"uint32": BytesToUint32,
	"uint64": BytesToUint64,

	"*syscontracts.NodeInfo":   BytesToNodeInfo,
	"*syscontracts.UpdateNode": BytesToUpdateNode,
}

func BytesToUpdateNode(curByte []byte) *syscontracts.UpdateNode {
	var info syscontracts.UpdateNode
	if err := json.Unmarshal(curByte, &info); nil != err {
		panic("BytesToUpdateNode:" + err.Error())
	}
	return &info
}

func BytesToNodeInfo(curByte []byte) *syscontracts.NodeInfo {
	var info syscontracts.NodeInfo
	if err := json.Unmarshal(curByte, &info); nil != err {
		panic("BytesToNodeInfo:" + err.Error())
	}
	return &info
}

func BytesToString(curByte []byte) string {
	return string(curByte)
}

func BytesToInt16(b []byte) int16 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int16
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return tmp
}

func BytesToInt32(b []byte) int32 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return tmp
}

func BytesToInt64(b []byte) int64 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int64
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return tmp
}

func BytesToUint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func BytesToUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}
