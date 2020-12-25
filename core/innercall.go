package core

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"
	"math/big"
)

var InnerCallFromAddress = common.HexToAddress("0x1000000000000000000000000000000000000000")

func innerCallContract(state *state.StateDB, bc *BlockChain, caller common.Address, contractAddr common.Address, input []byte) ([]byte, error) {
	// Create new call message
	msg := types.NewMessage(caller, &contractAddr, 1, big.NewInt(1), uint64(0xffffffffff), big.NewInt(1), input, false, types.NormalTxType)
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
	// Get the state
	state, err := bc.State()
	if state == nil || err != nil {
		return nil, err
	}
	input := common.GenCallData(funcName, funcParams)
	res, err := innerCallContract(state.Copy(), bc, InnerCallFromAddress, contractAddr, input)
	return res, err
}

func InnerCallContractWithState(state *state.StateDB, bc *BlockChain, caller common.Address, contractAddr common.Address, funcName string, funcParams []interface{}) ([]byte, error) {
	input := common.GenCallData(funcName, funcParams)
	return innerCallContract(state, bc, caller, contractAddr, input)
}
