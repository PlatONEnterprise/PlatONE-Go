package vm

import (
	"errors"
	"math/big"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

var (
	errNotCreator = errors.New("not creator of the contract")
)

type ContractDataProcessor struct {
	stateDB      StateDB
	caller       common.Address
	contractAddr common.Address
	blockNumber  *big.Int
}

func (d *ContractDataProcessor) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.UserManagementGas
}

// Run runs the precompiled contract
func (d *ContractDataProcessor) Run(input []byte) ([]byte, error) {
	fnName, ret, err := execSC(input, d.AllExportFns())
	if err != nil {
		if fnName == "" {
			fnName = "Notify"
		}
		d.emitEvent(fnName, operateFail, err.Error())
	}

	return ret, nil
}

func (d *ContractDataProcessor) setState(key, value []byte) {
	d.stateDB.SetState(d.contractAddr, key, value)
}
func (d *ContractDataProcessor) getState(key []byte) []byte {
	value := d.stateDB.GetState(d.contractAddr, key)
	return value
}

func (d *ContractDataProcessor) Caller() common.Address {
	return d.caller
}

func (d *ContractDataProcessor) emitEvent(topic string, code CodeType, msg string) {
	emitEvent(d.contractAddr, d.stateDB, d.blockNumber.Uint64(), topic, code, msg)
}

//for access control
func (d *ContractDataProcessor) AllExportFns() SCExportFns {
	return SCExportFns{
		"migrate": d.dataMigrate,
	}
}

func (d *ContractDataProcessor) dataMigrate(src common.Address, dest common.Address) (int32, error) {
	if d.stateDB.GetContractCreator(src) != d.Caller() {
		return -1, errNotCreator
	}
	if d.stateDB.GetContractCreator(dest) != d.Caller() {
		return -1, errNotCreator
	}
	if err := d.stateDB.CloneAccount(src, dest); err != nil {
		return -1, nil
	}
	return 0, nil
}

// TODO: export all storage k-v data of the contract
func (d *ContractDataProcessor) dataExport(addr common.Address) (int32, error) {
	if d.stateDB.GetContractCreator(addr) != d.Caller() {
		return -1, errNotCreator
	}
	return 0, nil
}

// TODO: import all storage k-v data to the contract
func (d *ContractDataProcessor) dataImport(addr common.Address) (int32, error) {
	if d.stateDB.GetContractCreator(addr) != d.Caller() {
		return -1, errNotCreator
	}
	return 0, nil
}
