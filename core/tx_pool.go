// Copyright 2014 The go-ethereum Authors
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
	"errors"
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/core/rawdb"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/prque"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/ethdb"
	"github.com/PlatONEnetwork/PlatONE-Go/event"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/metrics"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

const (
	// chainHeadChanSize is the size of channel listening to ChainHeadEvent.
	chainHeadChanSize = 10

	// txExtBufferSize is the size fo channel listening to txExt.
	txExtBufferSize = 4096

	DoneRst      = 0
	DoingRst     = 1
	DonePending  = 0
	DoingPending = 1
)

var (
	// ErrInvalidSender is returned if the transaction contains an invalid signature.
	ErrInvalidSender = errors.New("invalid sender")

	// ErrNonceTooLow is returned if the nonce of a transaction is lower than the
	// one present in the local chain.
	ErrNonceTooLow = errors.New("nonce too low")

	ErrTransactionRepeat = errors.New("transaction repeat")
	// ErrUnderpriced is returned if a transaction's gas price is below the minimum
	// configured for the transaction pool.
	ErrUnderpriced = errors.New("transaction underpriced")

	// ErrReplaceUnderpriced is returned if a transaction is attempted to be replaced
	// with a different one without the required price bump.
	ErrReplaceUnderpriced = errors.New("replacement transaction underpriced")

	// ErrInsufficientFunds is returned if the total cost of executing a transaction
	// is higher than the balance of the user's account.
	ErrInsufficientFunds = errors.New("insufficient funds for value")

	// ErrIntrinsicGas is returned if the transaction is specified to use less gas
	// than required to start the invocation.
	ErrIntrinsicGas = errors.New("intrinsic gas too low")

	// ErrGasLimit is returned if a transaction's requested gas limit exceeds the
	// maximum allowance of the current block.
	ErrGasLimit = errors.New("exceeds block gas limit")

	// ErrTxGasLimit is returned if a transaction's requested gas limit exceeds the
	// global maximum allowance of the transaction.
	ErrTransactionGasLimit = errors.New("exceeds transaction gas limit")

	// ErrNegativeValue is a sanity error to ensure noone is able to specify a
	// transaction with a negative value.
	ErrNegativeValue = errors.New("negative value")

	// ErrOversizedData is returned if the input data of a transaction is greater
	// than some meaningful limit a user might use. This is not a consensus error
	// making the transaction invalid, rather a DOS protection.
	ErrOversizedData = errors.New("oversized data")
)

var (
	evictionInterval    = time.Minute     // Time interval to check for evictable transactions
	statsReportInterval = 8 * time.Second // Time interval to report transaction pool stats
)

var (
	// Metrics for the pending pool
	pendingDiscardCounter   = metrics.NewRegisteredCounter("txpool/pending/discard", nil)
	pendingReplaceCounter   = metrics.NewRegisteredCounter("txpool/pending/replace", nil)
	pendingRateLimitCounter = metrics.NewRegisteredCounter("txpool/pending/ratelimit", nil) // Dropped due to rate limiting
	pendingNofundsCounter   = metrics.NewRegisteredCounter("txpool/pending/nofunds", nil)   // Dropped due to out-of-funds

	// Metrics for the queued pool
	queuedDiscardCounter   = metrics.NewRegisteredCounter("txpool/queued/discard", nil)
	queuedReplaceCounter   = metrics.NewRegisteredCounter("txpool/queued/replace", nil)
	queuedRateLimitCounter = metrics.NewRegisteredCounter("txpool/queued/ratelimit", nil) // Dropped due to rate limiting
	queuedNofundsCounter   = metrics.NewRegisteredCounter("txpool/queued/nofunds", nil)   // Dropped due to out-of-funds

	// General tx metrics
	invalidTxCounter     = metrics.NewRegisteredCounter("txpool/invalid", nil)
	underpricedTxCounter = metrics.NewRegisteredCounter("txpool/underpriced", nil)
)

// TxStatus is the current status of a transaction as seen by the pool.
type TxStatus uint

const (
	TxStatusUnknown TxStatus = iota
	TxStatusQueued
	TxStatusPending
	TxStatusIncluded
)

// blockChain provides the state of blockchain and current gas limit to do
// some pre checks in tx pool and event subscribers.
type txPoolBlockChain interface {
	CurrentBlock() *types.Block
	GetBlock(hash common.Hash, number uint64) *types.Block
	//StateAt(root common.Hash) (*state.StateDB, error)
	GetState(header *types.Header) (*state.StateDB, error)
	SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription
}

// TxPoolConfig are the configuration parameters of the transaction pool.
type TxPoolConfig struct {
	Locals    []common.Address // Addresses that should be treated by default as local
	NoLocals  bool             // Whether local transaction handling should be disabled
	Journal   string           // Journal of local transactions to survive node restarts
	Rejournal time.Duration    // Time interval to regenerate the local transaction journal

	PriceLimit uint64 // Minimum gas price to enforce for acceptance into the pool
	PriceBump  uint64 // Minimum price bump percentage to replace an already existing transaction (nonce)

	AccountSlots  uint64 // Number of executable transaction slots guaranteed per account
	GlobalSlots   uint64 // Maximum number of executable transaction slots for all accounts
	AccountQueue  uint64 // Maximum number of non-executable transaction slots permitted per account
	GlobalQueue   uint64 // Maximum number of non-executable transaction slots for all accounts
	GlobalTxCount uint64 // Maximum number of transactions for package

	Lifetime time.Duration // Maximum amount of time non-executable transaction are queued
}

