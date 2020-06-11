package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

type UserManagement struct {
	Contract *Contract
	Evm      *EVM
}

type UserInfo struct {
	Address common.Address
}

func (u *UserInfo) encode() ([]byte, error) {
	return rlp.EncodeToBytes(u)
}

//decode
func MakeUserInfo(data []byte) (*UserInfo, error) {
	var ui UserInfo
	if err := rlp.DecodeBytes(data, &ui); nil != err {
		return nil, err
	}

	return &ui, nil
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

// export function
func (u *UserManagement) registerRole(a, b string) ([]byte, error) {
	ui, err := u.doRegisterRole()
	if nil != err {
		return nil, err
	}

	return ui.encode()
}

//internal function
func (u *UserManagement) doRegisterRole() (*UserInfo, error) {

	//TODO implement
	panic("not implemented")
}

//for access control
func (u *UserManagement) AllExportFns() SCExportFns {
	return SCExportFns{
		"registerRole": u.registerRole,
		//TODO implement
	}
}
