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
	syscontracts.USER_MANAGEMENT_ADDRESS: &UserManagement{},
	syscontracts.NODE_MANAGEMENT_ADDRESS: &scNodeWrapper{},
}

//input format： hex.encode( rlp.encode( [][]byte{rlp.encode(txType), function name,rlp.encode(params[1]), rlp.encode(params[1])...} ) )
//old input format： hex.encode( rlp.encode( [][]byte{Int64ToBytes(txType), function name,BasicTypeToBytes(params[1]), BasicTypeToBytes(params[1])...} ) )
//旧的格式入参都是基础格式，只是简单的把数据按照内存格式以byte方法导出而已（有bug，平台依赖？？至少是有语言依赖的。）。
//TODO
func RunPlatONEPrecompiledSC(p PrecompiledContract, input []byte, contract *Contract, evm *EVM) (ret []byte, err error) {
	gas := p.RequiredGas(input)

	if contract.UseGas(gas) {
		switch p.(type) {
		case *UserManagement:
			um := &UserManagement{
				Contract: contract,
				Evm:      evm,
			}
			return um.Run(input)
		case *scNodeWrapper:
			node := newSCNodeWrapper()
			node.base.stateDB = evm.StateDB
			node.base.caller = contract.CallerAddress

			return node.Run(input)
		default:
			panic("system contract handler not found")
		}
	}

	return nil, ErrOutOfGas
}
