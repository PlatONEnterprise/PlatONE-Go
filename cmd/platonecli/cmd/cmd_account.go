package cmd

import (
	"encoding/json"
	"fmt"

	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platoneclient/precompiled"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"gopkg.in/urfave/cli.v1"
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
			UserAdd,
			UserUpdate,
			QueryUserCmd,
		},
	}

	TransferCmd = cli.Command{
		Name:      "transfer",
		Usage:     "Transfer value to another account",
		ArgsUsage: "<to> <value>",
		Action:    transfer,
		Flags:     globalCmdFlags,
		Description: `
		platonecli account transfer <to> <value>

The input value can be either hexadecimal ("0xDE0B6B3A7640000") or decimal format ("10") 
The unit conversion table are as follows:
<TODO>"`,
	}

	/*
		AdminUserCmd = cli.Command{
			Name:  "user",
			Usage: "Manage user accounts registered in the user platform",

			Subcommands: []cli.Command{
				UserAdd,
				UserUpdate,
			},
		}*/

	UserAdd = cli.Command{
		Name:      "add",
		Usage:     "Add a user to the user platform",
		ArgsUsage: "<address> <name> <tel> <email>",
		Action:    userAdd,
		Flags:     globalCmdFlags,
		Description: `
		platonecli account add <address> <name> <tel> <email>`,
	}

	UserUpdate = cli.Command{
		Name:      "update",
		Usage:     "Update the mobile and email info of a user",
		ArgsUsage: "<address>",
		Action:    userUpdate,
		Flags:     userUpdateCmdFlags,
		Description: `
		platonecli account update <address>`,
	}

	QueryUserCmd = cli.Command{
		Name:   "query",
		Usage:  "Query the user Info by user name or address",
		Action: queryUser,
		Flags:  userQueryCmdFlags,
		Description: `
		platonecli account query`,
	}
)

// todo: need testing
// transfer value from one account to another

func transfer(c *cli.Context) {
	to := c.Args().First()
	value := c.Args().Get(1)

	value = cmd_common.ChainParamConvert(value, "value").(string)
	toNew := cmd_common.ChainParamConvert(to, "to").(common.Address)

	result := clientCommonV2(c, nil, &toNew)
	fmt.Printf("result: %v\n", result[0])
}

type UserInfo struct {
	Address string `json:"address,omitempty,required"`
	Name string `json:"name,omitempty"`
}
func userAdd(c *cli.Context) {
	//var strMustArray = []string{"address", "name"}
	//strJson := combineJson(c, strMustArray, nil)
	//var strJson = c.Args().First()

	var userinfo UserInfo
	address := c.Args().First()
	userinfo.Address = address
	name := c.Args().Get(1)
	userinfo.Name = name
	bytes, _ := json.Marshal(userinfo)
	strJson := string(bytes)
	funcParams := []string{strJson}
	result := contractCall(c, funcParams, "addUser", precompile.UserManagementAddress)
	fmt.Printf("%s\n", result)
}

func userUpdate(c *cli.Context) {
	account := c.Args().First()
	paramValid(account, "address")

	strJson := "{\"mobile\":\"\",\"email\":\"\"}"
	str := combineJson(c, nil, []byte(strJson))

	funcParams := cmd_common.CombineFuncParams(account, str)

	result := contractCall(c, funcParams, "updateUserDescInfo", precompile.UserManagementAddress)
	fmt.Printf("%v\n", result)
}

func queryUser(c *cli.Context) {
	var funcName string
	var funcParams = make([]string, 0)

	user := c.String(UserIDFlags.Name)
	all := c.Bool(ShowAllFlags.Name)

	if len(c.Args()) > 1 {
		utils.Fatalf("please use one search key at a time")
	}

	switch {
	case user != "":
		isAddress := cmd_common.ParamParse(user, "user").(int32)
		if isAddress == cmd_common.CnsIsAddress {
			funcName = "getUserByAddress"
		} else {
			funcName = "getUserByName"
		}

		funcParams = []string{user}
	case all:
		funcName = "getAllUsers"

	default:
		utils.Fatalf("no search key provided\n")
	}

	result := contractCall(c, funcParams, funcName, precompile.UserManagementAddress)
	strResult := PrintJson([]byte(result.(string)))
	fmt.Printf("result:\n%s\n", strResult)
}
