package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

const (
	testTxHash  = "0x31bbe55da1e59a9a0b79204afc7d89ccc8a5cea9722b252036cdcc6286a334ee"
	TestAccount = "0x60ceca9c1290ee56b98d4e160ef0453f7c40d219"
	testResult  = "0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000002a30783165366334633432366466633365643234333566373236666439613063323939356466316166633700000000000000000000000000000000000000000000"
)

type f func(rw http.ResponseWriter, r *http.Request)

type JsonParam struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      int         `json:"id"`
}

type Response struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Id      int         `json:"id"`
	Error   struct {
		Code    int32  `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type Receipt struct {
	ContractAddress string `json:"contractAddress"`
	Logs            []struct {
		Address string   `json:"address"`
		Topics  []string `json:"topics"`
		Data    string   `json:"data"`
	} `json:"logs"`
	Status string `json:"status"`
}

// MockServer mocks a server for http testing
// It could either test the http post or the rpc calls
func MockServer(s string) *httptest.Server {

	var funcSelect f

	switch s {
	case "http":
		funcSelect = httpServer
	case "rpc":
		funcSelect = rpcCallsServer
	}

	return httptest.NewServer(http.HandlerFunc(funcSelect))
}

// httpServer mocks multiple http response for http post testing
var httpServer = func(rw http.ResponseWriter, r *http.Request) {
	var request = JsonParam{}
	var response interface{}
	//var response = Response{Id:1}

	data, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("http data is %s\n", data)
	_ = json.Unmarshal(data, &request)

	switch request.Method {
	case "test1":
		//rw.WriteHeader(http.StatusOK)
		rw.Write(nil)
		response = nil

	case "test2":
		rw.Header().Set("Content-type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		response = "test2"

	case "test3":
		rw.WriteHeader(http.StatusOK)
		response = "test3 Success"

	}

	rw.Header().Set("Content-type", "application/json")
	_ = json.NewEncoder(rw).Encode(response)

}

// rpcCallsServer mocks multiple responses for testing
// based on the specific RPC APIs (eth_call, eth_getTransactionCount, etc.)
var rpcCallsServer = func(rw http.ResponseWriter, r *http.Request) {
	var request = JsonParam{}
	var response = Response{
		Jsonrpc: "2.0",
		Id:      1,
	}

	// read the body of the http request and get the JSON RPC object
	data, _ := ioutil.ReadAll(r.Body)
	//fmt.Printf("rpc data is %s\n", data)
	_ = json.Unmarshal(data, &request)

	// mocks the responses based on the specific RPC calls
	switch request.Method {
	case "eth_getTransactionCount":
		response.Result = "0x17"
		response.Error.Code = 0
		response.Error.Message = "success"

	case "eth_call":
		var dataBytes = make([][]byte, 0)

		i := request.Params.([]interface{})[0]
		data := i.(map[string]interface{})["data"]

		tempBytes, _ := hexutil.Decode(data.(string))
		_ = rlp.DecodeBytes(tempBytes, &dataBytes)

		if string(dataBytes[2]) == "tofu" {
			response.Result = testResult
			response.Error.Code = 0
			response.Error.Message = "success"
		} else {
			response.Error.Code = 1
			response.Error.Message = "failed"
		}

	case "eth_sendTransaction":

		i := request.Params.([]interface{})[0]
		data := i.(map[string]interface{})["data"]
		tempBytes, _ := hexutil.Decode(data.(string))
		_, _, _, err := common.ParseWasmCodeRlpData(tempBytes)

		if err != nil {
			response.Error.Code = 1
			response.Error.Message = "failed"
		} else {
			response.Result = testTxHash
			response.Error.Code = 0
			response.Error.Message = "success"
		}

	case "eth_getTransactionReceipt":
		recpt := Receipt{Status: "0x1"}
		response.Result = recpt
		response.Error.Code = 0
		response.Error.Message = "success"
		/*
			recpt := Receipt{Status: "0x1",}
			response.Result = recpt
			response.Error.Code = 1
			response.Error.Message = "success"*/

	}

	// write the http response
	rw.Header().Set("Content-type", "application/json")
	rw.WriteHeader(http.StatusOK)
	err := json.NewEncoder(rw).Encode(response)
	if err != nil {
		utils.Fatalf("%s\n", err.Error())
	}
}
