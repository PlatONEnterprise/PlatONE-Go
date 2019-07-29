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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"
	"github.com/PlatONEnetwork/PlatONE-Go/life/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"math"
	"math/big"
	"strings"
)

var (
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
)

const CnsManagerAddr string = "0x0000000000000000000000000000000000000011"

var fwErr = errors.New("firewall error!")
var FirewallErr = errors.New("Permission Denied!")

var migErr = errors.New("migration error!")


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
	var ToAddr common.Address

	if contractName == "cnsManager" {
		return &addrProxy, nil
	}

	posOfColon := strings.Index(cnsName, ":")

	// The cnsName must be the format "Name:Version"
	if posOfColon == -1 {
		contractName = cnsName
		contractVer = "latest"

		var isSystemcontract bool = false
		for _, v := range common.SystemContractList {
			if v == contractName {
				isSystemcontract = true
				break
			}
		}

		callContract := func(conAddr common.Address, data []byte) []byte {
			res, _, err := evm.Call(vm.AccountRef(common.Address{}), conAddr, data, uint64(0xffffffffff), big.NewInt(0))
			if err != nil {
				return nil
			}
			return res
		}

		if isSystemcontract {
			ToAddr = common.SysCfg.GetContractAddress(contractName)
		} else {
			var fh string = "getContractAddress"
			callParams := []interface{}{contractName, "latest"}
			btsRes := callContract(common.HexToAddress(CnsManagerAddr), common.GenCallData(fh, callParams))
			strRes := common.CallResAsString(btsRes)
			if !(len(strRes) == 0 || common.IsHexZeroAddress(strRes)) {
				ToAddr = common.HexToAddress(strRes)
			}
		}

		return &ToAddr, nil
	} else {
		contractName = cnsName[:posOfColon]
		contractVer = cnsName[posOfColon+1:]
		if contractName == "" || contractVer == "" {
			return nil, errors.New("cns name do not has the right format")
		}

		params := []interface{}{contractName, contractVer}

		snapshot := evm.StateDB.Snapshot()
		ret, err := common.InnerCall(addrProxy, "getContractAddress", params)
		if err != nil {
			return nil, err
		}
		evm.StateDB.RevertToSnapshot(snapshot)

		toAddrStr := common.CallResAsString(ret)
		ToAddr = common.HexToAddress(toAddrStr)

		return &ToAddr, nil
	}
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
			log.Warn("Decode cnsRawData failed", "err", err)
			evm.StateDB.SetNonce(msg.From(), evm.StateDB.GetNonce(msg.From())+1)
			return nil, 0, true, nil
		}

		if len(cnsData) < 3 {
			evm.StateDB.SetNonce(msg.From(), evm.StateDB.GetNonce(msg.From())+1)
			return nil, 0, true, nil
		}

		addr, err := GetCnsAddr(evm, msg, string(cnsData[1]))
		if err != nil {
			log.Warn("GetCnsAddr failed", "err", err)
			evm.StateDB.SetNonce(msg.From(), evm.StateDB.GetNonce(msg.From())+1)
			return nil, 0, true, nil
		}

		msg.SetTo(*addr)

		cnsData = append(cnsData[:1], cnsData[2:]...)
		cnsRawData, err = rlp.EncodeToBytes(cnsData)
		if err != nil {
			log.Warn("Encode Cns Data failed", "err", err)
			evm.StateDB.SetNonce(msg.From(), evm.StateDB.GetNonce(msg.From())+1)
			return nil, 0, true, nil
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
	//mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.msg.Gas()), st.gasPrice)
	//if st.state.GetBalance(st.msg.From()).Cmp(mgval) < 0 {
	//	return errInsufficientBalanceForGas
	//}
	//if err := st.gp.SubGas(st.msg.Gas()); err != nil {
	//	return err
	//}
	//st.gas += st.msg.Gas()
	//
	//st.initialGas = st.msg.Gas()

	gas := uint64(common.SysCfg.GetTxGasLimit())
	if err := st.gp.SubGas(gas); err != nil {
		return err
	}
	st.gas += gas

	st.initialGas = gas
	//st.state.SubBalance(st.msg.From(), mgval)
	return nil
}

