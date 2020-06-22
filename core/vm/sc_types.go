package vm

import (
	"encoding/json"
	"fmt"
)

const (
	resultCodeSuccess       = 0
	resultCodeInternalError = 1
)

type result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func newResult(code int, msg string, data interface{}) *result {
	return &result{Code: code, Msg: msg, Data: data}
}

func newSuccessResult(data interface{}) *result {
	return newResult(resultCodeSuccess, "success", data)
}

func (res *result) String() string {
	b, err := json.Marshal(res)
	if err != nil {
		return fmt.Sprintf(`{"code":%d,"msg":"%s"}`, resultCodeInternalError, err.Error())
	}

	return string(b)
}
