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
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/PlatONnetwork/PlatON-Go/core/types"

	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/core/state"
	"github.com/PlatONnetwork/PlatON-Go/core/vm"
	"github.com/PlatONnetwork/PlatON-Go/log"
	"github.com/PlatONnetwork/PlatON-Go/params"
	"github.com/PlatONnetwork/PlatON-Go/rlp"
)

var (
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
)

const CnsManagerAddr string = "0x0000000000000000000000000000000000000011"

const FwPermissionNotAllowed = "Only contract creator can set the firewall data"
const FwInputInvalid = "Invalid input for firewall setting!"

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
	TxType() uint64
	SetTxType(uint64)
	SetNonce(uint64)

	GasPrice() *big.Int
	Gas() uint64
	Value() *big.Int

	Nonce() uint64
	CheckNonce() bool
	Data() []byte
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(data []byte, contractCreation, homestead bool) (uint64, error) {
	// Set the starting gas for the raw transaction
	var gas uint64
	if contractCreation && homestead {
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

func GetCnsAddr(evm *vm.EVM, msg Message, cnsName string) (*common.Address, error) {
	// TODO: 合约管理合约地址，后面设置为全局变量
	addrProxy := common.HexToAddress(CnsManagerAddr)

	if cnsName == "cnsManager" {
		return &addrProxy, nil
	}

	var contractName, contractVer string

	i := strings.Index(cnsName, ":")
	if i == -1 {
		contractName = cnsName
		contractVer = ""
	} else {
		contractName = cnsName[:i]
		contractVer = cnsName[i+1:]
	}

	paramArr := [][]byte{
		common.Int64ToBytes(int64(vm.CALL_CANTRACT_FLAG)),
		[]byte("getContractAddress"),
		[]byte(contractName),
		[]byte(contractVer),
	}
	paramBytes, _ := rlp.EncodeToBytes(paramArr)

	cnsMsg := types.NewMessage(msg.From(), &addrProxy, 0, new(big.Int), 0x99999, msg.GasPrice(), paramBytes, false, types.CnsTxType)
	gp := new(GasPool).AddGas(math.MaxUint64 / 2)

	snapshot := evm.StateDB.Snapshot()
	ret, _, _, err := NewStateTransition(evm, cnsMsg, gp).TransitionDb()
	evm.StateDB.RevertToSnapshot(snapshot)

	if err != nil {
		log.Error("vm applyMessage failed", err)
		return nil, errors.New("CNS error, getContractAddress failed")
	}

	retStr := string(ret)
	toAddrStr := string(retStr[strings.Index(retStr, "0x"):])
	ToAddr := common.HexToAddress(toAddrStr)

	return &ToAddr, nil
}

// ApplyMessage computes the new state by applying the given message
// against the old state within the environment.
//
// ApplyMessage returns the bytes returned by any EVM execution (if it took place),
// the gas used (which includes gas refunds) and an error if it failed. An error always
// indicates a core error meaning that the message would always fail for that particular
// state and would never be accepted within a block.
func ApplyMessage(evm *vm.EVM, msg Message, gp *GasPool) ([]byte, uint64, bool, error) {

	if msg.TxType() == types.CnsTxType {

		cnsRawData := msg.Data()
		var cnsData [][]byte

		if err := rlp.DecodeBytes(cnsRawData, &cnsData); err != nil {
			log.Debug("Decode cnsRawData failed, ", err)
			return nil, 0, true, nil
		}

		toAddr, err := GetCnsAddr(evm, msg, string(cnsData[1]))
		if err != nil {
			log.Debug("GetCnsAddr failed, ", err)
			return nil, 0, true, nil
		}
		msg.SetTo(*toAddr)

		cnsData = append(cnsData[:1], cnsData[2:]...)
		cnsRawData, _ = rlp.EncodeToBytes(cnsData)

		msg.SetData(cnsRawData)
		msg.SetTxType(types.NormalTxType)

		nonce := evm.StateDB.GetNonce(msg.From())
		msg.SetNonce(nonce)
	}
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
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.msg.Gas()), st.gasPrice)
	if st.state.GetBalance(st.msg.From()).Cmp(mgval) < 0 {
		return errInsufficientBalanceForGas
	}
	if err := st.gp.SubGas(st.msg.Gas()); err != nil {
		return err
	}
	st.gas += st.msg.Gas()

	st.initialGas = st.msg.Gas()
	st.state.SubBalance(st.msg.From(), mgval)
	return nil
}

func (st *StateTransition) preCheck() error {
	// Make sure this transaction's nonce is correct.
	if st.msg.CheckNonce() {
		nonce := st.state.GetNonce(st.msg.From())
		if nonce < st.msg.Nonce() {
			return ErrNonceTooHigh
		} else if nonce > st.msg.Nonce() {
			return ErrNonceTooLow
		}
	}
	return st.buyGas()
}

