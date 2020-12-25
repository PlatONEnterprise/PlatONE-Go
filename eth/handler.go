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

package eth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"math"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/consensus"
	"github.com/PlatONEnetwork/PlatONE-Go/core"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/eth/downloader"
	"github.com/PlatONEnetwork/PlatONE-Go/eth/fetcher"
	"github.com/PlatONEnetwork/PlatONE-Go/ethdb"
	"github.com/PlatONEnetwork/PlatONE-Go/event"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/p2p"
	"github.com/PlatONEnetwork/PlatONE-Go/p2p/discover"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"github.com/PlatONEnetwork/PlatONE-Go/rpc"
)

const (
	softResponseLimit = 2 * 1024 * 1024 // Target maximum size of returned blocks, headers or node data.
	estHeaderRlpSize  = 500             // Approximate size of an RLP encoded block header

	// txChanSize is the size of channel listening to NewTxsEvent.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096

	defaultTxsCacheSize      = 20
	defaultBroadcastInterval = 100 * time.Millisecond
)

var (
	daoChallengeTimeout = 15 * time.Second // Time allowance for a node to reply to the DAO handshake challenge
)

// errIncompatibleConfig is returned if the requested protocols and configs are
// not compatible (low protocol version restrictions and high requirements).
var errIncompatibleConfig = errors.New("incompatible configuration")

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

type ProtocolManager struct {
	networkID uint64

	fastSync  uint32 // Flag whether fast sync is enabled (gets disabled if we already have blocks)
	acceptTxs uint32 // Flag whether we're considered synchronised (enables transaction processing)

	txpool      txPool
	blockchain  *core.BlockChain
	chainconfig *params.ChainConfig
	maxPeers    int

	downloader *downloader.Downloader
	fetcher    *fetcher.Fetcher
	txFetcher  *fetcher.TxFetcher
	peers      *peerSet

	SubProtocols []p2p.Protocol

	eventMux      *event.TypeMux
	txsCh         chan core.NewTxsEvent
	txsCache      []*types.Transaction
	txsSub        event.Subscription
	minedBlockSub *event.TypeMuxSubscription

	prepareMinedBlockSub *event.TypeMuxSubscription
	blockSignatureSub    *event.TypeMuxSubscription

	// channels for fetcher, syncer, txsyncLoop
	newPeerCh   chan *peer
	txsyncCh    chan *txsync
	quitSync    chan struct{}
	noMorePeers chan struct{}

	// wait group is used for graceful shutdowns during downloading
	// and processing
	wg sync.WaitGroup

	engine consensus.Engine
}

