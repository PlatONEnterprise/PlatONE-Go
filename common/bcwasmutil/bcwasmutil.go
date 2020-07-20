package bcwasmutil

// unsigned_int is defined in bcwasm lib, which is used for DataStream serilize.
// When serilize a string to bytes, first pack the length of the string, then pack contents of the string.
// The length of the string is serlized with unsigned_int's methods defined below.
type UnsignedInt int32
func (i UnsignedInt)Bytes() []byte{
	val := i
	var ret []byte

	for val > 0 {
		b := val & 0x7f
		val = val >> 7
		if val > 0 {
			b |= 1 << 7
		}
		ret = append(ret, byte(b))
	}
	return ret
}

func (i UnsignedInt)Uint32() uint32{
	return uint32(i)
}

func (i UnsignedInt)Int32() int32{
	return int32(i)
}

func BytesToUnsignedInt(data []byte) (val UnsignedInt, pos int){
	for i, d := range data {
		val |= UnsignedInt(d & 0x7f)
		if d >> 7 == 0{
			pos = i
			break
		}
		val = val << 7
	}

	return val, pos
}

func SerilizString(str string) []byte{
	var ret []byte

	length := len(str)+1
	lenBytes := UnsignedInt(length).Bytes()

	ret = append(ret, lenBytes...)
	ret = append(ret, []byte(str)...)
	ret = append(ret, 0)

	return ret
}

func DeserilizeString(data []byte) string{
	length, pos := BytesToUnsignedInt(data)
	if int(length)> len(data) {
		return ""
	}

	return string(data[pos+1:int(length)-1])
}