// DefaultTxPoolConfig contains the default configurations for the transaction
// pool.
var DefaultTxPoolConfig = TxPoolConfig{
	Journal:   "transactions.rlp",
	Rejournal: time.Hour,

	PriceLimit: 1,
	PriceBump:  10,

	AccountSlots:  16,
	GlobalSlots:   4096,
	AccountQueue:  64,
	GlobalQueue:   1024,
	GlobalTxCount: 3000,

	Lifetime: 3 * time.Hour,
}

// sanitize checks the provided user configurations and changes anything that's
// unreasonable or unworkable.
func (config *TxPoolConfig) sanitize() TxPoolConfig {
	conf := *config
	if conf.Rejournal < time.Second {
		log.Warn("Sanitizing invalid txpool journal time", "provided", conf.Rejournal, "updated", time.Second)
		conf.Rejournal = time.Second
	}
	if conf.PriceLimit < 1 {
		log.Warn("Sanitizing invalid txpool price limit", "provided", conf.PriceLimit, "updated", DefaultTxPoolConfig.PriceLimit)
		conf.PriceLimit = DefaultTxPoolConfig.PriceLimit
	}
	if conf.PriceBump < 1 {
		log.Warn("Sanitizing invalid txpool price bump", "provided", conf.PriceBump, "updated", DefaultTxPoolConfig.PriceBump)
		conf.PriceBump = DefaultTxPoolConfig.PriceBump
	}
	return conf
}

// TxPool contains all currently known transactions. Transactions
// enter the pool when they are received from the network or submitted
// locally. They exit the pool when they are included in the blockchain.
//
// The pool separates processable transactions (which can be applied to the
// current state) and future transactions. Transactions move between those
// two states over time as they are received and processed.
type TxPool struct {
	config      TxPoolConfig
	chainconfig *params.ChainConfig
	extDb       ethdb.Database
	chain       txPoolBlockChain
	gasPrice    *big.Int
	txFeed      event.Feed
	scope       event.SubscriptionScope
	// modified by PlatONE
	chainHeadCh      chan *types.Block
	chainHeadEventCh chan ChainHeadEvent
	chainHeadSub     event.Subscription
	exitCh           chan struct{}
	signer           types.Signer
	mu               sync.RWMutex

	currentState  *state.StateDB      // Current state in the blockchain head
	pendingState  *state.ManagedState // Pending state tracking virtual nonces
	db            ethdb.Database
	currentMaxGas uint64 // Current gas limit for transaction caps

	locals  *accountSet // Set of local transaction to exempt from eviction rules
	journal *txJournal  // Journal of local transaction to back up to disk

	pending map[common.Address]*txQueuedMap // All currently processable transactions
	//queue   map[common.Address]*txQueuedMap    // Queued but non-processable transactions
	beats map[common.Address]time.Time // Last heartbeat from each known account
	all   *txLookup                    // All transactions to allow lookups
	//priced  *txPricedList                // All transactions sorted by price

	wg sync.WaitGroup // for shutdown sync

	homestead bool

	txExtBuffer chan *txExt
}

type txExt struct {
	tx    interface{} //*types.Transaction
	local bool
	txErr chan interface{}
}

// NewTxPool creates a new transaction pool to gather, sort and filter inbound
// transactions from the network.
//func NewTxPool(config TxPoolConfig, chainconfig *params.ChainConfig, chain blockChain) *TxPool {
func NewTxPool(config TxPoolConfig, chainconfig *params.ChainConfig, chain txPoolBlockChain, db ethdb.Database, extDb ethdb.Database) *TxPool {
	// Sanitize the input to ensure no vulnerable gas prices are set
	config = (&config).sanitize()

	// Create the transaction pool with its initial settings
	pool := &TxPool{
		extDb:       extDb,
		config:      config,
		chainconfig: chainconfig,
		chain:       chain,
		signer:      types.NewEIP155Signer(chainconfig.ChainID),
		pending:     make(map[common.Address]*txQueuedMap),
		//queue:       make(map[common.Address]*txQueuedMap),
		beats: make(map[common.Address]time.Time),
		all:   newTxLookup(),
		db:    db,
		// modified by PlatONE
		chainHeadEventCh: make(chan ChainHeadEvent, chainHeadChanSize),
		chainHeadCh:      make(chan *types.Block, chainHeadChanSize),
		exitCh:           make(chan struct{}),
		gasPrice:         new(big.Int).SetUint64(config.PriceLimit),
		txExtBuffer:      make(chan *txExt, txExtBufferSize),
	}
	pool.locals = newAccountSet(pool.signer)
	for _, addr := range config.Locals {
		log.Info("Setting new local account", "address", addr)
		pool.locals.add(addr)
	}
	//pool.priced = newTxPricedList(pool.all)
	pool.reset(nil, chain.CurrentBlock().Header())

	go pool.txExtBufferReadLoop()

	// If local transactions and journaling is enabled, load from disk
	if !config.NoLocals && config.Journal != "" {
		pool.journal = newTxJournal(config.Journal)

		if err := pool.journal.load(pool.AddLocals); err != nil {
			log.Warn("Failed to load transaction journal", "err", err)
		}
		if err := pool.journal.rotate(pool.local()); err != nil {
			log.Warn("Failed to rotate transaction journal", "err", err)
		}
	}
	// Subscribe events from blockchain
	// modified by PlatONE
	if pool.chainconfig.Istanbul != nil {
		pool.chainHeadSub = pool.chain.SubscribeChainHeadEvent(pool.chainHeadEventCh)
	}

	// Start the event loop and return
	pool.wg.Add(1)
	go pool.loop()

	return pool
}

