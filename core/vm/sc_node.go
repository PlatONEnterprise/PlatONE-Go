package vm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/vm"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"reflect"
	"sort"
)

const (
	NodeStatusNormal  = 1
	NodeStatusDeleted = 2
)

const (
	NodeTypeObserver  = 0
	NodeTypeValidator = 1
)

const (
	nodeNameMaxLenInCharacter = 50
	nodeDescMaxLenInCharacter = 100
)

const (
	keyOfNodesNameDB = "nodes-name-key"
	prefixNodeName   = "sc-node-name"
)

var (
	errParamsInvalid            = errors.New("the parameters invalid")
	errNoPermissionManageSCNode = errors.New("no permission to manage node system contract")
)

var (
	errNodeNameExist  = errors.New("node name exist")
	errPublicKeyExist = errors.New("publicKey exist")
	errNodeNotFound   = errors.New("node not found")
)

type NodeInfo struct {
	Name  string `json:"name,omitempty,required"` //全网唯一，不能重复。所有接口均以此为主键。 这个名称意义是？
	Owner string `json:"owner,omitempty"`         //todo 没有用到？删除？
	Desc  string `json:"desc,omitempty"`          //todo 没有用到？删除？
	Typ   int32  `json:"type,omitempty"`          // 0:观察者节点；1:共识节点
	// status 1为正常节点, 2为删除节点
	Status     int32  `json:"status,omitempty,required"`
	ExternalIP string `json:"externalIP,omitempty"` //todo 没有用到？删除？
	InternalIP string `json:"internalIP,omitempty,required"`
	PublicKey  string `json:"publicKey,omitempty,required"` //节点公钥，全网唯一，不能重复
	RpcPort    int32  `json:"rpcPort,omitempty"`            //todo 没有用到？删除？
	P2pPort    int32  `json:"p2pPort,omitempty,required"`
	// delay set validatorSet
	DelayNum uint64 `json:"delayNum,omitempty"` //共识节点延迟设置的区块高度 (可选, 默认实时设置)
}

type updateNode struct {
	Desc *string `json:"desc,omitempty"` //没有用到？删除？
	Typ  *int32  `json:"type,omitempty"` // 0:观察者节点；1:共识节点
	// status 1为正常节点, 2为删除节点
	Status *int32 `json:"status,omitempty,required"`
	// delay set validatorSet
	DelayNum *uint64 `json:"delayNum,omitempty"` //共识节点延迟设置的区块高度 (可选, 默认实时设置)
}

type eNode struct {
	PublicKey string
	IP        string
	Port      int32
}

func (en *eNode) String() string {
	return fmt.Sprintf("enode://%s@%s:%d", en.PublicKey, en.IP, en.Port)
}

func isValidUser(caller common.Address) bool {
	//todo
	panic("not implemented")
}

func hasAddNodePermission(caller common.Address) bool {
	//todo
	panic("not implemented")
}

func checkRequiredFieldsIsEmpty(node *NodeInfo) error {
	return common.CheckRequiredFieldsIsEmpty(node)
}

func checkNodeStatus(status int32) error {
	if status != NodeStatusDeleted &&
		status != NodeStatusNormal {
		return errors.New(
			fmt.Sprintf("The status of node must be DELETED(%d) or NORMAL(%d)",
				NodeStatusDeleted,
				NodeStatusNormal,
			))
	}

	return nil
}

func checkNodeType(typ int32) error {
	if typ != NodeTypeObserver &&
		typ != NodeTypeValidator {
		return errors.New(
			fmt.Sprintf("The type of node must be OBSERVER(%d) or VALIDATOR(%d)",
				NodeTypeObserver,
				NodeTypeValidator,
			))
	}

	return nil
}

func checkNodeDescLen(desc string) error {
	if len(bytes.Runes([]byte(desc))) > nodeDescMaxLenInCharacter {
		return errors.New(fmt.Sprintf("The length of node name must be less than %d", nodeDescMaxLenInCharacter))
	}
	return nil
}

