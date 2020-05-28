package main

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/packet"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/ctool/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"gopkg.in/urfave/cli.v1"
	"reflect"
)

var (
	// account
	AccountCmd = cli.Command{
		Name:      "account",
		Usage:     "Manage accounts",
		ArgsUsage: "",
		Category:  "account",
		Description: `
	`,
		Subcommands: []cli.Command{
			TransferCmd,
			RegisterUserCmd,
			RegisterRoleCmd,
			UpdateUserCmd,
			QueryUserCmd,
			StateUserCmd,
		},
	}

	TransferCmd = cli.Command{
		Name:      "transfer",
		Usage:     "Transfer value to another account",
		ArgsUsage: "<to> <value>",
		Action:    transfer,
		Flags:     globalCmdFlags,
		Description: `
		ctool account transfer <to> <value>

The input value can be either hexadecimal ("0xDE0B6B3A7640000") or decimal format ("10") 
The unit conversion table are as follows:
<TODO>"`,
	}

	RegisterUserCmd = cli.Command{
		Name:      "register-user",
		Usage:     "Register a user account to user platform.",
		ArgsUsage: "<account> <name> <tel> <email>",
		Action:    registerUser,
		Flags:     userRegisterCmdFlags,
		Description: `
		ctool account register-user <account> <name> <tel> <email>

The roles could be attached by --roles flag when registering the user account, 
the roles registration request will be approved once the user registration is accepted.`,
	}

	RegisterRoleCmd = cli.Command{
		Name:      "register-roles",
		Usage:     "Register roles for a user account",
		ArgsUsage: "<roles>",
		Action:    registerRole,
		Flags:     globalCmdFlags,
		Description: `
		ctool account register-roles <roles>

The roles are listed below:
chainCreator: the first account of the chain is defaulted to chainCreator. There is only one chainCreator in the chain.
chainAdmin: The chainAdmin has the right to add or delete ...<TODO>
contractAdmin: <TODO>
contractDeployer: The contractDeployer has the right to deploy and destroy the contracts
nodeAdmin: The nodeAdmin has the right to add, delete, and update the node to the nodelist.
`,
	}

	UpdateUserCmd = cli.Command{
		Name:      "update",
		Usage:     "Update the email and mobile info of a user account",
		ArgsUsage: "<account>",
		Action:    updateUser,
		Flags:     userUpdateCmdFlags,
		Description: `
		ctool account update <account>`,
	}

	QueryUserCmd = cli.Command{
		Name:   "query",
		Usage:  "Query the user Info by user name or address",
		Action: queryUser,
		Flags:  userQueryCmdFlags,
		Description: `
		ctool account query`,
	}

	StateUserCmd = cli.Command{
		Name:      "state",
		Usage:     "Trace a user's current registration state by user name or address",
		ArgsUsage: "<account>",
		Action:    stateUser,
		Flags:     globalCmdFlags,
		Description: `
		ctool account state <account>

The tracing has one of the following results:
1. the user application is under approving
2. the user application is rejected
3. the user is invalid
4. the user is valid: the user is a normal user(no role)
5. the user is valid: has Role(s): ...
6. the user is valid: has Role(s): ...
Role(s) in registration: ...`,
	}
)

// transfer value from one account to another
func transfer(c *cli.Context) {
	to := c.Args().First()
	value := c.Args().Get(1)

	value = utl.ChainParamConvert(value, "value").(string)
	toNew := utl.ChainParamConvert(to, "to").(common.Address)

	call := packet.NewContractCallDemo(nil, "", packet.TRANSFER)
	result := messageCall(c, call, &toNew, value, call.TxType)
	fmt.Printf("result: %v\n", result)
}

func registerUser(c *cli.Context) {

	var strMustArray = []string{"account", "name", "tel", "email"} // 必填
	var strJson = "{\"roles\":\"\",\"remark\":\"user platform application\"}"

	str := combineJson(c, strMustArray, []byte(strJson))
	funcParams := CombineFuncParams(str)

	result := contractCommon(c, funcParams, "registerUser", "__sys_UserRegister")
	fmt.Printf("result: %v\n", result)
}

func registerRole(c *cli.Context) {

	roles := c.Args().First()
	utl.ParamValid(roles, "roles")

	funcParams := CombineFuncParams(roles)

	result := contractCommon(c, funcParams, "registerRole", "__sys_RoleRegister")
	fmt.Printf("result: %v\n", result)
}

func updateUser(c *cli.Context) {
	account := c.Args().First()
	utl.ParamValid(account, "address")

	strJson := "{\"mobile\":\"\",\"email\":\"\"}"
	str := combineJson(c, nil, []byte(strJson))

	funcParams := CombineFuncParams(account, str)

	result := contractCommon(c, funcParams, "update", "__sys_UserManager")
	fmt.Printf("result: %v\n", result)
}

func queryUser(c *cli.Context) {
	var funcName string
	var contract string
	var funcParams []string

	user := c.String(UserIDFlags.Name)
	role := c.String(UserRoleFlag.Name)
	status := c.String(UserStatusFlag.Name)

	if len(c.Args()) > 1 {
		utils.Fatalf("please use one search key at a time")
	}

	switch {
	case user != "":
		isAddress := ParamParse(user, "user").(bool)
		if isAddress {
			funcName = "getAccountByAddress"
		} else {
			funcName = "getAccountByName"
		}

		contract = "__sys_UserManager"
		funcParams = []string{user}

	case role != "":
		if !utl.IsRoleMatch(role) {
			utils.Fatalf("invalid input role syntax\n")
		}
		funcName = "getAccountsByRole"
		contract = "__sys_RoleManager"
		funcParams = []string{role}

	case status != "":
		funcName = "getAccountsByStatus"
		contract = "__sys_UserRegister"
		funcParams = []string{"0", "10", status}

	default:
		utils.Fatalf("no search key provided\n")
	}

	result := contractCommon(c, funcParams, funcName, contract)
	utl.PrintJson([]byte(result.(string)))
}

//TODO state by user name
func stateUser(c *cli.Context) {
	account := c.Args().First()
	utl.ParamValid(account, "address")

	funcParams := CombineFuncParams(account)

	// check the user status if the user registration is approved
	result := contractCommon(c, funcParams, "getStatusByAddress", "__sys_UserRegister")
	switch result.(int32) {
	case 1:
		fmt.Printf("the user application is under approving\n")
		return
	case 3:
		fmt.Printf("the user application is rejected\n")
		return
	default:
		result = contractCommon(c, funcParams, "isValidUser", "__sys_UserManager")
	}

	// get the user roles if the user is valid
	if result.(int32) == 0 {
		fmt.Printf("the user is invalid\n")
		return
	} else {
		fmt.Printf("the user is valid: ")
		result = contractCommon(c, funcParams, "getRolesByAddress", "__sys_RoleManager")
	}

	// print the roles owned by the user
	resultBytes := []byte(result.(string))
	result2 := packet.ParseSysContractResult(resultBytes)
	if result2.Code == 0 {
		fmt.Printf("has Roles: %v\n", reflect.ValueOf(result2.Data))
	} else {
		fmt.Printf("the user is a normal user (no roles)\n")
	}

	// get the user roles in registration
	result = contractCommon(c, funcParams, "getRegisterInfoByAddress", "__sys_RoleRegister")
	resultBytes = []byte(result.(string))
	result2 = packet.ParseSysContractResult(resultBytes)
	if result2.Code == 0 {
		roles := result2.Data.(map[string]interface{})["requireRoles"]
		fmt.Printf("Roles in registration: %v\n", roles)
	}
}