func (pool *TxPool) txExtBufferReadLoop() {
	for {
		select {
		case ext := <-pool.txExtBuffer:
			err := pool.addTxExt(ext)
			ext.txErr <- err

		case <-pool.exitCh:
			return
		}
	}
}

// loop is the transaction pool's main event loop, waiting for and reacting to
// outside blockchain events as well as for various reporting and transaction
// eviction events.
func (pool *TxPool) loop() {
	defer pool.wg.Done()

	// Start the stats reporting and transaction eviction tickers
	var prevPending int

	report := time.NewTicker(statsReportInterval)
	defer report.Stop()

	evict := time.NewTicker(evictionInterval)
	defer evict.Stop()

	journal := time.NewTicker(pool.config.Rejournal)
	defer journal.Stop()

	// Track the previous head headers for transaction reorgs

	head := pool.chain.CurrentBlock()

	// Keep waiting for and reacting to the various events
	for {
		select {
		// Handle block
		case block := <-pool.chainHeadCh:
			if block != nil {
				pool.mu.Lock()
				if pool.chainconfig.IsHomestead(block.Number()) {
					pool.homestead = true
				}
				pool.reset(head.Header(), block.Header())
				head = block

				pool.mu.Unlock()
			}
		// Handle ChainHeadEvent
		case ev := <-pool.chainHeadEventCh:
			if ev.Block != nil {
				pool.mu.Lock()
				if pool.chainconfig.IsHomestead(ev.Block.Number()) {
					pool.homestead = true
				}
				pool.reset(head.Header(), ev.Block.Header())
				head = ev.Block

				pool.mu.Unlock()
			}

		case <-pool.exitCh:
			return

		// Handle stats reporting ticks
		case <-report.C:
			pool.mu.RLock()
			pending, _ := pool.stats()
			//stales := pool.priced.stales
			pool.mu.RUnlock()

			if pending != prevPending {
				log.Debug("Transaction pool status report", "executable", pending)
				prevPending = pending
			}

		// Handle inactive account transaction eviction
		case <-evict.C:
			pool.mu.Lock()
			for addr := range pool.pending {
				// Skip local transactions from the eviction mechanism
				if pool.locals.contains(addr) {
					continue
				}
				// Any non-locals old enough should be removed
				if time.Since(pool.beats[addr]) > pool.config.Lifetime {
					if pool.pending[addr] != nil {
						for _, tx := range pool.pending[addr].Get() {
							pool.removeTx(tx.Hash(), true)
						}
					}
				}
			}
			pool.mu.Unlock()

		// Handle local transaction journal rotation
		case <-journal.C:
			if pool.journal != nil {
				pool.mu.Lock()
				if err := pool.journal.rotate(pool.local()); err != nil {
					log.Warn("Failed to rotate local tx journal", "err", err)
				}
				pool.mu.Unlock()
			}
		}
	}
}

// lockedReset is a wrapper around reset to allow calling it in a thread safe
// manner. This method is only ever used in the tester!
func (pool *TxPool) lockedReset(oldHead, newHead *types.Header) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.reset(oldHead, newHead)
}

func (pool *TxPool) ForkedReset(origTress, newTress []*types.Block) {
	log.Debug("call ForkedReset()", "RoutineID", common.CurrentGoRoutineID(), "len(origTress)", len(origTress), "len(newTress)", len(newTress))

	pool.mu.Lock()
	defer pool.mu.Unlock()

	var reinject types.Transactions
	// Reorg seems shallow enough to pull in all transactions into memory
	var discarded, included types.Transactions

	for _, orig := range origTress {
		discarded = append(discarded, orig.Transactions()...)
	}
	for _, new := range newTress {
		included = append(included, new.Transactions()...)
	}

	reinject = types.TxDifference(discarded, included)

	// Initialize the internal state to the current head
	//
	statedb, err := pool.chain.GetState(newTress[len(newTress)-1].Header())
	if err != nil {
		log.Error("Failed to reset txpool state", "err", err)
		return
	}
	pool.currentState = statedb
	pool.pendingState = state.ManageState(statedb)
	pool.currentMaxGas = newTress[len(newTress)-1].Header().GasLimit

	// Inject any transactions discarded due to reorgs
	log.Debug("Reinjecting stale transactions", "count", len(reinject))
	senderCacher.recover(pool.signer, reinject)
	pool.addTxsLocked(reinject, false)

	// validate the pool of pending transactions, this will remove
	// any transactions that have been included in the block or
	// have been invalidated because of another transaction (e.g.
	// higher gas price)
	txs := pool.chain.CurrentBlock().Transactions()
	pool.demoteUnexecutables(txs)

	// Check the queue and move transactions over to the pending if possible
	// or remove those that have become invalid
	pool.promoteExecutables(nil)
}

