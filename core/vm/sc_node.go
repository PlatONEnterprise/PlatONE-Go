package vm

import (
	"encoding/json"
	"errors"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/vm"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
)

type NodeStatus int32

const (
	NODE_STATUS_NORMAL  NodeStatus = 1
	NODE_STATUS_DELETED NodeStatus = 2
)

type NodeType int32

const (
	NODE_TYPE_OBSERVER  NodeType = 0
	NODE_TYPE_VALIDATOR NodeType = 1
)

type NodeInfo struct {
	Name  string   `json:"name,omitempty,required"` //全网唯一，不能重复。所有接口均以此为主键。 这个名称意义是？
	Owner string   `json:"owner,omitempty"`         //没有用到？删除？
	Desc  string   `json:"desc,omitempty"`          //没有用到？删除？
	Typ   NodeType `json:"type,omitempty"`          // 0:观察者节点；1:共识节点
	// status 1为正常节点, 2为删除节点
	Status     NodeStatus `json:"status,omitempty,required"`
	ExternalIP string     `json:"externalIP,omitempty"` //没有用到？删除？
	InternalIP string     `json:"internalIP,omitempty,required"`
	PublicKey  string     `json:"publicKey,omitempty,required"` //节点公钥，全网唯一，不能重复
	RpcPort    int32      `json:"rpcPort,omitempty"`            //没有用到？删除？
	P2pPort    int32      `json:"p2pPort,omitempty,required"`
	// delay set validatorSet
	DelayNum uint64 `json:"delayNum,omitempty"` //共识节点延迟设置的区块高度 (可选, 默认实时设置)
}

type SCNode struct {
	stateDB StateDB
	address common.Address
	caller  common.Address
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

func IsValidUser(caller common.Address) bool {
	//todo
	panic("not implemented")
}

func HasAddNodePermission(caller common.Address) bool {
	//todo
	panic("not implemented")
}

func CheckRequiredFieldsIsEmpty(node *NodeInfo) error {
	return common.CheckRequiredFieldsIsEmpty(node)
}

func CheckNodeStatus(status NodeStatus) error {
	if status != NODE_STATUS_DELETED &&
		status != NODE_STATUS_NORMAL {
		return errors.New("The status of node must be DELETED(2) or NORMAL(1)")
	}

	return nil
}

func CheckNodeType(typ NodeType) error {
	if typ != NODE_TYPE_OBSERVER &&
		typ != NODE_TYPE_VALIDATOR {
		return errors.New("The type of node must be OBSERVER(0) or VALIDATOR(1)")
	}

	return nil
}

func CheckParamsOfAddNode(node *NodeInfo) error {
	if err := CheckRequiredFieldsIsEmpty(node); err != nil {
		return err
	}

	if err := CheckNodeStatus(node.Status); err != nil {
		return err
	}

	if err := CheckNodeType(node.Typ); err != nil {
		return err
	}

	//unique key check: publickey,name
	panic("not implemented")
}

func (n *SCNode) CheckPermissionForAdd() error {
	if !IsValidUser(n.caller) {
		log.Error("Failed to add node.", "error", n.caller.String()+" is invalid user.")
		return SC_ERR_NO_PERMISSION
	}

	if !HasAddNodePermission(n.caller) {
		log.Error("Failed to add node.", "error", n.caller.String()+" has no permission to add node.")
		return SC_ERR_NO_PERMISSION
	}

	return nil
}

func (n *SCNode) Add(node *NodeInfo) error {
	if err := n.CheckPermissionForAdd(); nil != err {
		return err
	}

	if err := CheckParamsOfAddNode(node); nil != err {
		log.Error("Failed to add node.", "error", err.Error(), "node", node)
		return err
	}

	encodedBin, err := json.Marshal(node)
	if err != nil {
		log.Error("Failed to add node.", "error", err.Error(), "node", node)
		return err
	}
	n.SetState([]byte(node.Name), encodedBin)
	log.Info("Add node success.", "node", string(encodedBin))

	return nil
}

func (n *SCNode) Update(name string, node *NodeInfo) error {
	panic("not implemented")
}

func (n *SCNode) GetAllNodes() []*NodeInfo {
	panic("not implemented")
}

func (n *SCNode) IsPublicKeyExist(pk string) bool {
	panic("not implemented")
}

type ENode struct {
	PublicKey string
	IP        string
	Port      string
}

func (n *SCNode) GetENodesOfAllNormalNodes() []*ENode {
	panic("not implemented")
}

func (n *SCNode) GetENodesOfAllDeletedNodes() []*ENode {
	panic("not implemented")
}

func (n *SCNode) GetNodes(query *NodeInfo) []*NodeInfo {

	panic("not implemented")
}

func (n *SCNode) NodesNum(query *NodeInfo) uint64 {
	panic("not implemented")
}
