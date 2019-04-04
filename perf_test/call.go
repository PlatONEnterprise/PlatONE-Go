package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/BCOSnetwork/BCOS-Go/cmd/ctool/core"
	"github.com/BCOSnetwork/BCOS-Go/log"
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

	num, _ := strconv.ParseInt(resp["result"].(string), 0, 64)

	return num
}

func getBlockTxNum(h int64) (num, timestamp int64) {
	var height interface{}
	height = "0x" + strconv.FormatInt(h, 16)
	params := []interface{}{height, true}
	r, err := Send(params, "eth_getBlockByNumber")
	if err != nil {
		fmt.Printf("send http post to get contract address error ")
		return -1, -1
	}

	var resp map[string]interface{}
	err = json.Unmarshal([]byte(r), &resp)
	if err != nil {
		panic(err)
	}

	txList := resp["result"].(map[string]interface{})["transactions"].([]interface{})

	timestamp, _ = strconv.ParseInt(resp["result"].(map[string]interface{})["timestamp"].(string), 0, 64)
	num = int64(len(txList))
	return
}

func getPerfResults() (interface{}) {
	var ret interface{}
	var err error
	if *stressTest == 1 {
		err, ret = invoke(*contractAddress, *abiPath, "getName()", *txType)
		if err != nil {
			panic(err)
		}
		return ret
	} else {
		var records map[string]interface{}

		err, ret = invoke(*contractAddress, *abiPath, "getRegisteredContracts(0,10000)", *txType)
		if err != nil {
			panic(err)
		}
		trimRet := []byte(ret.(string))
		l := len(trimRet) - 1
		for trimRet[l] == byte(0) {
			l--
		}

		if err := json.Unmarshal(trimRet[:l+1], &records); err != nil {
			panic(err)

		} else if records["code"] != 0 || records["msg"].(string) != "ok" {
			log.Error("contract inner error", "code", records["code"], "msg", records["msg"].(string))
		}

		registerContracts := int64(records["data"].(map[string]interface{})["total"].(float64)) - 1
		return registerContracts
	}
}