// NewProtocolManager returns a new Ethereum sub protocol manager. The Ethereum sub protocol manages peers capable
// with the Ethereum network.
func NewProtocolManager(config *params.ChainConfig, mode downloader.SyncMode, networkID uint64, mux *event.TypeMux,
	txpool txPool, engine consensus.Engine, blockchain *core.BlockChain, chaindb ethdb.Database) (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		networkID:   networkID,
		eventMux:    mux,
		txpool:      txpool,
		blockchain:  blockchain,
		chainconfig: config,
		peers:       newPeerSet(),
		newPeerCh:   make(chan *peer),
		noMorePeers: make(chan struct{}),
		txsyncCh:    make(chan *txsync),
		quitSync:    make(chan struct{}),
		engine:      engine,
	}

	if handler, ok := manager.engine.(consensus.Handler); ok {
		handler.SetBroadcaster(manager)
	}

	// Figure out whether to allow fast sync or not
	if mode == downloader.FastSync && blockchain.CurrentBlock().NumberU64() > 0 {
		log.Warn("Blockchain not empty, fast sync disabled")
		mode = downloader.FullSync
	}
	if mode == downloader.FastSync {
		manager.fastSync = uint32(1)
	}
	// Initiate a sub-protocol for every implemented version we can handle
	manager.SubProtocols = make([]p2p.Protocol, 0, len(ProtocolVersions))
	for i, version := range ProtocolVersions {
		// Compatible; initialise the sub-protocol
		version := version // Closure for the run
		manager.SubProtocols = append(manager.SubProtocols, p2p.Protocol{
			Name:    ProtocolNameArr[i],
			Version: version,
			Length:  ProtocolLengths[i],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				peer := manager.newPeer(int(version), p, rw)
				select {
				case manager.newPeerCh <- peer:
					manager.wg.Add(1)
					defer manager.wg.Done()
					return manager.handle(peer)
				case <-manager.quitSync:
					return p2p.DiscQuitting
				}
			},
			NodeInfo: func() interface{} {
				return manager.NodeInfo()
			},
			PeerInfo: func(id discover.NodeID) interface{} {
				if p := manager.peers.Peer(fmt.Sprintf("%x", id[:8])); p != nil {
					return p.Info()
				}
				return nil
			},
			UpdatePeer: func(info *common.NodeInfo) {
				if p := manager.peers.Peer(info.PublicKey); p != nil {
					p.setTypes(info.Types)
				}
			},
		})
	}
	if len(manager.SubProtocols) == 0 {
		return nil, errIncompatibleConfig
	}
	// Construct the different synchronisation mechanisms
	manager.downloader = downloader.New(mode, chaindb, manager.eventMux, blockchain, nil, manager.removePeer)

	validator := func(header *types.Header) error {
		return engine.VerifyHeader(blockchain, header, true)
	}
	heighter := func() uint64 {
		return blockchain.CurrentBlock().NumberU64()
	}
	inserter := func(blocks types.Blocks) (int, error) {
		// If fast sync is running, deny importing weird blocks
		if atomic.LoadUint32(&manager.fastSync) == 1 {
			log.Warn("Discarded bad propagated block", "number", blocks[0].Number(), "hash", blocks[0].Hash())
			return 0, nil
		}
		atomic.StoreUint32(&manager.acceptTxs, 1) // Mark initial sync done on any fetcher import
		return manager.blockchain.InsertChain(blocks)
	}

	manager.fetcher = fetcher.New(blockchain.GetBlockByHash, validator, manager.BroadcastBlock, heighter, inserter, manager.removePeer)

	fetchTx := func(peer string, hashes []common.Hash) error {
		p := manager.peers.Peer(peer)
		if p == nil {
			return errors.New("unknown peer")
		}
		return p.RequestTxs(hashes)
	}
	manager.txFetcher = fetcher.NewTxFetcher(txpool.Has, txpool.AddRemotes, fetchTx)

	return manager, nil
}

