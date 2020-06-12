package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/vm"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

const (
	NODE_STATUS_NORMAL  = 1
	NODE_STATUS_DELETED = 2

	NODE_TYPE_OBSERVER  = 0
	NODE_TYPE_VALIDATOR = 1
)

type NodeInfo struct {
	Name  string `json:"name,omitempty"`  //全网唯一，不能重复。所有接口均以此为主键。 这个名称意义是？
	Owner string `json:"owner,omitempty"` //没有用到？删除？
	Desc  string `json:"desc,omitempty"`  //没有用到？删除？
	Types int32  `json:"type,omitempty"`  // 0:观察者节点；1:共识节点
	// status 1为正常节点, 2为删除节点
	Status     int32  `json:"status,omitempty"`
	ExternalIP string `json:"externalIP,omitempty"` //没有用到？删除？
	InternalIP string `json:"internalIP,omitempty"`
	PublicKey  string `json:"publicKey,omitempty"` //节点公钥，全网唯一，不能重复
	RpcPort    int32  `json:"rpcPort,omitempty"`   //没有用到？删除？
	P2pPort    int32  `json:"p2pPort,omitempty"`
	// delay set validatorSet
	DelayNum uint64 `json:"delayNum,omitempty"` //共识节点延迟设置的区块高度 (可选, 默认实时设置)
}

type SCNode struct {
	stateDB StateDB
	address common.Address
}

func NewSCNode() *SCNode {
	return &SCNode{address: vm.NODE_MANAGEMENT_ADDRESS}
}

func (n *SCNode) SetState(key, value []byte) {
	n.stateDB.SetState(n.address, key, value)
}

func (n *SCNode) GetState(key []byte) []byte {
	return n.stateDB.GetState(n.address, key)
}

func (n *SCNode) RequiredGas(input []byte) uint64 {
	if IsEmpty(input) {
		return 0
	}
	return params.SCNodeGas
}

func (n *SCNode) Run(input []byte) ([]byte, error) {
	return execSC(input, n.AllExportFns())
}

func (n *SCNode) Add(node *NodeInfo) int {
	n.doAdd(node)
	panic("not implemented")
}

func (n *SCNode) doAdd(node *NodeInfo) error {
	panic("not implemented")
}

func (n *SCNode) Update(node *NodeInfo) int {
	n.doUpdate(node)
	panic("not implemented")
}

func (n *SCNode) doUpdate(node *NodeInfo) error {
	panic("not implemented")
}

func (n *SCNode) GetAllNodes() {
	n.doGetAllNodes()
	panic("not implemented")
}

func (n *SCNode) doGetAllNodes() []*NodeInfo {
	panic("not implemented")
}

//for c++ interface: validJoinNode
func (n *SCNode) IsPublicKeyExist(pk string) int {
	n.doIsPublicKeyExist(pk)
	panic("not implemented")
}

func (n *SCNode) doIsPublicKeyExist(pk string) bool {
	panic("not implemented")
}

type ENode struct {
	PublicKey string
	IP        string
	Port      string
}

//enode format: "enode://" + publicKey + "@" + ip:port
//for c++ sc interface: getNormalEnodeNodes
//C++这个接口返回的值是自定义格式，是不是应该采用通用的格式？都采用json会比较好？
func (n *SCNode) GetENodesOfAllNormalNodes() string {
	n.doGetENodesOfAllNormalNodes()
	panic("not implemented")
}

func (n *SCNode) doGetENodesOfAllNormalNodes() []*ENode {
	panic("not implemented")
}

func (n *SCNode) GetENodesOfAllDeletedNodes() string {
	n.doGetENodesOfAllDeletedNodes()

	panic("not implemented")
}

func (n *SCNode) doGetENodesOfAllDeletedNodes() []*ENode {
	panic("not implemented")
}

//TODO
//需要支持像C++合约一样组合搜索吗
func (n *SCNode) GetNodes(query *NodeInfo) string {
	panic("not implemented")
}

func (n *SCNode) doGetNodes(query *NodeInfo) []*NodeInfo {
	panic("not implemented")
}

func (n *SCNode) NodesNum(query *NodeInfo) uint64 {
	panic("not implemented")
}

//for access control
func (n *SCNode) AllExportFns() SCExportFns {
	return SCExportFns{
		"add":                  n.Add,
		"update":               n.Update,
		"getAllNodes":          n.GetAllNodes,
		"getNodes":             n.GetNodes,
		"getNormalEnodeNodes":  n.GetENodesOfAllNormalNodes(),
		"getDeletedEnodeNodes": n.GetENodesOfAllDeletedNodes,
		"validJoinNode":        n.IsPublicKeyExist,
		"nodesNum":             n.NodesNum,
	}
}
