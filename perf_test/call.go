package main

import (
	"encoding/json"
	"fmt"

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
