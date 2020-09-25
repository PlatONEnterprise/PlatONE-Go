package exterror

import "fmt"

type ExtError struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
}

func NewError(code int, msg interface{}) *ExtError {
	return &ExtError{Msg: msg, Code: code}
}

func (this *ExtError) Error() string {
	return fmt.Sprintf(`{"code":%d,"msg":"%s"}`, this.Code, this.Msg)
}