func addressCompare(addr1, addr2 common.Address) bool {
	return addr1.String() == addr2.String()
}

func fwCheck(stateDb vm.StateDB, contractAddr common.Address, caller common.Address, input []byte) bool {
	var data [][]byte
	if err := rlp.DecodeBytes(input, &data); err != nil {
		return false
	}
	if len(data) < 2 {
		fmt.Println("fw error: require function name")
		return false
	}
	funcName := string(data[1])

	var fwStatus state.FwStatus
	if addressCompare(stateDb.GetContractCreator(contractAddr), caller) {
		return true
	}

	if stateDb.IsFwOpened(contractAddr) == false {
		return true
	}

	fwStatus = stateDb.GetFwStatus(contractAddr)
	if !addressCompare(fwStatus.ContractAddress, contractAddr) {
		return false
	}

	/*
	* Reject List!
	 */
	for _, fwElem := range fwStatus.DeniedList {
		if fwElem.Addr.String() == state.FWALLADDR {
			if fwElem.FuncName == "*" {
				// 1. [*:*] and reject any address and any function access!
				return false
			} else if fwElem.FuncName == funcName {
				// 2. [*:funcName] and funcname matched!
				return false
			} else {
				// 3. [*:funcName] and funcname didn't match!
				continue
			}
		} else {
			if addressCompare(fwElem.Addr, caller) {
				if fwElem.FuncName == "*" {
					// 4. [address:*] and address matched!
					return false
				} else if fwElem.FuncName == funcName {
					// 5. [address:funcName] and both address&funcName matched!
					return false
				} else {
					// 6. [address:funcName] and address matched but funcname didn't match!
					continue
				}
			}
		}
	}

	/*
	* Accept List!
	 */
	for _, fwElem := range fwStatus.AcceptedList {
		if fwElem.Addr.String() == state.FWALLADDR {
			if fwElem.FuncName == "*" {
				// 1. [*:*] and allow any address and any function access!
				return true
			} else if fwElem.FuncName == funcName {
				// 2. [*:funcName] and funcname matched!
				return true
			} else {
				// 3. [*:funcName] and funcname didn't match!
				continue
			}
		} else {
			if addressCompare(fwElem.Addr, caller) {
				if fwElem.FuncName == "*" {
					// 4. [address:*] and address matched!
					return true
				} else if fwElem.FuncName == funcName {
					// 5. [address:funcName] and both address&funcName matched!
					return true
				} else {
					// 6. [address:funcName] and address matched but funcname didn't match!
					continue
				}
			}
		}
	}

	return false
}

