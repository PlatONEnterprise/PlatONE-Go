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
	return execSC(input, u.AllExportFns())
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
	if err != nil {
		if err == fwErrNotOwner {
			return int32(fwNoPermission), nil
		}
	}

	return int32(fwOpSuccess), nil
}

func (u *FwWrapper) closeFirewall(contractAddr common.Address) (int32, error) {
	err := u.base.closeFirewall(contractAddr)
	if err != nil {
		if err == fwErrNotOwner {
			return int32(fwNoPermission), nil
		}
	}

	return int32(fwOpSuccess), nil
}

func (u *FwWrapper) fwClear(contractAddr common.Address, action string) (int32, error) {
	err := u.base.fwClear(contractAddr, action)
	if err != nil {
		if err == fwErrNotOwner {
			return int32(fwNoPermission), nil
		}

		if err == state.ErrInvalidFwAction {
			return int32(fwInvalidArgument), nil
		}
	}

	return int32(fwOpSuccess), nil
}

func (u *FwWrapper) fwAdd(contractAddr common.Address, action, lst string) (int32, error) {
	err := u.base.fwAdd(contractAddr, action, lst)
	if err != nil {
		if err == fwErrNotOwner {
			return int32(fwNoPermission), nil
		}

		if err == state.ErrInvalidFwAction || err == ErrFwRule {
			return int32(fwInvalidArgument), nil
		}
	}

	return int32(fwOpSuccess), nil
}

func (u *FwWrapper) fwDel(contractAddr common.Address, action, lst string) (int32, error) {
	err := u.base.fwDel(contractAddr, action, lst)
	if err != nil {
		if err == fwErrNotOwner {
			return int32(fwNoPermission), nil
		}

		if err == state.ErrInvalidFwAction || err == ErrFwRule {
			return int32(fwInvalidArgument), nil
		}
	}

	return int32(fwOpSuccess), nil
}

func (u *FwWrapper) fwSet(contractAddr common.Address, act, lst string) (int32, error) {
	err := u.base.fwSet(contractAddr, act, lst)
	if err != nil {
		if err == fwErrNotOwner {
			return int32(fwNoPermission), nil
		}

		if err == state.ErrInvalidFwAction || err == ErrFwRule {
			return int32(fwInvalidArgument), nil
		}
	}

	return int32(fwOpSuccess), nil
}

func (u *FwWrapper) fwImport(contractAddr common.Address, data []byte) (int32, error) {
	err := u.base.fwImport(contractAddr, data)
	if err != nil {
		if err == fwErrNotOwner {
			return int32(fwNoPermission), nil
		}
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
