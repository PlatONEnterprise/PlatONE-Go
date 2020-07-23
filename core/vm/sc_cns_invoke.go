package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"strings"
)

type CnsInvoke struct {
	evm      *EVM
	caller   common.Address
	contract *Contract
}

func (c *CnsInvoke) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.CnsInvokeGas
}

func (c *CnsInvoke) Run(input []byte) ([]byte, error) {
	cnsRawData := input
	var cnsData [][]byte

	if err := rlp.DecodeBytes(cnsRawData, &cnsData); err != nil {
		log.Warn("Decode cnsRawData failed", "err", err)
		c.evm.StateDB.SetNonce(c.caller, c.evm.StateDB.GetNonce(c.caller)+1)
		return nil, err
	}

	if len(cnsData) < 3 {
		c.evm.StateDB.SetNonce(c.caller, c.evm.StateDB.GetNonce(c.caller)+1)
		return nil, nil
	}

	addr, err := c.getCnsAddr(string(cnsData[1]))
	if err != nil {
		log.Warn("GetCnsAddr failed", "err", err)
		c.evm.StateDB.SetNonce(c.caller, c.evm.StateDB.GetNonce(c.caller)+1)
		return nil, err
	}

	if *addr == ZeroAddress {
		return nil, nil
	}

	cnsData = append(cnsData[:1], cnsData[2:]...)
	cnsRawData, err = rlp.EncodeToBytes(cnsData)
	if err != nil {
		log.Warn("Encode Cns Data failed", "err", err)
		c.evm.StateDB.SetNonce(c.caller, c.evm.StateDB.GetNonce(c.caller)+1)
		return nil, err
	}

	//msg := inputRevert(input)
	res, _, err := c.evm.Call(AccountRef(c.caller), *addr, cnsRawData, c.contract.Gas, c.contract.value)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (c *CnsInvoke) getCnsAddr(cnsName string) (*common.Address, error) {
	//addrProxy := syscontracts.CnsManagementAddress

	var contractName, contractVer string
	var ToAddr common.Address

	posOfColon := strings.Index(cnsName, ":")

	// The cnsName must be the format "Name:Version"
	if posOfColon == -1 {
		contractName = cnsName
		contractVer = "latest"
	} else {
		contractName = cnsName[:posOfColon]
		contractVer = cnsName[posOfColon+1:]
	}

	//if contractName == "" || contractVer == "" {
	//	return nil, errors.New("cns name do not has the right format")
	//}
	//
	//if contractName == "cnsManager" {
	//	return &addrProxy, nil
	//}

	//var isSystemContract = false
	//for _, v := range common.SystemContractList {
	//	if v == contractName {
	//		isSystemContract = true
	//		break
	//	}
	//}
	//if isSystemContract {
	//	ToAddr = cnsSysContractsMap[contractName]
	//} else {
	ToAddr, err := getCnsAddress(c.evm.StateDB, contractName, contractVer)
	if err != nil {
		return nil, err
	}

	return &ToAddr, nil

}