func fwProcess(stateDb vm.StateDB, contractAddr common.Address, caller common.Address, input []byte) ([]byte, uint64, error) {
	var fwStatus state.FwStatus
	var err error
	var act state.Action
	var fwData [][]byte
	var funcName, listName, params string
	var list []state.FwElem

	if !addressCompare(stateDb.GetContractCreator(contractAddr), caller) {
		return []byte(FwPermissionNotAllowed), 0, nil
	}

	if err = rlp.DecodeBytes(input, &fwData); err != nil {
		return []byte(FwInputInvalid), 0, nil
	}

	// check parameters
	if len(fwData) < 2 {
		fmt.Println("fw error: require function name")
		return []byte(FwInputInvalid), 0, nil
	}
	funcName = string(fwData[1])
	if funcName == "__sys_FwOpen" || funcName == "__sys_FwClose" || funcName == "__sys_FwStatus" {
		if len(fwData) != 2 {
			fmt.Println("fw error: wrong function parameters")
			return []byte(FwInputInvalid), 0, nil
		}
	} else if funcName == "__sys_FwClear" {
		if len(fwData) != 3 {
			fmt.Println("fw error: wrong function parameters")
			return []byte(FwInputInvalid), 0, nil
		}

		listName = string(fwData[2])
		if listName == "Accept" {
			act = state.ACCEPT
		} else if listName == "Reject" {
			act = state.REJECT
		} else {
			fmt.Println("fw error: action is invalid")
			return []byte(FwInputInvalid), 0, nil
		}

	} else if funcName == "__sys_FwAdd" || funcName == "__sys_FwDel" || funcName == "__sys_FwSet" {
		if len(fwData) != 4 {
			fmt.Println("fw error: wrong function parameters")
			return []byte(FwInputInvalid), 0, nil
		}

		listName = string(fwData[2])
		params = string(fwData[3])
		if listName == "Accept" {
			act = state.ACCEPT
		} else if listName == "Reject" {
			act = state.REJECT
		} else {
			fmt.Println("fw error: action is invalid")
			return []byte(FwInputInvalid), 0, nil
		}

		elements := strings.Split(params, "|")
		for _, e := range elements {
			tmp := strings.Split(e, ":")
			if len(tmp) != 2 {
				return []byte(FwInputInvalid), 0, nil
			}

			addr := tmp[0]
			api := tmp[1]
			if addr == "*" {
				addr = state.FWALLADDR
			}
			fwElem := state.FwElem{Addr: common.HexToAddress(addr), FuncName: api}
			list = append(list, fwElem)
		}

	} else {
		fmt.Println("fw error: wrong function name")
		return []byte(FwInputInvalid), 0, nil
	}

	switch funcName {
	case "__sys_FwOpen":
		stateDb.OpenFirewall(contractAddr)
	case "__sys_FwClose":
		stateDb.CloseFirewall(contractAddr)
	case "__sys_FwClear":
		stateDb.FwClear(contractAddr, act)
	case "__sys_FwAdd":
		stateDb.FwAdd(contractAddr, act, list)
	case "__sys_FwDel":
		stateDb.FwDel(contractAddr, act, list)
	case "__sys_FwSet":
		stateDb.FwSet(contractAddr, act, list)
	default:
		// "__sys_FwStatus"
		fwStatus = stateDb.GetFwStatus(contractAddr)
	}

	var returnBytes []byte
	returnBytes, err = json.Marshal(fwStatus)
	if err != nil {
		fmt.Println("fwStatus Marshal error:", err)
		return []byte(FwInputInvalid), 0, nil
	}

	strHash := common.BytesToHash(common.Int32ToBytes(32))
	sizeHash := common.BytesToHash(common.Int64ToBytes(int64((len(returnBytes)))))
	var dataRealSize = len(returnBytes)
	if (dataRealSize % 32) != 0 {
		dataRealSize = dataRealSize + (32 - (dataRealSize % 32))
	}
	dataByt := make([]byte, dataRealSize)
	copy(dataByt[0:], returnBytes)

	finalData := make([]byte, 0)
	finalData = append(finalData, strHash.Bytes()...)
	finalData = append(finalData, sizeHash.Bytes()...)
	finalData = append(finalData, dataByt...)

	return finalData, 0, nil
}

// TransitionDb will transition the state by applying the current message and
// returning the result including the used gas. It returns an error if failed.
// An error indicates a consensus issue.
func (st *StateTransition) TransitionDb() (ret []byte, usedGas uint64, failed bool, err error) {
	// init initialGas value = txMsg.gas
	if err = st.preCheck(); err != nil {
		return
	}
	msg := st.msg
	sender := vm.AccountRef(msg.From())
	homestead := st.evm.ChainConfig().IsHomestead(st.evm.BlockNumber)
	contractCreation := msg.To() == nil

	// Pay intrinsic gas
	gas, err := IntrinsicGas(st.data, contractCreation, homestead)
	if err != nil {
		return nil, 0, false, err
	}
	if err = st.useGas(gas); err != nil {
		return nil, 0, false, err
	}

	var (
		evm = st.evm
		// vm errors do not effect consensus and are therefor
		// not assigned to err, except for insufficient balance
		// error.
		vmerr error
	)
	if contractCreation {
		ret, _, st.gas, vmerr = evm.Create(sender, st.data, st.gas, st.value)
	} else {
		if !fwCheck(evm.StateDB, st.to(), msg.From(), msg.Data()) {
			log.Debug("Calling contract was refused by firewall", "err", vmerr)
		} else {
			// Increment the nonce for the next transaction
			// If the transaction is cns-type, do not increment the nonce
			if msg.TxType() != types.CnsTxType {
				st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
			}
			if msg.TxType() == types.FwTxType {
				ret, st.gas, vmerr = fwProcess(evm.StateDB, st.to(), msg.From(), msg.Data())
			} else {
				ret, st.gas, vmerr = evm.Call(sender, st.to(), st.data, st.gas, st.value)
			}
		}
	}
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, vmerr
		}
	}
	st.refundGas()
	st.state.AddBalance(st.evm.Coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice))

	return ret, st.gasUsed(), vmerr != nil, err
}

func (st *StateTransition) refundGas() {
	// Apply refund counter, capped to half of the used gas.
	refund := st.gasUsed() / 2
	if refund > st.state.GetRefund() {
		refund = st.state.GetRefund()
	}
	st.gas += refund

	// Return ETH for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	st.state.AddBalance(st.msg.From(), remaining)

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) gasUsed() uint64 {
	return st.initialGas - st.gas
}
