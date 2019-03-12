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
	"github.com/PlatONnetwork/PlatON-Go/life/utils"
	"math"
	"math/big"
	"strings"

	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/core/state"
	"github.com/PlatONnetwork/PlatON-Go/core/types"
	"github.com/PlatONnetwork/PlatON-Go/core/vm"
	"github.com/PlatONnetwork/PlatON-Go/log"
	"github.com/PlatONnetwork/PlatON-Go/params"
	"github.com/PlatONnetwork/PlatON-Go/rlp"
)

var (
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
)

const CnsManagerAddr string = "0x0000000000000000000000000000000000000011"

var fwErr = errors.New("firewall error!")

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
	addrProxy := common.HexToAddress(CnsManagerAddr)

	var contractName, contractVer string

	posOfColon := strings.Index(cnsName, ":")

	// The cnsName must be the format "Name:Version"
	if posOfColon == -1 {
		contractName = cnsName
		contractVer = "latest"
	} else {
		contractName = cnsName[:posOfColon]
		contractVer = cnsName[posOfColon+1:]
	}
	if contractName == "" || contractVer == ""{
		return nil, errors.New("cns name do not has the right format")
	}

	if contractName == "cnsManager" {
		return &addrProxy, nil
	}

	params := []interface{}{contractName, contractVer}

	snapshot := evm.StateDB.Snapshot()
	ret := common.InnerCall(addrProxy, "getContractAddress", params)
	evm.StateDB.RevertToSnapshot(snapshot)

	toAddrStr := common.CallResAsString(ret)
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
			evm.StateDB.SetNonce(msg.From(), evm.StateDB.GetNonce(msg.From())+1)
			return nil,0,true,nil
		}

		if len(cnsData) < 3{
			log.Debug("cnsData < 3 ")
			evm.StateDB.SetNonce(msg.From(), evm.StateDB.GetNonce(msg.From())+1)
			return nil,0,true,nil
		}

		addr, err := GetCnsAddr(evm, msg, string(cnsData[1]))
		if err != nil {
			log.Debug("GetCnsAddr failed, ", err)
			evm.StateDB.SetNonce(msg.From(), evm.StateDB.GetNonce(msg.From())+1)
			return nil,0,true,nil
		}

		msg.SetTo(*addr)

		cnsData = append(cnsData[:1], cnsData[2:]...)
		cnsRawData, err = rlp.EncodeToBytes(cnsData)
		if err!= nil{
			log.Debug("Encode Cns Data failed, ", err)
			evm.StateDB.SetNonce(msg.From(), evm.StateDB.GetNonce(msg.From())+1)
			return nil,0,true,nil
		}

		msg.SetData(cnsRawData)
		msg.SetTxType(types.NormalTxType)
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

func (st *StateTransition) buyContractGas(contractAddr string) error {
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.msg.Gas()), st.gasPrice)

	addr := st.msg.From().String()
	params := []interface{}{addr}
	ret, _, err := st.doCallContract(contractAddr, "getBalance", params)
	if nil != err {
		fmt.Println(err)
		return err
	}

	balance := new(big.Int).SetBytes(ret)
	if balance.Cmp(mgval) < 0 {
		return errInsufficientBalanceForGas
	}
	if err := st.gp.SubGas(st.msg.Gas()); err != nil {
		return err
	}
	st.gas += st.msg.Gas()

	st.initialGas = st.msg.Gas()
	params = []interface{}{addr,mgval.Uint64()}
	_, _, err = st.doCallContract(contractAddr, "subBalance", params)
	if nil != err {
		fmt.Println(err)
		return err
	}

	return nil
}

func (st *StateTransition) preCheck() (string,bool,error) {
	// Make sure this transaction's nonce is correct.
	if st.msg.CheckNonce() {
		nonce := st.state.GetNonce(st.msg.From())
		if nonce < st.msg.Nonce() {
			return "",false, ErrNonceTooHigh
		} else if nonce > st.msg.Nonce() {
			return "", false, ErrNonceTooLow
		}
	}

	contractName,isUseContractToken,err := 	st.ifUseContractTokenAsFee()
	if nil != err{
		return "", false, err
	}

	if isUseContractToken{
		contractAddr,err := st.getContractAddr(contractName)
		if nil != err {
			//return "", false, err
			fmt.Println("getContractAddr failed,",err)//TODO format
		}else {
			return contractAddr, isUseContractToken, st.buyContractGas(contractAddr)
		}
	}

	return "", false,st.buyGas()
}

func addressCompare(addr1, addr2 common.Address) bool {
	return addr1.String() == addr2.String()
}

