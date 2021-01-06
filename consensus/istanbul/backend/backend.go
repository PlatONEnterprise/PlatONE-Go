// Copyright 2017 The go-ethereum Authors
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

package backend

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"
	"github.com/PlatONEnetwork/PlatONE-Go/p2p"
	"github.com/PlatONEnetwork/PlatONE-Go/params"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/consensus"
	"github.com/PlatONEnetwork/PlatONE-Go/consensus/istanbul"
	istanbulCore "github.com/PlatONEnetwork/PlatONE-Go/consensus/istanbul/core"
	"github.com/PlatONEnetwork/PlatONE-Go/consensus/istanbul/validator"
	"github.com/PlatONEnetwork/PlatONE-Go/core"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/ethdb"
	"github.com/PlatONEnetwork/PlatONE-Go/event"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	lru "github.com/hashicorp/golang-lru"
)

const (
	// fetcherID is the ID indicates the block is from Istanbul engine
	fetcherID = "istanbul"
)

// New creates an Ethereum backend for Istanbul core engine.
func New(config *params.IstanbulConfig, privateKey *ecdsa.PrivateKey, db ethdb.Database) consensus.Istanbul {
	// Allocate the snapshot caches and create the engine
	recents, _ := lru.NewARC(inmemorySnapshots)
	recentMessages, _ := lru.NewARC(inmemoryPeers)
	knownMessages, _ := lru.NewARC(inmemoryMessages)

	var address common.Address
	if privateKey == nil {
		address = common.BytesToAddress([]byte("0x0000000000000000000000000000000000000112"))
	} else {
		address = crypto.PubkeyToAddress(privateKey.PublicKey)
	}
	backend := &backend{
		config:           config,
		istanbulEventMux: new(event.TypeMux),
		msgFeed:          new(event.Feed),
		privateKey:       privateKey,
		address:          address,
		logger:           log.New(),
		db:               db,
		commitCh:         make(chan *types.Block, 1),
		recents:          recents,
		candidates:       make(map[common.Address]bool),
		coreStarted:      false,
		recentMessages:   recentMessages,
		knownMessages:    knownMessages,
	}
	backend.core = istanbulCore.New(backend, backend.config)
	return backend
}

// ----------------------------------------------------------------------------
// environment is the engine's current environment and holds all of the current state information.
type environment struct {
	//signer types.Signer
	block *types.Block

	state   *state.StateDB // apply state changes here
	tcount  int            // tx count in cycle
	gasPool *core.GasPool  // available gas used to pack transactions

	header   *types.Header
	txs      []*types.Transaction
	receipts []*types.Receipt
}

type backend struct {
	config           *params.IstanbulConfig
	istanbulEventMux *event.TypeMux
	msgFeed          *event.Feed
	privateKey       *ecdsa.PrivateKey
	address          common.Address
	core             istanbulCore.Engine
	logger           log.Logger
	db               ethdb.Database
	chain            consensus.ChainReader
	currentBlock     func() *types.Block
	current          *environment

	// the channels for istanbul engine notifications
	commitCh          chan *types.Block
	proposedBlockHash common.Hash
	sealMu            sync.Mutex
	coreStarted       bool
	coreMu            sync.RWMutex

	// Current list of candidates we are pushing
	candidates map[common.Address]bool
	// Protects the signer fields
	candidatesLock sync.RWMutex
	// Snapshots for recent block to speed up reorgs
	recents *lru.ARCCache

	// event subscription for ChainHeadEvent event
	broadcaster consensus.Broadcaster

	recentMessages *lru.ARCCache // the cache of peer's messages
	knownMessages  *lru.ARCCache // the cache of self messages
}

// Address implements istanbul.Backend.Address
func (sb *backend) Address() common.Address {
	return sb.address
}

// Validators implements istanbul.Backend.Validators
func (sb *backend) Validators(proposal istanbul.Proposal) istanbul.ValidatorSet {
	return sb.getValidators(proposal.Number().Uint64(), proposal.Hash())
}

// Broadcast implements istanbul.Backend.Broadcast
func (sb *backend) Broadcast(valSet istanbul.ValidatorSet, payload []byte) error {
	// send to others
	sb.Gossip(valSet, payload)
	// send to self
	msg := istanbul.MessageEvent{
		Payload: payload,
	}
	go sb.msgFeed.Send(msg)
	return nil
}

