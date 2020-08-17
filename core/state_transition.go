// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"errors"
	"math"
	"math/big"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"
	"github.com/PlatONEnetwork/PlatONE-Go/life/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

var (
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
)

var zeroAddress common.Address
var PermissionErr = errors.New("Permission Denied!")

/*
A state transition is a change made when a transaction is applied to the current world state
The state transitioning model does all the necessary work to work out a valid new state root.
1) Nonce handling
2) Pre pay gas
3) Create a new state object if the recipient is \0*32
4) Value transfer
== If contract creation ==
  4a) Attempt to run transaction data
  4b) If valid, use result as code for the new state object
== end ==
5) Run Script section
6) Derive new state root
*/
type StateTransition struct {
	gp         *GasPool
	msg        Message
	gas        uint64
	gasPrice   *big.Int
	initialGas uint64
	value      *big.Int
	data       []byte
	state      vm.StateDB
	evm        *vm.EVM
}

// Message represents a message sent to a contract.
type Message interface {
	From() common.Address
	//FromFrontier() (common.Address, error)
	To() *common.Address
	SetTo(common.Address)
	SetData([]byte)
	SetNonce(uint64)

	GasPrice() *big.Int
	Gas() uint64
	Value() *big.Int

	Nonce() uint64
	CheckNonce() bool
	Data() []byte
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(data []byte, contractCreation bool) (uint64, error) {
	// Set the starting gas for the raw transaction
	var gas uint64
	if contractCreation {
		gas = params.TxGasContractCreation
	} else {
		gas = params.TxGas
	}
	// Bump the required gas by the amount of transactional data
	if len(data) > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		// Make sure we don't exceed uint64 for all data combinations
		if (math.MaxUint64-gas)/params.TxDataNonZeroGas < nz {
			return 0, vm.ErrOutOfGas
		}
		gas += nz * params.TxDataNonZeroGas

		z := uint64(len(data)) - nz
		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, vm.ErrOutOfGas
		}
		gas += z * params.TxDataZeroGas
	}
	return gas, nil
}

// NewStateTransition initialises and returns a new state transition object.
func NewStateTransition(evm *vm.EVM, msg Message, gp *GasPool) *StateTransition {
	return &StateTransition{
		gp:       gp,
		evm:      evm,
		msg:      msg,
		gasPrice: msg.GasPrice(),
		value:    msg.Value(),
		data:     msg.Data(),
		state:    evm.StateDB,
	}
}

// ApplyMessage computes the new state by applying the given message
// against the old state within the environment.
//
// ApplyMessage returns the bytes returned by any EVM execution (if it took place),
// the gas used (which includes gas refunds) and an error if it failed. An error always
// indicates a core error meaning that the message would always fail for that particular
// state and would never be accepted within a block.
func ApplyMessage(evm *vm.EVM, msg Message, gp *GasPool) ([]byte, uint64, int64, bool, error) {
	return NewStateTransition(evm, msg, gp).TransitionDb()
}

// to returns the recipient of the message.
func (st *StateTransition) to() common.Address {
	if st.msg == nil || st.msg.To() == nil /* contract creation */ {
		return common.Address{}
	}
	return *st.msg.To()
}

func (st *StateTransition) useGas(amount uint64) error {
	if st.gas < amount {
		return vm.ErrOutOfGas
	}
	st.gas -= amount

	return nil
}

func (st *StateTransition) buyGas() error {
	gas := uint64(common.SysCfg.GetTxGasLimit())
	if err := st.gp.SubGas(gas); err != nil {
		return err
	}
	st.gas += gas

	st.initialGas = gas
	return nil
}

