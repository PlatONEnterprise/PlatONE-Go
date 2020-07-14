package vm

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
)

//system contract export functions
type (
	SCExportFn  interface{}
	SCExportFns map[string]SCExportFn //map[function name]function pointer
)

var PlatONEPrecompiledContracts = map[common.Address]PrecompiledContract{
	syscontracts.UserManagementAddress:        &UserManagement{},
	syscontracts.NodeManagementAddress:        &scNodeWrapper{},
	syscontracts.CnsManagementAddress:         &CnsManager{},
	syscontracts.ParameterManagementAddress:   &ParamManager{},
	syscontracts.FirewallManagementAddress:    &FireWall{},
	syscontracts.GroupManagementAddress:       &GroupManagement{},
	syscontracts.ContractDataProcessorAddress: &ContractDataProcessor{},
	syscontracts.GroupManagementAddress:       &GroupManagement{},
	syscontracts.CnsInvokeAddress:             &CnsInvoke{},
}

func RunPlatONEPrecompiledSC(p PrecompiledContract, input []byte, contract *Contract, evm *EVM) (ret []byte, err error) {
	defer func() {
		if err := recover(); nil != err {
			log.Error("failed to run precompiled system contract", "err", err)
			ret, err = nil, fmt.Errorf("failed to run precompiled system contract,err:%v", err)
		}
	}()

	gas := p.RequiredGas(input)

	if contract.UseGas(gas) {
		switch p.(type) {
		case *UserManagement:
			um := &UserManagement{
				state:   evm.StateDB,
				caller:  contract.Caller(),
				address: syscontracts.UserManagementAddress,
			}
			return um.Run(input)
		case *scNodeWrapper:
			node := newSCNodeWrapper(evm.StateDB)
			node.base.caller = evm.Origin
			node.base.blockNumber = evm.BlockNumber
			node.base.contractAddr = *contract.CodeAddr

			return node.Run(input)
		case *CnsManager:
			cns := newCnsManager(evm.StateDB)
			cns.callerAddr = contract.CallerAddress
			cns.isInit = evm.InitEntryID
			cns.origin = evm.Context.Origin

			return cns.Run(input)
		case *ParamManager:
			p := &ParamManager{
				stateDB:      evm.StateDB,
				contractAddr: contract.CodeAddr,
				caller:       evm.Context.Origin,
				blockNumber:  evm.BlockNumber,
			}
			return p.Run(input)
		case *FireWall:
			fw := &FireWall{
				db:           evm.StateDB,
				contractAddr: contract.self.Address(),
				caller:       contract.caller.Address(),
			}
			return fw.Run(input)
		case *GroupManagement:
			gm := &GroupManagement{
				state:   evm.StateDB,
				address: contract.self.Address(),
				caller:  contract.caller.Address(),
			}
			return gm.Run(input)
		case *ContractDataProcessor:
			dp := &ContractDataProcessor{
				state:   evm.StateDB,
				address: contract.self.Address(),
				caller:  contract.caller.Address(),
			}
			return dp.Run(input)
		case *CnsInvoke:
			ci := &CnsInvoke{
				evm:    evm,
				caller: contract.CallerAddress,
			}
			return ci.Run(input)
		default:
			panic("system contract handler not found")
		}
	}

	return nil, ErrOutOfGas
}