// Broadcast implements istanbul.Backend.Gossip
func (sb *backend) Gossip(valSet istanbul.ValidatorSet, payload []byte) error {
	hash := istanbul.RLPHash(payload)
	sb.knownMessages.Add(hash, true)

	targets := make(map[common.Address]bool)
	for _, val := range valSet.List() {
		if val.Address() != sb.Address() {
			targets[val.Address()] = true
		}
	}

	if sb.broadcaster != nil && len(targets) > 0 {
		ps := sb.broadcaster.FindPeers(targets)
		for addr, p := range ps {
			ms, ok := sb.recentMessages.Get(addr)
			var m *lru.ARCCache
			if ok {
				m, _ = ms.(*lru.ARCCache)
				if _, k := m.Get(hash); k {
					// This peer had this event, skip it
					continue
				}
			} else {
				m, _ = lru.NewARC(inmemoryMessages)
			}

			m.Add(hash, true)
			sb.recentMessages.Add(addr, m)

			go p.Send(istanbulMsg, payload)
		}
	}
	return nil
}

func (sb *backend) writeCommitedBlockWithState(block *types.Block) error {
	var (
		chain *core.BlockChain
		//receipts = make([]*types.Receipt, len(sb.current.receipts))
		logs   []*types.Log
		events []interface{}
		ok     bool
	)

	if chain, ok = sb.chain.(*core.BlockChain); !ok {
		return errors.New("sb.chain not a core.BlockChain")
	}
	if sb.current == nil {
		return errors.New("sb.current is nil")
	}
	if chain.HasBlock(block.Hash(), block.NumberU64()) {
		return nil
	}

	for _, receipt := range sb.current.receipts {
		//receipts[i] = new(types.Receipt)
		//*receipts[i] = *receipt
		// Update the block hash in all logs since it is now available and not when the
		// receipt/log of individual transactions were created.
		for _, log := range receipt.Logs {
			log.BlockHash = block.Hash()
		}
		logs = append(logs, receipt.Logs...)
	}

	now := time.Now()
	stat, err := chain.WriteBlockWithState(block, sb.current.receipts, sb.current.state, false)
	log.Info("write block with state ----------------------", "duration", time.Since(now))
	if err != nil {
		return err
	}
	go sb.EventMux().Post(core.NewMinedBlockEvent{Block: block})

	switch stat {
	case core.CanonStatTy:
		log.Debug("Prepare Events, WriteStatus=CanonStatTy")
		events = append(events, core.ChainEvent{Block: block, Hash: block.Hash(), Logs: logs})
		events = append(events, core.ChainHeadEvent{Block: block})
	case core.SideStatTy:
		log.Debug("Prepare Events, WriteStatus=SideStatTy")
		events = append(events, core.ChainSideEvent{Block: block})
	}

	chain.PostChainEvents(events, logs)
	return nil
}

// Commit implements istanbul.Backend.Commit
func (sb *backend) Commit(proposal istanbul.Proposal, seals [][]byte) error {
	// Check if the proposal is a valid block
	block := &types.Block{}
	block, ok := proposal.(*types.Block)
	if !ok {
		sb.logger.Error("Invalid proposal", "proposal", proposal)
		return errInvalidProposal
	}

	h := block.Header()
	// Append seals into extra-data
	err := writeCommittedSeals(h, seals)
	if err != nil {
		return err
	}
	// update block's header
	block = block.WithSeal(h)
	isEmpty := block.Transactions().Len() == 0
	isProduceEmptyBlock := common.SysCfg.IsProduceEmptyBlock()

	if !isEmpty || isProduceEmptyBlock {
		sb.logger.Info("Committed", "address", sb.Address(), "hash", proposal.Hash(), "number", proposal.Number().Uint64())
	}
	// - if the proposed and committed blocks are the same, send the proposed hash
	//   to commit channel, which is being watched inside the engine.Seal() function.
	// - otherwise, we try to insert the block.
	// -- if success, the ChainHeadEvent event will be broadcasted, try to build
	//    the next block and the previous Seal() will be stopped.
	// -- otherwise, a error will be returned and a round change event will be fired.
	if sb.proposedBlockHash == block.Hash() {
		sb.proposedBlockHash = common.Hash{}
		if err := sb.CheckFirstNodeCommitAtWrongTime(); err != nil {
			sb.commitCh <- nil
			return istanbulCore.ErrFirstCommitAtWrongTime
		}
		// feed block hash to Seal() and wait the Seal() result
		if isEmpty && !isProduceEmptyBlock {
			sb.commitCh <- nil
			return istanbulCore.ErrEmpty
		}
		sb.commitCh <- block
		return nil
	}

	if err := sb.CheckFirstNodeCommitAtWrongTime(); err != nil {
		return istanbulCore.ErrFirstCommitAtWrongTime
	}

	if sb.current != nil && sb.current.block != nil && sb.current.block.Hash() == block.Hash() {
		if isEmpty && !isProduceEmptyBlock {
			return istanbulCore.ErrEmpty
		}

		if err := sb.writeCommitedBlockWithState(block); err != nil {
			sb.logger.Error("writeCommitedBlockWithState() failed", "error", err.Error())
			return err
		}
	} else {
		if isEmpty && !isProduceEmptyBlock {
			return istanbulCore.ErrEmpty
		}

		if sb.broadcaster != nil {
			sb.broadcaster.Enqueue(fetcherID, block)
		}
	}
	return nil
}

