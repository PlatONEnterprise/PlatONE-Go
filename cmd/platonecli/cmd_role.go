package main

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"gopkg.in/urfave/cli.v1"
)

var(
	//role
	RoleCmd = cli.Command{
		Name: "role",
		Usage: "Manage role of accounts",
		Subcommands:  []cli.Command{
			RoleSetSuperAdmin,
			RoleTransferSuperAdmin,
			RoleAddChainAdmin,
			RoleDelChainAdmin,
			RoleAddGroupAdmin,
			RoleDelGroupAdmin,
			RoleAddNodeAdmin,
			RoleDelNodeAdmin,
			RoleAddContractAdmin,
			RoleDelContractAdmin,
			RoleAddContractDeployer,
			RoleDelContractDeployer,
			RoleGetAddrListOfRole,
			RoleHasRole,
			RoleGetRoles,
		},
	}
	RoleSetSuperAdmin = cli.Command{
		Name:      "setSuperAdmin",
		Usage:     "set the caller to superAdmin",
		ArgsUsage: "",
		Action:    setSuperAdmin,
		Flags:     roleCmdFlags,
		Description: `
		platonecli role setSuperAdmin

The caller will be set to superAdmin.`,
	}
	RoleTransferSuperAdmin = cli.Command{
		Name:      "transferSuperAdmin",
		Usage:     "transfer superAdmin to another account",
		ArgsUsage: "<address>",
		Action:    transferSuperAdmin,
		Flags:     roleCmdFlags,
		Description: `
		platonecli role transferSuperAdmin <address>

The caller should be a superAdmin, call this function to thransfer super to another`,
	}
	RoleAddChainAdmin = cli.Command{
		Name: "addChainAdmin",
		Usage: "add the account to ChainAdmin",
		ArgsUsage: "<address>",
		Action: addChainAdmin,
		Flags: roleCmdFlags,
		Description:`
		platonecli role addChainAdmin <address>
The caller should be a superAdmin, call this function to add the account to ChainAdmin`,
	}
	RoleDelChainAdmin = cli.Command{
		Name: "delChainAdmin",
		Usage: "del the account from ChainAdmin",
		ArgsUsage: "<address>",
		Action: delChainAdmin,
		Flags: roleCmdFlags,
		Description:`
		platonecli role addChainAdmin <address>
The caller should be a superAdmin, call this function to del the account from ChainAdmin`,
	}
	RoleAddGroupAdmin= cli.Command{
		Name: "addGroupAdmin",
		Usage: "add the account to GroupAdmin",
		ArgsUsage: "<address>",
		Action: addGroupAdmin,
		Flags: roleCmdFlags,
		Description:`
		platonecli role addGroupAdmin <address>
The caller should be a chainAdmin, call this function to add the account to GroupAdmin`,
	}
	RoleDelGroupAdmin= cli.Command{
		Name: "delGroupAdmin",
		Usage: "del the account from GroupAdmin",
		ArgsUsage: "<address>",
		Action: delGroupAdmin,
		Flags: roleCmdFlags,
		Description:`
		platonecli role delGroupAdmin <address>
The caller should be a chainAdmin, call this function to del the account from GroupAdmin`,
	}
	RoleAddNodeAdmin= cli.Command{
		Name: "addNodeAdmin",
		Usage: "add the account to NodeAdmin",
		ArgsUsage: "<address>",
		Action: addNodeAdmin,
		Flags: roleCmdFlags,
		Description:`
		platonecli role addNodeAdmin <address>
The caller should be a chainAdmin, call this function to add the account to NodeAdmin`,
	}
	RoleDelNodeAdmin= cli.Command{
		Name: "delNodeAdmin",
		Usage: "del the account from NodeAdmin",
		ArgsUsage: "<address>",
		Action: addNodeAdmin,
		Flags: roleCmdFlags,
		Description:`
		platonecli role delNodeAdmin <address>
The caller should be a chainAdmin, call this function to delete the account from NodeAdmin`,
	}
	RoleAddContractAdmin= cli.Command{
		Name: "addContractAdmin",
		Usage: "add the account to ContractAdmin",
		ArgsUsage: "<address>",
		Action: addContractAdmin,
		Flags: roleCmdFlags,
		Description:`
		platonecli role addContractAdmin <address>
The caller should be a chainAdmin, call this function to add the account to ContractAdmin`,
	}
	RoleDelContractAdmin= cli.Command{
		Name: "delContractAdmin",
		Usage: "del the account from ContractAdmin",
		ArgsUsage: "<address>",
		Action: delContractAdmin,
		Flags: roleCmdFlags,
		Description:`
		platonecli role delContractAdmin <address>
The caller should be a chainAdmin, call this function to delete the account from ContractAdmin`,
	}
	RoleAddContractDeployer= cli.Command{
		Name: "addContractDeployer",
		Usage: "add the account to ContractDeployer",
		ArgsUsage: "<address>",
		Action: addContractDeployer,
		Flags: roleCmdFlags,
		Description:`
		platonecli role addContractDeployer <address>
The caller should be a chainAdmin or contractAdmin, call this function to add the account to ContractDeployer`,
	}
	RoleDelContractDeployer= cli.Command{
		Name: "delContractDeployer",
		Usage: "del the account from ContractDeployer",
		ArgsUsage: "<address>",
		Action: delContractDeployer,
		Flags: roleCmdFlags,
		Description:`
		platonecli role delContractDeployer <address>
The caller should be a chainAdmin or contractAdmin, call this function to del the account from ContractDeployer`,
	}
	RoleGetAddrListOfRole = cli.Command{
		Name: "getAddrListOfRole",
		Usage: "get Address List Of the Role",
		ArgsUsage: "<role>",
		Action: getAddrListOfRole,
		Flags: roleCmdFlags,
		Description:`
		platonecli role getAddrListOfRole <role>
role can be "SUPER_ADMIN", "CHAIN_ADMIN", "GROUP_ADMIN", "NODE_ADMIN", "CONTRACT_ADMIN" or "CONTRACT_DEPLOYER"`,
	}
	RoleHasRole = cli.Command{
		Name: "hasRole",
		Usage: "check if the account has the role",
		ArgsUsage: "<address> <role>",
		Action: hasRole,
		Flags: roleCmdFlags,
		Description:`
		platonecli role hasRole <address> <role>
role can be "SUPER_ADMIN", "CHAIN_ADMIN", "GROUP_ADMIN", "NODE_ADMIN", "CONTRACT_ADMIN" or "CONTRACT_DEPLOYER"`,
	}
	RoleGetRoles = cli.Command{
		Name: "getRoles",
		Usage: "get roles of the account",
		ArgsUsage: "<address>",
		Action: getRoles,
		Flags: roleCmdFlags,
		Description:`
		platonecli role delContractDeployer <address>
The caller should be a chainAdmin or contractAdmin, call this function to del the account from ContractDeployer`,
	}
)

