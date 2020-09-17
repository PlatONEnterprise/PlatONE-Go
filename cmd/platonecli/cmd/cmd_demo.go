// This file is used to test the new command
// before the command is moved to the categories
package cmd

import (
	"fmt"

	"gopkg.in/urfave/cli.v1"
)

//================================2020.05==================================

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

// todo: combineJson refactory
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
	result := contractCall(c, funcParams, "add", "__sys_NodeManager")
	fmt.Printf("%s\n", result)
}
