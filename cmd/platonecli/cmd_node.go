package main

import (
	"fmt"

	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
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
		ArgsUsage: "<name> <publicKey> <externalIP> <internalIP>",
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

	var strJson = c.Args().First() // todo: add to the usage

	funcParams := []string{strJson}
	result := contractCall(c, funcParams, "add", nodeManagementAddress)
	fmt.Printf("result: %s\n", result)
}

func nodeDelete(c *cli.Context) {

	var str = "{\"status\":2}"

	name := c.Args().First()
	paramValid(name, "name")

	funcParams := CombineFuncParams(name, str)
	result := contractCall(c, funcParams, "update", nodeManagementAddress)
	fmt.Printf("result: %s\n", result)
}

func nodeUpdate(c *cli.Context) {

	// 可选(必填or必填)
	var strJson = "{\"type\":\"\",\"delayNum\":\"\",\"desc\":\"\"}"

	str := combineJson(c, nil, []byte(strJson))

	name := c.Args().First()
	paramValid(name, "name")

	funcParams := CombineFuncParams(name, str)
	result := contractCall(c, funcParams, "update", nodeManagementAddress)
	fmt.Printf("result: %s\n", result)
}

// TODO enode
func nodeQuery(c *cli.Context) {
	var strJson = "{\"type\":\"\",\"status\":\"\",\"name\":\"\",\"publicKey\":\"\"}"

	all := c.Bool(ShowAllFlags.Name)
	if all {
		result := contractCall(c, nil, "getAllNodes", nodeManagementAddress)
		utl.PrintJson([]byte(result.(string)))
		return
	}

	str := combineJson(c, nil, []byte(strJson))
	funcParams := CombineFuncParams(str)

	result := contractCall(c, funcParams, "getNodes", nodeManagementAddress)
	utl.PrintJson([]byte(result.(string)))
}

func nodeStat(c *cli.Context) {
	var strJson = "{\"type\":\"\",\"status\":\"\"}"

	str := combineJson(c, nil, []byte(strJson))
	funcParams := CombineFuncParams(str)

	result := contractCall(c, funcParams, "nodesNum", nodeManagementAddress)
	fmt.Printf("result: %v\n", result)
}
