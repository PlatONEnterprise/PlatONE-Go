package rest

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/accounts/abi"
	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"
)

const separation = "-"

// ====================== Utils Common ===============================
// todo: code optimization
func getDataParams(i interface{}) ([]string, error) {
	var funcParams []string
	if i == nil {
		return nil, nil
	}

	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, errors.New("todo")
	}

	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)

		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		if value.Kind() == reflect.Struct || value.Kind() == reflect.Interface {
			marshalBytes, _ := json.Marshal(value.Interface())
			funcParams = append(funcParams, string(marshalBytes))
			continue
		} else if value.Type().Kind() != reflect.String {
			return nil, errors.New("todo")
		}

		temp := value.String()
		temp = strings.TrimSpace(temp)
		if temp != "" {
			if strings.Index(temp, "(") == 0 && strings.LastIndex(temp, ")") == len(temp)-1 {
				/// temp = abi.TrimSpace(temp)
				funcParams = append(funcParams, abi.GetFuncParams(temp[1:len(temp)-1])...)
			} else {
				funcParams = append(funcParams, temp)
			}
		}
	}

	return funcParams, nil
}

func paramsCheck(i interface{}) bool {
	var valid = true
	if i == nil {
		return true
	}

	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Name
		val := v.Field(i)

		if val.Kind() == reflect.Struct || val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
			/// valid = paramsCheck(val.Interface())
		} else if val.Kind() != reflect.String {
			return false
		}

		value := v.Field(i).String()
		if value != "" {
			valid = cmd_common.ParamValidWrap(value, strings.ToLower(key))
		}
	}

	return valid
}

// UrlParamConvert convert e.g. chain-admin -> chainAdmin
func UrlParamConvert(str string) string {
	var slice = make([]byte, 0)
	var count int

	for {
		index := strings.Index(str[count:], separation)
		if index == -1 {
			slice = append(slice, str[count:]...)
			break
		}

		slice = append(slice, str[count:count+index]...)
		slice = append(slice, str[count+index+1]-32)
		count += index + 2
	}

	return string(slice)
}
