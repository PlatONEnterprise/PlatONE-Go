// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/consensus"
	"github.com/PlatONEnetwork/PlatONE-Go/consensus/misc"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"github.com/PlatONEnetwork/PlatONE-Go/rpc"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
		engine: engine,
	}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to
// the processor (coinbase).
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error) {
	var (
		receipts types.Receipts
		usedGas  = new(uint64)
		header   = block.Header()
		allLogs  []*types.Log
		gp       = new(GasPool).AddGas(block.GasLimit())
	)
	// Mutate the block and state according to any hard-fork specs
	if p.config.DAOForkSupport && p.config.DAOForkBlock != nil && p.config.DAOForkBlock.Cmp(block.Number()) == 0 {
		misc.ApplyDAOHardFork(statedb)
	}
	// Iterate over and process the individual transactios
	for i, tx := range block.Transactions() {
		rpc.MonitorWriteData(rpc.TransactionExecuteStartTime, tx.Hash().String(), "", p.bc.extdb)
		txHash := tx.Hash()
		statedb.Prepare(txHash, block.Hash(), i)
		log.Trace("Perform Transaction", "txHash", fmt.Sprintf("%x", txHash[:log.LogHashLen]), "blockNumber", block.Number())
		receipt, _, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		rpc.MonitorWriteData(rpc.TransactionExecuteEndTime, tx.Hash().String(), "", p.bc.extdb)
		if err != nil {
			rpc.MonitorWriteData(rpc.TransactionExecuteStatus, tx.Hash().String(), "false", p.bc.extdb)
			return nil, nil, 0, err
		}
		rpc.MonitorWriteData(rpc.TransactionExecuteStatus, tx.Hash().String(), "true", p.bc.extdb)
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}
	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	block, err := p.engine.Finalize(p.bc, header, statedb, block.Transactions(), receipts)
	if err != nil {
		return nil, nil, 0, err
	}
	return receipts, allLogs, *usedGas, nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return nil, 0, err
	}

	// Create a new context to be used in the EVM environment
	context := NewEVMContext(msg, header, bc, author)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, statedb, config, cfg)
	// Apply the transaction to the current state (included in the env)
	_, gas, gasPrice, failed, err := ApplyMessage(vmenv, msg, gp)

	if err != nil {
		switch err {
		case PermissionErr:
			data := [][]byte{}
			data = append(data, []byte(err.Error()))
			encodeData, _ := rlp.EncodeToBytes(data)
			topics := []common.Hash{common.BytesToHash(crypto.Keccak256([]byte("contract permission")))}
			log := &types.Log{
				Address:     msg.From(),
				Topics:      topics,
				Data:        encodeData,
				BlockNumber: vmenv.BlockNumber.Uint64(),
			}
			statedb.AddLog(log)
		default:
			return nil, 0, err
		}
	}
	if common.SysCfg.GetIsTxUseGas() {
		data := [][]byte{}
		data = append(data, []byte(common.Int64ToBytes(gasPrice)))
		encodeData, _ := rlp.EncodeToBytes(data)
		topics := []common.Hash{common.BytesToHash(crypto.Keccak256([]byte("GasPrice")))}
		log := &types.Log{
			Address:     msg.From(),
			Topics:      topics,
			Data:        encodeData,
			BlockNumber: vmenv.BlockNumber.Uint64(),
		}
		statedb.AddLog(log)
	}
	// Update the state with pending changes
	var root []byte
	if config.IsByzantium(header.Number) {
		statedb.Finalise(true)
	} else {
		root = statedb.IntermediateRoot(config.IsEIP158(header.Number)).Bytes()
	}
	*usedGas += gas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing whether the root touch-delete accounts.
	receipt := types.NewReceipt(root, failed, *usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil && err == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, statedb.GetNonce(msg.From())-1)
	}
	// Set the receipt logs and create a bloom for filtering

	receipt.Logs = statedb.GetLogs(tx.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	return receipt, gas, nil
}
