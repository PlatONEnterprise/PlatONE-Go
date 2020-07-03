package vm

import (
	"errors"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

var (
	errNotCreator = errors.New("not creator of the contract")
)

type ContractDataProcessor struct {
	state   StateDB
	caller  common.Address
	address common.Address
}

func (d *ContractDataProcessor) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.UserManagementGas
}

// Run runs the precompiled contract
func (d *ContractDataProcessor) Run(input []byte) ([]byte, error) {
	return execSC(input, d.AllExportFns())
}

func (d *ContractDataProcessor) setState(key, value []byte) {
	d.state.SetState(d.address, key, value)
}
func (d *ContractDataProcessor) getState(key []byte) []byte {
	value := d.state.GetState(d.address, key)
	return value
}

func (d *ContractDataProcessor) Caller() common.Address {
	return d.caller
}

//for access control
func (d *ContractDataProcessor) AllExportFns() SCExportFns {
	return SCExportFns{
		"migrate":d.dataMigrate,
	}
}

func (d *ContractDataProcessor) dataMigrate(src common.Address, dest common.Address) (int32, error){
	if d.state.GetContractCreator(src) != d.Caller() {
		return -1, errNotCreator
	}
	if d.state.GetContractCreator(dest) != d.Caller() {
		return -1, errNotCreator
	}
	if err := d.state.CloneAccount(src, dest); err != nil {
		return -1, nil
	}
	return 0, nil
}

// TODO: export all storage k-v data of the contract
func (d *ContractDataProcessor) dataExport(addr common.Address) (int32, error){
	if d.state.GetContractCreator(addr) != d.Caller() {
		return -1, errNotCreator
	}
	return 0, nil
}

// TODO: import all storage k-v data to the contract
func (d *ContractDataProcessor) dataImport(addr common.Address) (int32, error){
	if d.state.GetContractCreator(addr) != d.Caller() {
		return -1, errNotCreator
	}
	return 0, nil
}