package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

type SCNodeWrapper struct {
	*SCNode
}

func NewSCNodeWrapper() *SCNodeWrapper {
	return &SCNodeWrapper{NewSCNode()}
}

func (n *SCNodeWrapper) RequiredGas(input []byte) uint64 {
	if IsEmpty(input) {
		return 0
	}
	return params.SCNodeGas
}

func (n *SCNodeWrapper) Run(input []byte) ([]byte, error) {
	return execSC(input, n.AllExportFns())
}

func (n *SCNodeWrapper) Add(node *NodeInfo) int {
	n.Add(node)
	panic("not implemented")
}

func (n *SCNodeWrapper) Update(name string, node *NodeInfo) int {
	n.Update(name, node)
	panic("not implemented")
}

func (n *SCNodeWrapper) GetAllNodes() string {
	n.GetAllNodes()
	panic("not implemented")
}

//for c++ interface: validJoinNode
func (n *SCNodeWrapper) IsPublicKeyExist(pk string) int {
	n.IsPublicKeyExist(pk)
	panic("not implemented")
}

//enode format: "enode://" + publicKey + "@" + ip:port
//for c++ sc interface: getNormalEnodeNodes
//C++这个接口返回的值是自定义格式，是不是应该采用通用的格式？都采用json会比较好？
func (n *SCNodeWrapper) GetENodesOfAllNormalNodes() string {
	n.GetENodesOfAllNormalNodes()
	panic("not implemented")
}

func (n *SCNodeWrapper) GetENodesOfAllDeletedNodes() string {
	n.GetENodesOfAllDeletedNodes()

	panic("not implemented")
}

//TODO
//需要支持像C++合约一样组合搜索吗
func (n *SCNodeWrapper) GetNodes(query *NodeInfo) string {
	panic("not implemented")
}

func (n *SCNodeWrapper) NodesNum(query *NodeInfo) int {
	panic("not implemented")
}

//for access control
func (n *SCNodeWrapper) AllExportFns() SCExportFns {
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