// EventMux implements istanbul.Backend.EventMux
func (sb *backend) EventMux() *event.TypeMux {
	return sb.istanbulEventMux
}

// EventMux implements istanbul.Backend.EventMux
func (sb *backend) MsgFeed() *event.Feed {
	return sb.msgFeed
}

// makeCurrent creates a new environment for the current cycle.
func (sb *backend) makeCurrent(parentRoot common.Hash, header *types.Header) error {
	var (
		state *state.StateDB
		gp    = new(core.GasPool)
		chain *core.BlockChain
		err   error
		ok    bool
	)
	gp.AddGas(header.GasLimit)

	if chain, ok = sb.chain.(*core.BlockChain); !ok {
		return errors.New("invalid chainReader in consensus engine")
	}

	if state, err = chain.StateAt(parentRoot); err != nil {
		return err
	}

	env := &environment{
		//signer:  types.NewEIP155Signer(chain.Config().ChainID),
		state:   state,
		header:  header,
		gasPool: gp,
		txs:     make([]*types.Transaction, 0),
	}

	// Keep track of transactions which return errors so they can be removed
	env.tcount = 0
	sb.current = env
	return nil
}

func (sb *backend) excuteBlock(proposal istanbul.Proposal) error {
	var (
		block  *types.Block
		chain  *core.BlockChain
		parent *types.Block
		header *types.Header
		ok     bool
		err    error
	)

	if block, ok = proposal.(*types.Block); !ok {
		return errors.New("invalid proposal")
	}

	if chain, ok = sb.chain.(*core.BlockChain); !ok {
		return errors.New("sb.chain not a core.BlockChain")
	}

	header = block.Header()

	if parent = chain.GetBlockByHash(header.ParentHash); parent == nil {
		return errors.New("Proposal's parent block is not in current chain")
	}

	if err = sb.makeCurrent(parent.Root(), header); err != nil {
		return err
	} else {
		// Iterate over and process the individual transactios
		txsMap := make(map[common.Hash]struct{})
		for _, tx := range block.Transactions() {
			sb.current.state.Prepare(tx.Hash(), common.Hash{}, sb.current.tcount)
			snap := sb.current.state.Snapshot()
			if r := chain.GetReceiptsByHash(tx.Hash()); r != nil {
				return errors.New("Already executed tx")
			}
			if _, ok := txsMap[tx.Hash()]; ok {
				return errors.New("Repeated tx in one block")
			} else {
				txsMap[tx.Hash()] = struct{}{}
			}

			receipt, _, err := core.ApplyTransaction(chain.Config(), chain, &sb.address, sb.current.gasPool, sb.current.state, sb.current.header, tx, &sb.current.header.GasUsed, vm.Config{})
			if err != nil {
				sb.current.state.RevertToSnapshot(snap)
				return err
			}
			sb.current.txs = append(sb.current.txs, tx)
			sb.current.receipts = append(sb.current.receipts, receipt)
			sb.current.tcount++

			sb.current.state.Finalise(true)
		}

		cblock, err := sb.Finalize(chain, header, sb.current.state, block.Transactions(), sb.current.receipts)
		if err != nil {
			return err
		}

		if cblock.Root() != block.Root() {
			sb.current = nil
			return errors.New("Invalid block root")
		}
		sb.current.block = block
		sb.current.header = block.Header()
	}

	return nil
}