func fwCheck(stateDb vm.StateDB, contractAddr common.Address, caller common.Address, input []byte) bool {
	var data [][]byte
	// if this is a value transfer tx, just let it go!
	if len(input) == 0 && len(stateDb.GetCode(contractAddr)) == 0 {
		return true;
	}

	if err := rlp.DecodeBytes(input, &data); err != nil {
		return false
	}
	if len(data) < 2 {
		log.Debug("FW : Missing function name")
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
				log.Debug("FW : 1. Reject, pattern [*:*], reject any address and any function access!")
				return false
			} else if fwElem.FuncName == funcName {
				// 2. [*:funcName] and funcname matched!
				log.Debug("FW : 2. Reject, pattern [*:funcName], funcname matched!")
				return false
			} else {
				// 3. [*:funcName] and funcname didn't match!
				log.Debug("FW : 3. Reject pattern [*:funcName], funcname didn't match!")
				continue
			}
		} else {
			if addressCompare(fwElem.Addr, caller) {
				if fwElem.FuncName == "*" {
					// 4. [address:*] and address matched!
					log.Debug("FW : 4. Reject, pattern [address:*], address matched!")
					return false
				} else if fwElem.FuncName == funcName {
					// 5. [address:funcName] and both address&funcName matched!
					log.Debug("FW : 5. Reject, pattern [address:funcName], both address&funcName matched!")
					return false
				} else {
					// 6. [address:funcName] and address matched but funcname didn't match!
					log.Debug("FW : 6. Reject, pattern [address:funcName], address matched but funcname didn't match!")
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
				log.Debug("FW : 1. Accept, pattern [*:*], allow any address and any function access!")
				return true
			} else if fwElem.FuncName == funcName {
				// 2. [*:funcName] and funcname matched!
				log.Debug("FW : 2. Accept, pattern [*:funcName], funcname matched!")
				return true
			} else {
				// 3. [*:funcName] and funcname didn't match!
				log.Debug("FW : 3. Accept, pattern [*:funcName], funcname didn't match!")
				continue
			}
		} else {
			if addressCompare(fwElem.Addr, caller) {
				if fwElem.FuncName == "*" {
					// 4. [address:*] and address matched!
					log.Debug("FW : 4. Accept, pattern [address:*], address matched!")
					return true
				} else if fwElem.FuncName == funcName {
					// 5. [address:funcName] and both address&funcName matched!
					log.Debug("FW : 5. Accept, pattern [address:funcName], both address&funcName matched!")
					return true
				} else {
					// 6. [address:funcName] and address matched but funcname didn't match!
					log.Debug("FW : 6. Accept, pattern [address:funcName], address matched but funcname didn't match!")
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
		return nil, 0, fwErr
	}

	if err = rlp.DecodeBytes(input, &fwData); err != nil {
		return nil, 0, fwErr
	}

	// check parameters
	if len(fwData) < 2 {
		log.Debug("FW : error, require function name")
		return nil, 0, fwErr
	}
	funcName = string(fwData[1])
	if funcName == "__sys_FwOpen" || funcName == "__sys_FwClose" || funcName == "__sys_FwStatus" {
		if len(fwData) != 2 {
			log.Debug("FW : error, wrong function parameters")
			return nil, 0, fwErr
		}
	} else if funcName == "__sys_FwClear" {
		if len(fwData) != 3 {
			log.Debug("FW : error, wrong function parameters")
			return nil, 0, fwErr
		}

		listName = string(fwData[2])
		if listName == "Accept" {
			act = state.ACCEPT
		} else if listName == "Reject" {
			act = state.REJECT
		} else {
			log.Debug("FW : error, action is invalid")
			return nil, 0, fwErr
		}

	} else if funcName == "__sys_FwAdd" || funcName == "__sys_FwDel" || funcName == "__sys_FwSet" {
		if len(fwData) != 4 {
			log.Debug("FW : error, wrong function parameters")
			return nil, 0, fwErr
		}

		listName = string(fwData[2])
		params = string(fwData[3])
		if listName == "Accept" {
			act = state.ACCEPT
		} else if listName == "Reject" {
			act = state.REJECT
		} else {
			log.Debug("FW : error, action is invalid")
			return nil, 0, fwErr
		}

		elements := strings.Split(params, "|")
		for _, e := range elements {
			tmp := strings.Split(e, ":")
			if len(tmp) != 2 {
				return nil, 0, fwErr
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
		log.Debug("FW : error, wrong function name")
		return nil, 0, fwErr
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
		log.Debug("FW : fwStatus Marshal error:", err)
		return nil, 0, fwErr
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

func (st *StateTransition)ifUseContractTokenAsFee()(string,  bool, error){
	params := []interface{}{"__sys_ParamManager", "latest"}
		binParamMangerAddr, _, err := st.doCallContract(CnsManagerAddr, "getContractAddress", params)
	if nil != err {
		fmt.Println(err)
		return "",false, err
	}
	paramMangerAddr := utils.Bytes2string(binParamMangerAddr)
	fmt.Println("paramManagerAddr:",paramMangerAddr)

	if "0x0000000000000000000000000000000000000000" == paramMangerAddr {
		fmt.Println("paramManager contract address not found")//TODO
		return "", false, nil
	}

	params = []interface{}{}
	binContractName, _, err := st.doCallContract(paramMangerAddr, "getGasContractName", params)
	if nil != err {
		fmt.Println(err)
		return "",false, err
	}
	contractName := utils.Bytes2string(binContractName)
	fmt.Println("contractName: ",contractName)

	var isUseContractToken bool  = ("" != contractName)

	return contractName, isUseContractToken,nil
}

func (st *StateTransition) getContractAddr(contractName string) (feeContractAddr string, err error) {
	params := []interface{}{contractName, "latest"}
	var binFeeContractAddr []byte
	binFeeContractAddr, _, err = st.doCallContract(CnsManagerAddr, "getContractAddress", params)
	if nil != err {
		fmt.Println(err)
		return
	}
	feeContractAddr = utils.Bytes2string(binFeeContractAddr)

	if "0x0000000000000000000000000000000000000000" == feeContractAddr {
		err := errors.New("fee contract address not found")
		fmt.Println(err)
		return "",err
	}

	return
}

// TransitionDb will transition the state by applying the current message and
// returning the result including the used gas. It returns an error if failed.
// An error indicates a consensus issue.
func (st *StateTransition) TransitionDb() (ret []byte, usedGas uint64, failed bool, err error) {
	isUseContractToken := false
	feeContractAddr := ""

	// init initialGas value = txMsg.gas
	if feeContractAddr,isUseContractToken,err = st.preCheck(); err != nil {
		return nil,0,false,err
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
		if msg.TxType() != types.CnsTxType {
			st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
		}
		if !fwCheck(evm.StateDB, st.to(), msg.From(), msg.Data()) {
			vmerr = fwErr
			log.Debug("Calling contract was refused by firewall", "err", vmerr)
		} else {
			// Increment the nonce for the next transaction
			// If the transaction is cns-type, do not increment the nonce
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

	if isUseContractToken {
		err = st.refundContractGas(feeContractAddr)
	} else {
		st.refundGas()
		st.state.AddBalance(st.evm.Coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice))
	}

	return ret, st.gasUsed(), vmerr != nil, err
}

func (st *StateTransition)doCallContract(address, funcName string, funcParams []interface{}) (ret []byte, leftOverGas uint64, err error) {
	evm := st.evm
	msg := st.msg
	caller := vm.AccountRef(msg.From())
	gas := uint64(0x999999999)

	var txType int64 = vm.CALL_CANTRACT_FLAG // donot encode result in rlp
	paramArr := [][]byte{
		common.Int64ToBytes(txType),
		[]byte(funcName),
	}

	for _, v := range funcParams {
		p, e := common.ToBytes(v)
		if e != nil {
			err := fmt.Errorf("convert %v to string failed", v)
			fmt.Println(err) //TODO modify log format
			return nil, gas, err
		}
		paramArr = append(paramArr, p)
	}

	paramBytes, e := rlp.EncodeToBytes(paramArr)
	if e != nil {
		err := fmt.Errorf("rpl encode error,%s", e.Error())
		fmt.Println(err) //TODO modify log format
		return nil, gas, err
	}

	value := big.NewInt(0)
	//evm.StateDB.SetNonce(msg.From(), st.state.GetNonce(msg.From())+1)
	return evm.Call(caller, common.HexToAddress(address), paramBytes, gas, value)
}

func (st *StateTransition) refundContractGas(contractAddr string) error {
	// Apply refund counter, capped to half of the used gas.
	refund := st.gasUsed() / 2
	if refund > st.state.GetRefund() {
		refund = st.state.GetRefund()
	}
	st.gas += refund

	// Return ETH for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	//st.state.AddBalance(st.msg.From(), remaining)
	addr := st.msg.From().String()
	params := []interface{}{addr,remaining.Uint64()}
	_, _, err := st.doCallContract(contractAddr, "addBalance", params)
	if nil != err {
		fmt.Println(err)//TODO format
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
