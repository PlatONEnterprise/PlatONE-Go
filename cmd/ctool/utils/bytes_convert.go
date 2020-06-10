package utils

import (
	"bytes"
	"encoding/binary"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	ethmath "github.com/PlatONEnetwork/PlatONE-Go/common/math"
	"math"
	"math/big"
)

func BytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

func Int32ToBytes(n int32) []byte {
	tmp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func BytesToInt32(b []byte) int32 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	_ = binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int32(tmp)
}

func Int64ToBytes(n int64) []byte {
	tmp := int64(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func BytesToInt64(b []byte) int64 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int64
	_ = binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int64(tmp)
}

func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

/*
func BigToByte128(I *big.Int) ([]byte, bool) {
	if len(I.Bytes()) > 16 {
		return []byte{}, false
	}
	res := make([]byte, 16)

	if I.Sign() == -1 {
		// invert then add 1 equals to sub 1 then invert
		Iminus := new(big.Int).Neg(I)
		Iminus.Sub(Iminus, Big1)
		copy(res[16-len(Iminus.Bytes()):], Iminus.Bytes())
		for i := range res {
			res[i] ^= 0xff
		}
	} else {
		copy(res[16-len(I.Bytes()):], I.Bytes())
	}
	return res, true
}

func Byte128ToBig(b []byte, s bool) *big.Int {
	r := new(big.Int)
	if s && b[0]&0x80 != 0 {
		//invert b
		for i := range b {
			b[i] ^= 0xff
		}
		r.SetBytes(b)
		r.Add(r, Big1)
		r.Neg(r)
		return r
	}

	r.SetBytes(b)
	return r
}*/

func Float32ToBytes(float float32) []byte {
	bits := math.Float32bits(float)
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, bits)
	return b
}

func BytesToFloat32(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	return math.Float32frombits(bits)
}

func Float64ToBytes(float float64) []byte {
	bits := math.Float64bits(float)
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, bits)
	return b
}

func BytesToFloat64(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	return math.Float64frombits(bits)
}

/*
// move to common/inner_convert.go?
func CallResAsFloat128(bts []byte) *big.Float {
	if len(bts) < 32 {
		return &big.Float{}
	}

	bytes := bts[len(bts)-16:]
	low := binary.BigEndian.Uint64(bytes[8:])
	high := binary.BigEndian.Uint64(bytes[:8])

	F, _ := math2.NewFromBits(high, low).Big()

	return F
}*/

func BoolToBytes(b bool) []byte {
	buf := bytes.NewBuffer([]byte{})
	_ = binary.Write(buf, binary.BigEndian, b)
	return buf.Bytes()
}

// BytesConverter converts the bytes to the specific data type
func BytesConverter(source []byte, t string) interface{} {
	switch t {
	case "int32":
		return common.CallResAsInt32(source)
	case "int64":
		return common.CallResAsInt64(source)
	case "float32":
		return BytesToFloat32(source)
	case "float64":
		return BytesToFloat64(source)
	/*
	case "float128":
		return CallResAsFloat128(source)*/
	case "string":
		source = bytes.TrimRight(source, "\x00")

		if len(source) < 64 {
			return string(source[:])
		} else {
			return string(source[64:])
		}
	case "uint64":
		return common.CallResAsUint64(source)
	default:
		// return source
		return bytes.TrimRight(source, "\x00")
	}
}

// U256 converts a big Int into a 256bit EVM number.
func U256(n *big.Int) []byte {
	return ethmath.PaddedBigBytes(ethmath.U256(n), 32)
}