// reset retrieves the current state of the blockchain and ensures the content
// of the transaction pool is valid with regard to the chain state.
func (pool *TxPool) reset(oldHead, newHead *types.Header) {
	var oldHash common.Hash
	var oldNumber uint64
	if oldHead != nil {
		oldHash = oldHead.Hash()
		oldNumber = oldHead.Number.Uint64()
	}

	if oldHead != nil && newHead != nil && oldHead.Hash() == newHead.Hash() && oldHead.Number.Uint64() == newHead.Number.Uint64() {
		log.Debug("txpool needn't reset cause not changed", "RoutineID", common.CurrentGoRoutineID(), "oldHash", oldHash, "oldNumber", oldNumber, "newHash", newHead.Hash(), "newNumber", newHead.Number.Uint64())
		return
	}

	if newHead != nil {
		log.Debug("reset txpool", "RoutineID", common.CurrentGoRoutineID(), "oldHash", oldHash, "oldNumber", oldNumber, "newHash", newHead.Hash(), "newNumber", newHead.Number.Uint64())
	}
	// If we're reorging an old state, reinject all dropped transactions
	var reinject types.Transactions

	if oldHead != nil && oldHead.Hash() != newHead.ParentHash {
		// If the reorg is too deep, avoid doing it (will happen during fast sync)
		oldNum := oldHead.Number.Uint64()
		newNum := newHead.Number.Uint64()

		if depth := uint64(math.Abs(float64(oldNum) - float64(newNum))); depth > 64 {
			log.Debug("Skipping deep transaction reorg", "depth", depth)
		} else {
			// Reorg seems shallow enough to pull in all transactions into memory
			var discarded, included types.Transactions

			var (
				rem = pool.chain.GetBlock(oldHead.Hash(), oldHead.Number.Uint64())
				add = pool.chain.GetBlock(newHead.Hash(), newHead.Number.Uint64())
			)

			if rem == nil {
				log.Debug("cannot find oldHead", "hash", oldHead.Hash(), "number", oldHead.Number.Uint64())
			}
			if add == nil {
				log.Debug("cannot find newHead", "hash", newHead.Hash(), "number", newHead.Number.Uint64())
			}

			if rem != nil && add != nil {
				for rem.NumberU64() > add.NumberU64() {
					discarded = append(discarded, rem.Transactions()...)
					if rem = pool.chain.GetBlock(rem.ParentHash(), rem.NumberU64()-1); rem == nil {
						log.Error("Unrooted old chain seen by tx pool", "block", oldHead.Number, "hash", oldHead.Hash())
						return
					}
				}
				for add.NumberU64() > rem.NumberU64() {
					included = append(included, add.Transactions()...)
					if add = pool.chain.GetBlock(add.ParentHash(), add.NumberU64()-1); add == nil {
						log.Error("Unrooted new chain seen by tx pool", "block", newHead.Number, "hash", newHead.Hash())
						return
					}
				}
				for rem.Hash() != add.Hash() {
					discarded = append(discarded, rem.Transactions()...)
					if rem = pool.chain.GetBlock(rem.ParentHash(), rem.NumberU64()-1); rem == nil {
						log.Error("Unrooted old chain seen by tx pool", "block", oldHead.Number, "hash", oldHead.Hash())
						return
					}
					included = append(included, add.Transactions()...)
					if add = pool.chain.GetBlock(add.ParentHash(), add.NumberU64()-1); add == nil {
						log.Error("Unrooted new chain seen by tx pool", "block", newHead.Number, "hash", newHead.Hash())
						return
					}
				}
				reinject = types.TxDifference(discarded, included)
			}
		}
	}
	// Initialize the internal state to the current head
	if newHead == nil {
		newHead = pool.chain.CurrentBlock().Header() // Special case during testing
	}
	statedb, err := pool.chain.GetState(newHead)
	if err != nil {
		log.Error("Failed to reset txpool state", "newHeadHash", newHead.Hash(), "newHeadNumber", newHead.Number.Uint64(), "err", err)
		return
	}
	pool.currentState = statedb
	pool.pendingState = state.ManageState(statedb)
	pool.currentMaxGas = newHead.GasLimit

	// Inject any transactions discarded due to reorgs
	log.Debug("Reinjecting stale transactions", "count", len(reinject))
	senderCacher.recover(pool.signer, reinject)
	pool.addTxsLocked(reinject, true)

	// validate the pool of pending transactions, this will remove
	// any transactions that have been included in the block or
	// have been invalidated because of another transaction (e.g.
	// higher gas price)
	txs := pool.chain.CurrentBlock().Transactions()
	pool.demoteUnexecutables(txs)

	// Check the queue and move transactions over to the pending if possible
	// or remove those that have become invalid
	pool.promoteExecutables(nil)
}

// Stop terminates the transaction pool.
func (pool *TxPool) Stop() {
	// Unsubscribe all subscriptions registered from txpool
	pool.scope.Close()
	if pool.chainconfig.Istanbul != nil {
		pool.chainHeadSub.Unsubscribe()
	}
	close(pool.exitCh)

	pool.wg.Wait()

	if pool.journal != nil {
		pool.journal.close()
	}
	log.Info("Transaction pool stopped")
}

// SubscribeNewTxsEvent registers a subscription of NewTxsEvent and
// starts sending event to the given channel.
func (pool *TxPool) SubscribeNewTxsEvent(ch chan<- NewTxsEvent) event.Subscription {
	return pool.scope.Track(pool.txFeed.Subscribe(ch))
}

// GasPrice returns the current gas price enforced by the transaction pool.
func (pool *TxPool) GasPrice() *big.Int {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return new(big.Int).Set(pool.gasPrice)
}

// SetGasPrice updates the minimum price required by the transaction pool for a
// new transaction, and drops all transactions below this threshold.
func (pool *TxPool) SetGasPrice(price *big.Int) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.gasPrice = price

	log.Info("Transaction pool price threshold updated", "price", price)
}

// State returns the virtual managed state of the transaction pool.
func (pool *TxPool) State() *state.ManagedState {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.pendingState
}

// Stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (pool *TxPool) Stats() (int, int) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.stats()
}

// stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (pool *TxPool) stats() (int, int) {
	pending := 0
	for _, list := range pool.pending {
		if list != nil {
			pending += list.Len()
		}

	}

	return pending, 0
}

// Content retrieves the data content of the transaction pool, returning all the
// pending as well as queued transactions, grouped by account and sorted by nonce.
func (pool *TxPool) Content() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pending := make(map[common.Address]types.Transactions)
	for addr, list := range pool.pending {
		if list != nil {
			pending[addr] = list.Get()
		}
	}
	queued := make(map[common.Address]types.Transactions)
	//for addr, list := range pool.queue {
	//	if list != nil {
	//		queued[addr] = list.Get()
	//	}
	//}
	return pending, queued
}

