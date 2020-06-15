package vm

import (
	"reflect"
	"testing"
)

func TestF(t *testing.T) {
	t2 := struct {
		Name   string `json:"name,required"`
		Age    int    `json:"age"`
		gender int    `json:"gender"`
	}{
		Name: "elvindd",
		Age:  12,
	}
	vd := reflect.ValueOf(t2)
	v,ok :=vd.Type().Field(0).Tag.Lookup("json")
	t.Log(v)
	t.Log(ok)
}
