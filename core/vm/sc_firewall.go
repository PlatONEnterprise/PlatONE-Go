package vm

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

const (
	fwNoPermission    CodeType = 0
	fwInvalidArgument CodeType = 1
)

type FireWall struct {
	stateDB     StateDB
	caller      common.Address // msg.From()	contract.caller
	blockNumber *big.Int
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

func (u *FireWall) isOwner(contractAddr common.Address) bool {
	contractOwnerAddr := u.stateDB.GetContractCreator(contractAddr)
	callerAddr := u.caller

	if callerAddr.Hex() == contractOwnerAddr.Hex() {
		return true
	}

	return false
}

func (u *FireWall) openFirewall(contractAddr common.Address) (int32, error) {
	if !u.isOwner(contractAddr) {
		u.emitNotifyEvent(fwNoPermission, fwErrNotOwner.Error())
		return failure, fwErrNotOwner
	}

	u.stateDB.OpenFirewall(contractAddr)
	return success, nil
}

func (u *FireWall) closeFirewall(contractAddr common.Address) (int32, error) {
	if !u.isOwner(contractAddr) {
		u.emitNotifyEvent(fwNoPermission, fwErrNotOwner.Error())
		return failure, fwErrNotOwner
	}

	u.stateDB.CloseFirewall(contractAddr)
	return success, nil
}

func (u *FireWall) fwClear(contractAddr common.Address, action string) (int32, error) {
	if !u.isOwner(contractAddr) {
		u.emitNotifyEvent(fwNoPermission, fwErrNotOwner.Error())
		return failure, fwErrNotOwner
	}

	act, err := state.NewAction(action)
	if err != nil {
		u.emitNotifyEvent(fwInvalidArgument, err.Error())
		return failure, err
	}

	u.stateDB.FwClear(contractAddr, act)
	return success, nil
}

func (u *FireWall) fwAdd(contractAddr common.Address, action, lst string) (int32, error) {
	if !u.isOwner(contractAddr) {
		u.emitNotifyEvent(fwNoPermission, fwErrNotOwner.Error())
		return failure, fwErrNotOwner
	}

	act, err := state.NewAction(action)
	if err != nil {
		u.emitNotifyEvent(fwInvalidArgument, err.Error())
		return failure, err
	}

	list, err := convertToFwElem(lst)
	if err != nil {
		u.emitNotifyEvent(fwInvalidArgument, err.Error())
		return failure, err
	}

	u.stateDB.FwAdd(contractAddr, act, list)
	return success, nil
}

func (u *FireWall) fwDel(contractAddr common.Address, action, lst string) (int32, error) {
	if !u.isOwner(contractAddr) {
		u.emitNotifyEvent(fwNoPermission, fwErrNotOwner.Error())
		return failure, fwErrNotOwner
	}

	act, err := state.NewAction(action)
	if err != nil {
		u.emitNotifyEvent(fwInvalidArgument, err.Error())
		return failure, err
	}

	list, err := convertToFwElem(lst)
	if err != nil {
		u.emitNotifyEvent(fwInvalidArgument, err.Error())
		return failure, err
	}

	u.stateDB.FwDel(contractAddr, act, list)
	return success, nil
}

// todo: input arguments type
func (u *FireWall) fwSet(contractAddr common.Address, act, lst string) (int32, error) {
	if !u.isOwner(contractAddr) {
		u.emitNotifyEvent(fwNoPermission, fwErrNotOwner.Error())
		return failure, fwErrNotOwner
	}

	action, err := state.NewAction(act)
	if err != nil {
		u.emitNotifyEvent(fwInvalidArgument, err.Error())
		return failure, err
	}

	list, err := convertToFwElem(lst)
	if err != nil {
		u.emitNotifyEvent(fwInvalidArgument, err.Error())
		return failure, err
	}

	u.stateDB.FwSet(contractAddr, action, list)
	return success, nil
}

func (u *FireWall) fwImport(contractAddr common.Address, data []byte) (int32, error) {
	if !u.isOwner(contractAddr) {
		u.emitNotifyEvent(fwNoPermission, fwErrNotOwner.Error())
		return failure, fwErrNotOwner
	}

	err := u.stateDB.FwImport(contractAddr, data)
	return success, err
}

func (u *FireWall) getFwStatus(contractAddr common.Address) (string, error) {
	if !u.isOwner(contractAddr) {
		return "", fwErrNotOwner
	}

	fwStatus := u.stateDB.GetFwStatus(contractAddr)

	returnBytes, err := json.Marshal(fwStatus)
	if err != nil {
		errStr := fmt.Sprintf("FW : fwStatus Marshal error: %v", err)
		return "", errors.New(errStr)
	}
	return string(returnBytes), nil
}

func (u *FireWall) emitNotifyEvent(code CodeType, msg string) {
	topic := "Notify"
	emitEvent(syscontracts.FirewallManagementAddress, u.stateDB, u.blockNumber.Uint64(), topic, code, msg)
}

func convertToFwElem(l string) ([]state.FwElem, error) {
	var list = make([]state.FwElem, 0)

	elements := strings.Split(l, "|")
	for _, e := range elements {
		tmp := strings.Split(e, ":")
		if len(tmp) != 2 {
			/// log.Warn("FW : error, wrong function parameters!")
			return nil, errors.New("FW : error, incorrect firewall rule format")
		}

		addr := ZeroAddress
		addrStr := tmp[0]
		api := tmp[1]
		if addrStr == "*" {
			addr = state.FwWildchardAddr
		}
		fwElem := state.FwElem{Addr: addr, FuncName: api}
		list = append(list, fwElem)
	}

	return list, nil
}