func (st *StateTransition) buyContractGas(contractAddr common.Address) error {
	addr := st.msg.From().String()
	addr = strings.ToLower(addr)
	params := []interface{}{addr, st.msg.Gas()}

	ret, _, err := st.doCallContract(contractAddr, "checkBalance", params)
	if nil != err {
		log.Error("buyContractGas error", "err", err.Error())
		return err
	}
	if utils.BytesToInt64(ret) != 1 {
		log.Error("insufficientBalance error", "err", errInsufficientBalanceForGas)
		return errInsufficientBalanceForGas
	}

	if err := st.gp.SubGas(st.msg.Gas()); err != nil {
		log.Error("gas pool sub gas error", "err", err.Error())
		return err
	}

	st.gas += st.msg.Gas()

	st.initialGas = st.msg.Gas()
	usergas := st.msg.Gas()

	params = []interface{}{addr, usergas}

	_, _, err = st.doCallContract(contractAddr, "withHoldingFee", params)
	if nil != err {
		log.Error("withHoldingContractGas error", "err", err.Error())
		return err
	}
	return nil
}

func (st *StateTransition) preCheck() error {
	//set gasPrice = 0, for not sub Txfee and not add coinbase
	st.gasPrice = new(big.Int).SetInt64(0)

	return st.buyGas()
}

func (st *StateTransition) preContractGasCheck(contractAddr common.Address) error {
	return st.buyContractGas(contractAddr)
}
func (st *StateTransition) getGasPrice(contractAddr common.Address) (int64, error) {
	params := []interface{}{}
	ret, _, err := st.doCallContract(contractAddr, "getGasPrice", params)
	if nil != err {
		log.Error("get gas price error", "err", err.Error())
		return -1, err
	}
	gasPrice := utils.BytesToInt64(ret)

	return gasPrice, nil
}

func addressCompare(addr1, addr2 common.Address) bool {
	return strings.ToLower(addr1.String()) == strings.ToLower(addr2.String())
}

// 合约防火墙的检查：
//  1. 如果账户结构体code字段为空，pass
//  2. 如果账户data字段为空，pass
// 	3. 黑名单优先于白名单，后续只有不在黑名单列表，同时在白名单列表里的账户才能pass
func fwCheck(stateDb vm.StateDB, contractAddr common.Address, caller common.Address, input []byte) ([]byte, bool) {
	if stateDb.IsFwOpened(contractAddr) == false {
		return nil, true
	}

	// 如果账户结构体code字段为空或tx.data为空，pass
	if len(stateDb.GetCode(contractAddr)) == 0 || len(input) == 0 {
		return nil, true
	}

	var data [][]byte
	if err := rlp.DecodeBytes(input, &data); err != nil {
		log.Debug("FW : Input decode error")
		return vm.MakeReturnBytes([]byte("FW : Input decode error")), false
	}
	if len(data) < 2 {
		log.Debug("FW : Missing function name")
		return vm.MakeReturnBytes([]byte("FW : Missing function name")), false
	}
	funcName := string(data[1])

	if stateDb.GetContractCreator(contractAddr) == caller {
		return nil, true
	}

	fwStatus := stateDb.GetFwStatus(contractAddr)

	var fwLog string = "FW : Access to contract:" + contractAddr.String() + " by " + funcName + "is refused by firewall."

	if fwStatus.IsRejected(funcName, caller) {
		return vm.MakeReturnBytes([]byte(fwLog)), false
	}

	if fwStatus.IsAccepted(funcName, caller) {
		return nil, true
	}

	return vm.MakeReturnBytes([]byte(fwLog)), false
}

func (st *StateTransition) ifUseContractTokenAsFee() (common.Address, bool) {
	isUseContractToken := common.SysCfg.GetIsTxUseGas()
	contractAddr := common.SysCfg.GetGasContractAddress()
	if contractAddr == zeroAddress {
		return contractAddr, false
	}
	return contractAddr, isUseContractToken
}