// Verify implements istanbul.Backend.Verify
func (sb *backend) Verify(proposal istanbul.Proposal, isProposer bool) (time.Duration, error) {
	// Check if the proposal is a valid block
	block := &types.Block{}
	block, ok := proposal.(*types.Block)
	if !ok {
		sb.logger.Error("Invalid proposal", "proposal", proposal)
		return 0, errInvalidProposal
	}

	// check block body
	txnHash := types.DeriveSha(block.Transactions())
	//uncleHash := types.CalcUncleHash(block.Uncles())
	if txnHash != block.Header().TxHash {
		return 0, errMismatchTxhashes
	}
	//if uncleHash != nilUncleHash {
	//	return 0, errInvalidUncleHash
	//}

	// If this node is proposer and the proposal is mined by this node, need not to execute the block
	if (block.Coinbase() != sb.address) || !isProposer {
		//excute txs in block
		if err := sb.excuteBlock(proposal); err != nil {
			return 0, err
		}
	}

	// verify the header of proposed block
	err := sb.VerifyHeader(sb.chain, block.Header(), false)
	// ignore errEmptyCommittedSeals error because we don't have the committed seals yet
	if err == nil || err == errEmptyCommittedSeals {
		return 0, nil
	} else if err == consensus.ErrFutureBlock {
		headTime := block.Header().Time.Int64()
		return time.Unix(headTime/1000, (headTime%1000)*1e6).Sub(now()), consensus.ErrFutureBlock
	}
	return 0, err
}

// Sign implements istanbul.Backend.Sign
func (sb *backend) Sign(data []byte) ([]byte, error) {
	hashData := crypto.Keccak256([]byte(data))
	return crypto.Sign(hashData, sb.privateKey)
}

// CheckSignature implements istanbul.Backend.CheckSignature
func (sb *backend) CheckSignature(data []byte, address common.Address, sig []byte) error {
	signer, err := istanbul.GetSignatureAddress(data, sig)
	if err != nil {
		log.Error("Failed to get signer address", "err", err)
		return err
	}
	// Compare derived addresses
	if signer != address {
		return errInvalidSignature
	}
	return nil
}

// HasPropsal implements istanbul.Backend.HashBlock
func (sb *backend) HasPropsal(hash common.Hash, number *big.Int) bool {
	return sb.chain.GetHeader(hash, number.Uint64()) != nil
}

// GetProposer implements istanbul.Backend.GetProposer
func (sb *backend) GetProposer(number uint64) common.Address {
	if h := sb.chain.GetHeaderByNumber(number); h != nil {
		a, _ := sb.Author(h)
		return a
	}
	return common.Address{}
}

// ParentValidators implements istanbul.Backend.GetParentValidators
func (sb *backend) ParentValidators(proposal istanbul.Proposal) istanbul.ValidatorSet {
	if block, ok := proposal.(*types.Block); ok {
		return sb.getValidators(block.Number().Uint64()-1, block.ParentHash())
	}
	return validator.NewSet(nil, sb.config.ProposerPolicy)
}

func (sb *backend) getValidators(number uint64, hash common.Hash) istanbul.ValidatorSet {
	snap, err := sb.snapshot(sb.chain, number, hash, nil)
	if err != nil {
		return validator.NewSet(nil, sb.config.ProposerPolicy)
	}
	return snap.ValSet
}

func (sb *backend) LastProposal() (istanbul.Proposal, common.Address) {
	block := sb.currentBlock()

	var proposer common.Address
	if block.Number().Cmp(common.Big0) > 0 {
		var err error
		proposer, err = sb.Author(block.Header())
		if err != nil {
			sb.logger.Error("Failed to get block proposer", "err", err)
			return nil, common.Address{}
		}
	}

	// Return header only block here since we don't need block body
	return block, proposer
}

// SealHash returns the hash of a block prior to it being sealed.
func (sb *backend) SealHash(header *types.Header) common.Hash {
	return header.SealHash()
}

// Close implements consensus.Engine. It's a noop for cbft as there is are no background threads.
func (sb *backend) Close() error {
	return nil
}

func (sb *backend) ShouldSeal() bool {
	header := sb.currentBlock().Header()
	sb.getValidators(header.Number.Uint64(), header.Hash())
	return sb.core.CanPropose()
}

// Check if the first node of the network is allowed to produce blocks
func (sb *backend) CheckFirstNodeCommitAtWrongTime() error {
	block := sb.currentBlock()
	if block.NumberU64() != 0 || sb.config.FirstValidatorNode.ID.String() == "" {
		return nil
	}
	nodeId := sb.config.FirstValidatorNode.ID.String()
	// 1. self is the first node of validatorNodes in genesis
	// 2. The node startup specifies bootNodes,
	// 	  and if it is not specified itself, no block generation is performed.
	if p2p.IsSelfServerNode(nodeId) &&
		len(p2p.GetBootNodes()) != 0 && !p2p.IsNodeInBootNodes(nodeId) {
		return istanbulCore.ErrFirstCommitAtWrongTime
	}
	return nil
}
