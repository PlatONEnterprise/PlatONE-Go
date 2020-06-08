package vm

import (
	"github.com/magiconair/properties/assert"
	"reflect"
	"testing"
)

//TODO
func Test_retrieveFnNameAndParams(t *testing.T) {
	var input = ""
	fnName, fn, params, err := retrieveFnNameAndParams([]byte(input), (&UserManagement{}).AllExportFns())
	if nil != err {
		t.Error(err)
		return
	}

	assert.Equal(t, fnName, "", "function name is invalid")
	assert.Equal(t, reflect.ValueOf(fn).IsValid(), true, "function is invalid")
	assert.Equal(t, len(params), 3, "params length is invalid")
	assert.Equal(t, params, []reflect.Value{reflect.ValueOf(1), reflect.ValueOf("abc")}, "params length is invalid")
}

func Test_execSC(t *testing.T) {

}
