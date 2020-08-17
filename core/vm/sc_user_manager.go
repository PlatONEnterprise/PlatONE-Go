package vm

import (
	"math/big"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

type UserManagement struct {
	stateDB      StateDB
	caller       common.Address
	contractAddr common.Address
	blockNumber  *big.Int
}

func (u *UserManagement) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.UserManagementGas
}

// Run runs the precompiled contract
func (u *UserManagement) Run(input []byte) ([]byte, error) {
	fnName, ret, err := execSC(input, u.AllExportFns())
	if err != nil {
		if fnName == "" {
			fnName = "Notify"
		}
		u.emitEvent(fnName, operateFail, err.Error())
	}
	return ret, nil
}

func (u *UserManagement) setState(key, value []byte) {
	u.stateDB.SetState(u.contractAddr, key, value)
}
func (u *UserManagement) getState(key []byte) []byte {
	value := u.stateDB.GetState(u.contractAddr, key)
	return value
}

func (u *UserManagement) Caller() common.Address {
	return u.caller
}

func (u *UserManagement) returnSuccess(topic string) (int32, error) {
	u.emitEvent(topic, operateSuccess, "Success")
	return int32(operateSuccess), nil
}

func (u *UserManagement) returnFail(topic string, err error) (int32, error) {
	u.emitEvent(topic, operateFail, err.Error())
	returnErr := err
	// todo: in some cases, returnErr = nil
	return int32(operateFail), returnErr
}

func (u *UserManagement) emitEvent(topic string, code CodeType, msg string) {
	emitEvent(u.contractAddr, u.stateDB, u.blockNumber.Uint64(), topic, code, msg)
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
		"getRolesByName":    u.getRolesByName,
		"hasRole":           u.hasRole,

		"addUser":            u.addUser,
		"updateUserDescInfo": u.updateUserDescInfo,

		"getUserByAddress": u.getUserByAddress,
		"getUserByName":    u.getUserByName,
		"getAllUsers":      u.getAllUsers,
	}
}