func setSuperAdmin(c *cli.Context) {
	callUserManager(c, "setSuperAdmin", nil)
}

func transferSuperAdmin(c *cli.Context) {
	var addr = c.Args().First()
	if !common.IsHexAddress(addr){
		panic("the first argument should be hex address")
	}

	funcParams := []string{addr}
	callUserManager(c, "transferSuperAdmin", funcParams)
}

func addChainAdmin(c *cli.Context) {
	var addr = c.Args().First()
	if !common.IsHexAddress(addr){
		panic("the first argument should be hex address")
	}

	funcParams := []string{addr}
	callUserManager(c, "addChainAdminByAddress", funcParams)
}
func delChainAdmin(c *cli.Context) {
	var addr = c.Args().First()
	if !common.IsHexAddress(addr){
		panic("the first argument should be hex address")
	}

	funcParams := []string{addr}
	callUserManager(c, "delChainAdminByAddress", funcParams)
}

func addGroupAdmin(c *cli.Context) {
	var addr = c.Args().First()
	if !common.IsHexAddress(addr){
		panic("the first argument should be hex address")
	}

	funcParams := []string{addr}
	callUserManager(c, "addGroupAdminByAddress", funcParams)
}

func delGroupAdmin(c *cli.Context) {
	var addr = c.Args().First()
	if !common.IsHexAddress(addr){
		panic("the first argument should be hex address")
	}

	funcParams := []string{addr}
	callUserManager(c, "delGroupAdminByAddress", funcParams)
}

func addNodeAdmin(c *cli.Context) {
	var addr = c.Args().First()
	if !common.IsHexAddress(addr){
		panic("the first argument should be hex address")
	}

	funcParams := []string{addr}
	callUserManager(c, "addNodeAdminByAddress", funcParams)
}

func delNodeAdmin(c *cli.Context) {
	var addr = c.Args().First()
	if !common.IsHexAddress(addr){
		panic("the first argument should be hex address")
	}

	funcParams := []string{addr}
	callUserManager(c, "delNodeAdminByAddress", funcParams)
}

func addContractAdmin(c *cli.Context) {
	var addr = c.Args().First()
	if !common.IsHexAddress(addr){
		panic("the first argument should be hex address")
	}

	funcParams := []string{addr}
	callUserManager(c, "addContractAdminByAddress", funcParams)
}

func delContractAdmin(c *cli.Context) {
	var addr = c.Args().First()
	if !common.IsHexAddress(addr){
		panic("the first argument should be hex address")
	}

	funcParams := []string{addr}
	callUserManager(c, "delContractAdminByAddress", funcParams)
}


func addContractDeployer(c *cli.Context) {
	var addr = c.Args().First()
	if !common.IsHexAddress(addr){
		panic("the first argument should be hex address")
	}

	funcParams := []string{addr}
	callUserManager(c, "addContractDeployerByAddress", funcParams)
}

func delContractDeployer(c *cli.Context) {
	var addr = c.Args().First()
	if !common.IsHexAddress(addr){
		panic("the first argument should be hex address")
	}

	funcParams := []string{addr}
	callUserManager(c, "delContractDeployerByAddress", funcParams)
}

func getAddrListOfRole(c *cli.Context) {
	var role = c.Args().First()
	funcParams := []string{role}

	callUserManager(c, "getAddrListOfRole", funcParams)
}

func getRoles(c *cli.Context) {
	var addr = c.Args().First()
	if !common.IsHexAddress(addr){
		panic("the first argument should be hex address")
	}
	funcParams := []string{addr}

	callUserManager(c, "getRolesByAddress", funcParams)
}

func hasRole(c *cli.Context) {
	var addr = c.Args().First()
	var role = c.Args().Get(1)

	if !common.IsHexAddress(addr){
		panic("the first argument should be hex address")
	}
	funcParams := []string{addr, role}

	callUserManager(c, "hasRole", funcParams)
}

func callUserManager(c *cli.Context, funcName string, funcParams []string) {
	result := contractCommon(c, funcParams, funcName, userManagementAddress)
	fmt.Printf("result: %s\n", result)
}





