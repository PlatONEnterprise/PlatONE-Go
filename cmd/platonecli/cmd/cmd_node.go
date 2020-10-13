package cmd

import (
	"encoding/json"
	"fmt"
	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"strconv"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"

	"gopkg.in/urfave/cli.v1"
)

var (
	NodeCmd = cli.Command{
		Name:  "node",
		Usage: "Manage nodes in PlatONE network",

		Subcommands: []cli.Command{
			NodeAddCmd,
			NodeDeleteCmd,
			NodeQueryCmd,
			NodeStatCmd,
			NodeUpdateCmd,
		},
	}

	NodeAddCmd = cli.Command{
		Name:      "add",
		Usage:     "Add a node to the node list",
		ArgsUsage: "<name> <publicKey> <externalIP> <internalIP> <status>",
		Action:    nodeAdd,
		Flags:     nodeAddCmdFlags,
		Description: `
		platonecli admin node add <name> <publicKey> <externalIP> <internalIP>

The newly added nodes can only be observer type.`,
	}

	NodeDeleteCmd = cli.Command{
		Name:      "delete",
		Usage:     "Delete a node from the node list, the deleted node can no longer receiving and synchronizing blocks",
		ArgsUsage: "<name>",
		Action:    nodeDelete,
		Flags:     globalCmdFlags,
		Description: `
		platonecli admin node delete <name>`,
	}

	NodeUpdateCmd = cli.Command{
		Name:      "update",
		Usage:     "Update the description, delay number, and node type of a node",
		ArgsUsage: "<name>",
		Action:    nodeUpdate,
		Flags:     nodeUpdateCmdFlags,
		Description: `
		platonecli admin node update <name>`,
	}

	NodeQueryCmd = cli.Command{
		Name:   "query",
		Usage:  "Query the node Info by the search key provided",
		Action: nodeQuery,
		Flags:  nodeQueryCmdFlasg,
		Description: `
		platonecli admin node query

Except --all flag, other search keys can be combined.`,
	}

	NodeStatCmd = cli.Command{
		Name:   "stat",
		Usage:  "Statistic the node Info by the search key provided",
		Action: nodeStat,
		Flags:  nodeStatCmdFlags,
		Description: `
		platonecli admin node stat`,
	}
)

// 2020.7.6 modified, precompiled contract + combineJson deprecated
func nodeAdd(c *cli.Context) {
	var nodeinfo common.NodeInfo

	nodeinfo.Name = c.Args().First() // todo: add to the usage
	nodeinfo.PublicKey = c.Args().Get(1)
	nodeinfo.ExternalIP = c.Args().Get(2)
	nodeinfo.InternalIP = c.Args().Get(3)
	delayNum,_ := strconv.ParseInt(c.String(NodeDelayNumFlags.Name),10,32)
	nodeinfo.DelayNum = uint64(delayNum)
	p2pPort,_ := strconv.ParseInt(c.String(NodeP2pPortFlags.Name), 10,32)
	nodeinfo.P2pPort = int32(p2pPort)
	rpcPort,_ := strconv.ParseInt(c.String(NodeRpcPortFlags.Name), 10,32)
	nodeinfo.RpcPort = int32(rpcPort)
	nodeinfo.Desc = c.String(NodeDescFlags.Name)
	//status64,_ := strconv.ParseInt(c.Args().Get(4),10,32)
	//nodeinfo.Status = int32(status64)
	bytes, _ := json.Marshal(nodeinfo)
	strJson := string(bytes)

	funcParams := []string{strJson}
	result := contractCall(c, funcParams, "add", precompile.NodeManagementAddress)
	fmt.Printf("%s\n", result)
}

func nodeDelete(c *cli.Context) {

	var str = "{\"status\":2}"

	name := c.Args().First()
	paramValid(name, "name")

	funcParams := cmd_common.CombineFuncParams(name, str)
	result := contractCall(c, funcParams, "update", precompile.NodeManagementAddress)
	fmt.Printf("%s\n", result)
}

func nodeUpdate(c *cli.Context) {

	// 可选(必填or必填)
	var strJson = "{\"type\":\"\",\"delayNum\":\"\",\"desc\":\"\"}"

	str := combineJson(c, nil, []byte(strJson))

	name := c.Args().First()
	paramValid(name, "name")

	funcParams := cmd_common.CombineFuncParams(name, str)
	result := contractCall(c, funcParams, "update", precompile.NodeManagementAddress)
	fmt.Printf("%s\n", result)
}

// TODO enode
func nodeQuery(c *cli.Context) {
	var strJson = "{\"type\":\"\",\"status\":\"\",\"name\":\"\",\"publicKey\":\"\"}"

	all := c.Bool(ShowAllFlags.Name)
	if all {
		result := contractCall(c, nil, "getAllNodes", precompile.NodeManagementAddress)
		strResult := PrintJson([]byte(result.(string)))
		fmt.Printf("result:\n%s\n", strResult)
		return
	}

	str := combineJson(c, nil, []byte(strJson))
	funcParams := cmd_common.CombineFuncParams(str)

	result := contractCall(c, funcParams, "getNodes", precompile.NodeManagementAddress)
	strResult := PrintJson([]byte(result.(string)))
	fmt.Printf("result:\n%s\n", strResult)
}

func nodeStat(c *cli.Context) {
	var strJson = "{\"type\":\"\",\"status\":\"\"}"

	str := combineJson(c, nil, []byte(strJson))
	funcParams := cmd_common.CombineFuncParams(str)

	result := contractCall(c, funcParams, "nodesNum", precompile.NodeManagementAddress)
	fmt.Printf("result: %v\n", result)
}
