package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

//TODO
func Test_retrieveFnNameAndParams(t *testing.T) {
	fnNameInput := "registerRole"
	var input = MakeInput(fnNameInput, "v2.0.0", "abc")
	fnName, fn, params, err := retrieveFnAndParams(input, (&UserManagement{}).AllExportFns())
	if nil != err {
		t.Error(err)
		return
	}

	assert.Equal(t, fnNameInput, fnName, "function name is invalid")
	assert.Equal(t, true, reflect.ValueOf(fn).IsValid(), "function is invalid")
	assert.Equal(t, 2, len(params), "params length is invalid")
	assert.Equal(t, "v2.0.0", params[0].String(), "params invalid")
}

func Test_execSC(t *testing.T) {
	fnNameInput := "registerRole"
	var input = MakeInput(fnNameInput, "aaa", "bbb")
	ret, err := execSC(input, (&UserManagement{}).AllExportFns())
	if nil != err {
		t.Error(err)
		return
	}

	assert.Equal(t, []byte("aaabbb"), ret)

	input = MakeInput(fnNameInput, "bbb")
	_, err = execSC(input, (&UserManagement{}).AllExportFns())
	assert.EqualError(t, err, "The params number invalid")
}

// hex.encode( rlp.encode( [][]byte{rlp.encode(txType), function name,rlp.encode(params[1]), rlp.encode(params[1])...} ) )
func MakeInput(fnName string, params ...interface{}) []byte {
	input := make([][]byte, 0)
	txTyp, err := rlp.EncodeToBytes(E_INVOKE_CONTRACT)
	if nil != err {
		panic(err)
	}
	input = append(input, txTyp)

	input = append(input, []byte(fnName))

	for _, v := range params {
		param, err := rlp.EncodeToBytes(v)
		if nil != err {
			panic(err)
		}
		input = append(input, param)
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
