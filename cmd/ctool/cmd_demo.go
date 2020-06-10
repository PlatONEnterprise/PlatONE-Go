// This file is used to test the new command
// before the command is moved to the categories
package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"strings"
)

//=============================2020.06========================================

const (
	PARAM_MANAGER = "__sys_ParamManager"
	TX_GASLIMIT_MIN = "1277159600"
	TX_GASLIMIT_MAX = "2000000000"
	BLOCK_GASLIMIT_MIN = "1277159600"
	BLOCK_GASLIMIT_MAX = "20000000000"
)

func checkConfigParam(param, key string) bool {
	if param == "" {
		return false
	}

	switch key {
	case "TxGasLimit":
		// number check
		isInRange := (strings.Compare(TX_GASLIMIT_MIN, param) + strings.Compare(TX_GASLIMIT_MAX, param)) == 0 && len(param) == 10
		if !isInRange {
			fmt.Printf("the transaction gas limit should be within (%s, %s)\n", TX_GASLIMIT_MIN, TX_GASLIMIT_MAX)
			return false
		}
	case "BlockGasLimit":
		// number check
		// range check
	default:

	}

	return true
}

func setSysConfig(c *cli.Context) {


	txGasLimit := c.String(TxGasLimitFlags.Name)
	blockGasLimit := c.String(BlockGasLimitFlags.Name)

	setConfig(c, txGasLimit, "TxGasLimit")
	setConfig(c, blockGasLimit, "BlockGasLimit")

}

func setConfig(c *cli.Context, param, name string) {

	if !checkConfigParam(param, name){
		return
	}

	funcName := "set" + name
	funcParams := CombineFuncParams(param)

	result := contractCommon(c, funcParams, funcName, PARAM_MANAGER)
	fmt.Printf("result: %s\n", result)
}

func getSysConfig(c *cli.Context) {

	txGasLimit := c.Bool(GetTxGasLimitFlags.Name)
	blockGasLimit := c.Bool(GetBlockGasLimitFlags.Name)

	getConfig(c, txGasLimit, "getTxGasLimit")
	getConfig(c, blockGasLimit, "getBlockGasLimit")

}

func getConfig(c *cli.Context, isGet bool, funcName string) {

	if isGet {
		result := contractCommon(c, nil, funcName, PARAM_MANAGER)
		fmt.Printf("%s result: %v\n", funcName, result)
	}
}

//================================2020.05====================================================

type nodeJson struct {
	Name       string `json:"name"`
	Owner      string `json:"owner"`
	Desc       string `json:"desc"`
	Type       int    `json:"type"`
	Publickey  string `json:"publickey"`
	ExternalIP string `json:"externalIP"`
	InternalIP string `json:"internalIP"`
	RpcPort    int    `json:"rpcPort"`
	P2pPort    int    `json:"p2pPort"`
	Status     int    `json:"status"`
	DelayNum   int    `json:"delayNum,omitempty"`
}

// TODO combineJson refactory
func nodeAddDemo(c *cli.Context) {

	// required value
	var strMustArray = []string{"name", "publicKey", "externalIP", "internalIP"}

	// default or user input value
	var strConst = "\"owner\":\"todo\",\"status\":1,\"type\":0,"
	var strDefault = "\"rpcPort\":6791,\"p2pPort\":1800,\"desc\":\"add node to the list\","
	var strOption = "\"delayNum\":\"\""

	var strJson = fmt.Sprintf("{%s%s%s}", strConst, strDefault, strOption)

	/*
		var nodeJsonStr = nodeJson{
			Status: 1,
			Type: 0,
			RpcPort: 6791,
			P2pPort: 1800,
		}

		nodeJsonBytes, _ := json.Marshal(nodeJsonStr)*/

	// combine to json format
	str := combineJson(c, strMustArray, []byte(strJson))

	funcParams := []string{str}
	result := contractCommon(c, funcParams, "add", "__sys_NodeManager")
	fmt.Printf("result: %s\n", result)
}
