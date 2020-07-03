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
	syscontracts.USER_MANAGEMENT_ADDRESS:      &UserManagement{},
	syscontracts.NODE_MANAGEMENT_ADDRESS:      &scNodeWrapper{},
	syscontracts.CNS_MANAGEMENT_ADDRESS:       &CnsManager{},
	syscontracts.PARAMETER_MANAGEMENT_ADDRESS: &ParamManager{},
	syscontracts.FIREWALL_MANAGEMENT_ADDRESS:  &FireWall{},
}

func RunPlatONEPrecompiledSC(p PrecompiledContract, input []byte, contract *Contract, evm *EVM) (ret []byte, err error) {
	gas := p.RequiredGas(input)

	if contract.UseGas(gas) {
		switch p.(type) {
		case *UserManagement:
			um := &UserManagement{
				state:   evm.StateDB,
				caller:  contract.Caller(),
				address: syscontracts.USER_MANAGEMENT_ADDRESS,
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
		default:
			panic("system contract handler not found")
		}
	}

	return nil, ErrOutOfGas
}