// TransitionDb will transition the state by applying the current message and
// returning the result including the used gas. It returns an error if failed.
// An error indicates a consensus issue.
func (st *StateTransition) TransitionDb() (ret []byte, usedGas uint64, gasPrice int64, failed bool, err error) {
	var (
		evm = st.evm
		// vm errors do not effect consensus and are therefor
		// not assigned to err, except for insufficient balance
		// error.
		vmerr error
		//conAddr common.Address
		msg    = st.msg
		sender = vm.AccountRef(msg.From())
	)
	isCallSysParam := isCallParamManager(msg.To())
	feeContractAddr, isUseContractToken := st.ifUseContractTokenAsFee()
	isUseContractToken = isUseContractToken && msg.Nonce() != 0 && !isCallSysParam
	if isUseContractToken {
		if err = st.preContractGasCheck(feeContractAddr); err != nil {
			log.Error("PreContractGasCheck", "err:", err)
			return
		}
		if gasPrice, err = st.getGasPrice(feeContractAddr); err != nil {
			log.Error("getGasPrice from feeContractAddr", "err:", err)
			return
		}
	} else {
		if err = st.preCheck(); err != nil {
			return
		}
	}

	contractCreation := msg.To() == nil

	// Pay intrinsic gas
	gas, err := IntrinsicGas(st.data, contractCreation)
	log.Debug("IntrinsicGas amount", "IntrinsicGas:", gas)
	if err != nil {
		return nil, 0, gasPrice, false, err
	}
	if err = st.useGas(gas); err != nil {
		log.Error("GasLimitTooLow", "err:", err)
		return nil, 0, gasPrice, false, err
	}

	if contractCreation {
		allowDeployContract := checkContractDeployPermission(sender.Address(), evm)
		if !allowDeployContract {
			st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
			return nil, 0, gasPrice, true, PermissionErr
		}
		ret, _, st.gas, vmerr = evm.Create(sender, st.data, st.gas, st.value)
	} else {
		st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
		var pass bool
		if ret, pass = fwCheck(evm.StateDB, st.to(), msg.From(), msg.Data()); !pass {
			err = PermissionErr
			vmerr = PermissionErr
			log.Debug("Calling contract was refused by firewall", "err", vmerr)
		} else {
			// Increment the nonce for the next transaction
			// If the transaction is cns-type, do not increment the nonce
			ret, st.gas, vmerr = evm.Call(sender, st.to(), st.data, st.gas, st.value)
		}
	}
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, gasPrice, false, vmerr
		}
	}
	if isUseContractToken {
		err = st.refundContractGas(feeContractAddr)
	} else {
		st.refundGas()
	}
	return ret, st.gasUsed(), gasPrice, vmerr != nil, err
}

func (st *StateTransition) doCallContract(address common.Address, funcName string, funcParams []interface{}) (ret []byte, leftOverGas uint64, err error) {
	evm := st.evm
	msg := st.msg
	caller := vm.AccountRef(msg.From())
	gas := uint64(0x999999999)
	value := big.NewInt(0)

	data, err := common.GenerateWasmData(common.CallContractFlag, funcName, funcParams)
	if err != nil {
		log.Error(err.Error())
		return nil, gas, err
	}
	return evm.Call(caller, address, data, gas, value)
}

func (st *StateTransition) refundContractGas(contractAddr common.Address) error {
	// Apply refund counter, capped to half of the used gas.
	refund := st.gasUsed() / 2
	if refund > st.state.GetRefund() {
		refund = st.state.GetRefund()
	}
	st.gas += refund

	// Return ETH for remaining gas, exchanged at the original rate.
	//remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	//st.state.AddBalance(st.msg.From(), remaining)
	addr := st.msg.From().String()
	addr = strings.ToLower(addr)
	params := []interface{}{addr, st.gas}
	_, _, err := st.doCallContract(contractAddr, "refundFee", params)
	if nil != err {
		log.Warn("refundContractGas error", "err", err.Error()) //TODO format
		return err
	}

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(st.gas)
	return nil
}

func (st *StateTransition) refundGas() {
	// Apply refund counter, capped to half of the used gas.
	refund := st.gasUsed() / 2
	if refund > st.state.GetRefund() {
		refund = st.state.GetRefund()
	}
	st.gas += refund

	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) gasUsed() uint64 {
	return st.initialGas - st.gas
}

func checkContractDeployPermission(sender common.Address, evm *vm.EVM) bool {
	checkPermission := common.SysCfg.IfCheckContractDeployPermission()
	if checkPermission == 0 {
		return true
	}

	if vm.HasContractDeployPermission(evm.StateDB, sender) {
		return true
	}

	return false
}

func isCallParamManager(to *common.Address) bool {
	if to == nil {
		return false
	}

	return *to == syscontracts.ParameterManagementAddress
}
