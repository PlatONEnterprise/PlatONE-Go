package vm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"math/big"
	"reflect"
	"sort"
)

const (
	importOldDataSuccess CodeType = 0
	dataUnmarshalFail CodeType = 1
)
const (
	NodeStatusNormal  = uint32(1)
	NodeStatusDeleted = uint32(2)
)

const (
	NodeTypeObserver  = uint32(0)
	NodeTypeValidator = uint32(1)
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

const (
	addNodeSuccess      CodeType = 0
	addNodeBadParameter CodeType = 1
	addNodeNoPermission CodeType = 2
)

const (
	updateNodeSuccess      CodeType = 0
	updateNodeBadParameter CodeType = 1
	updateNodeNoPermission CodeType = 2
)

type eNode struct {
	PublicKey string
	IP        string
	Port      uint32
}

func (en *eNode) String() string {
	return fmt.Sprintf("enode://%s@%s:%d", en.PublicKey, en.IP, en.Port)
}

func checkRequiredFieldsIsEmpty(node *syscontracts.NodeInfo) error {
	return common.CheckRequiredFieldsIsEmpty(node)
}

func checkNodeStatus(status uint32) error {
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

func checkNodeType(typ uint32) error {
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

func fromNodes(nodes []*syscontracts.NodeInfo) []*eNode {
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
	stateDB      StateDB
	contractAddr common.Address
	caller       common.Address
	blockNumber  *big.Int
}

func NewSCNode(db StateDB) *SCNode {
	return &SCNode{stateDB: db, contractAddr: syscontracts.NodeManagementAddress, blockNumber: big.NewInt(0)}
}

func (n *SCNode) checkParamsOfAddNode(node *syscontracts.NodeInfo) error {
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
		if errNodeNotFound != err {
			return err
		}

		names = []string{}
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

func (n *SCNode) checkParamsOfUpdateNodeAndReturnUpdatedNode(name string, update *syscontracts.UpdateNode) (*syscontracts.NodeInfo, error) {
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
	query := &syscontracts.NodeInfo{}
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
	//internal call, do not check permission
	if common.IsHexZeroAddress(n.caller.String()) {
		return nil
	}

	if !hasNodeOpPermission(n.stateDB, n.caller) {
		log.Error("Failed to add node.", "error", n.caller.String()+" has no permission to add node.")
		return errNoPermissionManageSCNode
	}

	return nil
}

func (n *SCNode) add(node *syscontracts.NodeInfo) error {
	if err := n.checkPermissionForAdd(); nil != err {
		n.emitNotifyEvent(addNodeNoPermission, fmt.Sprintf("%s has no permission to add node info.", n.caller.String()))
		return errNoPermissionManageSCNode
	}

	if err := n.checkParamsOfAddNode(node); nil != err {
		n.emitNotifyEvent(addNodeBadParameter, fmt.Sprintf("node info is invalid. err:%s", err.Error()))
		log.Error("Failed to add node.", "error", err.Error(), "node", node)
		return errParamsInvalid
	}

	err := n.addName(node.Name)
	if err != nil {
		n.emitNotifyEvent(addNodeBadParameter, fmt.Sprintf("Failed to add node. err:%s", err.Error()))
		log.Error("Failed to add node.", "node", node, "error", err)
		return err
	}

	encodedBin, err := rlp.EncodeToBytes(node)
	if err != nil {
		n.emitNotifyEvent(addNodeBadParameter, fmt.Sprintf("Failed to add node. err:%s", err.Error()))
		log.Error("Failed to add node.", "error", err.Error(), "node", node.String())
		return err
	}
	n.setState(genNodeName(node.Name), encodedBin)

	n.emitNotifyEvent(addNodeSuccess, fmt.Sprintf("add node success. node:%s", node.String()))
	log.Info("add node success.", "node", node.String())

	return nil
}

func (n *SCNode) update(name string, update *syscontracts.UpdateNode) error {
	if err := n.checkPermissionForAdd(); nil != err {
		n.emitNotifyEvent(updateNodeNoPermission, fmt.Sprintf("%s no permission update node.", n.caller.String()))
		return err
	}

	node, err := n.checkParamsOfUpdateNodeAndReturnUpdatedNode(name, update)
	if err != nil {
		n.emitNotifyEvent(updateNodeBadParameter, fmt.Sprintf("parameter is invalid"))
		return err
	}

	encodedBin, err := rlp.EncodeToBytes(node)
	if err != nil {
		n.emitNotifyEvent(updateNodeBadParameter, fmt.Sprintf("parameter is invalid"))
		log.Error("Failed to update node.", "error", err.Error(), "update", update.String())
		return err
	}
	n.setState(genNodeName(node.Name), encodedBin)

	n.emitNotifyEvent(updateNodeSuccess, fmt.Sprintf("update node success. info:%s", update.String()))
	log.Info("update node success. ", "update info", update.String())

	return nil
}

//The slice must be sorted in ascending order
func (n *SCNode) isNameExist(names []string, name string) bool {
	index := sort.SearchStrings(names, name)
	//not found
	if index < len(names) && names[index] == name {
		return true
	}

	return false
}

func (n *SCNode) addName(name string) error {
	names, err := n.getNames()
	if err != nil {
		if errNodeNotFound != err {
			return err
		}

		names = []string{}
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

func (n *SCNode) GetAllNodes() ([]*syscontracts.NodeInfo, error) {
	return n.GetNodes(nil)
}

func (n *SCNode) getENodesOfAllNormalNodes() ([]*eNode, error) {
	query := new(syscontracts.NodeInfo)
	query.Status = NodeStatusNormal
	nodes, err := n.GetNodes(query)
	if err != nil {
		return nil, err
	}

	return fromNodes(nodes), nil
}

func (n *SCNode) getENodesOfAllDeletedNodes() ([]*eNode, error) {
	query := new(syscontracts.NodeInfo)
	query.Status = NodeStatusDeleted
	nodes, err := n.GetNodes(query)
	if err != nil {
		return nil, err
	}

	return fromNodes(nodes), nil
}

func (n *SCNode) getNodeByName(name string) (*syscontracts.NodeInfo, error) {
	bin := n.getState(genNodeName(name))
	if len(bin) == 0 {
		return nil, errNodeNotFound
	}

	var node syscontracts.NodeInfo
	err := rlp.DecodeBytes(bin, &node)
	if err != nil {
		return nil, err
	}

	return &node, nil
}

func (n *SCNode) GetNodes(query *syscontracts.NodeInfo) ([]*syscontracts.NodeInfo, error) {
	names, err := n.getNames()
	if err != nil {
		return nil, err
	}

	var nodes []*syscontracts.NodeInfo
	for _, name := range names {
		node, err := n.getNodeByName(name)
		if err != nil {
			return nil, err
		}

		if n.isMatch(node, query) {
			nodes = append(nodes, node)
		}
	}

	if len(nodes) == 0 {
		return nil, errNodeNotFound
	}

	return nodes, nil
}

func (n *SCNode) getNames() ([]string, error) {
	bin := n.getState(keyOfNodesNameDB)
	if len(bin) == 0 {
		return nil, errNodeNotFound
	}

	var names []string
	err := rlp.DecodeBytes(bin, &names)
	if err != nil {
		return nil, err
	}

	return names, nil
}

func (n *SCNode) nodesNum(query *syscontracts.NodeInfo) (int, error) {
	nodes, err := n.GetNodes(query)
	if err != nil {
		if errNodeNotFound == err {
			return 0, nil
		}

		return 0, err
	}

	return len(nodes), nil
}

func (n *SCNode) importOldNodesData(data string) error {
	str := []byte(data)
	nodes := make([]syscontracts.NodeInfo, 0)
	err := json.Unmarshal(str, &nodes)
	if err != nil {
		n.emitNotifyEvent(dataUnmarshalFail, fmt.Sprintf("old nodes data unmarshal fail"))
		return err
	}
	for index, _ := range nodes{
		//names, err := n.getNames()
		//if err != nil {
		//	if errNodeNotFound != err {
		//		return err
		//	}
		//
		//	names = []string{}
		//}
		//if n.isNameExist(names, nodes[index].Name) {
		//	n.update(nodes[index].Name, &nodes[index])
		//}
		n.add(&nodes[index])
	}
	if err != nil {
		return  err
	}
	n.emitNotifyEvent(importOldDataSuccess, fmt.Sprintf("import old nodes data success"))
	return nil
}


func (n *SCNode) setState(key string, value []byte) {
	n.stateDB.SetState(n.contractAddr, []byte(key), value)
}

func (n *SCNode) getState(key string) []byte {
	return n.stateDB.GetState(n.contractAddr, []byte(key))
}

func (n *SCNode) isMatch(node, query *syscontracts.NodeInfo) bool {
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

func (n *SCNode) emitNotifyEvent(code CodeType, msg string) {
	topic := "Notify"
	emitEvent(n.contractAddr, n.stateDB, n.blockNumber.Uint64(), topic, code, msg)
}
