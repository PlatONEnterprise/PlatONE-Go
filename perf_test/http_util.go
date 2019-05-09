package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/core"
)

func Send(params interface{}, action string) (string, error) {
	param := core.JsonParam{
		Jsonrpc: "2.0",
		Method:  action,
		Params:  params,
		Id:      1,
	}
	resp, err := HttpPost(param)
	if err != nil {
		panic(fmt.Sprintf("send http post error .\n %s" + err.Error()))
	}

	return resp, err
}

func HttpPost(param core.JsonParam) (string, error) {

	client := &http.Client{}
	req, _ := json.Marshal(param)
	reqNew := bytes.NewBuffer(req)

	//fmt.Println(string(reqNew.Bytes()))
	request, _ := http.NewRequest("POST", config.HttpUrl, reqNew)
	request.Header.Set("Content-type", "application/json")
	response, err := client.Do(request)
	//defer response.Body.Close()
	if response == nil || err != nil {
		panic(fmt.Sprintf("no response from node,%s", err.Error()))
	}
	if err == nil && response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		return string(body), nil
	} else {
		panic(fmt.Sprintf("http response status :%s", response.Status))
	}
	return "", err
}

func parseResponse(r string) *core.Response {
	var resp = core.Response{}
	err := json.Unmarshal([]byte(r), &resp)

	if err != nil {
		panic(fmt.Sprintf("parse result error ! error:%s \n", err.Error()))
	}

	if resp.Error.Code != 0 {
		fmt.Printf("send transaction error ,error:%v \n", resp.Error.Message)
		return nil
	}
	return &resp
}