// Pending retrieves all currently processable transactions, grouped by origin
// account and sorted by nonce. The returned transaction set is a copy and can be
// freely modified by calling code.
func (pool *TxPool) Pending() (map[common.Address]types.Transactions, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pending := make(map[common.Address]types.Transactions)
	for addr, list := range pool.pending {
		if list != nil {
			pending[addr] = list.Get()
		}
	}
	return pending, nil
}

// PendingLimited retrieves `pool.config.GlobalTxCount` processable transactions,
// grouped by origin account and stored by nonce. The returned transaction set
// is a copy and can be freely modified by calling code.
func (pool *TxPool) PendingLimited() (map[common.Address]types.Transactions, error) {
	now := time.Now()
	pool.mu.Lock()
	defer pool.mu.Unlock()

	log.Info("Pending txs before get", "txCnt", len(pool.pending))
	txCount := 0
	pending := make(map[common.Address]types.Transactions)
	for addr, list := range pool.pending {
		if list != nil {
			if list.Len() > 0 {
				pending[addr] = list.Get()
				txCount += len(pending[addr])
				if txCount >= int(pool.config.GlobalTxCount) {
					break
				}
			}
		}
	}
	log.Info("Get pending txs", "duration", time.Since(now), "txCnt", txCount)
	return pending, nil
}

// Locals retrieves the accounts currently considered local by the pool.
func (pool *TxPool) Locals() []common.Address {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.locals.flatten()
}

// local retrieves all currently known local transactions, grouped by origin
// account and sorted by nonce. The returned transaction set is a copy and can be
// freely modified by calling code.
func (pool *TxPool) local() map[common.Address]types.Transactions {
	txs := make(map[common.Address]types.Transactions)
	for addr := range pool.locals.accounts {
		if pending := pool.pending[addr]; pending != nil {
			txs[addr] = append(txs[addr], pending.Get()...)
		}
		//if queued := pool.queue[addr]; queued != nil {
		//	txs[addr] = append(txs[addr], queued.Get()...)
		//}
	}
	return txs
}

// validateTx checks whether a transaction is valid according to the consensus
// rules and adheres to some heuristic limits of the local node (price and size).
func (pool *TxPool) validateTx(tx *types.Transaction, local bool) error {
	if r, _, _, _ := rawdb.ReadReceipt(pool.db, tx.Hash()); r != nil {
		log.Error("Transaction Repeat", "old", r.TxHash.String(), "new", tx.Hash().String())
		return ErrTransactionRepeat
	}

	// Heuristic limit, reject transactions over 32KB to prevent DOS attacks
	// 32kb -> 1m
	if tx.Size() > 1024*1024 {
		return ErrOversizedData
	}
	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur if you create a transaction using the RPC.
	if tx.Value().Sign() < 0 {
		return ErrNegativeValue
	}
	// Make sure the transaction is signed properly
	from, err := types.Sender(pool.signer, tx)
	if err != nil {
		return ErrInvalidSender
	}
	// Drop non-local transactions under our own minimal accepted gas price
	local = local || pool.locals.contains(from) // account may be local even if the transaction arrived from the network

	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	if pool.currentState.GetBalance(from).Cmp(tx.Value()) < 0 {
		return ErrInsufficientFunds
	}
	if !isCallParamManager(tx.To()) && common.SysCfg.GetIsTxUseGas() && common.SysCfg.GetGasContractName() != "" {
		contractCreation := tx.To() == nil
		gas, err := IntrinsicGas(tx.Data(), contractCreation)
		log.Debug("IntrinsicGas amount", "IntrinsicGas:", gas)
		if err != nil {
			return err
		}
		if tx.Gas() < gas {
			log.Error("GasLimitTooLow", "err:", ErrIntrinsicGas)
			return ErrIntrinsicGas
		}

	}

	return nil
}

// add validates a transaction and inserts it into the non-executable queue for
// later pending promotion and execution. If the transaction is a replacement for
// an already pending or queued one, it overwrites the previous and returns this
// so outer code doesn't uselessly call promote.
//
// If a newly added transaction is marked as local, its sending account will be
// whitelisted, preventing any associated transaction from being dropped out of
// the pool due to pricing constraints.
func (pool *TxPool) add(tx *types.Transaction, local bool) (bool, error) {
	// If the transaction is already known, discard it
	hash := tx.Hash()
	if pool.all.Get(hash) != nil {
		log.Trace("Discarding already known transaction", "hash", hash)
		return false, fmt.Errorf("known transaction: %x", hash)
	}

	// If the transaction fails basic validation, discard it
	if err := pool.validateTx(tx, local); err != nil {
		log.Trace("Discarding invalid transaction", "hash", hash, "err", err)
		invalidTxCounter.Inc(1)
		return false, err
	}

	// If the transaction pool is full, discard underpriced transactions
	if uint64(pool.all.Count()) >= pool.config.GlobalSlots+pool.config.GlobalQueue {
		return false, errors.New("txpool is full")
	}
	// If the transaction is replacing an already pending one, do directly
	from, _ := types.Sender(pool.signer, tx) // already validated
	// New transaction isn't replacing a pending one, push into queue
	//replace, err := pool.enqueueTx(hash, tx)
	//if err != nil {
	//	return false, err
	//}

	pool.promoteTx(from, hash, tx)
	go pool.txFeed.Send(NewTxsEvent{types.Transactions{tx}})

	// Mark local addresses and journal local transactions
	if local {
		if !pool.locals.contains(from) {
			log.Info("Setting new local account", "address", from)
			pool.locals.add(from)
		}
	}
	pool.journalTx(from, tx)

	//log.Trace("Pooled new future transaction", "hash", hash, "from", from, "to", tx.To())
	return false, nil
}

