package common

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const JSONRequiredTag = "required"

func isValueZero(val reflect.Value) bool {
	return val.IsZero()
}

func CheckRequiredFields(isEmpty func(reflect.Value) bool, data interface{}) error {
	vData := reflect.ValueOf(data)
	if reflect.Ptr == vData.Type().Kind() {
		vData = vData.Elem()
		if reflect.Struct != vData.Type().Kind() {
			return errors.New("argument must be struct type or pointer of struct")
		}
	} else if reflect.Struct != vData.Type().Kind() {
		return errors.New("argument must be struct type or pointer of struct")
	}

	numFiled := vData.Type().NumField()
	for i := 0; i < numFiled; i++ {
		val := vData.Field(i)
		tag, ok := vData.Type().Field(i).Tag.Lookup("json")
		isRequiredFiled := strings.Index(tag, JSONRequiredTag) != -1
		if ok && isRequiredFiled {
			if isEmpty(val) {
				return errors.New(fmt.Sprintf("%s field is required,but value is empty.", vData.Type().Field(i).Name))
			}
		}
	}

	return nil
}

func CheckRequiredFieldsIsEmpty(data interface{}) error {
	return CheckRequiredFields(isValueZero, data)
}
