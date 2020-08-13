package vm

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

var ErrFwRule = errors.New("FW : error, incorrect firewall rule format")
var ErrFwRuleAddr = errors.New("FW : error, incorrect firewall rule address format")
var ErrFwRuleName = errors.New("FW : error, incorrect firewall rule api name format")

type FwWrapper struct {
	base *FireWall
}

func (u *FwWrapper) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.FireWall
}

// Run runs the precompiled contract
func (u *FwWrapper) Run(input []byte) ([]byte, error) {
	fnName, ret, err := execSC(input, u.AllExportFns())
	if err != nil {
		if fnName == "" {
			fnName = "Notify"
		}
		u.base.emitEvent(fnName, operateFail, err.Error())
	}

	return ret, nil
}

// for access control
func (u *FwWrapper) AllExportFns() SCExportFns {
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

func (u *FwWrapper) openFirewall(contractAddr common.Address) (int32, error) {
	err := u.base.openFirewall(contractAddr)

	switch err {
	case fwErrNotOwner:
		return int32(fwNoPermission), err
	}

	return int32(fwOpSuccess), nil
}

func (u *FwWrapper) closeFirewall(contractAddr common.Address) (int32, error) {
	err := u.base.closeFirewall(contractAddr)

	switch err {
	case fwErrNotOwner:
		return int32(fwNoPermission), err
	}

	return int32(fwOpSuccess), nil
}

func (u *FwWrapper) fwClear(contractAddr common.Address, action string) (int32, error) {
	err := u.base.fwClear(contractAddr, action)

	switch err {
	case fwErrNotOwner:
		return int32(fwNoPermission), err
	case state.ErrInvalidFwAction:
		return int32(fwInvalidArgument), nil
	}

	return int32(fwOpSuccess), nil
}

func (u *FwWrapper) fwAdd(contractAddr common.Address, action, lst string) (int32, error) {
	err := u.base.fwAdd(contractAddr, action, lst)

	switch err {
	case fwErrNotOwner:
		return int32(fwNoPermission), err
	case state.ErrInvalidFwAction, ErrFwRule, ErrFwRuleName, ErrFwRuleAddr:
		return int32(fwInvalidArgument), nil
	}

	return int32(fwOpSuccess), nil
}

func (u *FwWrapper) fwDel(contractAddr common.Address, action, lst string) (int32, error) {
	err := u.base.fwDel(contractAddr, action, lst)

	switch err {
	case fwErrNotOwner:
		return int32(fwNoPermission), err
	case state.ErrInvalidFwAction, ErrFwRule, ErrFwRuleName, ErrFwRuleAddr:
		return int32(fwInvalidArgument), nil
	}

	return int32(fwOpSuccess), nil
}

func (u *FwWrapper) fwSet(contractAddr common.Address, act, lst string) (int32, error) {
	err := u.base.fwSet(contractAddr, act, lst)

	switch err {
	case fwErrNotOwner:
		return int32(fwNoPermission), err
	case state.ErrInvalidFwAction, ErrFwRule, ErrFwRuleName, ErrFwRuleAddr:
		return int32(fwInvalidArgument), nil
	}

	return int32(fwOpSuccess), nil
}

func (u *FwWrapper) fwImport(contractAddr common.Address, data string) (int32, error) {
	err := u.base.fwImport(contractAddr, []byte(data))

	switch err {
	case fwErrNotOwner:
		return int32(fwNoPermission), err
	}

	return int32(fwOpSuccess), nil
}

func (u *FwWrapper) getFwStatus(contractAddr common.Address) (string, error) {
	fwStatus, err := u.base.getFwStatus(contractAddr)
	if err != nil {
		return "", err
	}

	returnBytes, err := json.Marshal(fwStatus)
	if err != nil {
		errStr := fmt.Sprintf("FW : fwStatus Marshal error: %v", err)
		return "", errors.New(errStr)
	}

	return string(returnBytes), nil
}
