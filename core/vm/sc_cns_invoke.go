package vm

import (
	"errors"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"math/big"
	"strings"
)

type CnsInvoke struct {
	evm    *EVM
	caller common.Address
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

	addr, err := c.GetCnsAddr(c.evm, string(cnsData[1]))
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
	res, _, err := c.evm.Call(AccountRef(common.Address{}), *addr, cnsRawData, uint64(0xffffffffff), big.NewInt(0))
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (c *CnsInvoke) GetCnsAddr(evm *EVM, cnsName string) (*common.Address, error) {
	addrProxy := syscontracts.CnsManagementAddress

	var contractName, contractVer string
	var ToAddr common.Address

	posOfColon := strings.Index(cnsName, ":")

	// The cnsName must be the format "Name:Version"
	if posOfColon == -1 {
		contractName = cnsName
		contractVer = "latest"

		if contractName == "cnsManager" {
			return &addrProxy, nil
		}

		var isSystemContract bool = false
		for _, v := range common.SystemContractList {
			if v == contractName {
				isSystemContract = true
				break
			}
		}

		callContract := func(conAddr common.Address, data []byte) []byte {
			res, _, err := evm.Call(AccountRef(common.Address{}), conAddr, data, uint64(0xffffffffff), big.NewInt(0))
			if err != nil {
				return nil
			}
			return res
		}

		if isSystemContract {
			ToAddr = cnsSysContractsMap[contractName]
		} else {
			var fh string = "getContractAddress"
			callParams := []interface{}{contractName, "latest"}
			btsRes := callContract(addrProxy, common.GenCallData(fh, callParams))
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

		if contractName == "cnsManager" {
			return &addrProxy, nil
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
