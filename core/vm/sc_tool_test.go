package vm

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common/byteutil"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type fakeClass struct{}

func (fc *fakeClass) Fn(name string, age int64) (string, error) {
	return fmt.Sprintf("%s+%d", name, age), nil
}

func (fc *fakeClass) allExportFns() SCExportFns {
	return SCExportFns{
		"Fn": fc.Fn,
	}
}

func Test_retrieveFnNameAndParams(t *testing.T) {
	fnNameInput := "Fn"
	name := "wanxiang"
	var age int64 = 3
	var input = MakeInput(fnNameInput, name, age)
	fnName, fn, params, err := retrieveFnAndParams(input, (&fakeClass{}).allExportFns())
	if nil != err {
		t.Error(err)
		return
	}

	assert.Equal(t, fnNameInput, fnName, "function name is invalid")
	assert.Equal(t, true, reflect.ValueOf(fn).IsValid(), "function is invalid")
	assert.Equal(t, 2, len(params), "params length is invalid")
	assert.Equal(t, name, params[0].String(), "params invalid")
	assert.Equal(t, age, params[1].Int(), "params invalid")
}

func Test_execSC(t *testing.T) {
	fnNameInput := "Fn"
	name := "wanxiang"
	var age int64 = 3
	var input = MakeInput(fnNameInput, name, age)

	ret, err := execSC(input, (&fakeClass{}).allExportFns())
	if nil != err {
		t.Error(err)
		return
	}

	ret2, err := (&fakeClass{}).Fn(name, age)
	assert.NoError(t, err)
	assert.Equal(t, []byte(ret2), ret)

	input = MakeInput(fnNameInput, "bbb")
	_, err = execSC(input, (&fakeClass{}).allExportFns())
	assert.Error(t, err, "The params number invalid")
}

func MakeInput(fnName string, params ...interface{}) []byte {
	input := make([][]byte, 0)

	txTyp := byteutil.Int64ToBytes(int64(E_INVOKE_CONTRACT))

	input = append(input, txTyp)
	input = append(input, []byte(fnName))

	for _, v := range params {
		switch v.(type) {
		case int, int8, int16, int32, int64:
			param := byteutil.Int64ToBytes(reflect.ValueOf(v).Int())
			input = append(input, param)
		case uint, uint8, uint16, uint32, uint64:
			param := byteutil.Uint64ToBytes(reflect.ValueOf(v).Uint())
			input = append(input, param)
		case string:
			input = append(input, []byte(v.(string)))
		default:
			panic("unsupported type")
		}
	}

	encodedInput, err := rlp.EncodeToBytes(input)
	if nil != err {
		panic(err)
	}

	return encodedInput
}

type TxType uint8

const (
	E_INVOKE_CONTRACT TxType = 1
)
