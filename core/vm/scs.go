package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
)

//system contract export functions
type (
	SCExportFn  interface{}
	SCExportFns map[string]SCExportFn //map[function name]function pointer
)

var PlatONEPrecompiledContracts = map[common.Address]PrecompiledContract{
	syscontracts.UserManagementAddress:      &UserManagement{},
	syscontracts.NodeManagementAddress:      &scNodeWrapper{},
	syscontracts.CnsManagementAddress:       &CnsManager{},
	syscontracts.ParameterManagementAddress: &ParamManager{},
	syscontracts.FirewallManagementAddress:  &FireWall{},
	syscontracts.GroupManagementAddress: &GroupManagement{},
	syscontracts.ContractDataProcessorAddress: &ContractDataProcessor{},
}

func RunPlatONEPrecompiledSC(p PrecompiledContract, input []byte, contract *Contract, evm *EVM) (ret []byte, err error) {
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
			node := newSCNodeWrapper(nil)
			node.base.stateDB = evm.StateDB
			node.base.caller = contract.CallerAddress
			node.base.blockNumber = evm.BlockNumber
			node.base.contractAddress = *contract.CodeAddr

			return node.Run(input)
		case *CnsManager:
			cns := &CnsManager{
				callerAddr: contract.CallerAddress,
				cMap:       NewCnsMap(evm.StateDB, contract.CodeAddr),
				isInit:     evm.InitEntryID,
				origin:     evm.Context.Origin,
			}
			return cns.Run(input)
		case *ParamManager:
			p := &ParamManager{
				state:      evm.StateDB,
				CodeAddr:   contract.CodeAddr,
				CallerAddr: contract.CallerAddress,
			}
			return p.Run(input)
		case *FireWall:
			fw := &FireWall{
				db:				evm.StateDB,
				contractAddr: 	contract.self.Address(),
				caller:			contract.caller.Address(),
			}
			return fw.Run(input)
		case *GroupManagement:
			gm := &GroupManagement{
				state:			evm.StateDB,
				address: 		contract.self.Address(),
				caller:			contract.caller.Address(),
			}
			return gm.Run(input)
		case *ContractDataProcessor:
			dp := &ContractDataProcessor{
				state:			evm.StateDB,
				address: 		contract.self.Address(),
				caller:			contract.caller.Address(),
			}
			return dp.Run(input)
		default:
			panic("system contract handler not found")
		}
	}

	return nil, ErrOutOfGas
}