func checkNodeNameLen(name string) error {
	if len(bytes.Runes([]byte(name))) > nodeNameMaxLenInCharacter {
		return errors.New(fmt.Sprintf("The length of node name must be less than %d", nodeNameMaxLenInCharacter))
	}
	return nil
}

func genNodeName(name string) string {
	return fmt.Sprintf("%s-%s", prefixNodeName, name)
}

func fromNodes(nodes []*NodeInfo) []*eNode {
	var enodes []*eNode
	for _, n := range nodes {
		enode := &eNode{}
		enode.PublicKey = n.PublicKey
		enode.Port = n.P2pPort
		enode.IP = n.InternalIP

		enodes = append(enodes, enode)
	}

	return enodes
}

type SCNode struct {
	stateDB StateDB
	address common.Address
	caller  common.Address
}

func NewSCNode() *SCNode {
	return &SCNode{address: vm.NODE_MANAGEMENT_ADDRESS}
}

func (n *SCNode) checkParamsOfAddNode(node *NodeInfo) error {
	if err := checkRequiredFieldsIsEmpty(node); err != nil {
		return err
	}

	if err := checkNodeStatus(node.Status); err != nil {
		return err
	}

	if err := checkNodeType(node.Typ); err != nil {
		return err
	}

	if err := checkNodeNameLen(node.Name); err != nil {
		return err
	}

	names, err := n.getNames()
	if err != nil {
		return err
	}
	if n.isNameExist(names, node.Name) {
		return errNodeNameExist
	}

	if err := CheckPublicKeyFormat(node.PublicKey); nil != err {
		return err
	}

	if err := n.checkPublicKeyExist(node.PublicKey); nil != err {
		return err
	}

	return nil
}

func (n *SCNode) checkParamsOfUpdateNodeAndReturnUpdatedNode(name string, update *updateNode) (*NodeInfo, error) {
	node, err := n.getNodeByName(name)
	if err != nil {
		return nil, err
	}

	if update.Status != nil {
		status := *update.Status
		if err := checkNodeStatus(status); err != nil {
			return nil, err
		}
		node.Status = status
	}

	if nil != update.Typ {
		typ := *update.Typ
		if err := checkNodeType(typ); err != nil {
			return nil, err
		}
		node.Typ = typ
	}

	if nil != update.Desc {
		desc := *update.Desc
		if err := checkNodeDescLen(desc); err != nil {
			return nil, err
		}
		node.Desc = desc
	}

	if nil != update.DelayNum {
		node.DelayNum = *update.DelayNum
	}

	return node, nil
}

func (n *SCNode) checkPublicKeyExist(pub string) error {
	query := &NodeInfo{}
	query.PublicKey = pub
	num, err := n.nodesNum(query)
	if err != nil {
		return err
	}

	if num > 0 {
		return errPublicKeyExist
	}

	return nil
}

func (n *SCNode) checkPermissionForAdd() error {
	if !isValidUser(n.caller) {
		log.Error("Failed to add node.", "error", n.caller.String()+" is invalid user.")
		return errNoPermissionManageSCNode
	}

	if !hasAddNodePermission(n.caller) {
		log.Error("Failed to add node.", "error", n.caller.String()+" has no permission to add node.")
		return errNoPermissionManageSCNode
	}

	return nil
}

func (n *SCNode) add(node *NodeInfo) error {
	if err := n.checkPermissionForAdd(); nil != err {
		return errNoPermissionManageSCNode
	}

	if err := n.checkParamsOfAddNode(node); nil != err {
		log.Error("Failed to add node.", "error", err.Error(), "node", node)
		return errParamsInvalid
	}

	err := n.addName(node.Name)
	if err != nil {
		log.Error("Failed to add node.", "node", node, "error", err)
		return err
	}

	encodedBin, err := json.Marshal(node)
	if err != nil {
		log.Error("Failed to add node.", "error", err.Error(), "node", node)
		return err
	}
	n.setState(genNodeName(node.Name), encodedBin)

	log.Info("add node success.", "node", string(encodedBin))

	return nil
}