// enqueueTx inserts a new transaction into the non-executable transaction queue.
//
// Note, this method assumes the pool lock is held!
func (pool *TxPool) enqueueTx(hash common.Hash, tx *types.Transaction) (bool, error) {
	// Try to insert the transaction into the future queue
	//from, _ := types.Sender(pool.signer, tx) // already validated

	//queue,ok := pool.queue[from]
	//if !ok{
	//	pool.queue[from] = newTxQueuedMap()
	//	queue = pool.queue[from]
	//}
	//
	//queue.Put(hash,tx)

	if pool.all.Get(hash) == nil {
		pool.all.Add(tx)
		//pool.priced.Put(tx)
	}
	//return old != nil, nil
	return false, nil
}

// journalTx adds the specified transaction to the local disk journal if it is
// deemed to have been sent from a local account.
func (pool *TxPool) journalTx(from common.Address, tx *types.Transaction) {
	// Only journal if it's enabled and the transaction is local
	if pool.journal == nil || !pool.locals.contains(from) {
		return
	}
	if err := pool.journal.insert(tx); err != nil {
		log.Warn("Failed to journal local transaction", "err", err)
	}
}

// promoteTx adds a transaction to the pending (processable) list of transactions
// and returns whether it was inserted or an older was better.
//
// Note, this method assumes the pool lock is held!
func (pool *TxPool) promoteTx(addr common.Address, hash common.Hash, tx *types.Transaction) bool {
	// Try to insert the transaction into the pending queue
	pending, ok := pool.pending[addr]
	if !ok {
		pool.pending[addr] = newTxQueuedMap()
		pending = pool.pending[addr]
	}

	pending.Put(hash, tx)

	if pool.all.Get(hash) == nil {
		pool.all.Add(tx)
	}
	// Set the potentially new pending nonce and notify any subsystems of the new tx
	pool.beats[addr] = time.Now()
	//pool.pendingState.SetNonce(addr, tx.Nonce()+1)

	return true
}

// AddLocal enqueues a single transaction into the pool if it is valid, marking
// the sender as a local one in the mean time, ensuring it goes around the local
// pricing constraints.
func (pool *TxPool) AddLocal(tx *types.Transaction) error {
	//return pool.addTxs(txs, true)
	errCh := make(chan interface{})
	txExt := &txExt{tx, !pool.config.NoLocals, errCh}
	pool.txExtBuffer <- txExt
	err := <-errCh
	if e, ok := err.(error); ok {
		return e
	} else {
		return nil
	}
}

// AddRemote enqueues a single transaction into the pool if it is valid. If the
// sender is not among the locally tracked ones, full pricing constraints will
// apply.
func (pool *TxPool) AddRemote(tx *types.Transaction) error {
	if len(pool.txExtBuffer) == txExtBufferSize {
		return errors.New("txpool is full")
	}
	//return pool.addTxs(txs, true)
	errCh := make(chan interface{}, 1)
	txExt := &txExt{tx, false, errCh}
	select {
	case <-pool.exitCh:
		return nil
	case pool.txExtBuffer <- txExt:
		return nil
	}
}

// AddLocals enqueues a batch of transactions into the pool if they are valid,
// marking the senders as a local ones in the mean time, ensuring they go around
// the local pricing constraints.
func (pool *TxPool) AddLocals(txs []*types.Transaction) []error {
	//return pool.addTxs(txs, true)
	errCh := make(chan interface{})
	txExt := &txExt{txs, !pool.config.NoLocals, errCh}
	pool.txExtBuffer <- txExt
	err := <-errCh
	if e, ok := err.([]error); ok {
		return e
	} else {
		return nil
	}
}

// get ext db
func (pool *TxPool) ExtendedDb() ethdb.Database {
	return pool.extDb
}

// AddRemotes enqueues a batch of transactions into the pool if they are valid.
// If the senders are not among the locally tracked ones, full pricing constraints
// will apply.
func (pool *TxPool) AddRemotes(txs []*types.Transaction) []error {
	//return pool.addTxs(txs, false)
	if len(pool.txExtBuffer) == txExtBufferSize {
		return []error{errors.New("txpool is full")}
	}
	errCh := make(chan interface{}, 1)
	txExt := &txExt{txs, false, errCh}
	select {
	case <-pool.exitCh:
		return nil
	case pool.txExtBuffer <- txExt:
		return nil
	}
}

func (pool *TxPool) RecoverTx(tx *types.Transaction) bool {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	from, _ := types.Sender(pool.signer, tx)
	return pool.recoverTx(tx, from, pool.locals.contains(from))
}

func (pool *TxPool) RecoverTxs(txs []*types.Transaction) []bool {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	results := make([]bool, len(txs))
	for i, tx := range txs {
		from, _ := types.Sender(pool.signer, tx)
		results[i] = pool.recoverTx(tx, from, pool.locals.contains(from))
	}
	return results
}

func (pool *TxPool) recoverTx(tx *types.Transaction, from common.Address, local bool) bool {
	// If the transaction is already known, discard it
	hash := tx.Hash()
	if pool.all.Get(hash) != nil {
		log.Trace("Discarding already known transaction", "hash", hash)
		return false
	}

	if local {
		if !pool.locals.contains(from) {
			log.Info("Setting new local account", "address", from)
			pool.locals.add(from)
		}
	}
	pool.journalTx(from, tx)
	pool.promoteTx(from, hash, tx)

	return true

}

