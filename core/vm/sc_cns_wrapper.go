package vm

import (
	"errors"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

var (
	errInvalidCallNotFromInit = errors.New("[CNS] cnsRegisterFromInit can only be called from init()")
	errInvalidCallFromInit    = errors.New("[CNS] cnsRegister can't be called from init()")
	errLowRegVersion          = errors.New("[CNS] Version must be larger than previous version")
	errNameAndVerReg          = errors.New("[CNS] name and version is already registered and activated in CNS")
	errNameReg                = errors.New("[CNS] Name is already registered")
	errNameAndVerUnReg        = errors.New("[CNS] Name and version didn't register before")
)

type CnsWrapper struct {
	base *CnsManager
}

func (cns *CnsWrapper) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.CnsManagerGas
}

// Run runs the precompiled contract
func (cns *CnsWrapper) Run(input []byte) ([]byte, error) {
	return execSC(input, cns.AllExportFns())
}

// for access control
func (cns *CnsWrapper) AllExportFns() SCExportFns {
	return SCExportFns{
		"cnsRegisterFromInit":             cns.cnsRegisterFromInit,
		"cnsRegister":                     cns.cnsRegister,
		"cnsRedirect":                     cns.cnsRedirect, // cnsUnregister is deprecated, replaced by cnsRedirect
		"getContractAddress":              cns.getContractAddress,
		"ifRegisteredByAddress":           cns.ifRegisteredByAddress,
		"ifRegisteredByName":              cns.ifRegisteredByName,
		"getRegisteredContracts":          cns.getRegisteredContractsByRange,
		"getRegisteredContractsByName":    cns.getRegisteredContractsByName, // getHistoryContractsByName -> getRegisteredContractsByName
		"getRegisteredContractsByAddress": cns.getRegisteredContractsByAddress,
		"getRegisteredContractsByOrigin":  cns.getRegisteredContractsByOrigin, // getContractInfoByAddress -> getRegisteredContractsByOrigin
	}
}

func (cns *CnsWrapper) cnsRegisterFromInit(name, version string) (int32, error) {
	err := cns.base.cnsRegisterFromInit(name, version)
	return cnsRegisterErrHandle(err)
}

func (cns *CnsWrapper) cnsRegister(name, version string, address common.Address) (int32, error) {
	err := cns.base.cnsRegister(name, version, address)
	return cnsRegisterErrHandle(err)
}

func cnsRegisterErrHandle(err error) (int32, error) {
	if err != nil {
		if err == errInvalidCallFromInit || err == errInvalidCallNotFromInit {
			return int32(cnsInvalidCall), nil
		}

		if err == errNotOwner {
			return int32(cnsNoPermission), nil
		}

		if err == errNameInvalid || err == errVersionInvalid || err == errLowRegVersion {
			return int32(cnsInvalidCall), nil
		}

		if err == errNameAndVerReg || err == errNameReg {
			return int32(cnsRegErr), nil
		}

	}

	return int32(cnsSuccess), nil
}

func (cns *CnsWrapper) cnsRedirect(name, version string) (int32, error) {
	err := cns.base.cnsRedirect(name, version)

	if err != nil {
		if err == errNameInvalid || err == errVersionInvalid {
			return int32(cnsInvalidCall), nil
		}

		if err == errNameAndVerUnReg {
			return int32(cnsRegErr), nil
		}

		if err == errNotOwner {
			return int32(cnsNoPermission), nil
		}

		return 0, err
	}

	return int32(cnsSuccess), nil
}

func (cns *CnsWrapper) getContractAddress(name, version string) (string, error) {
	addr, err := cns.base.getContractAddress(name, version)
	if err != nil {
		return "", err
	}

	return addr.String(), nil
}

func (cns *CnsWrapper) ifRegisteredByAddress(address common.Address) (int32, error) {
	isReg, _ := cns.base.ifRegisteredByAddress(address)

	if isReg {
		return int32(registered), nil
	}

	return int32(unregistered), nil
}

func (cns *CnsWrapper) ifRegisteredByName(name string) (int32, error) {
	isReg, err := cns.base.ifRegisteredByName(name)

	if err != nil {
		return int32(cnsInvalidArgument), nil
	}

	if isReg {
		return int32(registered), nil
	}

	return int32(unregistered), nil
}

func (cns *CnsWrapper) getRegisteredContractsByRange(head, size int) (string, error) {
	cnsInfoArray, err := cns.base.getRegisteredContractsByRange(head, size)
	if err != nil {
		return newInternalErrorResult(err).String(), nil
	}

	return newSuccessResult(cnsInfoArray).String(), nil
}

func (cns *CnsWrapper) getRegisteredContractsByName(name string) (string, error) {
	cnsInfoArray, err := cns.base.getRegisteredContractsByName(name)
	if err != nil {
		return newInternalErrorResult(err).String(), nil
	}

	return newSuccessResult(cnsInfoArray).String(), nil
}

func (cns *CnsWrapper) getRegisteredContractsByAddress(addr common.Address) (string, error) {
	cnsInfoArray, err := cns.base.getRegisteredContractsByAddress(addr)
	if err != nil {
		return newInternalErrorResult(err).String(), nil
	}

	return newSuccessResult(cnsInfoArray).String(), nil
}

func (cns *CnsWrapper) getRegisteredContractsByOrigin(origin common.Address) (string, error) {
	cnsInfoArray, err := cns.base.getRegisteredContractsByOrigin(origin)
	if err != nil {
		return newInternalErrorResult(err).String(), nil
	}

	return newSuccessResult(cnsInfoArray).String(), nil
}
