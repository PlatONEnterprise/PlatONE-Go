package vm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

// var fwProcessErr = errors.New("firewall process error!")
var fwErrNotOwner = errors.New("FW : error, only contract owner can set firewall setting!")

type FireWall struct {
	db           StateDB
	contractAddr common.Address // st.to()		contract.self
	caller       common.Address // msg.From()	contract.caller
}

func (u *FireWall) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.FireWall
}

// Run runs the precompiled contract
func (u *FireWall) Run(input []byte) ([]byte, error) {
	return execSC(input, u.AllExportFns())
}

// for access control
func (u *FireWall) AllExportFns() SCExportFns {
	return SCExportFns{
		"__sys_FwOpen":   u.openFirewall,
		"__sys_FwClose":  u.closeFirewall,
		"__sys_FwClear":  u.fwClear,
		"__sys_FwAdd":    u.fwAdd,
		"__sys_FwDel":    u.fwDel,
		"__sys_FwSet":    u.fwSet,
		"__sys_FwImport": u.fwImport,
		"__sys_FwStatus": u.getFwStatus,
		"__sys_FwExport": u.getFwStatus,
	}
}

func (u *FireWall) isOwner(contractAddr string) bool {
	contractOwnerAddr := u.db.GetContractCreator(common.HexToAddress(contractAddr))
	callerAddr := u.caller

	if callerAddr.Hex() == contractOwnerAddr.Hex() {
		return true
	} else {
		return false
	}
}

func (u *FireWall) openFirewall(contractAddr string) (int, error) {
	if !u.isOwner(contractAddr) {
		return failure, fwErrNotOwner
	}

	u.db.OpenFirewall(common.HexToAddress(contractAddr))
	return success, nil
}

func (u *FireWall) closeFirewall(contractAddr string) (int, error) {
	if !u.isOwner(contractAddr) {
		return failure, fwErrNotOwner
	}

	u.db.CloseFirewall(common.HexToAddress(contractAddr))
	return success, nil
}

func (u *FireWall) fwClear(contractAddr, action string) (int, error) {
	if !u.isOwner(contractAddr) {
		return failure, fwErrNotOwner
	}

	act, err := actConvert(action)
	if err != nil {
		return failure, err
	}

	u.db.FwClear(common.HexToAddress(contractAddr), act)
	return success, nil
}

func (u *FireWall) fwAdd(contractAddr, action, lst string) (int, error) {
	if !u.isOwner(contractAddr) {
		return failure, fwErrNotOwner
	}

	act, err := actConvert(action)
	if err != nil {
		return failure, err
	}

	list, err := listConvert(lst)
	if err != nil {
		return failure, err
	}

	u.db.FwAdd(common.HexToAddress(contractAddr), act, list)
	return success, nil
}

func (u *FireWall) fwDel(contractAddr, action, lst string) (int, error) {
	if !u.isOwner(contractAddr) {
		return failure, fwErrNotOwner
	}

	act, err := actConvert(action)
	if err != nil {
		return failure, err
	}

	list, err := listConvert(lst)
	if err != nil {
		return failure, err
	}

	u.db.FwDel(common.HexToAddress(contractAddr), act, list)
	return success, nil
}

func (u *FireWall) fwSet(contractAddr, action, lst string) (int, error) {
	if !u.isOwner(contractAddr) {
		return failure, fwErrNotOwner
	}

	act, err := actConvert(action)
	if err != nil {
		return failure, err
	}

	list, err := listConvert(lst)
	if err != nil {
		return failure, err
	}

	u.db.FwSet(common.HexToAddress(contractAddr), act, list)
	return success, nil
}

func (u *FireWall) fwImport(contractAddr string, data []byte) (int, error) {
	if !u.isOwner(contractAddr) {
		return failure, fwErrNotOwner
	}

	err := u.db.FwImport(common.HexToAddress(contractAddr), data)
	return success, err
}

func (u *FireWall) getFwStatus(contractAddr string) (string, error) {
	if !u.isOwner(contractAddr) {
		return "", fwErrNotOwner
	}

	fwStatus := u.db.GetFwStatus(common.HexToAddress(contractAddr))

	returnBytes, err := json.Marshal(fwStatus)
	if err != nil {
		/// log.Warn("FW : fwStatus Marshal error", "err", err)
		errStr := fmt.Sprintf("FW : fwStatus Marshal error: %v", err)
		return "", errors.New(errStr)
	}
	return string(makeReturnBytes(returnBytes)), nil
}

func listConvert(l string) ([]state.FwElem, error) {
	var list []state.FwElem

	elements := strings.Split(l, "|")
	for _, e := range elements {
		tmp := strings.Split(e, ":")
		if len(tmp) != 2 {
			/// log.Warn("FW : error, wrong function parameters!")
			return nil, errors.New("FW : error, incorrect firewall rule format")
		}

		addr := tmp[0]
		api := tmp[1]
		if addr == "*" {
			addr = state.FWALLADDR
		}
		fwElem := state.FwElem{Addr: common.HexToAddress(addr), FuncName: api}
		list = append(list, fwElem)
	}

	return list, nil
}

func actConvert(action string) (state.Action, error) {
	if strings.EqualFold(action, "ACCEPT") {
		return state.ACCEPT, nil
	} else if strings.EqualFold(action, "REJECT") {
		return state.REJECT, nil
	} else {
		return 0, errors.New("FW : error, action is invalid!") // todo fix the return value
	}
}

// todo optimize: code is duplicated, see state_transition.go
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