func (pm *ProtocolManager) removePeer(id string) {
	// Short circuit if the peer was already removed
	peer := pm.peers.Peer(id)
	if peer == nil {
		return
	}
	log.Debug("Removing Ethereum peer", "peer", id)

	// Unregister the peer from the downloader and Ethereum peer set
	pm.downloader.UnregisterPeer(id)
	pm.txFetcher.Drop(id)
	if err := pm.peers.Unregister(id); err != nil {
		log.Error("Peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
}

func (pm *ProtocolManager) Start(maxPeers int) {
	pm.maxPeers = maxPeers

	// broadcast transactions
	pm.txsCh = make(chan core.NewTxsEvent, txChanSize)
	pm.txsSub = pm.txpool.SubscribeNewTxsEvent(pm.txsCh)
	go pm.txBroadcastLoop()

	// broadcast mined blocks
	pm.minedBlockSub = pm.eventMux.Subscribe(core.NewMinedBlockEvent{})
	// broadcast prepare mined blocks
	pm.prepareMinedBlockSub = pm.eventMux.Subscribe(core.PrepareMinedBlockEvent{})
	go pm.minedBroadcastLoop()
	go pm.prepareMinedBlockcastLoop()

	// start sync handlers
	go pm.syncer()
	go pm.txsyncLoop()
}

func (pm *ProtocolManager) Stop() {
	log.Info("Stopping Ethereum protocol")

	pm.txsSub.Unsubscribe()        // quits txBroadcastLoop
	pm.minedBlockSub.Unsubscribe() // quits blockBroadcastLoop

	// Quit the sync loop.
	// After this send has completed, no new peers will be accepted.
	pm.noMorePeers <- struct{}{}

	// Quit fetcher, txsyncLoop.
	close(pm.quitSync)

	// Disconnect existing sessions.
	// This also closes the gate for any new registrations on the peer set.
	// sessions which are already established but not added to pm.peers yet
	// will exit when they try to register.
	pm.peers.Close()

	// Wait for all peer handler goroutines and the loops to come down.
	pm.wg.Wait()

	log.Info("Ethereum protocol stopped")
}

func (pm *ProtocolManager) newPeer(pv int, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return newPeer(pv, p, newMeteredMsgWriter(rw))
}

// handle is the callback invoked to manage the life cycle of an eth peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {
	// Ignore maxPeers if this is a trusted peer
	if pm.peers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted {
		return p2p.DiscTooManyPeers
	}
	p.Log().Debug("Ethereum peer connected", "name", p.Name())

	// Execute the Ethereum handshake
	var (
		genesis = pm.blockchain.Genesis()
		head    = pm.blockchain.CurrentHeader()
		hash    = head.Hash()
	)
	if err := p.Handshake(pm.networkID, head.Number, hash, genesis.Hash()); err != nil {
		p.Log().Debug("Ethereum handshake failed", "err", err)
		return err
	}
	if pm.blockchain.CurrentHeader().Number.Cmp(big.NewInt(0)) == 0 &&
		p.GetReplayParam().Pivot != 0 &&
		common.SysCfg.ReplayParam.Pivot == 0 {

		b, err := rlp.EncodeToBytes(p.GetReplayParam().Pivot)
		if err != nil {
			return err
		}
		pm.blockchain.Put(common.Sys_pivot_key, b)

		b, err = json.Marshal(p.GetReplayParam().OldSysContracts)
		if err != nil {
			return err
		}
		b, err = rlp.EncodeToBytes(b)
		if err != nil {
			return err
		}
		pm.blockchain.Put(common.Sys_old_system_contract_key, b)

		b, err = rlp.EncodeToBytes(p.GetReplayParam().OldSuperAdmin)
		if err != nil {
			return err
		}
		pm.blockchain.Put(common.Sys_old_super_admin_key, b)

		core.InitBlockReplayConfig(pm.blockchain, common.SysCfg)
	}

	if rw, ok := p.rw.(*meteredMsgReadWriter); ok {
		rw.Init(p.version)
	}
	// Register the peer locally
	if err := pm.peers.Register(p); err != nil {
		p.Log().Error("Ethereum peer registration failed", "err", err)
		return err
	}
	defer pm.removePeer(p.id)

	// Register the peer in the downloader. If the downloader considers it banned, we disconnect
	if err := pm.downloader.RegisterPeer(p.id, p.version, p); err != nil {
		return err
	}
	// Propagate existing transactions. new transactions appearing
	// after this will be sent via broadcasts.
	if p.IsConsensus() {
		pm.syncTransactionHashes(p)
	}

	// main loop. handle incoming messages.
	for {
		if err := pm.handleMsg(p); err != nil {
			p.Log().Error("Ethereum message handling failed", "err", err)
			return err
		}
	}
}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (pm *ProtocolManager) handleMsg(p *peer) error {
	// Read the next message from the remote peer, and ensure it's fully consumed
	msg, err := p.rw.ReadMsg()
	if err != nil {
		p.Log().Error("read peer message error", "err", err)
		return err
	}
	if msg.Size > ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	defer msg.Discard()

	if handler, ok := pm.engine.(consensus.Handler); ok {
		pubKey, err := p.ID().Pubkey()
		if err != nil {
			return err
		}
		addr := crypto.PubkeyToAddress(*pubKey)
		handled, err := handler.HandleMsg(addr, msg)
		if handled {
			return err
		}
	}
	// Handle the message depending on its contents
	switch {
	case msg.Code == StatusMsg:
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")

	// Block header query, collect the requested headers and reply
	case msg.Code == GetBlockHeadersMsg:
		if pm.isUnNormalBootNodesAtPeer(p) {
			log.Warn("the bootNode is not a normal node. cancel exemption", "bootNode id: ", pm.peers.Peer(p.id).Peer.ID().String())
			break
		}
		// Decode the complex header query
		var query getBlockHeadersData
		if err := msg.Decode(&query); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		hashMode := query.Origin.Hash != (common.Hash{})
		first := true
		maxNonCanonical := uint64(100)

		// Gather headers until the fetch or network limits is reached
		var (
			bytes   common.StorageSize
			headers []*types.Header
			unknown bool
		)
		for !unknown && len(headers) < int(query.Amount) && bytes < softResponseLimit && len(headers) < downloader.MaxHeaderFetch {
			// Retrieve the next header satisfying the query
			var origin *types.Header
			if hashMode {
				if first {
					first = false
					origin = pm.blockchain.GetHeaderByHash(query.Origin.Hash)
					if origin != nil {
						query.Origin.Number = origin.Number.Uint64()
					}
				} else {
					origin = pm.blockchain.GetHeader(query.Origin.Hash, query.Origin.Number)
				}
			} else {
				origin = pm.blockchain.GetHeaderByNumber(query.Origin.Number)
			}
			if origin == nil {
				break
			}
			headers = append(headers, origin)
			bytes += estHeaderRlpSize

			// Advance to the next header of the query
			switch {
			case hashMode && query.Reverse:
				// Hash based traversal towards the genesis block
				ancestor := query.Skip + 1
				if ancestor == 0 {
					unknown = true
				} else {
					query.Origin.Hash, query.Origin.Number = pm.blockchain.GetAncestor(query.Origin.Hash, query.Origin.Number, ancestor, &maxNonCanonical)
					unknown = (query.Origin.Hash == common.Hash{})
				}
			case hashMode && !query.Reverse:
				// Hash based traversal towards the leaf block
				var (
					current = origin.Number.Uint64()
					next    = current + query.Skip + 1
				)
				if next <= current {
					infos, _ := json.MarshalIndent(p.Peer.Info(), "", "  ")
					p.Log().Warn("GetBlockHeaders skip overflow attack", "current", current, "skip", query.Skip, "next", next, "attacker", infos)
					unknown = true
				} else {
					if header := pm.blockchain.GetHeaderByNumber(next); header != nil {
						nextHash := header.Hash()
						expOldHash, _ := pm.blockchain.GetAncestor(nextHash, next, query.Skip+1, &maxNonCanonical)
						if expOldHash == query.Origin.Hash {
							query.Origin.Hash, query.Origin.Number = nextHash, next
						} else {
							unknown = true
						}
					} else {
						unknown = true
					}
				}
			case query.Reverse:
				// Number based traversal towards the genesis block
				if query.Origin.Number >= query.Skip+1 {
					query.Origin.Number -= query.Skip + 1
				} else {
					unknown = true
				}

			case !query.Reverse:
				// Number based traversal towards the leaf block
				query.Origin.Number += query.Skip + 1
			}
		}
		return p.SendBlockHeaders(headers)

	case msg.Code == BlockHeadersMsg:
		// A batch of headers arrived to one of our previous requests
		var headers []*types.Header
		if err := msg.Decode(&headers); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}

		// Filter out any explicitly requested headers, deliver the rest to the downloader
		filter := len(headers) == 1
		if filter {
			// Irrelevant of the fork checks, send the header to the fetcher just in case
			headers = pm.fetcher.FilterHeaders(p.id, headers, time.Now())
		}
		if len(headers) > 0 || !filter {
			err := pm.downloader.DeliverHeaders(p.id, headers)
			if err != nil {
				log.Debug("Failed to deliver headers", "err", err)
			}
		}

	case msg.Code == GetBlockBodiesMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather blocks until the fetch or network limits is reached
		var (
			hash   common.Hash
			bytes  int
			bodies []rlp.RawValue
		)
		for bytes < softResponseLimit && len(bodies) < downloader.MaxBlockFetch {
			// Retrieve the hash of the next block
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested block body, stopping if enough was found
			if data := pm.blockchain.GetBodyRLP(hash); len(data) != 0 {
				bodies = append(bodies, data)
				bytes += len(data)
			}
		}
		return p.SendBlockBodiesRLP(bodies)

	case msg.Code == BlockBodiesMsg:
		// A batch of block bodies arrived to one of our previous requests
		var request blockBodiesData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver them all to the downloader for queuing
		transactions := make([][]*types.Transaction, len(request))

		for i, body := range request {
			transactions[i] = body.Transactions
		}
		// Filter out any explicitly requested bodies, deliver the rest to the downloader
		filter := len(transactions) > 0
		if filter {
			transactions = pm.fetcher.FilterBodies(p.id, transactions, time.Now())
		}
		if len(transactions) > 0 || !filter {
			err := pm.downloader.DeliverBodies(p.id, transactions)
			if err != nil {
				log.Debug("Failed to deliver bodies", "err", err)
			}
		}

	case msg.Code == GetNodeDataMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather state data until the fetch or network limits is reached
		var (
			hash  common.Hash
			bytes int
			data  [][]byte
		)
		for bytes < softResponseLimit && len(data) < downloader.MaxStateFetch {
			// Retrieve the hash of the next state entry
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested state entry, stopping if enough was found
			if entry, err := pm.blockchain.TrieNode(hash); err == nil {
				data = append(data, entry)
				bytes += len(entry)
			}
		}
		return p.SendNodeData(data)

	case msg.Code == NodeDataMsg:
		// A batch of node state data arrived to one of our previous requests
		var data [][]byte
		if err := msg.Decode(&data); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver all to the downloader
		if err := pm.downloader.DeliverNodeData(p.id, data); err != nil {
			log.Debug("Failed to deliver node state data", "err", err)
		}

	case msg.Code == GetReceiptsMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather state data until the fetch or network limits is reached
		var (
			hash     common.Hash
			bytes    int
			receipts []rlp.RawValue
		)
		for bytes < softResponseLimit && len(receipts) < downloader.MaxReceiptFetch {
			// Retrieve the hash of the next block
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested block's receipts, skipping if unknown to us
			results := pm.blockchain.GetReceiptsByHash(hash)
			if results == nil {
				if header := pm.blockchain.GetHeaderByHash(hash); header == nil || header.ReceiptHash != types.EmptyRootHash {
					continue
				}
			}
			// If known, encode and queue for response packet
			if encoded, err := rlp.EncodeToBytes(results); err != nil {
				log.Error("Failed to encode receipt", "err", err)
			} else {
				receipts = append(receipts, encoded)
				bytes += len(encoded)
			}
		}
		return p.SendReceiptsRLP(receipts)

	case msg.Code == ReceiptsMsg:
		// A batch of receipts arrived to one of our previous requests
		var receipts [][]*types.Receipt
		if err := msg.Decode(&receipts); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver all to the downloader
		if err := pm.downloader.DeliverReceipts(p.id, receipts); err != nil {
			log.Debug("Failed to deliver receipts", "err", err)
		}

	case msg.Code == NewBlockHashesMsg:
		var announces newBlockHashesData
		if err := msg.Decode(&announces); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}

		// Mark the hashes as present at the remote node
		for _, block := range announces {
			p.MarkBlock(block.Hash)
			log.Debug("Received a message[NewBlockHashesMsg]------------", "GoRoutineID", common.CurrentGoRoutineID(), "receiveAt", msg.ReceivedAt.Unix(), "peerId", p.id, "hash", block.Hash, "number", block.Number)
		}
		// Schedule all the unknown hashes for retrieval
		unknown := make(newBlockHashesData, 0, len(announces))
		for _, block := range announces {
			if !pm.blockchain.HasBlock(block.Hash, block.Number) {
				unknown = append(unknown, block)
			}
		}
		for _, block := range unknown {
			pm.fetcher.Notify(p.id, block.Hash, block.Number, time.Now(), p.RequestOneHeader, p.RequestBodies)
		}

	case msg.Code == NewBlockMsg:
		// Retrieve and decode the propagated block
		var request newBlockData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		request.Block.ReceivedAt = msg.ReceivedAt
		request.Block.ReceivedFrom = p

		log.Debug("Received a message[NewBlockMsg]------------", "GoRoutineID", common.CurrentGoRoutineID(), "receiveAt", request.Block.ReceivedAt.Unix(), "peerId", p.id, "hash", request.Block.Hash(), "number", request.Block.NumberU64())

		// Mark the peer as owning the block and schedule it for import
		p.MarkBlock(request.Block.Hash())
		pm.fetcher.Enqueue(p.id, request.Block)

		if !p2p.BootNodesNotExempt {
			if hDiff := request.Block.NumberU64() - pm.blockchain.CurrentBlock().NumberU64(); hDiff == 1 || hDiff == 0 {
				log.Info("enable boot nodes not exemption: ", "local", "at new block msg")
				p2p.BootNodesNotExempt = true
			}
		}
		// Assuming the block is importable by the peer, but possibly not yet done so,
		// calculate the head hash and block number that the peer truly must have.
		var (
			trueHead = request.Block.ParentHash()
			trueBn   = new(big.Int).Sub(request.Block.Number(), big.NewInt(1))
		)
		// Update the peers block number if better than the previous

		if _, bn := p.Head(); trueBn.Cmp(bn) > 0 {
			p.SetHead(trueHead, trueBn)

			// Schedule a sync if above ours. Note, this will not fire a sync for a gap of
			// a singe block (as the true TD is below the propagated block), however this
			// scenario should easily be covered by the fetcher.
			currentBlock := pm.blockchain.CurrentBlock()
			if trueBn.Cmp(currentBlock.Number()) > 0 {
				go pm.synchronise(p)
			}
		}

	case msg.Code == TxHashesMsg:
		// New transaction announcement arrived, make sure we have
		// a valid and fresh chain to handle them
		if atomic.LoadUint32(&pm.acceptTxs) == 0 {
			break
		}
		var hashes []common.Hash
		if err := msg.Decode(&hashes); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Schedule all the unknown hashes for retrieval
		for _, hash := range hashes {
			p.MarkTransaction(hash)
		}
		pm.txFetcher.Notify(p.id, hashes)

	case msg.Code == GetPooledTxMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather transactions until the fetch or network limits is reached
		var (
			hash   common.Hash
			bytes  int
			hashes []common.Hash
			txs    []rlp.RawValue
		)
		for bytes < softResponseLimit {
			// Retrieve the hash of the next block
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested transaction, skipping if unknown to us
			tx := pm.txpool.Get(hash)
			if tx == nil {
				continue
			}
			// If known, encode and queue for response packet
			if encoded, err := rlp.EncodeToBytes(tx); err != nil {
				log.Error("Failed to encode transaction", "err", err)
			} else {
				hashes = append(hashes, hash)
				txs = append(txs, encoded)
				bytes += len(encoded)
			}
		}
		return p.SendPooledTransactionsRLP(hashes, txs)

	case msg.Code == TxMsg || msg.Code == PooledTxMsg:
		// Transactions arrived, make sure we have a valid and fresh chain to handle them
		if atomic.LoadUint32(&pm.acceptTxs) == 0 {
			break
		}
		// Transactions can be processed, parse all of them and deliver to the pool
		var txs []*types.Transaction
		if err := msg.Decode(&txs); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		for i, tx := range txs {
			// Validate and mark the remote transaction
			if tx == nil {
				return errResp(ErrDecode, "transaction %d is nil", i)
			}

			// This node is the node that receives the broadcast transaction
			strTx := tx.Hash().String()
			tx.RouterMark()
			rpc.MonitorWriteData(rpc.TransactionReceiveTime, strTx, "", pm.txpool.ExtendedDb())
			rpc.MonitorWriteData(rpc.TransactionReceiveNode, strTx, "false", pm.txpool.ExtendedDb())
			p.MarkTransaction(tx.Hash())
		}
		go pm.txFetcher.Enqueue(p.id, txs, msg.Code == PooledTxMsg)

	case msg.Code == PrepareBlockMsg:
		// Retrieve and decode the propagated block
		var request prepareBlockData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		log.Debug("Received a broadcast message[PrepareBlockMsg]------------", "GoRoutineID", common.CurrentGoRoutineID(), "peerId", p.id, "hash", request.Block.Hash(), "number", request.Block.NumberU64())

		request.Block.ReceivedAt = msg.ReceivedAt
		request.Block.ReceivedFrom = p

		// Preliminary check block
		if err := pm.engine.VerifyHeader(pm.blockchain, request.Block.Header(), true); err != nil {
			log.Error("Failed to VerifyHeader in PrepareBlockMsg,discard this msg", "err", err)
			return nil
		}
		if pm.blockchain.HasBlock(request.Block.Hash(), request.Block.NumberU64()) {
			log.Warn("Block already in blockchain,discard this msg", "err", err)
			return nil
		}
	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

func (pm *ProtocolManager) Enqueue(id string, block *types.Block) {
	pm.fetcher.Enqueue(id, block)
}

// BroadcastBlock will either propagate a block to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (pm *ProtocolManager) BroadcastBlock(block *types.Block, propagate bool) {
	hash := block.Hash()
	peers := pm.peers.PeersWithoutBlock(hash)
	//var peers []*peer
	//if _, ok := pm.engine.(consensus.Bft); ok {
	//	peers = pm.peers.PeersWithoutConsensus(pm.engine)
	//} else {
	//	peers = pm.peers.PeersWithoutBlock(hash)
	//}

	// If propagation is requested, send to a subset of the peer
	if propagate {
		// Calculate the TD of the block (it's not imported yet, so block.Td is not valid)
		if parent := pm.blockchain.GetBlock(block.ParentHash(), block.NumberU64()-1); parent != nil {
		} else {
			log.Warn("Propagating dangling block", "number", block.Number(), "hash", hash)
			return
		}
		// Send the block to a subset of our peers
		transfer := peers[:int(math.Sqrt(float64(len(peers))))]
		for _, peer := range transfer {
			peer.AsyncSendNewBlock(block)
		}
		log.Trace("Propagated block", "hash", fmt.Sprintf("%x", hash[:log.LogHashLen]), "blockNumber", block.Number(), "recipients", len(transfer), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
		return
	}
	// Otherwise if the block is indeed in out own chain, announce it
	if pm.blockchain.HasBlock(hash, block.NumberU64()) {
		for _, peer := range peers {
			peer.AsyncSendNewBlockHash(block)
		}
		log.Trace("Announced block", "hash", fmt.Sprintf("%x", hash[:log.LogHashLen]), "blockNumber", block.Number(), "recipients", len(peers), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
	}
}

func (pm *ProtocolManager) MulticastConsensus(a interface{}) {
	// Consensus node peer
	peers := pm.peers.PeersWithConsensus(pm.engine)
	if peers == nil || len(peers) <= 0 {
		log.Error("consensus peers is empty")
	}

	if block, ok := a.(*types.Block); ok {
		for _, peer := range peers {
			log.Warn("~ Send a broadcast message[PrepareBlockMsg]------------",
				"peerId", peer.id, "Hash", block.Hash(), "Number", block.Number())
			peer.AsyncSendPrepareBlock(block)
		}
	}
}

// BroadcastTxs will propagate a batch of transactions to all peers which are not known to
// already have the given transaction.
func (pm *ProtocolManager) BroadcastTxs(txs types.Transactions) {
	var txset = make(map[*peer]types.Transactions)
	var hashSet = make(map[*peer][]common.Hash)

	// Broadcast transactions to a batch of peers not knowing about it
	consensusPeers := pm.peers.ConsensusPeers()
	for _, tx := range txs {
		txHash := tx.Hash()
		if tx.FromRemote() {
			transfer := consensusPeers[:int(math.Sqrt(float64(len(consensusPeers))))]
			for _, peer := range transfer {
				if !peer.knownTxs.Contains(txHash) {
					hashSet[peer] = append(hashSet[peer], tx.Hash())
				}
			}
			log.Trace("Broadcast transaction", "hash", fmt.Sprintf("%x", txHash[:log.LogHashLen]), "recipients", len(transfer))
		} else {
			for _, peer := range consensusPeers {
				if !peer.knownTxs.Contains(txHash) {
					txset[peer] = append(txset[peer], tx)
				}
			}
			log.Trace("Broadcast transaction", "hash", fmt.Sprintf("%x", txHash[:log.LogHashLen]), "recipients", len(consensusPeers))
		}
	}

	// FIXME include this again: peers = peers[:int(math.Sqrt(float64(len(peers))))]
	for peer, txs := range txset {
		peer.AsyncSendTransactions(txs)
	}
	for peer, hashes := range hashSet {
		peer.AsyncSendPooledTransactionHashes(hashes)
	}
}

// Mined broadcast loop
func (pm *ProtocolManager) minedBroadcastLoop() {
	// automatically stops if unsubscribe
	for obj := range pm.minedBlockSub.Chan() {
		if ev, ok := obj.Data.(core.NewMinedBlockEvent); ok {
			pm.BroadcastBlock(ev.Block, true)  // First propagate block to peers
			pm.BroadcastBlock(ev.Block, false) // Only then announce to the rest
		}
	}
	/*
		for {
			select {
			case event :=  <- pm.minedBlockSub.Chan():
				if ev, ok := event.Data.(core.NewMinedBlockEvent); ok {
					pm.BroadcastBlock(ev.Block, true)  // First propagate block to peers
					pm.BroadcastBlock(ev.Block, false) // Only then announce to the rest
				}
			case event :=  <- pm.prepareMinedBlockSub.Chan():
				if ev, ok := event.Data.(core.PrepareMinedBlockEvent); ok {
					pm.MulticastConsensus(ev.Block)  // propagate block to consensus peers
				}
			case event :=  <- pm.blockSignatureSub.Chan():
				if ev, ok := event.Data.(core.BlockSignatureEvent); ok {
					pm.MulticastConsensus(ev.BlockSignature)  // propagate blockSignature to consensus peers
				}
			}
		}
	*/
}

func (pm *ProtocolManager) prepareMinedBlockcastLoop() {
	for obj := range pm.prepareMinedBlockSub.Chan() {
		if ev, ok := obj.Data.(core.PrepareMinedBlockEvent); ok {
			pm.MulticastConsensus(ev.Block) // propagate block to consensus peers
		}
	}
}

func (pm *ProtocolManager) txBroadcastLoop() {
	timer := time.NewTimer(defaultBroadcastInterval)

	for {
		select {
		case event := <-pm.txsCh:
			pm.txsCache = append(pm.txsCache, event.Txs...)
			if len(pm.txsCache) >= defaultTxsCacheSize {
				log.Trace("broadcast txs", "count", len(pm.txsCache))
				pm.BroadcastTxs(pm.txsCache)
				pm.txsCache = make([]*types.Transaction, 0)
				timer.Reset(defaultBroadcastInterval)
			}
		case <-timer.C:
			if len(pm.txsCache) > 0 {
				log.Trace("broadcast txs", "count", len(pm.txsCache))
				pm.BroadcastTxs(pm.txsCache)
				pm.txsCache = make([]*types.Transaction, 0)
			}
			timer.Reset(defaultBroadcastInterval)

		// Err() channel will be closed when unsubscribing.
		case <-pm.txsSub.Err():
			return
		}
	}
}

// NodeInfo represents a short summary of the Ethereum sub-protocol metadata
// known about the host peer.
type NodeInfo struct {
	Network uint64              `json:"network"` // Ethereum network ID (1=Frontier, 2=Morden, Ropsten=3, Rinkeby=4)
	Genesis common.Hash         `json:"genesis"` // SHA3 hash of the host's genesis block
	Config  *params.ChainConfig `json:"config"`  // Chain configuration for the fork rules
	Head    common.Hash         `json:"head"`    // SHA3 hash of the host's best owned block
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (pm *ProtocolManager) NodeInfo() *NodeInfo {
	currentBlock := pm.blockchain.CurrentBlock()
	return &NodeInfo{
		Network: pm.networkID,
		Genesis: pm.blockchain.Genesis().Hash(),
		Config:  pm.blockchain.Config(),
		Head:    currentBlock.Hash(),
	}
}

func (self *ProtocolManager) FindPeers(targets map[common.Address]bool) map[common.Address]consensus.Peer {
	m := make(map[common.Address]consensus.Peer)
	for _, p := range self.peers.Peers() {
		pubKey, err := p.ID().Pubkey()
		if err != nil {
			continue
		}
		addr := crypto.PubkeyToAddress(*pubKey)
		if targets[addr] {
			m[addr] = p
		}
	}
	return m
}
