package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

type UserManagement struct {
	Contract *Contract
	Evm      *EVM
}

func (u *UserManagement) RequiredGas(input []byte) uint64 {
	if IsEmpty(input) {
		return 0
	}
	return params.UserManagementGas
}

// Run runs the precompiled contract
func (u *UserManagement) Run(input []byte) ([]byte, error) {
	return execSC(input, u.AllExportFns())
}

func (u *UserManagement) setState(key, value []byte) {
	u.Evm.StateDB.SetState(*u.Contract.CodeAddr, key, value)
}
func (u *UserManagement) getState(key []byte) []byte {
	return u.Evm.StateDB.GetState(*u.Contract.CodeAddr, key)
}

func (u *UserManagement) Caller() common.Address{
	return u.Contract.Caller()
}


//for access control
func (u *UserManagement) AllExportFns() SCExportFns {
	return SCExportFns{
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
		"getAllUser": u.getAllUser,
	}
}