// addTx enqueues a single transaction into the pool if it is valid.
func (pool *TxPool) addTx(tx *types.Transaction, local bool) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.addTxLocked(tx, local)
}

func (pool *TxPool) addTxLocked(tx *types.Transaction, local bool) error {

	// Try to inject the transaction and update any state
	replace, err := pool.add(tx, local)
	if err != nil {
		return err
	}
	// If we added a new transaction, run promotion checks and return
	if !replace {
		from, _ := types.Sender(pool.signer, tx) // already validated
		pool.promoteExecutables([]common.Address{from})
	}
	return nil
}

func (pool *TxPool) addTxExt(txExt *txExt) interface{} {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if tx, ok := txExt.tx.(*types.Transaction); ok {
		err := pool.addTxLocked(tx, txExt.local)
		if txExt.local && err != nil {
			from, _ := types.Sender(pool.signer, tx)
			log.Warn("Nonce tracking, add local tx to pool", "from", from, "err", err, "nonce", pool.currentState.GetNonce(from), "tx.Hash", tx.Hash(), "tx.Nonce()", tx.Nonce())
		}
		return err
	}

	if txs, ok := txExt.tx.([]*types.Transaction); ok {
		return pool.addTxsLocked(txs, txExt.local)
	}

	return nil
}

// addTxs attempts to queue a batch of transactions if they are valid.
func (pool *TxPool) addTxs(txs []*types.Transaction, local bool) []error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.addTxsLocked(txs, local)
}

// addTxsLocked attempts to queue a batch of transactions if they are valid,
// whilst assuming the transaction pool lock is already held.
func (pool *TxPool) addTxsLocked(txs []*types.Transaction, local bool) []error {
	// Add the batch of transaction, tracking the accepted ones
	dirty := make(map[common.Address]struct{})
	errs := make([]error, len(txs))

	for i, tx := range txs {
		var replace bool
		if replace, errs[i] = pool.add(tx, local); errs[i] == nil && !replace {
			from, _ := types.Sender(pool.signer, tx) // already validated
			dirty[from] = struct{}{}
		}
	}
	// Only reprocess the internal state if something was actually added
	if len(dirty) > 0 {
		addrs := make([]common.Address, 0, len(dirty))
		for addr := range dirty {
			addrs = append(addrs, addr)
		}
		pool.promoteExecutables(addrs)
	}
	return errs
}

// Status returns the status (unknown/pending/queued) of a batch of transactions
// identified by their hashes.
func (pool *TxPool) Status(hashes []common.Hash) []TxStatus {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	status := make([]TxStatus, len(hashes))
	for i, hash := range hashes {
		if tx := pool.all.Get(hash); tx != nil {
			from, _ := types.Sender(pool.signer, tx) // already validated
			if pool.pending[from] != nil {
				status[i] = TxStatusPending
			} else {
				status[i] = TxStatusQueued
			}
		}
	}
	return status
}

// Get returns a transaction if it is contained in the pool
// and nil otherwise.
func (pool *TxPool) Get(hash common.Hash) *types.Transaction {
	return pool.all.Get(hash)
}

// removeTx removes a single transaction from the queue, moving all subsequent
// transactions back to the future queue.
func (pool *TxPool) removeTx(hash common.Hash, outofbound bool) {
	// Fetch the transaction we wish to delete
	tx := pool.all.Get(hash)
	if tx == nil {
		return
	}
	addr, _ := types.Sender(pool.signer, tx) // already validated during insertion

	// Remove it from the list of known transactions
	pool.all.Remove(hash)
	if outofbound {
		//pool.priced.Removed()
	}
	// Remove the transaction from the pending lists and reset the account nonce
	if pending := pool.pending[addr]; pending != nil {
		pending.Remove(hash)
		if pending.Len() == 0 {
			delete(pool.beats, addr)
			delete(pool.pending, addr)
		}
	}
}

// promoteExecutables moves transactions that have become processable from the
// future queue to the set of pending transactions. During this process, all
// invalidated transactions (low nonce, low balance) are deleted.
func (pool *TxPool) promoteExecutables(accounts []common.Address) {
	// If the pending limit is overflown, start equalizing allowances
	pending := uint64(0)
	for _, list := range pool.pending {
		pending += uint64(list.Len())
	}
	if pending > pool.config.GlobalSlots {
		pendingBeforeCap := pending
		// Assemble a spam order to penalize large transactors first
		spammers := prque.New(nil)
		for addr, list := range pool.pending {
			// Only evict transactions from high rollers
			if !pool.locals.contains(addr) && uint64(list.Len()) > pool.config.AccountSlots {
				spammers.Push(addr, int64(list.Len()))
			}
		}
		// Gradually drop transactions from offenders
		offenders := []common.Address{}
		for pending > pool.config.GlobalSlots && !spammers.Empty() {
			// Retrieve the next offender if not local address
			offender, _ := spammers.Pop()
			offenders = append(offenders, offender.(common.Address))

			// Equalize balances until all the same or below threshold
			if len(offenders) > 1 {
				// Calculate the equalization threshold for all current offenders
				threshold := pool.pending[offender.(common.Address)].Len()

				// Iteratively reduce all offenders until below limit or threshold reached
				for pending > pool.config.GlobalSlots && pool.pending[offenders[len(offenders)-2]].Len() > threshold {
					for i := 0; i < len(offenders)-1; i++ {
						if list, ok := pool.pending[offenders[i]]; ok {
							if list.Len() == 0 {
								delete(pool.pending, offenders[i])
								continue
							}
							txs := list.Get()
							tx := txs[txs.Len()-1]
							// Drop the transaction from the global pools too
							hash := tx.Hash()
							pool.all.Remove(hash)
							list.Remove(hash)

							log.Trace("Removed fairness-exceeding pending transaction", "hash", hash)
							pending--
						}
					}
				}
			}
		}
		// If still above threshold, reduce to limit or min allowance
		if pending > pool.config.GlobalSlots && len(offenders) > 0 {
			for pending > pool.config.GlobalSlots && uint64(pool.pending[offenders[len(offenders)-1]].Len()) > pool.config.AccountSlots {
				for _, addr := range offenders {
					if list, ok := pool.pending[addr]; ok {
						if list.Len() == 0 {
							delete(pool.pending, addr)
							continue
						}
						txs := list.Get()
						tx := txs[txs.Len()-1]
						// Drop the transaction from the global pools too
						hash := tx.Hash()
						pool.all.Remove(hash)
						list.Remove(hash)
						pending--
					}
				}
			}
		}
		pendingRateLimitCounter.Inc(int64(pendingBeforeCap - pending))
	}
}

