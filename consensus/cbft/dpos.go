package cbft

import (
	"errors"
	"bytes"
	"github.com/BCOSnetwork/BCOS-Go/common"
	"github.com/BCOSnetwork/BCOS-Go/core"
	"github.com/BCOSnetwork/BCOS-Go/crypto"
	"github.com/BCOSnetwork/BCOS-Go/log"
	"github.com/BCOSnetwork/BCOS-Go/p2p/discover"
)

type dpos struct {
	rootNodeID        discover.NodeID
	primaryNodeList   []discover.NodeID
	chain             *core.BlockChain
	lastCycleBlockNum uint64
	startTimeOfEpoch  int64 // A round of consensus start time is usually the block time of the last block at the end of the last round of consensus;
							// if it is the first round, it starts from 1970.1.1.0.0.0.0. Unit: second
}

var (
	errGetInitialNodesListFailed = errors.New("get initial node list failed")
)

func convert(initialNodes []string) []discover.NodeID {
	NodeIDList := make([]discover.NodeID, 0, len(initialNodes))
	for _, value := range initialNodes {
		if nodeID, error := discover.HexID(value); error == nil {
			NodeIDList = append(NodeIDList, nodeID)
		}
	}
	return NodeIDList
}

func newDpos(rootID discover.NodeID, initialNodes []discover.NodeID) *dpos {
	dpos := &dpos{
		rootNodeID:        rootID,
		primaryNodeList:   initialNodes,
		lastCycleBlockNum: 0,
	}
	return dpos
}


// Determine whether the current node is a consensus node
func (d *dpos) IsPrimary(addr common.Address) bool {
	for _, node := range d.primaryNodeList {
		pub, err := node.Pubkey()
		if err != nil || pub == nil {
			log.Error("nodeID.ID.Pubkey error!")
		}
		address := crypto.PubkeyToAddress(*pub)

		return bytes.Equal(address[:], addr[:])
	}
	return false
}

func (d *dpos) NodeIndex(nodeID discover.NodeID) int64 {
	for idx, node := range d.primaryNodeList {
		if node == nodeID {
			return int64(idx)
		}
	}
	return int64(-1)
}

func (d *dpos) LastCycleBlockNum() uint64 {
	// Get the block height at the end of the final round of consensus
	return d.lastCycleBlockNum
}

func (d *dpos) SetLastCycleBlockNum(blockNumber uint64) {
	// Set the block height at the end of the last round of consensus
	d.lastCycleBlockNum = blockNumber
}

// Returns the current consensus node address list
/*func (b *dpos) ConsensusNodes() []discover.Node {
	return b.primaryNodeList
}
*/
// Determine whether a node is the current or previous round of election consensus nodes
/*func (b *dpos) CheckConsensusNode(id discover.NodeID) bool {
	nodes := b.ConsensusNodes()
	for _, node := range nodes {
		if node.ID == id {
			return true
		}
	}
	return false
}*/

// Determine whether the current node is the current or previous round of election consensus nodes
/*func (b *dpos) IsConsensusNode() (bool, error) {
	return true, nil
}
*/

func (d *dpos) StartTimeOfEpoch() int64 {
	return d.startTimeOfEpoch
}

func (d *dpos) SetStartTimeOfEpoch(startTimeOfEpoch int64) {
	// Set the block time at the end of the last round of consensus
	d.startTimeOfEpoch = startTimeOfEpoch
	log.Info("Set the block time at the end of the last round of consensus", "startTimeOfEpoch", startTimeOfEpoch)
}
