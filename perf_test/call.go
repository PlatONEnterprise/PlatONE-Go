package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/BCOSnetwork/BCOS-Go/cmd/ctool/core"
)

// 判断合约是否成功上链
func getContractByAddress(addr string) bool {
	params := []string{addr, "latest"}
	r, err := Send(params, "eth_getCode")
	if err != nil {
		fmt.Printf("send http post to get contract address error ")
		return false
	}

	var resp = core.Response{}
	err = json.Unmarshal([]byte(r), &resp)
	if err != nil {
		fmt.Printf("parse eth_getCode result error ! \n %s", err.Error())
		return false
	}

	if resp.Error.Code != 0 {
		fmt.Printf("eth_getCode error ,error:%v", resp.Error.Message)
		return false
	}
	//fmt.Printf("trasaction hash: %s\n", resp.Result)

	if resp.Result != "" && len(resp.Result) > 2 {
		return true
	} else {
		return false
	}
}

func getTxByHash(hash string) bool {

	params := []string{hash}
	r, err := Send(params, "eth_getTransactionByHash")
	if err != nil {
		fmt.Printf("send http post to get contract address error ")
		return false
	}

	var resp map[string]interface{}
	err = json.Unmarshal([]byte(r), &resp)
	if err != nil {
		panic(err)
	}

	if resp["result"].(map[string]interface{})["blockNumber"] != nil {
		return true
	} else {
		return false
	}
}

func getCurrentBlockNum() int64 {

	params := []string{}
	r, err := Send(params, "eth_blockNumber")
	if err != nil {
		fmt.Printf("send http post to get contract address error ")
		return -1
	}

	var resp map[string]interface{}
	err = json.Unmarshal([]byte(r), &resp)
	if err != nil {
		panic(err)
	}

	d, _ := strconv.ParseInt(resp["result"].(string), 0, 64)
	return d
}

func getBlockTxNum(h int64) int64 {
	var height interface{}
	height = "0x" + strconv.FormatInt(h, 16)
	params := []interface{}{height, true}
	r, err := Send(params, "eth_getBlockByNumber")
	if err != nil {
		fmt.Printf("send http post to get contract address error ")
		return -1
	}

	var resp map[string]interface{}
	err = json.Unmarshal([]byte(r), &resp)
	if err != nil {
		panic(err)
	}

	tmp := resp["result"].(map[string]interface{})["transactions"].([]interface{})

	return int64(len(tmp))
}
