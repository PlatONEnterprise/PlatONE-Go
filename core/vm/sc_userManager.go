package vm

import "github.com/PlatONEnetwork/PlatONE-Go/common"

type UserManagement struct {
	Contract *Contract
	Evm      *EVM
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
		"registerRole": u.addContractAdminByAddress,
		//TODO implement
	}
}

