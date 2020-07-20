package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

type UserManagement struct {
	state    StateDB
	caller   common.Address
	address common.Address
}

func (u *UserManagement) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.UserManagementGas
}

// Run runs the precompiled contract
func (u *UserManagement) Run(input []byte) ([]byte, error) {
	return execSC(input, u.AllExportFns())
}

func (u *UserManagement) setState(key, value []byte) {
	u.state.SetState(u.address, key, value)
}
func (u *UserManagement) getState(key []byte) []byte {
	value := u.state.GetState(u.address, key)
	return value
}

func (u *UserManagement) Caller() common.Address {
	return u.caller
}

//for access control
func (u *UserManagement) AllExportFns() SCExportFns {
	return SCExportFns{
		"setSuperAdmin":                u.setSuperAdmin,
		"transferSuperAdminByAddress":  u.transferSuperAdminByAddress,
		"transferSuperAdminByName":     u.transferSuperAdminByName,
		"addChainAdminByAddress":       u.addChainAdminByAddress,
		"addChainAdminByName":          u.addChainAdminByName,
		"addGroupAdminByAddress":       u.addGroupAdminByAddress,
		"addGroupAdminByName":          u.addGroupAdminByName,
		"addNodeAdminByAddress":        u.addNodeAdminByAddress,
		"addNodeAdminByName":           u.addNodeAdminByName,
		"addContractAdminByAddress":    u.addContractAdminByAddress,
		"addContractAdminByName":       u.addContractAdminByName,
		"addContractDeployerByAddress": u.addContractDeployerByAddress,
		"addContractDeployerByName":    u.addContractDeployerByName,

		"delChainAdminByAddress":       u.delChainAdminByAddress,
		"delChainAdminByName":          u.delChainAdminByName,
		"delGroupAdminByAddress":       u.delGroupAdminByAddress,
		"delGroupAdminByName":          u.delGroupAdminByName,
		"delNodeAdminByAddress":        u.delNodeAdminByAddress,
		"delNodeAdminByName":           u.delNodeAdminByName,
		"delContractAdminByAddress":    u.delContractAdminByAddress,
		"delContractAdminByName":       u.delContractAdminByName,
		"delContractDeployerByAddress": u.delContractDeployerByAddress,
		"delContractDeployerByName":    u.delContractDeployerByName,

		"getAddrListOfRole": u.getAddrListOfRoleStr,
		"getRolesByAddress": u.getRolesByAddress,
		"getRolesByName": u.getRolesByName,
		"hasRole": u.hasRole,

		"addUser":            u.addUser,
		"updateUserDescInfo": u.updateUserDescInfo,

		"getUserByAddress": u.getUserByAddress,
		"getUserByName":    u.getUserByName,
		"getAllUsers":      u.getAllUsers,
	}
}
