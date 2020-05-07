package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"io/ioutil"
	"net/http"
	"strings"
)

// JsonParam, JSON-RPC request
type JsonParam struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      int         `json:"id"`
}

// Response, the response of JSON-RPC
type RpcResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Id      int         `json:"id"`
	Error   struct {
		Code    int32  `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

var url string

// SetHttpUrl set the url for HttpPost
func SetHttpUrl(str string) {
	if strings.HasPrefix(str, "http://") {
		str = str[7:]
	}

	if !IsUrl(str) {
		utils.Fatalf(ErrParamInValidSyntax, "url")
	}
	url = str
}

// NewRpcJson new a JsonParam object
func NewRpcJson(action string, params interface{}) JsonParam {

	if action == "" {
		// error?
	}

	param := JsonParam{
		Jsonrpc: "2.0",
		Method:  action,
		Params:  params,
		Id:      1,
	}

	return param
}

// HttpPost post a http request, parse the response and return the body
func HttpPost(param JsonParam) (string, error) {

	client := &http.Client{}
	req, _ := json.Marshal(param)
	reqNew := bytes.NewBuffer(req)

	request, _ := http.NewRequest("POST", "http://"+url, reqNew)
	request.Header.Set("Content-type", "application/json")
	response, err := client.Do(request)

	switch {
	case response == nil && err != nil:
		return "", fmt.Errorf(ErrHttpNoResponseFormat, err.Error())
	case err == nil && response == nil:
		return "", nil	// TODO no response
	case err == nil && response.StatusCode == 200:
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			utils.Fatalf(ErrTODO, DEFAULT_LOG_DIRT)
		}
		return string(body), nil
	default:
		return "", fmt.Errorf(ErrHttpResponseStatusFormat, response.Status)
	}
}

// ParseRpcResponse parse the response of the RPC-JSON and
// return the result field of Response object if there is no error
func ParseRpcResponse(r string) (interface{}, error) {
	var resp = RpcResponse{}

	if r == "" {
		return "", errors.New("no rpc response")
	}

	err := json.Unmarshal([]byte(r), &resp)
	Logger.Printf("the rpc response is %+v\n", resp)

	switch {
	case err != nil:
		LogErr.Printf(ErrUnmarshalBytesFormat, "http response", err.Error())
		return nil, fmt.Errorf(ErrTODO, DEFAULT_LOG_DIRT)
	case resp.Error.Code != 0:
		return nil, fmt.Errorf(resp.Error.Message)
	default:
		return resp.Result, nil
	}
}
