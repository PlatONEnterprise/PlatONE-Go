package vm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/vm"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"reflect"
	"sort"
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

const NODE_DESC_MAX_LEN_IN_CHARACTOR = 100

func CheckNodeDescLen(desc string) error {
	if len(bytes.Runes([]byte(desc))) > NODE_DESC_MAX_LEN_IN_CHARACTOR {
		return errors.New(fmt.Sprintf("The length of node name must be less than %d", NODE_DESC_MAX_LEN_IN_CHARACTOR))
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

const NODE_NAME_MAX_LEN_IN_CHARACTOR = 50

func CheckNodeNameLen(name string) error {
	if len(bytes.Runes([]byte(name))) > NODE_NAME_MAX_LEN_IN_CHARACTOR {
		return errors.New(fmt.Sprintf("The length of node name must be less than %d", NODE_NAME_MAX_LEN_IN_CHARACTOR))
	}
	return nil
}

func CheckPublicKeyFormat(pub string) error {
	_, err := crypto.UnmarshalPubkey([]byte(pub))
	if err != nil {
		return err
	}

	return nil
}

var E_NODE_NAME_EXIST = errors.New("node name exist")

func (n *SCNode) CheckParamsOfAddNode(node *NodeInfo) error {
	if err := CheckRequiredFieldsIsEmpty(node); err != nil {
		return err
	}

	if err := CheckNodeStatus(node.Status); err != nil {
		return err
	}

	if err := CheckNodeType(node.Typ); err != nil {
		return err
	}

	if err := CheckNodeNameLen(node.Name); err != nil {
		return err
	}

	names, err := n.GetNames()
	if err != nil {
		return err
	}
	if n.IsNameExist(names, node.Name) {
		return E_NODE_NAME_EXIST
	}

	if err := CheckPublicKeyFormat(node.PublicKey); nil != err {
		return err
	}

	if err := n.CheckPublicKeyExist(node.PublicKey); nil != err {
		return err
	}

	return nil
}

func (n *SCNode) CheckParamsOfUpdateNodeAndReturnUpdatedNode(name string, update map[string]interface{}) (*NodeInfo, error) {
	node, err := n.GetNodeByName(name)
	if err != nil {
		return nil, err
	}
	//todo 这些也许应该在wrapper里面处理
	istatus, ok := update["status"]
	if ok {
		status, ok := istatus.(NodeStatus)
		if !ok {
			return nil, errors.New("the params is invalid")
		}

		if err := CheckNodeStatus(status); err != nil {
			return nil, err
		}
		node.Status = status
	}

	itype, ok := update["type"]
	if ok {
		typ, ok := itype.(NodeType)
		if !ok {
			return nil, errors.New("the params is invalid")
		}

		if err := CheckNodeType(typ); err != nil {
			return nil, err
		}
		node.Typ = typ
	}

	idesc, ok := update["desc"]
	if ok {
		desc, ok := idesc.(string)
		if !ok {
			return nil, errors.New("the params is invalid")
		}

		if err := CheckNodeDescLen(desc); err != nil {
			return nil, err
		}
		node.Desc = desc
	}

	idelayNum, ok := update["delayNum"]
	if ok {
		delayNum, ok := idelayNum.(uint64)
		if !ok {
			return nil, errors.New("the params is invalid")
		}
		node.DelayNum = delayNum
	}

	return node, nil
}

var E_PUBLICKEY_EXIST = errors.New("publicKey exist")

func (n *SCNode) CheckPublicKeyExist(pub string) error {
	query := &NodeInfo{}
	query.PublicKey = pub
	num, err := n.NodesNum(query)
	if err != nil {
		return err
	}

	if num > 0 {
		return E_PUBLICKEY_EXIST
	}

	return nil
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

	if err := n.CheckParamsOfAddNode(node); nil != err {
		log.Error("Failed to add node.", "error", err.Error(), "node", node)
		return err
	}

	encodedBin, err := json.Marshal(node)
	if err != nil {
		log.Error("Failed to add node.", "error", err.Error(), "node", node)
		return err
	}
	n.SetState(GenNodeName(node.Name), encodedBin)

	err = n.AddName(node.Name)
	if err != nil {
		log.Error("Failed to add node.", "node", node, "error", err)
		return err
	}
	log.Info("Add node success.", "node", string(encodedBin))

	return nil
}

func (n *SCNode) Update(name string, update map[string]interface{}) error {
	if err := n.CheckPermissionForAdd(); nil != err {
		return err
	}

	node, err := n.CheckParamsOfUpdateNodeAndReturnUpdatedNode(name, update)
	if err != nil {
		return err
	}

	encodedBin, err := json.Marshal(node)
	if err != nil {
		log.Error("Failed to update node.", "error", err.Error(), "update", update)
		return err
	}
	n.SetState(GenNodeName(node.Name), encodedBin)

	return nil
}

//const (
//	KeyTotalCountsOfNodes = "total-counts-nodes"
//
//	ShardingNum       = 1000
//	PrefixShardingKey = "prefix-sharding-key"
//)
//
//func (n *SCNode) ShardingCount() (uint64, error) {
//	count, err := n.TotalCounts()
//	if err != nil {
//		return 0, err
//	}
//
//	if count%ShardingNum > 0 {
//		return count/ShardingNum + 1, nil
//	}
//
//	return count / ShardingNum, nil
//}
//
//func (n *SCNode) GetNames() ([]string, error) {
//	shardNum, err := n.ShardingCount()
//	if err != nil {
//		return nil, err
//	}
//
//	var names []string
//	for i := uint64(0); i < shardNum; i++ {
//		val := n.GetState([]byte(fmt.Sprintf("%s-%d", PrefixShardingKey, i)))
//		var shardNames []string
//		err := json.Unmarshal(val, &shardNames)
//		if err != nil {
//			return nil, err
//		}
//
//		names = append(names, shardNames...)
//	}
//
//	return names, nil
//}
//
//func (n *SCNode) TotalCounts() (uint64, error) {
//	bin := n.GetState([]byte(KeyTotalCountsOfNodes))
//
//	var count uint64
//	err := rlp.DecodeBytes(bin, &count)
//	if nil != err {
//		return 0, err
//	}
//
//	return count, nil
//}
//
//func (n *SCNode) IncrementTotalCounts() error {
//	count, err := n.TotalCounts()
//	if nil != err {
//		return err
//	}
//
//	encodedCount, err := rlp.EncodeToBytes(count + 1)
//	if nil != err {
//		return err
//	}
//	n.SetState([]byte(KeyTotalCountsOfNodes), encodedCount)
//
//	return nil
//}

const (
	NodesNameKey = "nodes-name-key"
)

func (n *SCNode) GetNames() ([]string, error) {
	bin := n.GetState([]byte(NodesNameKey))
	if len(bin) == 0 {
		return []string{}, nil
	}

	var names []string
	err := rlp.DecodeBytes(bin, &names)
	if err != nil {
		return nil, err
	}

	return names, nil
}

//var (
//	E_NODE_NAME_EXIST = errors.New("node name exist")
//)
//
//func (n *SCNode) CheckNameExist(name string) (existedNames []string, err error) {
//	names, err := n.GetNames()
//	if err != nil {
//		return nil, err
//	}
//
//	if n.IsNameExist(names, name) {
//		return names, E_NODE_NAME_EXIST
//	}
//
//	return names,nil
//}

//The slice must be sorted in ascending order
func (n *SCNode) IsNameExist(names []string, name string) bool {
	index := sort.SearchStrings(names, name)
	//not found
	if len(names) == index {
		return false
	}

	return true
}

func (n *SCNode) AddName(name string) error {
	names, err := n.GetNames()
	if err != nil {
		return err
	}

	if n.IsNameExist(names, name) {
		log.Info("node exist.", "name", name)
		return nil
	}

	names = append(names, name)
	sort.Strings(names)
	encodedNames, err := rlp.EncodeToBytes(names)
	if err != nil {
		return err
	}

	n.SetState([]byte(NodesNameKey), encodedNames)

	return nil
}

func (n *SCNode) GetAllNodes() ([]*NodeInfo, error) {
	return n.GetNodes(nil)
}

type ENode struct {
	PublicKey string
	IP        string
	Port      int32
}

func FromNodes(nodes []*NodeInfo) []*ENode {
	var enodes []*ENode
	for _, n := range nodes {
		enode := &ENode{}
		enode.PublicKey = n.PublicKey
		enode.Port = n.P2pPort
		enode.IP = n.InternalIP

		enodes = append(enodes, enode)
	}

	return enodes
}

func (n *SCNode) GetENodesOfAllNormalNodes() ([]*ENode, error) {
	query := new(NodeInfo)
	query.Status = NODE_STATUS_NORMAL
	nodes, err := n.GetNodes(query)
	if err != nil {
		return nil, err
	}

	return FromNodes(nodes), nil
}

func (n *SCNode) GetENodesOfAllDeletedNodes() ([]*ENode, error) {
	query := new(NodeInfo)
	query.Status = NODE_STATUS_DELETED
	nodes, err := n.GetNodes(query)
	if err != nil {
		return nil, err
	}

	return FromNodes(nodes), nil
}

const (
	PrefixNodeName = "sc-node-name"
)

func GenNodeName(name string) []byte {
	return []byte( fmt.Sprintf("%s-%s", PrefixNodeName, name))
}

var E_NODE_NOT_FOUND = errors.New("node not found")

func (n *SCNode) GetNodeByName(name string) (*NodeInfo, error) {
	bin := n.GetState(GenNodeName(name))
	if len(bin) == 0 {
		return nil, E_NODE_NOT_FOUND
	}

	var node NodeInfo
	err := json.Unmarshal(bin, &node)
	if err != nil {
		return nil, err
	}

	return &node, nil
}

func (n *SCNode) GetNodes(query *NodeInfo) ([]*NodeInfo, error) {
	names, err := n.GetNames()
	if err != nil {
		return nil, err
	}

	var nodes []*NodeInfo
	for _, name := range names {
		node, err := n.GetNodeByName(name)
		if err != nil {
			return nil, err
		}

		if isMatch(node, query) {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

func isMatch(node, query *NodeInfo) bool {
	if nil == query {
		return true
	}

	vquery := reflect.ValueOf(query).Elem()
	vnode := reflect.ValueOf(node).Elem()
	for i := 0; i < vquery.Type().NumField(); i++ {
		if !vquery.Field(i).IsZero() {
			if !(vnode.Field(i).CanInterface() &&
				vquery.Field(i).CanInterface() &&
				reflect.DeepEqual(vquery.Field(i).Interface(), vnode.Field(i).Interface())) {

				return false
			}
		}
	}

	return true
}

func (n *SCNode) NodesNum(query *NodeInfo) (int, error) {
	nodes, err := n.GetNodes(query)
	if err != nil {
		return 0, err
	}

	return len(nodes), nil
}
