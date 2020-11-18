package core

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"
	"math/big"
)

var InnerCallFromAddress = common.HexToAddress("0x1000000000000000000000000000000000000000")

func innerCallContractReadOnly(bc *BlockChain, contractAddr common.Address, input []byte) ([]byte, error) {
	// Create new call message
	msg := types.NewMessage(InnerCallFromAddress, &contractAddr, 1, big.NewInt(1), uint64(0xffffffffff), big.NewInt(1), input, false, types.NormalTxType)

	// Get the state
	state, err := bc.State()
	if state == nil || err != nil {
		return nil, err
	}
	header := bc.CurrentHeader()

	context := NewEVMContext(msg, header, bc, nil)
	evm := vm.NewEVM(context, state, bc.Config(), vm.Config{})

	res, _, err := evm.Call(vm.AccountRef(msg.From()), *msg.To(), input, msg.Gas(), big.NewInt(0))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func InnerCallContractReadOnly(bc *BlockChain, contractAddr common.Address, funcName string, funcParams []interface{}) ([]byte, error) {
	input := common.GenCallData(funcName, funcParams)
	return innerCallContractReadOnly(bc, contractAddr, input)
}
