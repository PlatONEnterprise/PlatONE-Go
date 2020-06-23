package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

type UserManagement struct {
	Contract *Contract
	Evm      *EVM
	state  StateDB
	caller common.Address
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
	log.Warn("(u *UserManagement) setState", "key", key, "value", value)
	u.state.SetState(u.caller, key, value)
}
func (u *UserManagement) getState(key []byte) []byte {
	value := u.state.GetState(u.caller, key)
	log.Warn("(u *UserManagement) getState", "key", key, "value", value)
	return value
}

func (u *UserManagement) Caller() common.Address{
	return u.caller
}

//for access control
func (u *UserManagement) AllExportFns() SCExportFns {
	return SCExportFns{
		"setSuperAdmin": u.setSuperAdmin,
		"transferSuperAdminByAddress": u.transferSuperAdminByAddress,
		"transferSuperAdminByName": u.transferSuperAdminByName,
		"addChainAdminByAddress":u.addChainAdminByAddress,
		"addChainAdminByName": u.addChainAdminByName,
		"addNodeAdminByAddress": u.addNodeAdminByAddress,
		"addNodeAdminByName": u.addNodeAdminByName,
		"addContractAdminByAddress": u.addContractAdminByAddress,
		"addContractAdminByName": u.addContractAdminByName,
		"addContractDeployerByAddress":u.addContractDeployerByAddress,
		"addContractDeployerByName": u.addContractDeployerByName,

		"getAddrListOfRole":u.getAddrListOfRole,
		"getRolesByAddress": u.getRolesByAddress,

		"addUser":u.addUser,
		"updateUserDescInfo":u.updateUserDescInfo,

		"getUserByAddress": u.getUserByAddress,
		"getUserByName": u.getUserByName,
		"getAllUsers": u.getAllUsers,
	}
}

