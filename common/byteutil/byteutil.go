package byteutil

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"reflect"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
)

func ConvertBytesTo(input []byte, targetType string) reflect.Value {
	v, ok := Bytes2X_CMD[targetType]
	if !ok {
		panic("unsupported type")
	}

	return reflect.ValueOf(v).Call([]reflect.Value{reflect.ValueOf(input)})[0]
}

var Bytes2X_CMD = map[string]interface{}{
	"string": BytesToString,

	"uint16": BytesToUint16,
	"uint32": BytesToUint32,
	"uint64": BytesToUint64,

	"int":   BytesToInt,
	"int16": BytesToInt16,
	"int32": BytesToInt32,
	"int64": BytesToInt64,

	"*syscontracts.NodeInfo":     BytesToNodeInfo,
	"*syscontracts.UpdateNode":   BytesToUpdateNode,
	"*syscontracts.UserInfo":     BytesToUserInfo,
	"*syscontracts.UserDescInfo": BytesToUserDescInfo,
	"common.Address":             HexBytesToAddress,
}

func HexBytesToAddress(curByte []byte) common.Address {
	return common.HexToAddress(string(curByte))
}
func BytesToUpdateNode(curByte []byte) *syscontracts.UpdateNode {
	var info syscontracts.UpdateNode
	if err := json.Unmarshal(curByte, &info); nil != err {
		panic("BytesToUpdateNode:" + err.Error() + " bytes:" + string(curByte))
	}
	return &info
}

func BytesToNodeInfo(curByte []byte) *syscontracts.NodeInfo {
	var info syscontracts.NodeInfo
	if err := json.Unmarshal(curByte, &info); nil != err {
		panic("BytesToNodeInfo:" + err.Error() + " bytes:" + string(curByte))
	}
	return &info
}

func BytesToUserInfo(curByte []byte) *syscontracts.UserInfo {
	var info syscontracts.UserInfo
	if err := json.Unmarshal(curByte, &info); nil != err {
		panic("BytesToUserInfo:" + err.Error() + " bytes:" + string(curByte))
	}
	return &info
}

func BytesToUserDescInfo(curByte []byte) *syscontracts.UserDescInfo {
	var info syscontracts.UserDescInfo
	if err := json.Unmarshal(curByte, &info); nil != err {
		panic("BytesToUserInfo:" + err.Error() + " bytes:" + string(curByte))
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

func BytesToInt(b []byte) int {
	return int(BytesToInt32(b))
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

func Int64ToBytes(i int64) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, &i)
	return buf.Bytes()
}

func Int32ToBytes(i int32) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, &i)
	return buf.Bytes()
}

func Uint64ToBytes(n uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, n)
	return buf
}

func BoolToBytes(b bool) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, b)
	return buf.Bytes()
}
