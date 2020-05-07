package main

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/packet"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
	"strings"
)

var (
	// admin
	AdminCmd = cli.Command{
		Name:  "admin",
		Usage: "Manage administrator rights",

		Subcommands: []cli.Command{
			NodeCmd,
			AdminContractCmd,
			AdminUserCmd,
			AdminSupCmd,
			SysConfigCmd,
		},
	}

	//-----------------------------------------------
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
		ctool admin node add <name> <publicKey> <externalIP> <internalIP>

The newly added nodes can only be observer type.`,
	}

	NodeDeleteCmd = cli.Command{
		Name:      "delete",
		Usage:     "Delete a node from the node list, the deleted node can no longer receiving and synchronizing blocks",
		ArgsUsage: "<name>",
		Action:    nodeDelete,
		Flags:     globalCmdFlags,
		Description: `
		ctool admin node delete <name>`,
	}

	NodeUpdateCmd = cli.Command{
		Name:      "update",
		Usage:     "Update the description, delay number, and node type of a node",
		ArgsUsage: "<name>",
		Action:    nodeUpdate,
		Flags:     nodeUpdateCmdFlags,
		Description: `
		ctool admin node update <name>`,
	}

	NodeQueryCmd = cli.Command{
		Name:  "query",
		Usage: "Query the node Info by the search key provided",
		Action: nodeQuery,
		Flags:  nodeQueryCmdFlasg,
		Description: `
		ctool admin node query

Except --all flag, other search keys can be combined.`,
	}

	NodeStatCmd = cli.Command{
		Name:   "stat",
		Usage:  "Statistic the node Info by the search key provided",
		Action: nodeStat,
		Flags:  nodeStatCmdFlags,
		Description: `
		ctool admin node stat`,
	}

	//--------------------------------------------------
	AdminContractCmd = cli.Command{
		Name:  "contract",
		Usage: "Manage contract deployers <TODO: deprecated?>",
		Subcommands: []cli.Command{
			ContractApprove,
			ContractAdd,
			ContractDelete,
			ContractList,
		},
	}

	ContractApprove = cli.Command{
		Name:      "approve",
		Usage:     "show a refactoring contractDeployer approve demo",
		ArgsUsage: "<address> <operation>",
		Action:    contractApprove,
		Flags:     globalCmdFlags,
		Description: `
		ctool admin contract approve <address> <operation>`,
	}

	ContractAdd = cli.Command{
		Name:      "add",
		Usage:     "show a refactoring contractDeployer add demo",
		ArgsUsage: "<name> <address>",
		Action:    contractAdd,
		Flags:     globalCmdFlags,
		Description: `
		ctool admin contract add <name> <address>`,
	}

	ContractDelete = cli.Command{
		Name:      "delete",
		Usage:     "show a refactoring contractDeployer delete demo",
		ArgsUsage: "<address> <roles>",
		Action:    contractDelete,
		Flags:     globalCmdFlags,
		Description: `
		ctool admin contract delete <address> <roles>`,
	}

	ContractList = cli.Command{
		Name:   "list",
		Usage:  "show a refactoring contractDeployer list demo",
		Action: contractList,
		Flags:  contractAdminCmdFlags,
		Description: `
		ctool admin contract list`,
	}

	//-------------------------------------------------
	AdminUserCmd = cli.Command{
		Name:  "user",
		Usage: "Manager user accounts registered in the user platform",

		Subcommands: []cli.Command{
			UserApprove,
			UserAdd,
			UserDelete,
			UserEnable,
			UserDisable,
			UserUpdate,
			UserList,
		},
	}

	UserApprove = cli.Command{
		Name:      "approve",
		Usage:     "Approve the user registration",
		ArgsUsage: "<address> <operation>",
		Action:    userApprove,
		Flags:     globalCmdFlags,
		Description: `
		ctool admin user approve <address> <operation>`,
	}

	UserAdd = cli.Command{
		Name:      "add",
		Usage:     "Add a user to the user platform",
		ArgsUsage: "<address> <name> <tel> <email>",
		Action:    userAdd,
		Flags:     globalCmdFlags,
		Description: `
		ctool admin user add <address> <name> <tel> <email>`,
	}

	UserDelete = cli.Command{
		Name:      "delete",
		Usage:     "Delete a user from user platform. The user becomes invalid and ",
		ArgsUsage: "<address>",
		Action:    userDelete,
		Flags:     globalCmdFlags,
		Description: `
		ctool admin user delete <address>`,
	}

	UserEnable = cli.Command{
		Name:      "enable",
		Usage:     "Enable a user from disable",
		ArgsUsage: "<address>",
		Action:    userEnable,
		Flags:     globalCmdFlags,
		Description: `
		ctool admin user enable <address>`,
	}

	UserDisable = cli.Command{
		Name:      "disable",
		Usage:     "Disable a user. The user becomes invalid. The disabled user can be enabled by the enable command",
		ArgsUsage: "<address>",
		Action:    userDisable,
		Flags:     globalCmdFlags,
		Description: `
		ctool admin user disable <address>`,
	}

	UserUpdate = cli.Command{
		Name:      "update",
		Usage:     "Update the mobile and email info of a user",
		ArgsUsage: "<address>",
		Action:    userUpdate,
		Flags:     userUpdateCmdFlags,
		Description: `
		ctool admin user update <address>`,
	}

	UserList = cli.Command{
		Name:  "list",
		Usage: "<TODO>",
		Action: userList,
		Flags:  userListCmdFlags,
		Description: `
		ctool admin user list`,
	}

	//------------------------------------------------
	AdminSupCmd = cli.Command{
		Name:  "sup",
		Usage: "<TODO>",

		Subcommands: []cli.Command{
			SupApprove,
			SupAdd,
			SupDelete,
			SupList,
		},
	}

	SupApprove = cli.Command{
		Name:      "approve",
		Usage:     "Approve the role registration of a user",
		ArgsUsage: "<address> <operation>",
		Action:    supApprove,
		Flags:     globalCmdFlags,
		Description: `
		ctool admin sup approve <address> <operation>`,
	}

	SupAdd = cli.Command{
		Name:      "add",
		Usage:     "Add a role to a user",
		ArgsUsage: "<name> <address> <roles>",
		Action:    supAdd,
		Flags:     globalCmdFlags,
		Description: `
		ctool admin sup add <name> <address> <roles>
The format of the <roles> is '["<role1>, <role2>, ..."]'`,
	}

	SupDelete = cli.Command{
		Name:      "delete",
		Usage:     "Delete a role of a user",
		ArgsUsage: "<address> <roles>",
		Action:    supDelete,
		Flags:     globalCmdFlags,
		Description: `
		ctool admin sup delete <address> <roles>`,
	}

	SupList = cli.Command{
		Name:   "list",
		Usage:  "<TODO>",
		Action: supList,
		Flags:  supAdminCmdFlags,
		Description: `
		ctool admin sup list`,
	}

	//-----------------------------------------
	SysConfigCmd = cli.Command{
		Name:  "sysconfig",
		Usage: "<TODO: currently unavailable>",

		Subcommands: []cli.Command{
			ConsensusConfig,
			AccessControlConfig,
			GasFeeConfig,
		},
	}

	ConsensusConfig = cli.Command{
		Name:  "consensus",
		Usage: "<TODO: currently unavailable>",
		//ArgsUsage: "",
		//Action: ,
		Flags: globalCmdFlags,
	}

	AccessControlConfig = cli.Command{
		Name:  "add",
		Usage: "<TODO: currently unavailable>",
		//ArgsUsage: "",
		//Action: ,
		Flags: globalCmdFlags,
	}

	GasFeeConfig = cli.Command{
		Name:  "gasfee",
		Usage: "<TODO: currently unavailable>",
		//ArgsUsage: "",
		//Action: ,
		Flags: globalCmdFlags,
	}
)

func nodeAdd(c *cli.Context) {

	// required value
	var strMustArray = []string{"name", "publicKey", "externalIP", "internalIP"}

	// default or user input value
	var strConst = "\"owner\":\"todo\",\"status\":1,\"type\":0,"
	var strDefault = "\"rpcPort\":6791,\"p2pPort\":1800,\"desc\":\"add node to the list\","
	var strOption = "\"delayNum\":\"\""

	var strJson = fmt.Sprintf("{%s%s%s}", strConst, strDefault, strOption)

	// combine to json format
	str := combineJson(c, strMustArray, []byte(strJson))

	funcParams := []string{str}
	result := contractCommon(c, funcParams, "add", "__sys_NodeManager")
	fmt.Printf("result: %s\n", result)
}

func nodeDelete(c *cli.Context) {

	var str = "{\"status\":2}"

	name := c.Args().First()
	utl.ParamValid(name, "name")

	funcParams := CombineFuncParams(name, str)
	result := contractCommon(c, funcParams, "update", "__sys_NodeManager")
	fmt.Printf("result: %s\n", result)
}

func nodeUpdate(c *cli.Context) {

	// 可选(必填or必填)
	var strJson = "{\"type\":\"\",\"delayNum\":\"\",\"desc\":\"\"}"

	str := combineJson(c, nil, []byte(strJson))

	name := c.Args().First()
	utl.ParamValid(name, "name")

	funcParams := CombineFuncParams(name, str)
	result := contractCommon(c, funcParams, "update", "__sys_NodeManager")
	fmt.Printf("result: %s\n", result)
}

// TODO enode
func nodeQuery(c *cli.Context) {
	var strJson = "{\"type\":\"\",\"status\":\"\",\"name\":\"\",\"publicKey\":\"\"}"

	all := c.Bool("all")
	if all {
		result := contractCommon(c, nil, "getAllNodes", "__sys_NodeManager")
		utl.PrintJson([]byte(result.(string)))
		return
	}

	str := combineJson(c, nil, []byte(strJson))
	funcParams := CombineFuncParams(str)

	result := contractCommon(c, funcParams, "getNodes", "__sys_NodeManager")
	utl.PrintJson([]byte(result.(string)))
}

func nodeStat(c *cli.Context) {
	var strJson = "{\"type\":\"\",\"status\":\"\"}"

	str := combineJson(c, nil, []byte(strJson))
	funcParams := CombineFuncParams(str)

	result := contractCommon(c, funcParams, "nodesNum", "__sys_NodeManager")
	fmt.Printf("result: %v\n", result)
}

//------------------------------------------------------------------------------------
func contractApprove(c *cli.Context) {
	supApprove(c)
}

func contractAdd(c *cli.Context) {
	name := c.Args().First()
	account := c.Args().Get(1)

	utl.ParamValid(name, "name")
	utl.ParamValid(account, "address")

	funcParams := CombineFuncParams(name, account, "[\"contractDeployer\"]")

	result := contractCommon(c, funcParams, "addRole", "__sys_RoleManager")
	fmt.Printf("result: %s\n", result)
}

func contractDelete(c *cli.Context) {
	account := c.Args().First()
	utl.ParamValid(account, "address")

	funcParams := CombineFuncParams(account, "[\"contractDeployer\"]")

	result := contractCommon(c, funcParams, "removeRole", "__sys_RoleManager")
	fmt.Printf("result: %s\n", result)
}

// future feature
// func contractAudit(c *cli.Context){}

//TODO refactory
func contractList(c *cli.Context) {
	var result interface{}

	del := c.Bool("delete")
	approve := c.Bool("approve")

	if del && approve {
		utils.Fatalf("please select one operation at one time")
	}

	switch {
	case approve:
		funcParams := []string{
			"1",  // status
			"0",  // pageNum
			"10", // pageSize
		}

		result = contractCommon(c, funcParams, "getRegisterInfosByStatus", "__sys_RoleRegister")
		result = packet.ExtractContractData(result.(string), "contractDeployer")
	case del:
		funcParams := []string{"contractDeployer"}
		result = contractCommon(c, funcParams, "getAccountsByRole", "__sys_RoleManager")
	default:
		panic(fmt.Sprintf(utl.PanicUnexpSituation, "contractList"))
	}

	utl.PrintJson([]byte(result.(string)))
}

//---------------------------------------------------------------------------
func userApprove(c *cli.Context) {
	account := c.Args().First()
	statusString := c.Args().Get(1)

	utl.ParamValid(account, "address")
	status := ParamParse(statusString, "operation").(string)

	funcParams := CombineFuncParams(account, status)
	result := contractCommon(c, funcParams, "approve", "__sys_UserRegister")
	fmt.Printf("result: %s\n", result)
}

func userAdd(c *cli.Context) {
	var strMustArray = []string{"address", "name", "mobile", "email"}

	str := combineJson(c, strMustArray, nil)

	funcParams := []string{str}
	result := contractCommon(c, funcParams, "addUser", "__sys_UserManager")
	fmt.Printf("result: %s\n", result)
}

func userDelete(c *cli.Context) {
	account := c.Args().First()
	utl.ParamValid(account, "address")

	funcParams := []string{account}
	result := contractCommon(c, funcParams, "delUser", "__sys_UserManager")
	fmt.Printf("result: %s\n", result)
}

func userEnable(c *cli.Context) {
	account := c.Args().First()
	utl.ParamValid(account, "address")

	funcParams := []string{account}
	result := contractCommon(c, funcParams, "enable", "__sys_UserManager")
	fmt.Printf("result: %s\n", result)
}

func userDisable(c *cli.Context) {
	account := c.Args().First()
	utl.ParamValid(account, "address")

	funcParams := []string{account}
	result := contractCommon(c, funcParams, "disable", "__sys_UserManager")
	fmt.Printf("result: %s\n", result)
}

func userUpdate(c *cli.Context) {
	updateUser(c)
}

//TODO ?
func userList(c *cli.Context) {
	var funcParams []string
	var funcName string
	var contract string

	// del enable disable
	// _ = contractCommon(c, []string{}, "getAccountBystatus", "__sys_UserRegister")

	action := c.Bool("approve")
	switch {
	case action:
		status := "1"
		funcParams = CombineFuncParams("0", "10", status)
		funcName = "getAccountsByStatus"
		contract = "__sys_UserRegister"
	default:
		utils.Fatalf("no flag provided\n")
	}

	result := contractCommon(c, funcParams, funcName, contract)
	utl.PrintJson([]byte(result.(string)))
}

//------------------------------------------------------------------------------
func supApprove(c *cli.Context) {
	account := c.Args().First()
	statusString := c.Args().Get(1)

	utl.ParamValid(account, "address")
	status := ParamParse(statusString, "operation").(string)

	funcParams := CombineFuncParams(account, status)
	result := contractCommon(c, funcParams, "approveRole", "__sys_RoleRegister")
	fmt.Printf("result: %s\n", result)
}

func supAdd(c *cli.Context) {
	name := c.Args().First()
	account := c.Args().Get(1)
	roles := c.Args().Get(2)

	roles = utl.TrimSpace(roles)
	utl.ParamValid(name, "name")
	utl.ParamValid(account, "address")
	utl.ParamValid(roles, "roles")

	funcParams := CombineFuncParams(name, account, roles)
	result := contractCommon(c, funcParams, "addRole", "__sys_RoleManager")
	fmt.Printf("result: %s\n", result)
}

func supDelete(c *cli.Context) {
	account := c.Args().First()
	roles := c.Args().Get(1)

	roles = utl.TrimSpace(roles)
	utl.ParamValid(account, "address")
	utl.ParamValid(roles, "roles")

	funcParams := CombineFuncParams(account, roles)

	result := contractCommon(c, funcParams, "removeRole", "__sys_RoleManager")
	fmt.Printf("result: %s\n", result)
}

func supList(c *cli.Context) {
	var result interface{}
	role := c.String("delete")
	approve := c.Bool("approve")

	if role != "" && approve {
		utils.Fatalf("please select one operation at one time")
	}

	switch {
	case approve:
		funcParams := []string{
			"1",  // status
			"0",  // pageNum
			"10", // pageSize
		}

		result = contractCommon(c, funcParams, "getRegisterInfosByStatus", "__sys_RoleRegister")
	case role != "":
		role = strings.TrimSpace(role)
		if !utl.IsRoleMatch(role){
			utils.Fatalf("the input role is invalid\n")
		}

		funcParams := []string{role}
		result = contractCommon(c, funcParams, "getAccountsByRole", "__sys_RoleManager")
	}

	utl.PrintJson([]byte(result.(string)))
}

//----------------------------------------------
//TODO system parameter contract
//func sysconfigCon(c *cli.Context) {}