func (st *StateTransition) buyContractGas(contractAddr string) error {
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.msg.Gas()), st.gasPrice)

	addr := st.msg.From().String()
	params := []interface{}{addr}
	ret, _, err := st.doCallContract(contractAddr, "getBalance", params)
	if nil != err {
		log.Warn("buyContractGas error", "err", err.Error())
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
	params = []interface{}{addr, mgval.Uint64()}
	_, _, err = st.doCallContract(contractAddr, "subBalance", params)
	if nil != err {
		fmt.Println(err)
		return err
	}

	return nil
}

func (st *StateTransition) preCheck() error {
	//set gasPrice = 0, for not sub Txfee and not add coinbase
	st.gasPrice = new(big.Int).SetInt64(0)

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

func (st *StateTransition) preContractGasCheck(contractAddr string) error {
	// Make sure this transaction's nonce is correct.
	if st.msg.CheckNonce() {
		nonce := st.state.GetNonce(st.msg.From())
		if nonce < st.msg.Nonce() {
			return ErrNonceTooHigh
		} else if nonce > st.msg.Nonce() {
			return ErrNonceTooLow
		}
	}

	return st.buyContractGas(contractAddr)
}

func addressCompare(addr1, addr2 common.Address) bool {
	return strings.ToLower(addr1.String()) == strings.ToLower(addr2.String())
}

func makeReturnBytes(ret []byte) []byte {

	strHash := common.BytesToHash(common.Int32ToBytes(32))
	sizeHash := common.BytesToHash(common.Int64ToBytes(int64((len(ret)))))
	var dataRealSize = len(ret)
	if (dataRealSize % 32) != 0 {
		dataRealSize = dataRealSize + (32 - (dataRealSize % 32))
	}
	dataByt := make([]byte, dataRealSize)
	copy(dataByt[0:], ret)

	finalData := make([]byte, 0)
	finalData = append(finalData, strHash.Bytes()...)
	finalData = append(finalData, sizeHash.Bytes()...)
	finalData = append(finalData, dataByt...)

	return finalData
}

// 合约防火墙的检查：
//  1. 如果账户结构体code字段为空，pass
//  2. 如果账户data字段为空，pass
// 	3. 黑名单优先于白名单，后续只有不在黑名单列表，同时在白名单列表里的账户才能pass
func fwCheck(stateDb vm.StateDB, contractAddr common.Address, caller common.Address, input []byte) ([]byte, bool) {
	var data [][]byte
	// 如果账户结构体code字段为空，pass
	if len(stateDb.GetCode(contractAddr)) == 0 {
		return nil, true
	}

	// 如果账户data字段为空，pass
	if len(input) == 0 {
		return nil, true
	}

	if err := rlp.DecodeBytes(input, &data); err != nil {
		log.Debug("FW : Input decode error")
		return makeReturnBytes([]byte("FW : Input decode error")), false
	}
	if len(data) < 2 {
		log.Debug("FW : Missing function name")
		return makeReturnBytes([]byte("FW : Missing function name")), false
	}
	funcName := string(data[1])

	var fwStatus state.FwStatus
	if addressCompare(stateDb.GetContractCreator(contractAddr), caller) {
		return nil, true
	}

	if stateDb.IsFwOpened(contractAddr) == false {
		return nil, true
	}

	fwStatus = stateDb.GetFwStatus(contractAddr)

	var fwLog string = "FW : Access to contract:" + contractAddr.String() + " by " + funcName + "is refused by firewall."
	/*
	* Reject List!
	 */
	for _, fwElem := range fwStatus.DeniedList {
		if addressCompare(fwElem.Addr, common.HexToAddress(state.FWALLADDR)) {
			if fwElem.FuncName == "*" {
				// 1. [*:*] and reject any address and any function access!
				log.Debug("FW : 1. Reject, pattern [*:*], reject any address and any function access!")
				return makeReturnBytes([]byte(fwLog)), false
			} else if fwElem.FuncName == funcName {
				// 2. [*:funcName] and funcname matched!
				log.Debug("FW : 2. Reject, pattern [*:funcName], funcname matched!")
				return makeReturnBytes([]byte(fwLog)), false
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
					return makeReturnBytes([]byte(fwLog)), false
				} else if fwElem.FuncName == funcName {
					// 5. [address:funcName] and both address&funcName matched!
					log.Debug("FW : 5. Reject, pattern [address:funcName], both address&funcName matched!")
					return makeReturnBytes([]byte(fwLog)), false
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
		if addressCompare(fwElem.Addr, common.HexToAddress(state.FWALLADDR)) {
			if fwElem.FuncName == "*" {
				// 1. [*:*] and allow any address and any function access!
				log.Debug("FW : 1. Accept, pattern [*:*], allow any address and any function access!")
				return nil, true
			} else if fwElem.FuncName == funcName {
				// 2. [*:funcName] and funcname matched!
				log.Debug("FW : 2. Accept, pattern [*:funcName], funcname matched!")
				return nil, true
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
					return nil, true
				} else if fwElem.FuncName == funcName {
					// 5. [address:funcName] and both address&funcName matched!
					log.Debug("FW : 5. Accept, pattern [address:funcName], both address&funcName matched!")
					return nil, true
				} else {
					// 6. [address:funcName] and address matched but funcname didn't match!
					log.Debug("FW : 6. Accept, pattern [address:funcName], address matched but funcname didn't match!")
					continue
				}
			}
		}
	}

	return makeReturnBytes([]byte(fwLog)), false
}

// 只允许合约创建者设置合约的防火墙规则
func fwProcess(stateDb vm.StateDB, contractAddr common.Address, caller common.Address, input []byte) ([]byte, uint64, error) {
	var fwStatus state.FwStatus
	var err error
	var act state.Action
	var fwData [][]byte
	var funcName, listName, params string
	var list []state.FwElem

	if !addressCompare(stateDb.GetContractCreator(contractAddr), caller) {
		log.Warn("FW : error, only contract owner can set firewall setting!")
		return nil, 0, fwErr
	}

	if err = rlp.DecodeBytes(input, &fwData); err != nil {
		log.Warn("FW : error, fwData decoded failure!")
		return nil, 0, fwErr
	}

	// check parameters
	if len(fwData) < 2 {
		log.Warn("FW : error, require function name!")
		return nil, 0, fwErr
	}
	funcName = string(fwData[1])
	if funcName == "__sys_FwOpen" || funcName == "__sys_FwClose" || funcName == "__sys_FwStatus" {
		if len(fwData) != 2 {
			log.Warn("FW : error, wrong function parameters!")
			return nil, 0, fwErr
		}
	} else if funcName == "__sys_FwClear" {
		if len(fwData) != 3 {
			log.Warn("FW : error, wrong function parameters!")
			return nil, 0, fwErr
		}

		listName = string(fwData[2])
		if listName == "Accept" {
			act = state.ACCEPT
		} else if listName == "Reject" {
			act = state.REJECT
		} else {
			log.Warn("FW : error, action is invalid!")
			return nil, 0, fwErr
		}

	} else if funcName == "__sys_FwAdd" || funcName == "__sys_FwDel" || funcName == "__sys_FwSet" {
		if len(fwData) != 4 {
			log.Warn("FW : error, wrong function parameters!")
			return nil, 0, fwErr
		}

		listName = string(fwData[2])
		params = string(fwData[3])
		if listName == "Accept" {
			act = state.ACCEPT
		} else if listName == "Reject" {
			act = state.REJECT
		} else {
			log.Warn("FW : error, action is invalid!")
			return nil, 0, fwErr
		}

		elements := strings.Split(params, "|")
		for _, e := range elements {
			tmp := strings.Split(e, ":")
			if len(tmp) != 2 {
				log.Warn("FW : error, wrong function parameters!")
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

	} else if funcName == "__sys_FwImport" {
		if len(fwData) != 3 {
			log.Warn("FW : error, wrong function parameters!")
			return nil, 0, fwErr
		}
	} else {
		log.Warn("FW : error, wrong function name!")
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
	case "__sys_FwImport":
		stateDb.FwImport(contractAddr, fwData[2])
	default:
		// "__sys_FwStatus"
		fwStatus = stateDb.GetFwStatus(contractAddr)
	}

	var returnBytes []byte
	returnBytes, err = json.Marshal(fwStatus)
	if err != nil {
		log.Warn("FW : fwStatus Marshal error", "err", err)
		return nil, 0, fwErr
	}
	return makeReturnBytes(returnBytes), 0, nil
}

func (st *StateTransition) ifUseContractTokenAsFee() (string, bool, error) {
	params := []interface{}{"__sys_ParamManager", "latest"}
	binParamMangerAddr, _, err := st.doCallContract(CnsManagerAddr, "getContractAddress", params)
	if nil != err {
		log.Warn("ifUseContractTokenAsFee error", "err", err.Error())
		return "", false, err
	}
	paramMangerAddr := utils.Bytes2string(binParamMangerAddr)

	if "0x0000000000000000000000000000000000000000" == paramMangerAddr {
		//fmt.Println("paramManager contract address not found")//TODO
		return "", false, nil
	}

	params = []interface{}{}
	binContractName, _, err := st.doCallContract(paramMangerAddr, "getGasContractName", params)
	if nil != err {
		log.Warn("st.doCallContract error", "err", err.Error())
		return "", false, err
	}
	contractName := utils.Bytes2string(binContractName)

	var isUseContractToken bool = ("" != contractName)

	contractAddr := ""
	if isUseContractToken {
		contractAddr, err = st.getContractAddr(contractName)
		if nil != err {
			log.Warn("getContractAddr error", "err", err.Error())
			return "", false, err
		}
	}

	return contractAddr, isUseContractToken, nil
}

func (st *StateTransition) getContractAddr(contractName string) (feeContractAddr string, err error) {
	params := []interface{}{contractName, "latest"}
	var binFeeContractAddr []byte
	binFeeContractAddr, _, err = st.doCallContract(CnsManagerAddr, "getContractAddress", params)
	if nil != err {
		log.Warn("getContractAddr fail", "err", err.Error())
		return
	}
	feeContractAddr = utils.Bytes2string(binFeeContractAddr)

	if "0x0000000000000000000000000000000000000000" == feeContractAddr {
		err := errors.New("fee contract address not found")
		return "", err
	}

	return
}

// TransitionDb will transition the state by applying the current message and
// returning the result including the used gas. It returns an error if failed.
// An error indicates a consensus issue.
func (st *StateTransition) TransitionDb() (ret []byte, usedGas uint64, failed bool, err error) {
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

	var (
		isUseContractToken = false
		feeContractAddr    = ""
	)

	//TODO comment temporarily for performance test
	//feeContractAddr, isUseContractToken,err = st.ifUseContractTokenAsFee()
	//if nil != err{
	//	st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
	//	return nil, 0, true, nil
	//}
	//
	//if isUseContractToken{
	//	// init initialGas value = txMsg.gas
	//	if err = st.preContractGasCheck(feeContractAddr); err != nil {
	//		st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
	//		return nil, 0, true, nil
	//	}
	//}else {
	//	// init initialGas value = txMsg.gas
	if err = st.preCheck(); err != nil {
		return
	}
	//}

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
	if contractCreation {
		//check sender permisson
		res, err := checkSenderPermission(sender.Address(), evm)
		if err != nil {
			st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
			log.Debug("VM returned with error", "err", vmerr)
			return nil, 0, true, nil
		}
		if !res {
			st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
			return nil, 0, true, FirewallErr
		}

		if msg.TxType() == types.MigTxType {
			ret, _, st.gas, vmerr = evm.MigCreate(sender, st.data, st.gas, st.value)
		} else {
			ret, _, st.gas, vmerr = evm.Create(sender, st.data, st.gas, st.value)
		}
		//st.state.OpenFirewall(conAddr)
	} else {
		if msg.TxType() != types.CnsTxType {
			st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
		}
		var pass bool
		if ret, pass = fwCheck(evm.StateDB, st.to(), msg.From(), msg.Data()); !pass {
			err = FirewallErr
			vmerr = FirewallErr
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
		vmerr = st.refundContractGas(feeContractAddr)
	} else {
		st.refundGas()
		//st.state.AddBalance(st.evm.Coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice))
	}

	return ret, st.gasUsed(), vmerr != nil, err
}

func (st *StateTransition) doCallContract(address, funcName string, funcParams []interface{}) (ret []byte, leftOverGas uint64, err error) {
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
			log.Error(err.Error())
			return nil, gas, err
		}
		paramArr = append(paramArr, p)
	}

	paramBytes, e := rlp.EncodeToBytes(paramArr)
	if e != nil {
		err := fmt.Errorf("rpl encode error,%s", e.Error())
		log.Error(err.Error())
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
	params := []interface{}{addr, remaining.Uint64()}
	_, _, err := st.doCallContract(contractAddr, "addBalance", params)
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

	// Return ETH for remaining gas, exchanged at the original rate.
	//remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	//st.state.AddBalance(st.msg.From(), remaining)

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) gasUsed() uint64 {
	return st.initialGas - st.gas
}

func checkSenderPermission(sender common.Address, evm *vm.EVM) (bool, error) {
	allowAny := common.SysCfg.IfCheckContractDeployPermission()
	if allowAny == 0 {
		return true, nil
	}

	valid, err := isValidUser(sender, evm)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, nil
	}
	conAddr, found := getContractAddr("__sys_RoleManager")
	if !found {
		return true, nil
	}
	res, err := callContractByFw(conAddr, "getRolesByAddress", sender, evm)
	if err != nil {
		return false, err
	}
	roles := string(res[:])
	if strings.Contains(roles, "contractAdmin") || strings.Contains(roles, "chainCreator") ||
		strings.Contains(roles, "chainAdmin") || strings.Contains(roles, "contractDeployer") {
		return true, nil
	}
	return false, nil

}

func isValidUser(sender common.Address, evm *vm.EVM) (bool, error) {

	conAddr, found := getContractAddr("__sys_UserManager")
	if !found {
		return true, nil
	}
	res, err := callContractByFw(conAddr, "isValidUser", sender, evm)
	if err != nil {
		return false, err
	}
	valid := common.CallResAsInt64(res)
	if valid > 0 {
		return true, nil
	}
	return false, nil
}

func getContractAddr(cn string) (common.Address, bool) {
	conAddr := common.SysCfg.GetContractAddress(cn)
	if (conAddr == common.Address{}) {
		return common.Address{}, false
	}
	return conAddr, true
}

func callContractByFw(conAddr common.Address, fn string, sender common.Address, evm *vm.EVM) ([]byte, error) {
	useraddr := hex.EncodeToString(sender[:])
	callParams := []interface{}{useraddr}
	data := common.GenCallData(fn, callParams)
	res, _, err := evm.Call(vm.AccountRef(common.Address{}), conAddr, data, uint64(0xffffffffff), big.NewInt(0))
	return res, err

}