// demoteUnexecutables removes invalid and processed transactions from the pools
// executable/pending queue and any subsequent transactions that become unexecutable
// are moved back into the future queue.
func (pool *TxPool) demoteUnexecutables(txs types.Transactions) {
	// Iterate over all accounts and demote any non-executable transactions
	for addr, list := range pool.pending {
		// Drop all transactions that are deemed too old (low nonce)
		if list == nil || list.Len() == 0 {
			continue
		}
		for _, tx := range txs {
			hash := tx.Hash()
			log.Trace("Removed old pending transaction", "hash", hash)
			list.Remove(hash)
			pool.all.Remove(hash)
		}

		// drop all transactions that do not have enough balance
		for _, tx := range list.Get() {
			bal := pool.currentState.GetBalance(addr)
			if tx.Value().Cmp(bal) < 0 {
				hash := tx.Hash()
				log.Trace("Removed unpayable queued transaction", "hash", hash)
				list.Remove(hash)
				pool.all.Remove(hash)

			}
		}

		if list.Len() == 0 {
			delete(pool.pending, addr)
		}
	}
}

// addressByHeartbeat is an account address tagged with its last activity timestamp.
type addressByHeartbeat struct {
	address   common.Address
	heartbeat time.Time
}

type addressesByHeartbeat []addressByHeartbeat

func (a addressesByHeartbeat) Len() int           { return len(a) }
func (a addressesByHeartbeat) Less(i, j int) bool { return a[i].heartbeat.Before(a[j].heartbeat) }
func (a addressesByHeartbeat) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// accountSet is simply a set of addresses to check for existence, and a signer
// capable of deriving addresses from transactions.
type accountSet struct {
	accounts map[common.Address]struct{}
	signer   types.Signer
	cache    *[]common.Address
}

// newAccountSet creates a new address set with an associated signer for sender
// derivations.
func newAccountSet(signer types.Signer) *accountSet {
	return &accountSet{
		accounts: make(map[common.Address]struct{}),
		signer:   signer,
	}
}

// contains checks if a given address is contained within the set.
func (as *accountSet) contains(addr common.Address) bool {
	_, exist := as.accounts[addr]
	return exist
}

// containsTx checks if the sender of a given tx is within the set. If the sender
// cannot be derived, this method returns false.
func (as *accountSet) containsTx(tx *types.Transaction) bool {
	if addr, err := types.Sender(as.signer, tx); err == nil {
		return as.contains(addr)
	}
	return false
}

// add inserts a new address into the set to track.
func (as *accountSet) add(addr common.Address) {
	as.accounts[addr] = struct{}{}
	as.cache = nil
}

// flatten returns the list of addresses within this set, also caching it for later
// reuse. The returned slice should not be changed!
func (as *accountSet) flatten() []common.Address {
	if as.cache == nil {
		accounts := make([]common.Address, 0, len(as.accounts))
		for account := range as.accounts {
			accounts = append(accounts, account)
		}
		as.cache = &accounts
	}
	return *as.cache
}

// txLookup is used internally by TxPool to track transactions while allowing lookup without
// mutex contention.
//
// Note, although this type is properly protected against concurrent access, it
// is **not** a type that should ever be mutated or even exposed outside of the
// transaction pool, since its internal state is tightly coupled with the pools
// internal mechanisms. The sole purpose of the type is to permit out-of-bound
// peeking into the pool in TxPool.Get without having to acquire the widely scoped
// TxPool.mu mutex.
type txLookup struct {
	all  map[common.Hash]*types.Transaction
	lock sync.RWMutex
}

// newTxLookup returns a new txLookup structure.
func newTxLookup() *txLookup {
	return &txLookup{
		all: make(map[common.Hash]*types.Transaction),
	}
}

// Range calls f on each key and value present in the map.
func (t *txLookup) Range(f func(hash common.Hash, tx *types.Transaction) bool) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	for key, value := range t.all {
		if !f(key, value) {
			break
		}
	}
}

// Get returns a transaction if it exists in the lookup, or nil if not found.
func (t *txLookup) Get(hash common.Hash) *types.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.all[hash]
}

// Count returns the current number of items in the lookup.
func (t *txLookup) Count() int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return len(t.all)
}

// Add adds a transaction to the lookup.
func (t *txLookup) Add(tx *types.Transaction) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.all[tx.Hash()] = tx
}

// Remove removes a transaction from the lookup.
func (t *txLookup) Remove(hash common.Hash) {
	t.lock.Lock()
	defer t.lock.Unlock()

	delete(t.all, hash)
}