func (n *SCNode) update(name string, update *updateNode) error {
	if err := n.checkPermissionForAdd(); nil != err {
		return err
	}

	node, err := n.checkParamsOfUpdateNodeAndReturnUpdatedNode(name, update)
	if err != nil {
		return err
	}

	encodedBin, err := json.Marshal(node)
	if err != nil {
		log.Error("Failed to update node.", "error", err.Error(), "update", update)
		return err
	}
	n.setState(genNodeName(node.Name), encodedBin)

	return nil
}

//The slice must be sorted in ascending order
func (n *SCNode) isNameExist(names []string, name string) bool {
	index := sort.SearchStrings(names, name)
	//not found
	if len(names) == index {
		return false
	}

	return true
}

func (n *SCNode) addName(name string) error {
	names, err := n.getNames()
	if err != nil {
		return err
	}

	if n.isNameExist(names, name) {
		log.Error("node exist.", "name", name)
		return errors.New("node exist.")
	}

	names = append(names, name)
	sort.Strings(names)                           //must sort names before set to DB
	encodedNames, err := rlp.EncodeToBytes(names) //todo C++合约里面这个存储格式是自定义的，自定义格式：name|name
	if err != nil {
		return err
	}

	n.setState(keyOfNodesNameDB, encodedNames)

	return nil
}

func (n *SCNode) GetAllNodes() ([]*NodeInfo, error) {
	return n.GetNodes(nil)
}

func (n *SCNode) getENodesOfAllNormalNodes() ([]*eNode, error) {
	query := new(NodeInfo)
	query.Status = NodeStatusNormal
	nodes, err := n.GetNodes(query)
	if err != nil {
		return nil, err
	}

	return fromNodes(nodes), nil
}

func (n *SCNode) getENodesOfAllDeletedNodes() ([]*eNode, error) {
	query := new(NodeInfo)
	query.Status = NodeStatusDeleted
	nodes, err := n.GetNodes(query)
	if err != nil {
		return nil, err
	}

	return fromNodes(nodes), nil
}

func (n *SCNode) getNodeByName(name string) (*NodeInfo, error) {
	bin := n.getState(genNodeName(name))
	if len(bin) == 0 {
		return nil, errNodeNotFound
	}

	var node NodeInfo
	err := json.Unmarshal(bin, &node)
	if err != nil {
		return nil, err
	}

	return &node, nil
}

func (n *SCNode) GetNodes(query *NodeInfo) ([]*NodeInfo, error) {
	names, err := n.getNames()
	if err != nil {
		return nil, err
	}

	var nodes []*NodeInfo
	for _, name := range names {
		node, err := n.getNodeByName(name)
		if err != nil {
			return nil, err
		}

		if n.isMatch(node, query) {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

func (n *SCNode) getNames() ([]string, error) {
	bin := n.getState(keyOfNodesNameDB)
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

func (n *SCNode) nodesNum(query *NodeInfo) (int, error) {
	nodes, err := n.GetNodes(query)
	if err != nil {
		return 0, err
	}

	return len(nodes), nil
}

func (n *SCNode) setState(key string, value []byte) {
	n.stateDB.SetState(n.address, []byte(key), value)
}

func (n *SCNode) getState(key string) []byte {
	return n.stateDB.GetState(n.address, []byte(key))
}

func (n *SCNode) isMatch(node, query *NodeInfo) bool {
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
//func (n *SCNode) getNames() ([]string, error) {
//	shardNum, err := n.ShardingCount()
//	if err != nil {
//		return nil, err
//	}
//
//	var names []string
//	for i := uint64(0); i < shardNum; i++ {
//		val := n.getState([]byte(fmt.Sprintf("%s-%d", PrefixShardingKey, i)))
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
//	bin := n.getState([]byte(KeyTotalCountsOfNodes))
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
//	n.setState([]byte(KeyTotalCountsOfNodes), encodedCount)
//
//	return nil
//}

//var (
//	errNodeNameExist = errors.New("node name exist")
//)
//
//func (n *SCNode) CheckNameExist(name string) (existedNames []string, err error) {
//	names, err := n.getNames()
//	if err != nil {
//		return nil, err
//	}
//
//	if n.isNameExist(names, name) {
//		return names, errNodeNameExist
//	}
//
//	return names,nil
//}
